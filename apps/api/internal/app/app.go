package app

import (
	"filnest/internal/auth"
	"filnest/internal/file"
	"filnest/internal/folder"
	"filnest/internal/photo"
	"filnest/internal/platform/config"
	"filnest/internal/platform/httpx"
	"filnest/internal/platform/mailer"
	"filnest/internal/platform/objectstore"
	"filnest/internal/platform/postgres"
	"filnest/internal/search"
	"filnest/internal/storage"
	"filnest/internal/trash"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(cfg config.Config, pool *pgxpool.Pool) (*gin.Engine, error) {
	obj, err := objectstore.NewFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	return NewWithDeps(cfg, pool, mailer.New(cfg), obj), nil
}

func NewWithMailer(cfg config.Config, pool *pgxpool.Pool, m mailer.Mailer) *gin.Engine {
	return NewWithDeps(cfg, pool, m, objectstore.NewMemory())
}

func NewWithDeps(cfg config.Config, pool *pgxpool.Pool, m mailer.Mailer, obj objectstore.ObjectStore) *gin.Engine {
	store := postgres.NewStore(pool)
	tokens := auth.NewTokens(cfg.JWTSecret)
	authSvc := auth.NewService(cfg, store, tokens, m)
	authHandlers := auth.NewHandler(authSvc)

	folderRepo := postgres.NewFolderRepository(store)
	folderSvc := folder.NewService(folderRepo)
	folderHandlers := folder.NewHandler(folderSvc)

	fileRepo := postgres.NewFileRepository(store)
	quota := postgres.NewQuotaStore(store)
	fileSvc := file.NewService(fileRepo, folderRepo, quota, obj)
	fileHandlers := file.NewHandler(fileSvc)

	trashRepo := postgres.NewTrashRepository(store)
	trashSvc := trash.NewService(trashRepo, quota, obj)
	trashHandlers := trash.NewHandler(trashSvc)

	photoRepo := postgres.NewPhotoRepository(store)
	photoSvc := photo.NewService(photoRepo)
	photoHandlers := photo.NewHandler(photoSvc)

	searchRepo := postgres.NewSearchRepository(store)
	searchSvc := search.NewService(searchRepo)
	searchHandlers := search.NewHandler(searchSvc)

	storageHandlers := storage.NewHandler(quota)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(httpx.RequestLog())
	if len(cfg.CORSAllowedOrigins) > 0 {
		engine.Use(httpx.CORS(cfg.CORSAllowedOrigins))
	}
	v1 := engine.Group("/api/v1")
	authHandlers.RegisterRoutes(v1)

	storage := []gin.HandlerFunc{
		authHandlers.AuthRequired(),
		authHandlers.EmailVerifiedRequired(),
	}
	folderHandlers.RegisterRoutes(v1, storage...)
	fileHandlers.RegisterRoutes(v1, storage...)
	trashHandlers.RegisterRoutes(v1, storage...)
	photoHandlers.RegisterRoutes(v1, storage...)
	searchHandlers.RegisterRoutes(v1, storage...)
	storageHandlers.RegisterRoutes(v1, storage...)

	return engine
}

func NewTrashService(pool *pgxpool.Pool, obj objectstore.ObjectStore) *trash.Service {
	store := postgres.NewStore(pool)
	return trash.NewService(postgres.NewTrashRepository(store), postgres.NewQuotaStore(store), obj)
}
