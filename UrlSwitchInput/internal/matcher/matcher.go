package matcher

import (
	"fmt"
	"regexp"
	"strings"

	"UrlSwitchInput/internal/config"
)

// URLMatcher URL匹配器
type URLMatcher struct {
	rules      []config.Rule
	regexCache map[string]*regexp.Regexp // 缓存编译好的正则表达式
}

// NewURLMatcher 创建新的URL匹配器
func NewURLMatcher(rules []config.Rule) *URLMatcher {
	matcher := &URLMatcher{
		rules:      rules,
		regexCache: make(map[string]*regexp.Regexp),
	}

	// 预编译正则表达式
	matcher.precompileRegex()

	return matcher
}

// MatchResult 匹配结果
type MatchResult struct {
	Matched   bool
	Rule      config.Rule
	MatchType config.MatchType
	Details   string // 匹配详情
}

// Match 匹配URL
func (m *URLMatcher) Match(url string) *MatchResult {
	for _, rule := range m.rules {
		if !rule.Enabled {
			continue
		}

		matched, details := m.matchByType(url, rule)
		if matched {
			return &MatchResult{
				Matched:   true,
				Rule:      rule,
				MatchType: rule.MatchType,
				Details:   details,
			}
		}
	}

	return &MatchResult{
		Matched: false,
		Details: "未匹配任何规则",
	}
}

// matchByType 根据匹配类型进行匹配
func (m *URLMatcher) matchByType(url string, rule config.Rule) (bool, string) {
	switch rule.MatchType {
	case config.MatchTypeRegex:
		return m.regexMatch(url, rule)
	case config.MatchTypeExact:
		return m.exactMatch(url, rule)
	case config.MatchTypeKeyword:
		return m.keywordMatch(url, rule)
	case config.MatchTypeWildcard:
		return m.wildcardMatch(url, rule)
	default:
		// 兼容旧版本，自动判断匹配类型
		return m.autoMatch(url, rule)
	}
}

// regexMatch 正则表达式匹配
func (m *URLMatcher) regexMatch(url string, rule config.Rule) (bool, string) {
	regex, exists := m.regexCache[rule.URLPattern]
	if !exists {
		var err error
		regex, err = regexp.Compile(rule.URLPattern)
		if err != nil {
			return false, fmt.Sprintf("正则表达式编译失败: %v", err)
		}
		m.regexCache[rule.URLPattern] = regex
	}

	matches := regex.FindStringSubmatch(url)
	if len(matches) > 0 {
		if len(matches) > 1 {
			// 有捕获组
			return true, fmt.Sprintf("正则匹配成功，捕获组: %v", matches[1:])
		}
		return true, "正则匹配成功"
	}

	return false, "正则匹配失败"
}

// exactMatch 精准匹配
func (m *URLMatcher) exactMatch(url string, rule config.Rule) (bool, string) {
	if url == rule.URLPattern {
		return true, "精准匹配成功"
	}
	return false, "精准匹配失败"
}

// keywordMatch 关键词匹配
func (m *URLMatcher) keywordMatch(url string, rule config.Rule) (bool, string) {
	// 支持多个关键词，用逗号分隔
	keywords := m.parseKeywords(rule.URLPattern)

	matchedKeywords := []string{}
	for _, keyword := range keywords {
		keyword = strings.TrimSpace(keyword)
		if keyword == "" {
			continue
		}

		// 支持大小写不敏感匹配
		if strings.Contains(strings.ToLower(url), strings.ToLower(keyword)) {
			matchedKeywords = append(matchedKeywords, keyword)
		}
	}

	if len(matchedKeywords) > 0 {
		if len(matchedKeywords) == len(keywords) {
			return true, fmt.Sprintf("所有关键词匹配成功: %v", matchedKeywords)
		} else {
			// 部分匹配，根据配置决定是否算作匹配成功
			return true, fmt.Sprintf("部分关键词匹配成功: %v", matchedKeywords)
		}
	}

	return false, "关键词匹配失败"
}

// wildcardMatch 通配符匹配
func (m *URLMatcher) wildcardMatch(url string, rule config.Rule) (bool, string) {
	// 将通配符模式转换为正则表达式
	regexPattern := m.wildcardToRegex(rule.URLPattern)

	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return false, fmt.Sprintf("通配符转换失败: %v", err)
	}

	if regex.MatchString(url) {
		return true, "通配符匹配成功"
	}

	return false, "通配符匹配失败"
}

// autoMatch 自动判断匹配类型（兼容旧版本）
func (m *URLMatcher) autoMatch(url string, rule config.Rule) (bool, string) {
	pattern := rule.URLPattern

	// 检查是否为正则表达式
	if m.isRegexPattern(pattern) {
		return m.regexMatch(url, config.Rule{
			URLPattern: pattern,
			MatchType:  config.MatchTypeRegex,
		})
	}

	// 检查是否包含通配符
	if strings.Contains(pattern, "*") {
		return m.wildcardMatch(url, config.Rule{
			URLPattern: pattern,
			MatchType:  config.MatchTypeWildcard,
		})
	}

	// 默认使用关键词匹配
	return m.keywordMatch(url, config.Rule{
		URLPattern: pattern,
		MatchType:  config.MatchTypeKeyword,
	})
}

// precompileRegex 预编译正则表达式
func (m *URLMatcher) precompileRegex() {
	for _, rule := range m.rules {
		if rule.MatchType == config.MatchTypeRegex {
			regex, err := regexp.Compile(rule.URLPattern)
			if err == nil {
				m.regexCache[rule.URLPattern] = regex
			}
		}
	}
}

// parseKeywords 解析关键词（支持逗号分隔）
func (m *URLMatcher) parseKeywords(pattern string) []string {
	// 支持逗号、分号、管道符分隔
	separators := []string{",", ";", "|"}

	keywords := []string{pattern}
	for _, sep := range separators {
		if strings.Contains(pattern, sep) {
			keywords = strings.Split(pattern, sep)
			break
		}
	}

	// 清理空白字符
	var cleanKeywords []string
	for _, keyword := range keywords {
		if cleaned := strings.TrimSpace(keyword); cleaned != "" {
			cleanKeywords = append(cleanKeywords, cleaned)
		}
	}

	return cleanKeywords
}

// wildcardToRegex 将通配符模式转换为正则表达式
func (m *URLMatcher) wildcardToRegex(pattern string) string {
	// 转义正则表达式特殊字符，但保留 *
	escaped := regexp.QuoteMeta(pattern)
	// 将转义后的 \* 替换为 .*
	regexPattern := strings.ReplaceAll(escaped, "\\*", ".*")
	return "^" + regexPattern + "$"
}

// isRegexPattern 判断是否为正则表达式模式
func (m *URLMatcher) isRegexPattern(pattern string) bool {
	// 检查是否包含正则表达式特殊字符（排除通配符*）
	regexChars := []string{"(", ")", "[", "]", "{", "}", "^", "$", "+", "?", "|", "\\"}
	for _, char := range regexChars {
		if strings.Contains(pattern, char) {
			return true
		}
	}
	return false
}

// GetMatchTypeDescription 获取匹配类型描述
func GetMatchTypeDescription(matchType config.MatchType) string {
	switch matchType {
	case config.MatchTypeRegex:
		return "正则表达式匹配：支持复杂的正则表达式模式"
	case config.MatchTypeExact:
		return "精准匹配：URL必须完全相同"
	case config.MatchTypeKeyword:
		return "关键词匹配：URL包含指定关键词即可匹配（支持多个关键词用逗号分隔）"
	case config.MatchTypeWildcard:
		return "通配符匹配：支持*通配符"
	default:
		return "未知匹配类型"
	}
}
