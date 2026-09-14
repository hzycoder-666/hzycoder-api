package routes

import (
	"github.com/gin-gonic/gin"
	"hzycoder.com/lion/internal/handler"
	"hzycoder.com/lion/internal/middleware"
	"hzycoder.com/lion/internal/model"
	"hzycoder.com/lion/internal/repository"
)

type Deps struct {
	JWTSecret       string
	UserRepo        repository.UserRepository
	LoginLimiter    *middleware.RateLimiter
	RegisterLimiter *middleware.RateLimiter
	Health          *handler.HealthHandler
	Auth            *handler.AuthHandler
	Users           *handler.UserHandler
	Questions       *handler.QuestionHandler
	Papers          *handler.PaperHandler
	Attempts        *handler.AttemptHandler
}

func SetupRouter(deps Deps) *gin.Engine {
	r := gin.New()

	r.Use(middleware.RecoveryWithBizError())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.Logger())
	r.Use(middleware.BizErrorHandler())

	r.GET("/healthz", deps.Health.Healthz)
	r.GET("/readyz", deps.Health.Readyz)

	r.GET("/docs", handler.DocsRedoc)
	r.GET("/docs/swagger.json", handler.DocsSpec)
	r.GET("/docs/redoc.standalone.js", handler.DocsRedocJS)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", deps.LoginLimiter.Limit(), deps.Auth.Login)
			auth.POST("/register", deps.RegisterLimiter.Limit(), deps.Auth.Register)
		}

		v1 := api.Group("/v1")
		v1.Use(middleware.AuthMiddleware(deps.JWTSecret))
		{
			v1.GET("/users/me", deps.Users.Me)
			v1.GET("/users/:username", middleware.RequireRole(deps.UserRepo, model.RoleAdmin), deps.Users.GetByUsername)

			// 试卷系统-用户端（练习模式）
			v1.GET("/papers", deps.Papers.ListPublished)
			v1.GET("/papers/:id", deps.Papers.GetPublished)
			v1.POST("/papers/:id/submit", deps.Attempts.Submit)
			v1.GET("/attempts", deps.Attempts.List)
			v1.GET("/attempts/:id", deps.Attempts.GetByID)

			// 试卷系统-管理端
			admin := v1.Group("/admin", middleware.RequireRole(deps.UserRepo, model.RoleAdmin))
			{
				admin.POST("/questions", deps.Questions.Create)
				admin.GET("/questions", deps.Questions.List)
				admin.GET("/questions/:id", deps.Questions.GetByID)
				admin.PUT("/questions/:id", deps.Questions.Update)
				admin.DELETE("/questions/:id", deps.Questions.Delete)

				admin.POST("/papers", deps.Papers.Create)
				admin.GET("/papers", deps.Papers.List)
				admin.GET("/papers/:id", deps.Papers.GetByID)
				admin.PUT("/papers/:id", deps.Papers.Update)
				admin.POST("/papers/:id/publish", deps.Papers.Publish)
				admin.POST("/papers/:id/offline", deps.Papers.Offline)
				admin.GET("/papers/:id/attempts", deps.Attempts.AdminListByPaper)
				admin.GET("/papers/:id/stats", deps.Attempts.AdminStats)
			}
		}
	}

	return r
}
