package service

import (
	"context"
	"fmt"
	"geminiApi/config"
	"geminiApi/internal/provider"
	"log"
	"sync"
	"time"
)

// providerNode 包装了 LLMProvider 及其健康状态
type providerNode struct {
	provider  provider.LLMProvider
	isHealthy bool
	failCount int
}

// RouterService 管理多个 LLM 提供商，并实现负载均衡和故障转移 (熔断重试)
type RouterService struct {
	nodes   []*providerNode
	mu      sync.Mutex
	current int // 记录当前轮询到的索引
}

// NewRouterService 初始化路由网关
func NewRouterService() *RouterService {
	return &RouterService{
		nodes:   make([]*providerNode, 0),
		current: 0,
	}
}

// SetupRouterServiceFromConfig 根据配置初始化路由网关，自动注册所有的 Gemini API Key
func SetupRouterServiceFromConfig(ctx context.Context, cfg *config.Config) (*RouterService, error) {
	routerSvc := NewRouterService()
	
	if len(cfg.GeminiAPIKeys) == 0 {
		return nil, fmt.Errorf("没有配置任何 Gemini API Key")
	}

	for i, apiKey := range cfg.GeminiAPIKeys {
		providerName := fmt.Sprintf("Gemini-Key-%d", i)
		// 这里可以使用配置文件中的默认模型，如果没有配置，则使用 "gemini-2.0-flash"
		p, err := provider.NewGeminiProvider(ctx, apiKey, providerName, "gemini-2.0-flash")
		if err != nil {
			log.Printf("[警告] 无法初始化 Provider %s: %v", providerName, err)
			continue
		}
		routerSvc.AddProvider(p)
	}
	
	if len(routerSvc.nodes) == 0 {
		return nil, fmt.Errorf("所有 Provider 初始化均失败，请检查 API Key 配置")
	}

	// 启动后台健康检查任务
	go routerSvc.startHealthCheckTask()

	return routerSvc, nil
}

// AddProvider 注册一个新的提供商 (例如一个新的 Gemini API Key 实例)
func (r *RouterService) AddProvider(p provider.LLMProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodes = append(r.nodes, &providerNode{
		provider:  p,
		isHealthy: true,
		failCount: 0,
	})
	log.Printf("Router 注册 Provider: %s", p.GetName())
}

// getNextNode 获取下一个健康的提供商节点 (Round-Robin 策略)
func (r *RouterService) getNextNode() *providerNode {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.nodes) == 0 {
		return nil
	}

	totalNodes := len(r.nodes)
	for i := 0; i < totalNodes; i++ {
		node := r.nodes[r.current]
		r.current = (r.current + 1) % totalNodes
		if node.isHealthy {
			return node
		}
	}

	// 如果所有节点都不健康，进行容错处理，返回当前节点，死马当活马医
	log.Printf("[Router 警告] 所有节点均处于不健康状态，尝试放行请求...")
	node := r.nodes[r.current]
	r.current = (r.current + 1) % totalNodes
	return node
}

// GenerateText 带有自动重试与降级机制的请求方法
func (r *RouterService) GenerateText(ctx context.Context, prompt string) (string, error) {
	if len(r.nodes) == 0 {
		return "", fmt.Errorf("无可用的模型提供商，请检查配置")
	}

	totalProviders := len(r.nodes)
	var lastErr error

	// 尝试所有的 Provider，直到有一个成功 (或者全部失败)
	for i := 0; i < totalProviders; i++ {
		node := r.getNextNode()
		if node == nil {
			break
		}
		p := node.provider

		log.Printf("--- Router 路由至: %s ---", p.GetName())
		result, err := p.GenerateText(ctx, prompt)
		
		if err == nil {
			// 请求成功，重置失败计数，直接返回
			r.mu.Lock()
			node.failCount = 0
			node.isHealthy = true
			r.mu.Unlock()
			return result, nil
		}

		// 如果失败，增加失败计数
		r.mu.Lock()
		node.failCount++
		// 如果连续失败达到阈值(例如 2 次)，标记为不健康
		if node.failCount >= 2 && node.isHealthy {
			node.isHealthy = false
			log.Printf("[Router 熔断隔离] %s 连续失败 %d 次，已被隔离", p.GetName(), node.failCount)
		}
		r.mu.Unlock()

		// 记录错误并继续下一个循环 (触发降级重试逻辑)
		log.Printf("[Router 警告] %s 请求失败: %v, 准备切换至备用模型...", p.GetName(), err)
		lastErr = err
	}

	// 如果循环结束依然没有成功，说明所有 Provider 都挂了
	return "", fmt.Errorf("所有模型提供商均不可用，最后一次错误: %w", lastErr)
}

// startHealthCheckTask 后台定期检查不健康的节点，尝试将其恢复
func (r *RouterService) startHealthCheckTask() {
	ticker := time.NewTicker(60 * time.Second) // 每 60 秒检查一次
	defer ticker.Stop()

	for range ticker.C {
		r.mu.Lock()
		// 找出需要探活的节点 (为了不在持有锁时做耗时的网络请求，先把它们挑出来)
		var unhealthyNodes []*providerNode
		for _, node := range r.nodes {
			if !node.isHealthy {
				unhealthyNodes = append(unhealthyNodes, node)
			}
		}
		r.mu.Unlock()

		if len(unhealthyNodes) == 0 {
			continue // 大家都很好，继续休息
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		for _, node := range unhealthyNodes {
			err := node.provider.CheckHealth(ctx)
			
			r.mu.Lock()
			if err == nil {
				// 探活成功，恢复节点
				node.isHealthy = true
				node.failCount = 0
				log.Printf("[Router 探活成功] %s 已恢复，重新加入路由池", node.provider.GetName())
			} else {
				log.Printf("[Router 探活失败] %s 仍未恢复: %v", node.provider.GetName(), err)
			}
			r.mu.Unlock()
		}
		cancel()
	}
}
