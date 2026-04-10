package provider

import "context"

// LLMProvider 定义了与大语言模型交互的统一接口
type LLMProvider interface {
	// GenerateText 根据给定的 prompt 生成文本回复
	GenerateText(ctx context.Context, prompt string) (string, error)

	// GetName 返回该 Provider 的名称标识 (用于日志和路由追踪)
	GetName() string

	// CheckHealth 检查提供商当前是否可用 (用于探活)
	CheckHealth(ctx context.Context) error
}
