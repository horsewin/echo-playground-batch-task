package batch

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/horsewin/echo-playground-batch-task/internal/common/config"
	"github.com/horsewin/echo-playground-batch-task/internal/common/database"
	"github.com/horsewin/echo-playground-batch-task/internal/model"
	"github.com/horsewin/echo-playground-batch-task/internal/repository"
)

// NotificationRepository は通知の永続化を担当するインターフェースです
type NotificationRepository interface {
	CreateNotifications(records []model.NotificationRecord) error
	Create(tx *sql.Tx, record *model.NotificationRecord) error
}

// NotificationBatchService は通知バッチ処理を担当します
type NotificationBatchService struct {
	db               *database.DB
	notificationRepo NotificationRepository
	cfg              *config.Config
}

// NewNotificationBatchService は新しいNotificationBatchServiceを作成します
func NewNotificationBatchService(cfg *config.Config) (*NotificationBatchService, error) {
	db, err := database.NewDB(cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	return &NotificationBatchService{
		db:               db,
		notificationRepo: repository.NewNotificationRepository(db.DB),
		cfg:              cfg,
	}, nil
}

// Close は終了処理を行います
func (s *NotificationBatchService) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Run は通知バッチ処理を実行します
func (s *NotificationBatchService) Run(ctx context.Context, notifications []model.Notification) error {
	log.Printf("Starting notification batch process for %d notifications...", len(notifications))

	// 処理開始時刻を記録
	startTime := time.Now()

	// 通知をレコードに変換
	records := make([]model.NotificationRecord, len(notifications))
	for i, notification := range notifications {
		// Dataフィールドの型をチェック
		if _, ok := notification.Data.(map[string]interface{}); !ok {
			return fmt.Errorf("invalid notification data format")
		}

		records[i] = model.NotificationRecord{
			UserID:    notification.UserID,
			Title:     "通知",
			Message:   "新しい通知が届きました",
			IsRead:    false,
			Type:      notification.Type,
			CreatedAt: notification.CreatedAt,
			UpdatedAt: notification.CreatedAt,
		}
	}

	// 通知レコードを作成
	if err := s.notificationRepo.CreateNotifications(records); err != nil {
		return fmt.Errorf("failed to create notifications: %w", err)
	}

	// 処理終了時刻を記録し、実行時間を計算
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	log.Printf("Notification batch process completed successfully. Duration: %v", duration)
	return nil
}
