package request

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

// ─── 配置 ───────────────────────────────────────────────

type Config struct {
	BaseURL       string
	Timeout       time.Duration // 整体超时
	RetryCount    int           // 重试次数
	RetryWaitTime time.Duration // 重试间隔
	AuthToken     string        // Bearer Token（可选）
	UserAgent     string        // UA 标识
	Debug         bool          // 是否开启 debug 日志
}

func DefaultConfig(baseURL string) Config {
	return Config{
		BaseURL:       baseURL,
		Timeout:       10 * time.Second,
		RetryCount:    3,
		RetryWaitTime: 500 * time.Millisecond,
		UserAgent:     "MyApp/1.0",
	}
}

// ─── Client ─────────────────────────────────────────────

type ApiClient struct {
	client *resty.Client
	logger *slog.Logger
}

func NewApiClient(cfg Config, logger *slog.Logger) *ApiClient {
	if logger == nil {
		logger = slog.Default()
	}

	c := resty.New().
		SetBaseURL(cfg.BaseURL).
		SetTimeout(cfg.Timeout).
		SetRetryCount(cfg.RetryCount).
		SetRetryWaitTime(cfg.RetryWaitTime).
		SetRetryMaxWaitTime(10 * time.Second).
		SetDebug(cfg.Debug)

	// 默认请求头
	c.SetHeader("Accept", "application/json")
	c.SetHeader("Content-Type", "application/json")
	if cfg.UserAgent != "" {
		c.SetHeader("User-Agent", cfg.UserAgent)
	}
	if cfg.AuthToken != "" {
		c.SetAuthToken(cfg.AuthToken)
	}

	// 重试条件：仅 5xx 和网络错误重试，4xx 不重试
	c.AddRetryCondition(func(resp *resty.Response, err error) bool {
		if err != nil {
			return true
		}
		return resp.StatusCode() >= 500 || resp.StatusCode() == http.StatusTooManyRequests
	})

	api := &ApiClient{client: c, logger: logger}
	api.setupHooks()
	return api
}

// ─── 钩子：日志 & 追踪 ──────────────────────────────────

func (a *ApiClient) setupHooks() {
	// 请求前：注入 Request-ID、记录开始时间
	a.client.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
		req.SetHeader("X-Request-ID", uuid.NewString())
		req.SetContext(context.WithValue(req.Context(), ctxKeyStartTime, time.Now()))
		return nil
	})

	// 响应后：记录耗时、状态码、错误
	a.client.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		start, _ := resp.Request.Context().Value(ctxKeyStartTime).(time.Time)
		duration := time.Since(start)

		a.logger.Info("http_request",
			slog.String("method", resp.Request.Method),
			slog.String("url", resp.Request.URL),
			slog.Int("status", resp.StatusCode()),
			slog.Duration("duration", duration),
			slog.String("request_id", resp.Request.Header.Get("X-Request-ID")),
		)

		if resp.IsError() {
			a.logger.Error("http_error",
				slog.Int("status", resp.StatusCode()),
				slog.String("body", resp.String()),
			)
		}
		return nil
	})
}

type ctxKey string

const ctxKeyStartTime ctxKey = "start_time"

// ─── 便捷方法 ────────────────────────────────────────────

// Get 发起 GET 请求，自动将响应解析到 out
func (a *ApiClient) Get(ctx context.Context, path string, out any, params map[string]string) error {
	req := a.client.R().SetContext(ctx)
	if params != nil {
		req.SetQueryParams(params)
	}
	if out != nil {
		req.SetResult(out)
	}

	resp, err := req.Get(path)
	return a.wrapError(resp, err)
}

// Post 发起 POST 请求
func (a *ApiClient) Post(ctx context.Context, path string, body any, out any) error {
	req := a.client.R().SetContext(ctx).SetBody(body)
	if out != nil {
		req.SetResult(out)
	}

	resp, err := req.Post(path)
	return a.wrapError(resp, err)
}

// Put 发起 PUT 请求
func (a *ApiClient) Put(ctx context.Context, path string, body any, out any) error {
	req := a.client.R().SetContext(ctx).SetBody(body)
	if out != nil {
		req.SetResult(out)
	}

	resp, err := req.Put(path)
	return a.wrapError(resp, err)
}

// Delete 发起 DELETE 请求
func (a *ApiClient) Delete(ctx context.Context, path string) error {
	resp, err := a.client.R().SetContext(ctx).Delete(path)
	return a.wrapError(resp, err)
}

// ─── 错误处理 ────────────────────────────────────────────

type APIError struct {
	StatusCode int
	Body       string
	Err        error
}

func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("request failed: %v", e.Err)
	}
	return fmt.Sprintf("api error: status=%d, body=%s", e.StatusCode, e.Body)
}

func (a *ApiClient) wrapError(resp *resty.Response, err error) error {
	if err != nil {
		return &APIError{Err: err}
	}
	if resp.IsError() {
		return &APIError{
			StatusCode: resp.StatusCode(),
			Body:       string(resp.Body()),
		}
	}
	return nil
}

// ─── 底层访问（特殊场景） ────────────────────────────────

// Raw 返回底层 resty.Request，用于文件上传等复杂场景
func (a *ApiClient) Raw(ctx context.Context) *resty.Request {
	return a.client.R().SetContext(ctx)
}

// SetAuthToken 动态切换 Token（如多租户场景）
func (a *ApiClient) SetAuthToken(token string) {
	a.client.SetAuthToken(token)
}
