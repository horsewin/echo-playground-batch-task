package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/horsewin/echo-playground-batch-task/internal/config"
	"github.com/horsewin/echo-playground-batch-task/internal/service"
)

func main() {
	// 設定の読み込み
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// コンテキストの作成
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// シグナルハンドリング
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// バッチサービスの初期化
	batchService := service.NewBatchService(cfg)

	// バッチ処理の実行
	go func() {
		if err := batchService.Run(ctx); err != nil {
			log.Printf("Batch process error: %v", err)
			cancel()
		}
	}()

	// シグナル待ち受け
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
		cancel()
	case <-ctx.Done():
		log.Println("Context cancelled")
	}
}
