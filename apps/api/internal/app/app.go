package app

import (
	"filnest/internal/auth"
	"filnest/internal/file"
	"filnest/internal/folder"
	"filnest/internal/platform/config"
	"filnest/internal/platform/mailer"
	"filnest/internal/platform/objectstore"
	"filnest/internal/platform/postgres"

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

	fileSvc := file.NewService(postgres.NewFileRepository(store), folderRepo, postgres.NewQuotaStore(store), obj)
	fileHandlers := file.NewHandler(fileSvc)

	engine := gin.New()
	engine.Use(gin.Recovery())
	v1 := engine.Group("/api/v1")
	authHandlers.RegisterRoutes(v1)

	storage := []gin.HandlerFunc{
		authHandlers.AuthRequired(),
		authHandlers.EmailVerifiedRequired(),
	}
	folderHandlers.RegisterRoutes(v1, storage...)
	fileHandlers.RegisterRoutes(v1, storage...)

	return engine
}
