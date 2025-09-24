//go:build windows
// +build windows

package ime

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// Windows API 常量
const (
	WM_IME_CONTROL    = 0x283
	IMC_GETOPENSTATUS = 0x0005 // 获取输入法状态
	IMC_SETOPENSTATUS = 0x0006 // 设置输入法状态
)

// Controller 输入法控制器
type Controller struct {
	user32                  *windows.LazyDLL
	imm32                   *windows.LazyDLL
	procGetForegroundWindow *windows.LazyProc
	procSendMessageW        *windows.LazyProc
	procImmGetDefaultIMEWnd *windows.LazyProc
}

// InputMethod 输入法状态
type InputMethod int

const (
	English InputMethod = 0 // 英文输入法
	Chinese InputMethod = 1 // 中文输入法
)

// NewController 创建新的输入法控制器
func NewController() *Controller {
	user32 := windows.NewLazySystemDLL("user32.dll")
	imm32 := windows.NewLazySystemDLL("imm32.dll")

	return &Controller{
		user32:                  user32,
		imm32:                   imm32,
		procGetForegroundWindow: user32.NewProc("GetForegroundWindow"),
		procSendMessageW:        user32.NewProc("SendMessageW"),
		procImmGetDefaultIMEWnd: imm32.NewProc("ImmGetDefaultIMEWnd"),
	}
}

// getForegroundWindow 获取前台窗口
func (c *Controller) getForegroundWindow() uintptr {
	hwnd, _, _ := c.procGetForegroundWindow.Call()
	return hwnd
}

// immGetDefaultIMEWnd 获取默认IME窗口
func (c *Controller) immGetDefaultIMEWnd(hwnd uintptr) uintptr {
	hIME, _, _ := c.procImmGetDefaultIMEWnd.Call(hwnd)
	return hIME
}

// sendMessage 发送Windows消息
func (c *Controller) sendMessage(hwnd uintptr, msg uint32, wparam, lparam uintptr) uintptr {
	ret, _, _ := c.procSendMessageW.Call(hwnd, uintptr(msg), wparam, lparam)
	return ret
}

// GetCurrentStatus 获取当前输入法状态
func (c *Controller) GetCurrentStatus() (InputMethod, error) {
	// 获取前台窗口
	hwnd := c.getForegroundWindow()
	if hwnd == 0 {
		return English, fmt.Errorf("无法获取前台窗口")
	}

	// 获取对应的 IME 窗口
	hIME := c.immGetDefaultIMEWnd(hwnd)
	if hIME == 0 {
		return English, fmt.Errorf("无法获取 IME 窗口")
	}

	// 检测当前状态
	status := c.sendMessage(hIME, WM_IME_CONTROL, IMC_GETOPENSTATUS, 0)

	if status == 1 {
		return Chinese, nil
	}
	return English, nil
}

// SetInputMethod 设置输入法状态
func (c *Controller) SetInputMethod(method InputMethod) error {
	// 获取前台窗口
	hwnd := c.getForegroundWindow()
	if hwnd == 0 {
		return fmt.Errorf("无法获取前台窗口")
	}

	// 获取对应的 IME 窗口
	hIME := c.immGetDefaultIMEWnd(hwnd)
	if hIME == 0 {
		return fmt.Errorf("无法获取 IME 窗口")
	}

	// 设置输入法状态
	c.sendMessage(hIME, WM_IME_CONTROL, IMC_SETOPENSTATUS, uintptr(method))
	return nil
}

// SwitchToEnglish 切换到英文输入法
func (c *Controller) SwitchToEnglish() error {
	current, err := c.GetCurrentStatus()
	if err != nil {
		return err
	}

	if current == Chinese {
		return c.SetInputMethod(English)
	}

	return nil // 已经是英文状态，无需切换
}

// SwitchToChinese 切换到中文输入法
func (c *Controller) SwitchToChinese() error {
	current, err := c.GetCurrentStatus()
	if err != nil {
		return err
	}

	if current == English {
		return c.SetInputMethod(Chinese)
	}

	return nil // 已经是中文状态，无需切换
}

// GetStatusString 获取状态字符串描述
func (c *Controller) GetStatusString() string {
	status, err := c.GetCurrentStatus()
	if err != nil {
		return "未知"
	}

	switch status {
	case Chinese:
		return "中文"
	case English:
		return "英文"
	default:
		return "未知"
	}
}
