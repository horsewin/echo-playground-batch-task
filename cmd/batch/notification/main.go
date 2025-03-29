package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/horsewin/echo-playground-batch-task/internal/common/config"
	"github.com/horsewin/echo-playground-batch-task/internal/model"
	"github.com/horsewin/echo-playground-batch-task/internal/service/batch"
)

func main() {
	// 設定を読み込む
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 通知バッチサービスを作成
	notificationService, err := batch.NewNotificationBatchService(cfg)
	if err != nil {
		log.Fatalf("Failed to create notification batch service: %v", err)
	}
	defer notificationService.Close()

	// コンテキストを作成
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// シグナルハンドリング
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 通知バッチ処理を実行
	go func() {
		// TODO: 実際の通知データを取得する処理を実装
		notifications := []model.Notification{} // ここに実際の通知データを設定

		if err := notificationService.Run(ctx, notifications); err != nil {
			log.Printf("Failed to run notification batch: %v", err)
			cancel()
		}
	}()

	// シグナルを待機
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
		cancel()
	case <-ctx.Done():
		log.Println("Context cancelled")
	}
}
