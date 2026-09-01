package middleware

import (
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	"hzycoder.com/go-gin-template/internal/auth"
	"hzycoder.com/go-gin-template/internal/model"
	"hzycoder.com/go-gin-template/pkg/response"
)

func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			response.Abort(ctx, "missing token")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			response.Abort(ctx, "invalid token")
			return
		}

		claims, err := auth.ParseToken([]byte(secret), parts[1])
		if err != nil {
			slog.Warn("jwt parse failed", "error", err)
			response.Abort(ctx, "invalid token")
			return
		}

		ctx.Set("user_id", claims.UserID)
		ctx.Set("role", claims.Role)
		ctx.Next()
	}
}

func RequireRole(allowedRoles ...model.Role) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		currentRole, exists := ctx.Get("role")
		if !exists {
			response.Abort(ctx, "invalid role")
			return
		}

		for _, allowed := range allowedRoles {
			if currentRole == allowed {
				ctx.Next()
				return
			}
		}

		response.Abort(ctx, "forbidden")
	}
}

func GetUserID(c *gin.Context) int64 {
	userID, _ := c.Get("user_id")
	id, _ := userID.(int64)
	return id
}
