package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"hzycoder.com/go-gin-template/internal/handler"
)

type fakeHealthChecker struct {
	err error
}

func (f fakeHealthChecker) PingContext(context.Context) error {
	return f.err
}

func TestHealthz(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	health := handler.NewHealthHandler(fakeHealthChecker{err: errors.New("db down")})
	r.GET("/healthz", health.Healthz)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestReadyz(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	health := handler.NewHealthHandler(fakeHealthChecker{})
	r.GET("/readyz", health.Readyz)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestReadyzUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	health := handler.NewHealthHandler(fakeHealthChecker{err: errors.New("db down")})
	r.GET("/readyz", health.Readyz)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", w.Code)
	}
}
