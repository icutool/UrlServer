//go:build windows
// +build windows

package main

import (
	"log"
	"net/http"

	"UrlSwitchInput/internal/config"
	"UrlSwitchInput/internal/handler"
	"UrlSwitchInput/internal/ime"
	"UrlSwitchInput/internal/notification"
)

func main() {
	// 加载配置文件
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatal("加载配置文件失败:", err)
	}

	// 初始化输入法控制器
	imeController := ime.NewController()

	// 初始化通知服务
	notifier := notification.NewNotifier()

	// 创建处理器
	urlHandler := handler.NewURLHandler(cfg, imeController, notifier)

	// 设置路由
	http.HandleFunc("/url", urlHandler.HandleURL)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Println("服务启动在端口 :7887")
	if err := http.ListenAndServe(":7887", nil); err != nil {
		log.Fatal("启动服务器失败:", err)
	}
}
