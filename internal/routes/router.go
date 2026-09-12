package routes

import (
	"github.com/gin-gonic/gin"
	"hzycoder.com/lion/internal/handler"
	"hzycoder.com/lion/internal/middleware"
	"hzycoder.com/lion/internal/model"
)

type Deps struct {
	JWTSecret string
	Health    *handler.HealthHandler
	Auth      *handler.AuthHandler
	Users     *handler.UserHandler
}

func SetupRouter(deps Deps) *gin.Engine {
	r := gin.New()

	r.Use(middleware.RecoveryWithBizError())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.Logger())
	r.Use(middleware.BizErrorHandler())

	r.GET("/healthz", deps.Health.Healthz)
	r.GET("/readyz", deps.Health.Readyz)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", deps.Auth.Login)
			auth.POST("/register", deps.Auth.Register)
		}

		v1 := api.Group("/v1")
		v1.Use(middleware.AuthMiddleware(deps.JWTSecret))
		{
			v1.GET("/users/me", deps.Users.Me)
			v1.GET("/users/:username", middleware.RequireRole(model.RoleAdmin), deps.Users.GetByUsername)
		}
	}

	return r
}
