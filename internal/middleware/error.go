package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"hzycoder.com/lion/pkg/response"
)

func RecoveryWithBizError() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		slog.Error("panic recovered",
			"error", fmt.Sprintf("%+v", recovered),
			"stack", string(debug.Stack()))
		response.AbortWithCode(c, 500, response.CodeSystemError, "internal server error")
	})
}

func BizErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Writer.Written() || len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		var bizErr *response.BizError
		var valErr validator.ValidationErrors

		logAttrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", c.FullPath()),
			slog.String("client_ip", c.ClientIP()),
			slog.String("user_agent", c.GetHeader("User-Agent")),
		}

		switch {
		case errors.As(err, &bizErr):
			logAttrs = append(logAttrs,
				slog.Int("biz_code", bizErr.Code),
				slog.String("biz_msg", bizErr.Message),
			)
			slog.LogAttrs(c.Request.Context(), slog.LevelWarn, "business error", logAttrs...)
			response.HandleError(c, bizErr)
		case errors.As(err, &valErr):
			msg := response.ParseValidationError(valErr)
			logAttrs = append(logAttrs,
				slog.String("validate_err", valErr.Error()),
				slog.String("err_msg", msg),
			)
			slog.LogAttrs(c.Request.Context(), slog.LevelWarn, "validation failed", logAttrs...)
			response.AbortWithCode(c, 400, response.CodeParamInvalid, msg)
		default:
			logAttrs = append(logAttrs, slog.String("err_detail", err.Error()))
			slog.LogAttrs(c.Request.Context(), slog.LevelError, "internal error", logAttrs...)
			response.HandleError(c, err)
		}
	}
}
