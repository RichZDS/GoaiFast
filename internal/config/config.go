package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GeminiAPIKeys []string `yaml:"GEMINI_API_KEYS"` // 提供多个api key，用于负载均衡
	ProxyURL      string   `yaml:"PROXY_URL"`       // 本地代理配置
}

var GlobalConfig *Config

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	if len(config.GeminiAPIKeys) == 0 {
		return nil, fmt.Errorf("GEMINI_API_KEYS not found or empty in config file")
	}

	GlobalConfig = &config
	return &config, nil
}
