package app

import (
	"filnest/internal/auth"
	"filnest/internal/folder"
	"filnest/internal/platform/config"
	"filnest/internal/platform/mailer"
	"filnest/internal/platform/postgres"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(cfg config.Config, pool *pgxpool.Pool) *gin.Engine {
	return NewWithMailer(cfg, pool, mailer.New(cfg))
}

func NewWithMailer(cfg config.Config, pool *pgxpool.Pool, m mailer.Mailer) *gin.Engine {
	store := postgres.NewStore(pool)
	tokens := auth.NewTokens(cfg.JWTSecret)
	authSvc := auth.NewService(cfg, store, tokens, m)
	authHandlers := auth.NewHandler(authSvc)

	folderSvc := folder.NewService(postgres.NewFolderRepository(store))
	folderHandlers := folder.NewHandler(folderSvc)

	engine := gin.New()
	engine.Use(gin.Recovery())
	v1 := engine.Group("/api/v1")
	authHandlers.RegisterRoutes(v1)

	storage := []gin.HandlerFunc{
		authHandlers.AuthRequired(),
		authHandlers.EmailVerifiedRequired(),
	}
	folderHandlers.RegisterRoutes(v1, storage...)

	return engine
}
