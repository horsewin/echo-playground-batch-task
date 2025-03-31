package batch

import (
	"context"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/horsewin/echo-playground-batch-task/internal/common/config"
	"github.com/horsewin/echo-playground-batch-task/internal/common/database"
	"github.com/horsewin/echo-playground-batch-task/internal/model"
	"github.com/horsewin/echo-playground-batch-task/internal/repository"
)

// NotificationBatchService は通知バッチ処理を担当します
type NotificationBatchService struct {
	args             []model.Notification
	db               *database.DB
	notificationRepo repository.NotificationRepository
	petRepo          repository.PetRepository
	cfg              *config.Config
}

// NewNotificationBatchService は新しいNotificationBatchServiceを作成します
func NewNotificationBatchService(cfg *config.Config) (*NotificationBatchService, error) {
	db, err := database.NewDB(cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	// database.DBをrepository.DBに変換
	repoDb := &repository.DB{DB: db.DB}

	return &NotificationBatchService{
		db:               db,
		notificationRepo: repository.NewNotificationRepository(repoDb),
		petRepo:          repository.NewPetRepository(repoDb),
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

// setArgs は通知バッチ処理の引数を設定します
func (s *NotificationBatchService) SetArgs(args []model.Notification) {
	s.args = args
}

// Run は通知バッチ処理を実行します
func (s *NotificationBatchService) Run(ctx context.Context) error {
	notifications := s.args
	log.Printf("Starting notification batch process for %d notifications...", len(notifications))

	// 処理開始時刻を記録
	startTime := time.Now()

	// ペット名を取得
	petNameMap, err := s.getPetNameMap(ctx, notifications)
	if err != nil {
		return err
	}

	// 通知をレコードに変換
	records := make([]model.NotificationRecord, len(notifications))
	for i, notification := range notifications {
		record, err := notification.ToNotificationRecord(petNameMap)
		if err != nil {
			return err
		}
		records[i] = *record
	}

	// 通知レコードを作成
	if err := s.notificationRepo.CreateNotifications(ctx, records); err != nil {
		return fmt.Errorf("failed to create notifications: %w", err)
	}

	// 処理終了時刻を記録し、実行時間を計算
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	log.Printf("Notification batch process completed successfully. Duration: %v", duration)
	return nil
}

// 通知データに含まれる情報からペット名を取得する
// N+1とならないように先に重複がないペットIDを取得をしておく
// 1. 重複がないペットIDを取得
// 2. ペットIDからペット名を取得してMapとして保持する
func (s *NotificationBatchService) getPetNameMap(ctx context.Context, notifications []model.Notification) (map[string]string, error) {
	petIDs := make([]string, 0)
	petNameMap := make(map[string]string)
	for _, notification := range notifications {
		// Dataフィールドの型をチェック
		data, ok := notification.Data.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid notification data format")
		}

		petID, ok := data["pet_id"].(string)
		if !ok {
			return nil, fmt.Errorf("pet_id is not a string")
		}

		// petIDが重複している場合はスキップ
		if slices.Contains(petIDs, petID) {
			continue
		}

		petIDs = append(petIDs, petID)
	}

	// ペット名を取得
	for _, petID := range petIDs {
		petName, err := s.petRepo.GetNameByID(ctx, petID)
		if err != nil {
			return nil, err
		}
		petNameMap[petID] = petName
	}

	return petNameMap, nil
}
