# Gemini API - Gin Web Server (标准重构版)

这是一个基于 Go 语言 [Gin 框架](https://github.com/gin-gonic/gin) 开发的 Web 服务示例。项目采用了严格的 **Go 社区标准项目目录结构 (`golang-standards/project-layout`)**，展示了如何构建一个具备高可用性、自动熔断与探活恢复的 LLM API 网关。

## 🚀 项目特性

- **Go 社区标准规范**: 代码严密遵循 `cmd`, `internal`, `configs` 目录分离，解耦配置、HTTP层与核心业务逻辑。
- **高可用网关 (API Gateway)**: 
  - **负载均衡**: 对配置的多个 API Key 进行轮询调度 (Round Robin)。
  - **被动隔离 (熔断机制)**: 当某个 API Key 连续调用失败时，会将其临时隔离，防止阻塞用户请求。
  - **自动探活与恢复**: 后台协程定时对“不健康”的节点发送轻量测试，一旦恢复正常，自动重新加入路由池中。
- **配置与安全分离**: 使用 `.gitignore` 保护包含真实凭据的 `config.yml` 不被提交，使用 `config.yml.example` 作为团队协同模板。

## 🛠️ 环境要求

- **Go**: 1.25.0 或更高版本
- **网络**: 能够访问 Google Gemini API 服务（如需直接使用国内网络，可自行配置代理）

## 📦 安装与运行

### 1. 克隆项目与安装依赖

```bash
git clone <your-repo-url>
cd geminiApi
go mod tidy
```

### 2. 配置说明

复制配置文件示例并重命名为真实配置文件：

```bash
cp configs/config.yml.example configs/config.yml
```

编辑 `configs/config.yml`，填入你的多个 API Keys：

```yaml
GEMINI_API_KEYS: 
  - "YOUR_API_KEY_1"
  - "YOUR_API_KEY_2"
```

### 3. 运行服务

由于项目入口被迁移至 `cmd/api` 下，需要在根目录执行以下命令启动：

```bash
go run cmd/api/main.go
```

或者编译成可执行文件后运行：

```bash
go build -o gemini-api cmd/api/main.go
./gemini-api
```

服务启动后，默认监听端口为 `:8080`。你会在控制台看到 `Gateway 注册 Provider` 相关的日志。

## 📁 目录结构与核心函数说明

### 1. 入口与配置
- `cmd/api/main.go` 
  - **`main()`**: 程序的唯一入口。负责依次：加载本地配置、初始化 GatewayService (API 网关)、初始化 HTTP Handler、注册 Gin 路由并最终启动 `:8080` 端口监听。
- `configs/config.yml`
  - 存放应用程序运行时的真实配置文件（不受 Git 追踪）。
- `internal/config/config.go`
  - **`LoadConfig()`**: 负责读取 `configs/config.yml` 文件并反序列化到 `Config` 结构体内存对象中，供全局使用。

### 2. 外部 API 交互层 (Provider)
将对大模型厂商的 API 调用抽象为统一标准。
- `internal/provider/provider.go`
  - **`LLMProvider` (接口)**: 定义了 `GenerateText` (生成文本)、`GetName` (获取提供商标识)、`CheckHealth` (健康检查) 三个标准方法，方便未来无缝接入 OpenAI、Claude 等。
- `internal/provider/gemini_provider.go`
  - **`NewGeminiProvider()`**: 利用官方 SDK 创建 Gemini Client 实例。
  - **`GenerateText()`**: 调用 Gemini 模型 SDK 发起真实的文本生成请求。
  - **`CheckHealth()`**: 为了避免探活消耗过多 token，向大模型发送极短文本（"hi"），只要不返回网络或 Auth 错误即可判定 Key 存活。

### 3. 核心业务层 (Service / Gateway)
- `internal/service/gateway_service.go`
  - **`SetupGatewayServiceFromConfig()`**: 遍历配置文件中的所有 API Key，批量初始化对应的 Provider，并开启一个独立的探活后台协程。
  - **`AddProvider()`**: 内部方法，包装 Provider 并附带 `isHealthy` 与 `failCount` 等状态加入到网关节点池中。
  - **`getNextNode()`**: 实现基于 Round-Robin（轮询）策略的负载均衡，并且在获取节点时**主动跳过不健康节点**。
  - **`GenerateText()`**: 网关的生成包装方法。发起请求时，若请求成功则重置节点失败计数；若失败则累加，**连续失败 2 次以上触发熔断被动隔离**，并自动切至备用 Key 重试，保证了用户的无感体验。
  - **`startHealthCheckTask()`**: 后台运行的探活协程（`Goroutine`）。每隔 1 分钟扫描一遍所有被隔离的（`!isHealthy`）节点，并主动调用其 `CheckHealth()`。一旦成功便清空错误计数，让该 API Key 重新回到调度池中。

### 4. HTTP 接口层 (Handler & Router)
- `internal/handler/gemini_handler.go`
  - **`NewGeminiHandler()`**: 控制器初始化，注入底层的 `GatewayService` 以便调用业务逻辑。
  - **`GenerateText()`**: 解析用户请求传入的 `?prompt=xxx` 参数，调用 Gateway 获取大模型结果并组装为 JSON 返回。如果参数为空则返回 `400 Bad Request`。
  - **`Ping()`**: `/ping` 接口的基础回调，验证服务本身是否存活。
- `internal/router/router.go`
  - **`NewRouter()`**: 将各种 Handler 绑定到对应的 URL 路由上，例如创建 `/v1` 路由组，关联 `/v1/generate`。

## 🔌 API 接口详情

| 方法 | 路径 | URL 参数 | 说明 |
| :--- | :--- | :--- | :--- |
| GET | `/ping` | - | Web 服务本身的健康检查，返回 `pong` |
| GET | `/v1/generate` | `?prompt=你的问题` | 调用底层大模型网关生成内容 |

**调用示例**：
```bash
curl "http://localhost:8080/v1/generate?prompt=你好"
```
预期响应：
```json
{
  "prompt": "你好",
  "response": "你好！有什么我可以帮您的吗？"
}
```
