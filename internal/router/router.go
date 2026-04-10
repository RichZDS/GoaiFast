package router

import (
	"geminiApi/internal/handler"

	"github.com/gin-gonic/gin"
)

func NewRouter(h *handler.GeminiHandler) *gin.Engine {
	r := gin.Default()

	// 健康检查
	r.GET("/ping", h.Ping)

	// Gemini 路由组
	v1 := r.Group("/v1")
	{
		// 定义 GET 请求
		v1.GET("/generate", h.GenerateText)
	}

	return r
}
