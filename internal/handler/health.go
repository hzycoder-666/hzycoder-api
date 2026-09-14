package handler

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"hzycoder.com/lion/pkg/response"
)

type HealthChecker interface {
	PingContext(ctx context.Context) error
}

type HealthHandler struct {
	startedAt time.Time
	db        HealthChecker
}

func NewHealthHandler(db HealthChecker) *HealthHandler {
	return &HealthHandler{
		startedAt: time.Now(),
		db:        db,
	}
}

// Healthz
// @Summary 存活探针
// @Description 返回服务存活状态与运行时长，不检查依赖
// @Tags 系统
// @Produce json
// @Success 200 {object} response.Resp
// @Router /healthz [get]
func (h *HealthHandler) Healthz(c *gin.Context) {
	response.Success(c, gin.H{
		"status": "ok",
		"uptime": time.Since(h.startedAt).String(),
	})
}

// Readyz
// @Summary 就绪探针
// @Description 检查数据库连通性，依赖异常时返回 503
// @Tags 系统
// @Produce json
// @Success 200 {object} response.Resp
// @Failure 503 {object} response.Resp
// @Router /readyz [get]
func (h *HealthHandler) Readyz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if h.db != nil {
		if err := h.db.PingContext(ctx); err != nil {
			response.AbortWithCode(c, 503, response.CodeSystemError, "service unavailable")
			return
		}
	}

	response.Success(c, gin.H{"status": "ready"})
}
