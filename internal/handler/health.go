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

func (h *HealthHandler) Healthz(c *gin.Context) {
	response.Success(c, gin.H{
		"status": "ok",
		"uptime": time.Since(h.startedAt).String(),
	})
}

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
