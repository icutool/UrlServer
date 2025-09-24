//go:build windows
// +build windows

package notification

import (
	"github.com/go-toast/toast"
)

// Notifier 通知服务
type Notifier struct{}

// NewNotifier 创建新的通知服务
func NewNotifier() *Notifier {
	return &Notifier{}
}

// SendNotification 发送通知
func (n *Notifier) SendNotification(title, message string) error {
	notification := toast.Notification{
		AppID:   "UrlSwitchInput",
		Title:   title,
		Message: message,
		Icon:    "", // 可以设置图标路径
		Actions: []toast.Action{
			{Type: "protocol", Label: "确定", Arguments: ""},
		},
	}

	return notification.Push()
}

// SendURLMatchNotification 发送URL匹配通知
func (n *Notifier) SendURLMatchNotification(ruleName, url, action string) error {
	title := "URL规则匹配"
	message := ""

	if action != "" {
		message = "规则: " + ruleName + "\nURL: " + url + "\n操作: " + action
	} else {
		message = "规则: " + ruleName + "\nURL: " + url
	}

	return n.SendNotification(title, message)
}

// SendIMESwitchNotification 发送输入法切换通知
func (n *Notifier) SendIMESwitchNotification(from, to string) error {
	title := "输入法切换"
	message := "从 " + from + " 切换到 " + to

	return n.SendNotification(title, message)
}
