package middleware

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"hzycoder.com/lion/internal/auth"
	"hzycoder.com/lion/internal/model"
	"hzycoder.com/lion/pkg/response"
)

// RoleProvider 按用户 ID 查询当前角色；授权以数据库为准，角色变更或注销立即生效
type RoleProvider interface {
	FindRoleByID(ctx context.Context, userID int64) (model.Role, error)
}

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

func RequireRole(provider RoleProvider, allowedRoles ...model.Role) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, ok := GetUserID(ctx)
		if !ok {
			response.Abort(ctx, "missing user id")
			return
		}

		role, err := provider.FindRoleByID(ctx.Request.Context(), userID)
		if errors.Is(err, sql.ErrNoRows) {
			response.Abort(ctx, "user not found")
			return
		}
		if err != nil {
			slog.Error("lookup user role failed", "user_id", userID, "error", err)
			response.AbortWithCode(ctx, http.StatusInternalServerError, response.CodeSystemError)
			return
		}

		for _, allowed := range allowedRoles {
			if role == allowed {
				ctx.Next()
				return
			}
		}

		response.AbortWithCode(ctx, http.StatusForbidden, response.CodeForbidden)
	}
}

func GetUserID(c *gin.Context) (int64, bool) {
	v, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}
