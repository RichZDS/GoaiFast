package main

import (
	"context"
	"geminiApi/config"
	"geminiApi/internal/controller"
	"geminiApi/internal/service"
	"geminiApi/router"
	"log"
)

func main() {
	// 1. 初始化配置 (config)
	cfg, err := config.LoadConfig("config.yml")
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	ctx := context.Background()

	// 2. 初始化路由网关 (RouterService)
	// 所有的 API Key 轮询、模型实例化的复杂逻辑被隐藏在 SetupRouterServiceFromConfig 中
	routerSvc, err := service.SetupRouterServiceFromConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("初始化路由网关失败: %v", err)
	}

	// 3. 初始化控制层 (controller)，将路由网关注入
	geminiCtrl := controller.NewGeminiController(routerSvc)

	// 4. 初始化 API 路由层 (router/api)
	r := router.NewRouter(geminiCtrl)

	log.Println("服务器正在启动，监听端口 :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("服务器运行出错: %v", err)
	}
}
