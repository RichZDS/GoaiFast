package router

import (
	"geminiApi/internal/controller"

	"github.com/gin-gonic/gin"
)

func NewRouter(ctrl *controller.GeminiController) *gin.Engine {
	r := gin.Default()

	// 健康检查
	r.GET("/ping", ctrl.Ping)

	// Gemini 路由组
	v1 := r.Group("/v1")
	{
		// 定义 GET 请求
		v1.GET("/generate", ctrl.GenerateText)
	}

	return r
}
