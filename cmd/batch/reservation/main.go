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

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/horsewin/echo-playground-batch-task/internal/common/config"
	"github.com/horsewin/echo-playground-batch-task/internal/common/utils"
	"github.com/horsewin/echo-playground-batch-task/internal/service/batch"
)

func main() {
	// コマンドライン引数のパース
	timeout := flag.Duration("timeout", 5*time.Minute, "バッチ処理のタイムアウト時間")
	flag.Parse()

	// 最後の引数として渡されたタスクトークンを取得
	// ENV=LOCALの場合はタスクトークンを取得しない
	taskToken := "DUMMY_TASK_TOKEN"
	if os.Getenv("ENV") != "LOCAL" {
		taskToken = flag.Arg(len(flag.Args()) - 1)
		if taskToken == "" {
			log.Fatalf("Task token is required")
		}
	}

	// 設定の読み込み
	cfg, err := config.LoadConfig(taskToken)
	if err != nil {
		log.Fatalf("Failed to load config: %v\nStack trace:\n%s", err, debug.Stack())
	}

	// Step Functionsクライアントの初期化
	var sfnClient *sfn.Client
	if os.Getenv("ENV") != "LOCAL" {
		awsCfg, err := awsconfig.LoadDefaultConfig(context.Background())
		if err != nil {
			log.Fatalf("Failed to load AWS config: %v\nStack trace:\n%s", err, debug.Stack())
		}
		sfnClient = sfn.NewFromConfig(awsCfg)
	}

	// サービスの初期化
	service, err := batch.NewReservationBatchService(cfg, sfnClient)
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

			// ローカル環境以外の場合のみStep Functionsのエラー通知を行う
			if os.Getenv("ENV") != "LOCAL" && sfnClient != nil {
				input := &sfn.SendTaskFailureInput{
					TaskToken: aws.String(taskToken),
					Error:     aws.String("Batch process failed"),
				}

				_, err := sfnClient.SendTaskFailure(ctx, input)
				if err != nil {
					log.Printf("Failed to send task failure: %v\nStack trace:\n%s", err, debug.Stack())
				}
			}

			os.Exit(1)
		}
		log.Println("Batch process completed successfully")
	}
}
