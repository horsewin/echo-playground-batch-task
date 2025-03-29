package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/horsewin/echo-playground-batch-task/internal/common/config"
	"github.com/horsewin/echo-playground-batch-task/internal/model"
	"github.com/horsewin/echo-playground-batch-task/internal/service/batch"
)

func main() {
	// コマンドライン引数のパース
	taskToken := flag.String("task-token", "", "Step Functions task token")
	flag.Parse()

	// タスクトークンの検証
	if *taskToken == "" {
		log.Fatal("Task token is required")
	}

	// 設定を読み込む
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// タスクトークンを設定に反映
	cfg.SFN.TaskToken = *taskToken

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
		// タスクトークンから通知データを生成
		notifications, err := generateNotificationsFromTaskToken(*taskToken)
		if err != nil {
			log.Printf("Failed to generate notifications: %v", err)
			cancel()
			return
		}

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

// generateNotificationsFromTaskToken はタスクトークンから通知データを生成します
func generateNotificationsFromTaskToken(taskToken string) ([]model.Notification, error) {
	// タスクトークンから通知データを取得する処理を実装
	// この例では、タスクトークンをJSONとして解析し、通知データを生成します
	var input struct {
		Events []struct {
			UserID              string    `json:"user_id"`
			ReservationDateTime time.Time `json:"reservation_date_time"`
			PetID               string    `json:"pet_id"`
			CreatedAt           time.Time `json:"created_at"`
		} `json:"events"`
	}

	if err := json.Unmarshal([]byte(taskToken), &input); err != nil {
		return nil, fmt.Errorf("failed to parse task token: %w", err)
	}

	notifications := make([]model.Notification, len(input.Events))
	for i, event := range input.Events {
		notifications[i] = model.NewReservationNotification(model.ReservationEvent{
			UserID:    event.UserID,
			DateTime:  event.ReservationDateTime,
			PetID:     event.PetID,
			CreatedAt: event.CreatedAt,
		})
	}

	return notifications, nil
}
