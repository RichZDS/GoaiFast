package controller

import (
	"geminiApi/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GeminiController struct {
	routerService *service.RouterService
}

func NewGeminiController(rs *service.RouterService) *GeminiController {
	return &GeminiController{routerService: rs}
}

// GenerateText 获取 prompt 并调用路由网关
func (ctrl *GeminiController) GenerateText(c *gin.Context) {
	prompt := c.Query("prompt")
	if prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt 不能为空"})
		return
	}

	result, err := ctrl.routerService.GenerateText(c.Request.Context(), prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"prompt":   prompt,
		"response": result,
	})
}

// Ping 健康检查
func (ctrl *GeminiController) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
