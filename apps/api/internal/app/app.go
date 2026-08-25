package app

import (
	"context"
	"net/http"
	"time"

	"filvault/internal/auth"
	"filvault/internal/file"
	"filvault/internal/folder"
	"filvault/internal/photo"
	"filvault/internal/platform/config"
	"filvault/internal/platform/httpx"
	"filvault/internal/platform/mailer"
	"filvault/internal/platform/objectstore"
	"filvault/internal/platform/postgres"
	"filvault/internal/search"
	"filvault/internal/storage"
	"filvault/internal/trash"

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
	photoSvc := photo.NewService(photoRepo, obj)
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
	engine.GET("/healthz", healthz(pool))
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

func healthz(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func NewTrashService(pool *pgxpool.Pool, obj objectstore.ObjectStore) *trash.Service {
	store := postgres.NewStore(pool)
	return trash.NewService(postgres.NewTrashRepository(store), postgres.NewQuotaStore(store), obj)
}
