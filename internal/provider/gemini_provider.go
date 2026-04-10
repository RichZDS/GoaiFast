package provider

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/genai"
)

// GeminiProvider 实现了 LLMProvider 接口
type GeminiProvider struct {
	client *genai.Client
	name   string // 例如 "Gemini-Key-1"
	model  string // 默认使用的模型，如 "gemini-2.0-flash"
}

// NewGeminiProvider 创建一个新的 Gemini 客户端实例
func NewGeminiProvider(ctx context.Context, apiKey string, name string, model string) (*GeminiProvider, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client for %s: %v", name, err)
	}

	if model == "" {
		model = "gemini-2.0-flash" // 默认模型
	}

	return &GeminiProvider{
		client: client,
		name:   name,
		model:  model,
	}, nil
}

// GenerateText 实现 LLMProvider 接口
func (p *GeminiProvider) GenerateText(ctx context.Context, prompt string) (string, error) {
	log.Printf("[Provider: %s] 正在请求模型: %s", p.name, p.model)

	result, err := p.client.Models.GenerateContent(
		ctx,
		p.model,
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("[%s] error generating content: %w", p.name, err)
	}

	return result.Text(), nil
}

// GetName 实现 LLMProvider 接口
func (p *GeminiProvider) GetName() string {
	return p.name
}

// CheckHealth 实现 LLMProvider 接口
func (p *GeminiProvider) CheckHealth(ctx context.Context) error {
	// 发送一个极短的文本请求以验证 API Key 是否可用
	_, err := p.client.Models.GenerateContent(
		ctx,
		p.model,
		genai.Text("hi"),
		nil,
	)
	if err != nil {
		return fmt.Errorf("[%s] health check failed: %w", p.name, err)
	}
	return nil
}
