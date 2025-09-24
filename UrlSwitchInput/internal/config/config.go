package config

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Config 配置结构
type Config struct {
	Rules []Rule `json:"rules"`
}

// MatchType 匹配类型
type MatchType string

const (
	MatchTypeRegex    MatchType = "regex"    // 正则表达式匹配
	MatchTypeExact    MatchType = "exact"    // 精准匹配
	MatchTypeKeyword  MatchType = "keyword"  // 关键词匹配
	MatchTypeWildcard MatchType = "wildcard" // 通配符匹配（兼容旧版本）
)

// Rule 匹配规则
type Rule struct {
	Name        string    `json:"name"`
	URLPattern  string    `json:"url_pattern"` // URL匹配模式
	MatchType   MatchType `json:"match_type"`  // 匹配类型：regex, exact, keyword, wildcard
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
}

// LoadConfig 加载配置文件
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// GetEnabledRules 获取启用的规则
func (c *Config) GetEnabledRules() []Rule {
	var enabledRules []Rule
	for _, rule := range c.Rules {
		if rule.Enabled {
			// 为旧配置设置默认匹配类型
			if rule.MatchType == "" {
				// 根据模式自动判断匹配类型
				if c.containsRegexChars(rule.URLPattern) {
					rule.MatchType = MatchTypeRegex
				} else if strings.Contains(rule.URLPattern, "*") {
					rule.MatchType = MatchTypeWildcard
				} else {
					rule.MatchType = MatchTypeKeyword
				}
			}
			enabledRules = append(enabledRules, rule)
		}
	}
	return enabledRules
}

// ValidateRules 验证规则配置
func (c *Config) ValidateRules() []error {
	var errors []error

	for i, rule := range c.Rules {
		// 检查必填字段
		if rule.Name == "" {
			errors = append(errors, fmt.Errorf("规则 %d: 名称不能为空", i+1))
		}
		if rule.URLPattern == "" {
			errors = append(errors, fmt.Errorf("规则 %d: URL模式不能为空", i+1))
		}

		// 验证匹配类型
		if rule.MatchType != "" {
			switch rule.MatchType {
			case MatchTypeRegex, MatchTypeExact, MatchTypeKeyword, MatchTypeWildcard:
				// 有效的匹配类型
			default:
				errors = append(errors, fmt.Errorf("规则 %d: 无效的匹配类型 '%s'", i+1, rule.MatchType))
			}
		}

		// 验证正则表达式
		if rule.MatchType == MatchTypeRegex {
			if _, err := regexp.Compile(rule.URLPattern); err != nil {
				errors = append(errors, fmt.Errorf("规则 %d: 无效的正则表达式 '%s': %v", i+1, rule.URLPattern, err))
			}
		}
	}

	return errors
}

// containsRegexChars 检查是否包含正则表达式特殊字符
func (c *Config) containsRegexChars(pattern string) bool {
	regexChars := []string{"(", ")", "[", "]", "{", "}", "^", "$", "+", "?", "|", "\\"}
	for _, char := range regexChars {
		if strings.Contains(pattern, char) {
			return true
		}
	}
	return false
}
