package app

import (
	"context"
	"net/http"
	"time"

	"filvault/internal/activity"
	"filvault/internal/auth"
	"filvault/internal/call"
	"filvault/internal/chat"
	"filvault/internal/file"
	"filvault/internal/folder"
	"filvault/internal/photo"
	"filvault/internal/platform/config"
	"filvault/internal/platform/httpx"
	"filvault/internal/platform/mailer"
	"filvault/internal/platform/objectstore"
	"filvault/internal/platform/postgres"
	"filvault/internal/search"
	"filvault/internal/share"
	"filvault/internal/sharelink"
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

	activityRepo := postgres.NewActivityRepository(store)
	activityRecorder := activity.NewRecorder(activityRepo)
	activitySvc := activity.NewService(activityRepo)
	activityHandlers := activity.NewHandler(activitySvc)

	authSvc := auth.NewService(cfg, store, tokens, m, activityRecorder)
	authHandlers := auth.NewHandler(authSvc)

	folderRepo := postgres.NewFolderRepository(store)
	folderSvc := folder.NewService(folderRepo, activityRecorder)
	folderHandlers := folder.NewHandler(folderSvc)

	fileRepo := postgres.NewFileRepository(store)
	quota := postgres.NewQuotaStore(store)
	fileSvc := file.NewService(fileRepo, folderRepo, quota, obj, activityRecorder)
	fileHandlers := file.NewHandler(fileSvc)

	trashRepo := postgres.NewTrashRepository(store)
	trashSvc := trash.NewService(trashRepo, quota, obj, activityRecorder)
	trashHandlers := trash.NewHandler(trashSvc)

	photoRepo := postgres.NewPhotoRepository(store)
	photoSvc := photo.NewService(photoRepo, obj)
	photoHandlers := photo.NewHandler(photoSvc)

	searchRepo := postgres.NewSearchRepository(store)
	searchSvc := search.NewService(searchRepo)
	searchHandlers := search.NewHandler(searchSvc)

	shareRepo := postgres.NewShareRepository(store)
	shareSvc := share.NewService(shareRepo, obj, m, activityRecorder)
	shareHandlers := share.NewHandler(shareSvc)

	shareLinkRepo := postgres.NewShareLinkRepository(store)
	shareLinkSvc := sharelink.NewService(shareLinkRepo, obj, activityRecorder)
	shareLinkHandlers := sharelink.NewHandler(shareLinkSvc)

	chatRepo := postgres.NewChatRepository(store)
	chatSvc := chat.NewService(chatRepo, store, fileRepo, quota, obj)
	if cfg.LiveKitAPIKey != "" && cfg.LiveKitAPISecret != "" {
		callTokenGen := call.NewTokenGenerator(cfg.LiveKitAPIKey, cfg.LiveKitAPISecret, 6*time.Hour)
		chatSvc.SetCallConfig(callTokenGen, cfg.LiveKitPublicURL)
	}
	chatHandlers := chat.NewHandler(chatSvc)

	storageHandlers := storage.NewHandler(quota)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(httpx.RequestLog())
	metrics := newRequestMetrics()
	engine.Use(metrics.Middleware())
	if len(cfg.CORSAllowedOrigins) > 0 {
		engine.Use(httpx.CORS(cfg.CORSAllowedOrigins))
	}
	engine.GET("/healthz", healthz(pool))
	engine.GET("/readyz", readiness(pool))
	engine.GET("/metrics", metrics.Handler)
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
	shareHandlers.RegisterRoutes(v1, storage...)
	chatHandlers.RegisterRoutes(v1, storage...)
	storageHandlers.RegisterRoutes(v1, storage...)
	shareLinkHandlers.RegisterOwnerRoutes(v1, storage...)
	activityHandlers.Register(engine, authHandlers.AuthRequired(), authHandlers.EmailVerifiedRequired())

	public := v1.Group("", httpx.NewIPRateLimiter(cfg.PublicShareRateLimitPerMin).Middleware())
	shareLinkHandlers.RegisterPublicRoutes(public)

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

func readiness(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}
		var migrationsReady bool
		if err := pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.tables
				WHERE table_schema = current_schema()
					AND table_name = 'schema_migrations'
			)
		`).Scan(&migrationsReady); err != nil || !migrationsReady {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	}
}

func NewTrashService(pool *pgxpool.Pool, obj objectstore.ObjectStore) *trash.Service {
	store := postgres.NewStore(pool)
	recorder := activity.NewRecorder(postgres.NewActivityRepository(store))
	return trash.NewService(postgres.NewTrashRepository(store), postgres.NewQuotaStore(store), obj, recorder)
}

func NewActivityRepository(pool *pgxpool.Pool) activity.Repository {
	return postgres.NewActivityRepository(postgres.NewStore(pool))
}
