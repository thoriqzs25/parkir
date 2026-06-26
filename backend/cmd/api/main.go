package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	authsvc "github.com/thoriqzs/PARKIR/backend/internal/auth"
	"github.com/thoriqzs/PARKIR/backend/internal/config"
	"github.com/thoriqzs/PARKIR/backend/internal/db"
	authdomain "github.com/thoriqzs/PARKIR/backend/internal/domain/auth"
	"github.com/thoriqzs/PARKIR/backend/internal/domain/health"
	"github.com/thoriqzs/PARKIR/backend/internal/domain/locations"
	"github.com/thoriqzs/PARKIR/backend/internal/domain/rates"
	"github.com/thoriqzs/PARKIR/backend/internal/domain/roles"
	"github.com/thoriqzs/PARKIR/backend/internal/domain/sessions"
	"github.com/thoriqzs/PARKIR/backend/internal/domain/shifts"
	"github.com/thoriqzs/PARKIR/backend/internal/domain/transactions"
	"github.com/thoriqzs/PARKIR/backend/internal/domain/users"
	"github.com/thoriqzs/PARKIR/backend/internal/logger"
	"github.com/thoriqzs/PARKIR/backend/internal/middleware"
	"github.com/thoriqzs/PARKIR/backend/internal/permissions"
	"github.com/thoriqzs/PARKIR/backend/internal/store"
)

func main() {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("backend/.env")

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.LogLevel)

	pool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := db.Migrate(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
		log.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	authService, err := authsvc.NewService(cfg.JWTPrivateKeyPath, cfg.JWTPublicKeyPath)
	if err != nil {
		log.Error("failed to initialize auth service", "error", err)
		os.Exit(1)
	}

	store := store.New(pool)
	permResolver := permissions.NewResolver(pool)

	router := gin.New()
	router.Use(middleware.RequestLogger(log))
	router.Use(middleware.Recovery(log))
	router.Use(middleware.CORS(cfg))

	health.RegisterRoutes(router, pool)

	authHandler := authdomain.NewHandler(authService, store)

	// Public API routes (no auth required)
	public := router.Group("/api/v1")
	{
		authHandler.RegisterPublicRoutes(public)
	}

	// Protected API routes (auth required)
	api := router.Group("/api/v1")
	api.Use(middleware.Auth(authService, permResolver))
	{
		authHandler.RegisterProtectedRoutes(api)

		users.NewHandler(store).RegisterRoutes(api)
		roles.NewHandler(store).RegisterRoutes(api)

		locations.NewHandler(store).RegisterRoutes(api)
		rates.NewHandler(store).RegisterRoutes(api.Group("/locations"), api.Group("/rates"))

		sessions.NewHandler(store).RegisterRoutes(api)
		transactions.NewHandler(store).RegisterRoutes(api)
		shifts.NewHandler(store).RegisterRoutes(api)
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: router,
	}

	go func() {
		log.Info("starting server", "port", cfg.Port, "env", cfg.Environment)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server shutdown error", "error", err)
	}

	log.Info("server stopped")
}
