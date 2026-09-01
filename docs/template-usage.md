# 模板使用清单

从这个模板创建新项目后，建议按下面顺序整理。

## 1. 修改模块名

```bash
go mod edit -module github.com/your-name/your-project
```

然后把代码里的旧模块路径全部替换掉：

```text
hzycoder.com/go-gin-template
```

## 2. 初始化本地配置

```bash
make init-config
```

这会创建：

```text
config/config.yaml
.env
```

这两个文件只用于本地开发，不提交到 Git。

## 3. 修改数据库配置

编辑 `config/config.yaml` 中的 MySQL DSN：

```yaml
database:
  dsn: "root:password@tcp(127.0.0.1:3306)/go_gin_template?charset=utf8mb4&parseTime=True&loc=Local"
```

## 4. 初始化数据库

先创建数据库，再执行 `sql/ddl.sql`。

## 5. 检查项目

```bash
make tidy
make test
```

## 6. 开始新增业务模块

推荐按现有分层增加文件：

```text
internal/model/example.go
internal/repository/example.go
internal/service/example.go
internal/handler/example.go
```

然后在 `internal/routes/router.go` 注册路由，在 `cmd/server/main.go` 组装依赖。

## 7. 上线前检查

- 已替换 JWT secret
- 数据库账号权限已收敛
- 本地配置未提交
- 测试通过
- Docker 构建通过
- 根据业务需要补充迁移、限流、监控和链路追踪
