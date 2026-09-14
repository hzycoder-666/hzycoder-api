#!/usr/bin/env bash
# 重新生成 OpenAPI 文档（docs/swagger.json），供 /docs Redoc 页面使用。
# 用法: bash scripts/gen_docs.sh   （修改 handler 注解后执行）
set -euo pipefail

cd "$(dirname "$0")/.."

SWAG="$(go env GOPATH)/bin/swag"
if [ ! -x "$SWAG" ]; then
  echo "swag 未安装，正在安装..." >&2
  go install github.com/swaggo/swag/cmd/swag@latest
fi

"$SWAG" init -g main.go -d cmd/server,internal -o docs --parseDependency --parseInternal
echo "已生成 docs/swagger.json，重启服务后访问 http://localhost:8080/docs"
