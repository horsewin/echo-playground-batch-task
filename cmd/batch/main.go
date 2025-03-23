package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/horsewin/echo-playground-batch-task/internal/config"
	"github.com/horsewin/echo-playground-batch-task/internal/service"
)

func main() {
	// 設定の読み込み
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		os.Exit(1)
	}

	// コンテキストの作成（タイムアウト付き）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// シグナルハンドリング
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// エラーチャネル
	errChan := make(chan error, 1)

	// バッチサービスの初期化と実行
	batchService := service.NewBatchService(cfg)

	go func() {
		errChan <- batchService.Run(ctx)
	}()

	// シグナルまたは完了を待機
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
		cancel()
		os.Exit(1)
	case err := <-errChan:
		if err != nil {
			log.Printf("Batch process failed: %v", err)
			os.Exit(1)
		}
		log.Println("Batch process completed successfully")
		os.Exit(0)
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Batch process timed out")
			os.Exit(1)
		}
	}
}
