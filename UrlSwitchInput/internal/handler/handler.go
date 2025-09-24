package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"UrlSwitchInput/internal/config"
	"UrlSwitchInput/internal/ime"
	"UrlSwitchInput/internal/matcher"
	"UrlSwitchInput/internal/notification"
)

// URLHandler URL处理器
type URLHandler struct {
	config        *config.Config
	imeController *ime.Controller
	notifier      *notification.Notifier
	matcher       *matcher.URLMatcher
}

// URLRequest 请求结构
type URLRequest struct {
	URL string `json:"url"`
}

// URLResponse 响应结构
type URLResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	Matched     bool   `json:"matched"`
	RuleName    string `json:"rule_name,omitempty"`
	IMEStatus   string `json:"ime_status,omitempty"`
	IMESwitched bool   `json:"ime_switched,omitempty"`
}

// NewURLHandler 创建新的URL处理器
func NewURLHandler(cfg *config.Config, imeController *ime.Controller, notifier *notification.Notifier) *URLHandler {
	return &URLHandler{
		config:        cfg,
		imeController: imeController,
		notifier:      notifier,
		matcher:       matcher.NewURLMatcher(cfg.GetEnabledRules()),
	}
}

// HandleURL 处理URL请求
func (h *URLHandler) HandleURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendErrorResponse(w, http.StatusMethodNotAllowed, "只支持POST方法")
		return
	}

	var req URLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "无效的JSON格式")
		return
	}

	if req.URL == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "URL不能为空")
		return
	}

	log.Printf("收到URL请求: %s", req.URL)

	// 匹配URL规则
	result := h.matcher.Match(req.URL)

	response := URLResponse{
		Success: true,
		Message: "处理完成",
		Matched: result.Matched,
	}

	if result.Matched {
		response.RuleName = result.Rule.Name
		log.Printf("URL匹配规则: %s", result.Rule.Name)

		// 获取当前输入法状态
		currentStatus, err := h.imeController.GetCurrentStatus()
		if err != nil {
			log.Printf("获取输入法状态失败: %v", err)
			response.IMEStatus = "获取失败"
		} else {
			response.IMEStatus = h.getStatusString(currentStatus)

			// 如果当前是中文输入法，切换到英文
			if currentStatus == ime.Chinese {
				if err := h.imeController.SwitchToEnglish(); err != nil {
					log.Printf("切换输入法失败: %v", err)
					response.Message = "规则匹配成功，但切换输入法失败"
				} else {
					response.IMESwitched = true
					log.Printf("已切换输入法: 中文 → 英文")

					// 发送输入法切换通知
					if err := h.notifier.SendIMESwitchNotification("中文", "英文"); err != nil {
						log.Printf("发送输入法切换通知失败: %v", err)
					}
				}
			}
		}

		// 发送匹配通知
		action := ""
		if response.IMESwitched {
			action = "已切换到英文输入法"
		} else if currentStatus == ime.English {
			action = "输入法已是英文状态"
		}

		if err := h.notifier.SendURLMatchNotification(result.Rule.Name, req.URL, action); err != nil {
			log.Printf("发送匹配通知失败: %v", err)
		}
	} else {
		log.Printf("URL未匹配任何规则")
		response.Message = "未匹配任何规则"
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

// sendErrorResponse 发送错误响应
func (h *URLHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := URLResponse{
		Success: false,
		Message: message,
	}
	h.sendJSONResponse(w, statusCode, response)
}

// sendJSONResponse 发送JSON响应
func (h *URLHandler) sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("编码JSON响应失败: %v", err)
	}
}

// getStatusString 获取状态字符串
func (h *URLHandler) getStatusString(status ime.InputMethod) string {
	switch status {
	case ime.Chinese:
		return "中文"
	case ime.English:
		return "英文"
	default:
		return "未知"
	}
}
