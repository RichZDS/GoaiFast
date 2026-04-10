# Gemini API - Gin Web Server

这是一个基于 Go 语言 [Gin 框架](https://github.com/gin-gonic/gin) 开发的 Web 服务示例，用于演示如何构建高性能的 API 接口。

## 🚀 项目特性

- **高性能**: 基于 Gin 框架，路由响应极快。
- **多模型路由网关 (Model Router)**: 支持配置多个 API Key 实例，内部实现负载均衡 (Round Robin) 与自动熔断/重试降级机制，增强 API 调用的稳定性。
- **配置驱动**: 使用 `config.yml` 管理敏感信息和应用配置，支持多个 `GEMINI_API_KEYS`。
- **现代化 Go**: 使用 Go 1.25+ 版本特性，接口式编程，方便未来横向扩展至 OpenAI、Claude 等。
- **易于扩展**: 采用标准的 Router -> Controller -> Service 结构。

## 🛠️ 环境要求

- **Go**: 1.25.0 或更高版本
- **网络**: 能够访问 Google Gemini API 服务（如需集成 Gemini）

## 📦 安装与运行

### 1. 克隆项目

```bash
git clone <your-repo-url>
cd geminiApi
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置说明

在根目录下创建或编辑 `config.yml`：

```yaml
# 可以配置单个或多个 API Key，框架会自动进行负载均衡和故障重试
GEMINI_API_KEYS: 
  - "YOUR_API_KEY_1"
  - "YOUR_API_KEY_2"
```

### 4. 运行服务

```bash
go run main.go
```

服务启动后，默认监听端口为 `:8080`。你会在控制台看到 `Router 注册 Provider: Gemini-Key-0` 的日志。

## 📁 项目结构

```text
.
├── config/             # 配置管理 (config.yml 加载)
│   └── config.go
├── internal/
│   ├── controller/     # 控制层 (获取参数，调用 Router Service，返回 JSON)
│   │   └── gemini_controller.go
│   ├── provider/       # 大模型抽象层 (定义统一接口，封装各家 API)
│   │   ├── provider.go
│   │   └── gemini_provider.go
│   ├── service/        # 业务逻辑层 (负责多 Provider 的负载均衡与熔断切换)
│   │   └── router_service.go
│   └── model/          # 数据模型层 (结构体定义)
├── middleware/         # Gin 中间件
├── router/             # 路由配置
│   └── router.go
├── config.yml          # 本地配置文件 (需包含 GEMINI_API_KEYS)
├── go.mod              # Go 模块定义
├── main.go             # 程序入口 (初始化路由网关并注入)
└── README.md           # 项目说明文档
```

## 🔌 API 接口说明

| 方法 | 路径 | 参数 | 说明 |
| :--- | :--- | :--- | :--- |
| GET | `/ping` | - | 健康检查，返回 `pong` |
| GET | `/v1/generate` | `prompt` | 使用 Gemini 生成内容 |
