package main

import (
	"log"

	"github.com/cymoo/pebble/internal/app"
	"github.com/cymoo/pebble/internal/config"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 创建应用实例
	application := app.New(cfg)

	// 初始化应用
	if err := application.Initialize(); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// 运行应用
	if err := application.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
