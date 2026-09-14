package handler

import (
	"embed"
	"net/http"

	"github.com/gin-gonic/gin"
	"hzycoder.com/lion/docs"
)

//go:embed static/redoc.standalone.js
var redocFS embed.FS

const redocHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Lion 试卷系统 API 文档</title>
  <style>
    body { margin: 0; padding: 0; }
  </style>
</head>
<body>
  <redoc spec-url="/docs/swagger.json"></redoc>
  <script src="/docs/redoc.standalone.js"></script>
</body>
</html>`

// DocsRedoc 渲染 Redoc 文档页面（本地打包，无需外网 CDN）
func DocsRedoc(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(redocHTML))
}

// DocsSpec 输出 OpenAPI JSON（由 swag init 生成，经 docs.SwaggerInfo 注入版本信息）
func DocsSpec(c *gin.Context) {
	jsonStr := docs.SwaggerInfo.ReadDoc()
	c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(jsonStr))
}

// DocsRedocJS 提供本地打包的 redoc.standalone.js
func DocsRedocJS(c *gin.Context) {
	data, err := redocFS.ReadFile("static/redoc.standalone.js")
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Data(http.StatusOK, "application/javascript; charset=utf-8", data)
}
