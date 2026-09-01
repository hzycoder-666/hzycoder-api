package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hzycoder.com/go-gin-template/internal/config"
	"hzycoder.com/go-gin-template/internal/database"
	"hzycoder.com/go-gin-template/internal/handler"
	"hzycoder.com/go-gin-template/internal/repository"
	"hzycoder.com/go-gin-template/internal/routes"
	"hzycoder.com/go-gin-template/internal/service"
	"hzycoder.com/go-gin-template/pkg/logger"

	_ "hzycoder.com/go-gin-template/pkg/utils"
)

func main() {
	configPath := flag.String("config", defaultConfigPath(), "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		panic(err)
	}

	logger.Init(cfg.Log.Level)

	if cfg.JWT.Secret == "" {
		panic("jwt secret is required")
	}

	db, err := database.Open(*cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret, time.Duration(cfg.JWT.Expire)*time.Second)

	r := routes.SetupRouter(routes.Deps{
		JWTSecret: cfg.JWT.Secret,
		Health:    handler.NewHealthHandler(db),
		Auth:      handler.NewAuthHandler(authService),
		Users:     handler.NewUserHandler(userService),
	})

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("server started", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("server shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}
}

func defaultConfigPath() string {
	if path := os.Getenv("APP_CONFIG"); path != "" {
		return path
	}
	return "config/config.yaml"
}
