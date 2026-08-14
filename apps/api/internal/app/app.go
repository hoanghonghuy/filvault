package app

import (
	"filnest/internal/auth"
	"filnest/internal/platform/config"
	"filnest/internal/platform/postgres"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(cfg config.Config, pool *pgxpool.Pool) *gin.Engine {
	store := postgres.NewStore(pool)
	tokens := auth.NewTokens(cfg.JWTSecret)
	svc := auth.NewService(cfg, store, tokens)
	handlers := auth.NewHandler(svc)

	engine := gin.New()
	engine.Use(gin.Recovery())
	v1 := engine.Group("/api/v1")
	handlers.RegisterRoutes(v1)
	return engine
}
