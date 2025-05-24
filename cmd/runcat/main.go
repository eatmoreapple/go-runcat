package main

import (
	"embed"
	"log"
	"os"
	"os/signal"
	"syscall"

	. "github.com/eatmoreapple/go-runcat/internal/app"
)

//go:embed assets
var assets embed.FS

func main() {
	// 创建应用程序实例
	app, err := NewApp(assets)
	if err != nil {
		log.Println("Failed to create application:", err)
		os.Exit(1)
	}

	// 运行应用程序
	if err = app.Run(); err != nil {
		log.Println("Failed to run application:", err)
		os.Exit(1)
	}

	// 等待信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
