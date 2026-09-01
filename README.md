# Go Gin Template

这是一个最小生产级 Go Web API 模板，技术栈为 **Gin + sqlx + MySQL + Viper + slog + JWT**。

它的目标不是做成大而全的脚手架，而是保留真实后端项目需要的基本骨架，同时让代码足够清楚，适合学习和继续扩展业务。

## 使用模板后的第一步

从 GitHub 的 **Use this template** 创建新仓库后，建议先做这几件事：

```bash
go mod edit -module your-company.com/your-project
make init-config
make tidy
make test
```

然后全局替换旧模块路径：

```text
hzycoder.com/go-gin-template
```

替换成你的新模块路径，例如：

```text
github.com/your-name/questionnaire-api
```

再根据实际数据库修改：

```text
config/config.yaml
.env
```

真实密钥、数据库密码、线上配置不要提交到 Git。

## 项目特点

- 分层结构清晰：`handler -> service -> repository -> database`
- 使用 `sqlx` 编写显式 SQL，适合学习 SQL，同时减少手动 `Scan` 的样板代码
- 在 `cmd/server/main.go` 中手动组装依赖，不依赖全局数据库对象
- Repository 使用接口边界，Service 可以用 fake repository 做单元测试
- 使用 Go 标准库 `slog` 输出 JSON 结构化日志
- 统一 API 响应格式和业务错误处理
- 错误响应同时包含业务码和正确的 HTTP 状态码
- 使用 Gin validator 做请求参数校验
- 内置用户注册、登录、JWT 鉴权示例
- 提供 `/healthz` 存活检查和 `/readyz` 就绪检查
- 提供基础安全响应头中间件
- 支持优雅关闭
- 提供 `Dockerfile`、`Makefile`、SQL 初始化脚本
- 提供 GitHub Actions 测试工作流
- 只提交安全配置样例，不提交真实密钥

## 目录结构

```text
.
├── .github/workflows/ci.yml        # GitHub Actions 测试工作流
├── cmd/server/main.go              # 程序入口，加载配置、初始化依赖、启动 HTTP 服务
├── config/config.example.yaml      # 配置样例
├── docs/template-usage.md          # 模板使用清单
├── internal
│   ├── auth                         # JWT 生成与解析
│   ├── config                       # 配置加载
│   ├── database                     # MySQL/sqlx 初始化
│   ├── handler                      # HTTP 请求处理层
│   ├── middleware                   # Gin 中间件
│   ├── model                        # 数据模型
│   ├── repository                   # SQL 数据访问层
│   ├── routes                       # 路由注册
│   └── service                      # 业务逻辑层
├── pkg
│   ├── logger                       # slog 初始化
│   ├── response                     # 统一响应和业务错误码
│   └── utils                        # 通用工具
├── sql/ddl.sql                      # 数据库初始化 SQL
├── Dockerfile
├── Makefile
├── go.mod
└── go.sum
```

## 配置

先从样例创建本地配置文件：

```bash
make init-config
```

等价于：

```bash
cp config/config.example.yaml config/config.yaml
cp .env.example .env
```

Windows PowerShell 可以使用：

```powershell
Copy-Item config/config.example.yaml config/config.yaml
Copy-Item .env.example .env
```

程序默认读取：

```text
config/config.yaml
```

也可以通过启动参数或环境变量指定配置文件：

```bash
go run ./cmd/server -config config/config.example.yaml
APP_CONFIG=config/config.example.yaml go run ./cmd/server
```

配置项也支持通过 `APP_` 前缀的环境变量覆盖：

```bash
APP_SERVER_PORT=8080
APP_DATABASE_DSN='root:password@tcp(127.0.0.1:3306)/go_gin_template?charset=utf8mb4&parseTime=True&loc=Local'
APP_JWT_SECRET='replace-with-a-long-random-secret'
APP_LOG_LEVEL=info
```

注意：真实环境的数据库密码、JWT 密钥等敏感信息不要提交到 Git。`config/config.yaml` 和 `.env` 已经在 `.gitignore` 中忽略。

## 数据库

先创建数据库：

```sql
CREATE DATABASE go_gin_template CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

然后执行：

```text
sql/ddl.sql
```

当前模板只内置一张 `users` 表，用于演示注册、登录和 JWT 鉴权。

Repository 层使用 `sqlx.GetContext` 和 `sqlx.NamedExecContext`，SQL 仍然是手写的，例如：

```go
err := db.GetContext(ctx, &user, `
    SELECT id, username, password, nickname, role, created_at, updated_at
    FROM users
    WHERE username = ?
`, username)
```

这样既能保留 SQL 的透明度，又不用写大量重复的 `row.Scan(...)`。

## 启动

下载依赖：

```bash
go mod download
```

启动服务：

```bash
go run ./cmd/server
```

或者使用 Make：

```bash
make run
```

## 接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/healthz` | 存活检查，只表示进程还活着 |
| `GET` | `/readyz` | 就绪检查，会检查数据库连接 |
| `POST` | `/api/auth/register` | 用户注册 |
| `POST` | `/api/auth/login` | 用户登录 |
| `GET` | `/api/v1/users/me` | 获取当前登录用户信息 |
| `GET` | `/api/v1/users/:username` | 按用户名查询用户，仅管理员可访问 |

注册请求示例：

```json
{
  "username": "demo001",
  "password": "Demo123!",
  "confirm_password": "Demo123!"
}
```

登录请求示例：

```json
{
  "username": "demo001",
  "password": "Demo123!"
}
```

需要鉴权的接口使用 Bearer Token：

```http
Authorization: Bearer <token>
```

## 响应格式

成功响应：

```json
{
  "code": 0,
  "msg": "success",
  "data": {}
}
```

失败响应：

```json
{
  "code": 1001,
  "msg": "参数错误"
}
```

错误处理由 `internal/middleware/error.go` 统一收口：

- 参数校验错误返回 `400`
- 未授权返回 `401`
- 禁止访问返回 `403`
- 资源不存在返回 `404`
- 系统错误返回 `500`

## 日志

日志使用 Go 标准库 `slog`，默认输出 JSON 格式。

请求日志由 `internal/middleware/logger.go` 记录，包含：

- 请求方法
- 请求路径
- query 参数
- HTTP 状态码
- 请求耗时

日志等级通过配置控制：

```yaml
log:
  level: info
```

支持：`debug`、`info`、`warn`、`error`。

## 测试

运行全部测试：

```bash
make test
```

或直接使用：

```bash
go test ./...
```

Go 的单元测试通常和被测试代码放在同一个 package 目录下，例如：

```text
internal/auth/jwt_test.go
internal/handler/health_test.go
internal/service/auth_test.go
```

当前 Service 测试使用 fake repository，不依赖真实数据库。

## 构建

```bash
make build
```

构建产物默认输出到：

```text
bin/server
```

## Docker

构建镜像：

```bash
docker build -t go-gin-template .
```

运行容器时建议用环境变量或挂载配置文件传入真实配置。

## 生产前检查

在把这个模板用于真实项目之前，建议至少完成以下事情：

- 替换所有样例密钥，不要使用默认 JWT secret
- 如果历史提交中出现过真实密钥，立即轮换这些密钥
- 为数据库配置备份、账号权限和连接数限制
- 项目开始频繁改表后，引入数据库迁移工具
- 为 Repository 增加基于临时数据库的集成测试
- 根据业务需要补充限流、跨域、链路追踪和监控指标
- 根据部署环境完善 Docker Compose、Kubernetes 或 CI/CD 配置

## 适用场景

这个模板适合：

- 学习 Go Web 后端项目结构
- 学习 Gin + SQL/sqlx 的组合
- 快速启动一个小中型 REST API 项目
- 作为更复杂脚手架之前的清晰基础版

它不追求像 Nunu 那样大而全，也暂时没有代码生成、Wire、Redis、迁移系统和完整部署编排。当前重点是：**小、清楚、能跑、可测试、容易继续扩展**。

