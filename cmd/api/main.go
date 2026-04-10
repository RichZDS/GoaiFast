package main

import (
	"context"
	"geminiApi/internal/config"
	"geminiApi/internal/handler"
	"geminiApi/internal/router"
	"geminiApi/internal/service"
	"log"
)

func main() {
	// 1. 初始化配置 (config)
	// 因为 main.go 现在在 cmd/api/ 下运行，配置文件在项目的 configs/ 目录下
	cfg, err := config.LoadConfig("configs/config.yml")
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	ctx := context.Background()

	// 2. 初始化 API Key 网关 (GatewayService)
	// 所有的 API Key 轮询、模型实例化的复杂逻辑被隐藏在 SetupGatewayServiceFromConfig 中
	gatewaySvc, err := service.SetupGatewayServiceFromConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("初始化 API 网关失败: %v", err)
	}

	// 3. 初始化 HTTP 请求处理层 (handler)，将网关注入
	geminiHandler := handler.NewGeminiHandler(gatewaySvc)

	// 4. 初始化 HTTP 路由层 (router/api)
	r := router.NewRouter(geminiHandler)

	log.Println("服务器正在启动，监听端口 :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("服务器运行出错: %v", err)
	}
}
