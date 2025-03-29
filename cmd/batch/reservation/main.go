package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"runtime/debug"

	"github.com/horsewin/echo-playground-batch-task/internal/common/config"
	"github.com/horsewin/echo-playground-batch-task/internal/common/utils"
	"github.com/horsewin/echo-playground-batch-task/internal/service/batch"
)

func main() {
	// コマンドライン引数のパース
	timeout := flag.Duration("timeout", 5*time.Minute, "バッチ処理のタイムアウト時間")
	flag.Parse()

	// 設定の読み込み
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v\nStack trace:\n%s", err, debug.Stack())
	}

	log.Printf("Env mode: %s", os.Getenv("ENV"))

	// サービスの初期化
	service, err := batch.NewReservationBatchService(cfg)
	if err != nil {
		log.Fatalf("Failed to create service: %v\nStack trace:\n%s", err, debug.Stack())
	}
	defer service.Close()

	// コンテキストの作成
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// シグナルハンドリングの設定
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// バッチ処理の実行
	errChan := make(chan error, 1)
	go func() {
		errChan <- utils.RunWithTimeout(ctx, *timeout, service.Run)
	}()

	// シグナルまたはエラーの待機
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
		cancel()
	case err := <-errChan:
		if err != nil {
			log.Printf("Batch process failed: %v\nStack trace:\n%s", err, debug.Stack())
			os.Exit(1)
		}
		log.Println("Batch process completed successfully")
	}
}
