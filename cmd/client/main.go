package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"hzycoder.com/lion/pkg/request"
)

// 业务接口普遍遵循 code/message/data 信封结构，泛型让 Data 按具体类型解码
type response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// 导航信息按需声明字段即可，json 解码会忽略多余字段
type navData struct {
	IsLogin  bool   `json:"isLogin"`
	UserID   int64  `json:"mid"`
	Username string `json:"uname"`
	Face     string `json:"face"`
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := request.DefaultConfig("https://api.bilibili.com")
	cfg.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/152.0.0.0 Safari/537.36"
	apiClient := request.NewApiClient(cfg, log)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result response[navData]
	err := apiClient.Get(ctx, "/x/web-interface/nav", &result, nil)
	if err != nil {
		var apiErr *request.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode != 0 {
			log.Error("http error", slog.Int("status", apiErr.StatusCode), slog.String("body", apiErr.Body))
		} else {
			log.Error("request failed", slog.Any("err", err))
		}
		return
	}

	// HTTP 200 不代表业务成功：bilibili 把业务错误放在 body 的 code 里
	if result.Code != 0 {
		log.Error("api business error", slog.Int("code", result.Code), slog.String("message", result.Message))
		return
	}

	log.Info("nav ok",
		slog.Bool("isLogin", result.Data.IsLogin),
		slog.Int64("mid", result.Data.UserID),
		slog.String("uname", result.Data.Username),
		slog.String("face", result.Data.Face),
	)
}
