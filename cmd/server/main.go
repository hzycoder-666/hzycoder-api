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

	"hzycoder.com/lion/internal/config"
	"hzycoder.com/lion/internal/database"
	"hzycoder.com/lion/internal/handler"
	"hzycoder.com/lion/internal/repository"
	"hzycoder.com/lion/internal/routes"
	"hzycoder.com/lion/internal/service"
	"hzycoder.com/lion/pkg/logger"

	_ "hzycoder.com/lion/pkg/utils"

	_ "hzycoder.com/lion/docs"
)

// @title Lion 试卷系统 API
// @version 1.0
// @description 试卷系统后端服务：题库管理、手动组卷与发布、在线答题、自动评分、成绩查询与统计。
// @host localhost:8080
// @BasePath /api
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 登录或注册接口签发的 JWT，格式：Bearer <token>
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
	questionRepo := repository.NewQuestionRepository(db)
	paperRepo := repository.NewPaperRepository(db)
	attemptRepo := repository.NewAttemptRepository(db)

	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret, time.Duration(cfg.JWT.Expire)*time.Second)
	questionService := service.NewQuestionService(questionRepo)
	paperService := service.NewPaperService(paperRepo, questionRepo)
	attemptService := service.NewAttemptService(attemptRepo, paperRepo)

	r := routes.SetupRouter(routes.Deps{
		JWTSecret: cfg.JWT.Secret,
		Health:    handler.NewHealthHandler(db),
		Auth:      handler.NewAuthHandler(authService),
		Users:     handler.NewUserHandler(userService),
		Questions: handler.NewQuestionHandler(questionService),
		Papers:    handler.NewPaperHandler(paperService),
		Attempts:  handler.NewAttemptHandler(attemptService),
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
