package batch

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/horsewin/echo-playground-batch-task/internal/common/config"
	"github.com/horsewin/echo-playground-batch-task/internal/common/database"
	"github.com/horsewin/echo-playground-batch-task/internal/common/utils"
	"github.com/horsewin/echo-playground-batch-task/internal/model"
	"github.com/horsewin/echo-playground-batch-task/internal/repository"
)

type ReservationBatchService struct {
	db              *database.DB
	reservationRepo *repository.ReservationRepository
	sfnClient       *sfn.Client
	cfg             *config.Config
}

// NewReservationBatchService ... 予約バッチサービスを作成する
func NewReservationBatchService(cfg *config.Config) (*ReservationBatchService, error) {
	db, err := database.NewDB(cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	var sfnClient *sfn.Client
	// ローカル環境以外の場合のみAWS SDKの設定を行う
	if os.Getenv("ENV") != "LOCAL" {
		// AWS SDKの設定を読み込む
		awsCfg, err := awsconfig.LoadDefaultConfig(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}

		// Step Functionsクライアントを作成
		sfnClient = sfn.NewFromConfig(awsCfg)
	}

	return &ReservationBatchService{
		db:              db,
		reservationRepo: repository.NewReservationRepository(db.DB),
		sfnClient:       sfnClient,
		cfg:             cfg,
	}, nil
}

// Close ... 終了処理
func (s *ReservationBatchService) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Run ... 予約バッチ処理を実行する
func (s *ReservationBatchService) Run(ctx context.Context) error {
	log.Println("Starting reservation batch process...")

	// 処理開始時刻を記録
	startTime := time.Now()

	// バッチ処理を実行
	events, err := s.processReservationsByStatus("pending")
	if err != nil {
		return utils.GetStackWithError(fmt.Errorf("failed to process pending reservations: %w", err))
	}

	// イベントを発行
	if err := s.sendTaskSuccess(ctx, events); err != nil {
		return utils.GetStackWithError(fmt.Errorf("failed to send task success: %w", err))
	}

	// 処理終了時刻を記録し、実行時間を計算
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	log.Printf("Reservation batch process completed successfully. Duration: %v", duration)
	return nil
}

// processReservationsByStatus は、指定されたステータスの予約を処理します
func (s *ReservationBatchService) processReservationsByStatus(status string) ([]model.ReservationEvent, error) {
	// 指定されたステータスの予約を取得
	reservations, err := s.reservationRepo.GetReservationsByStatus(status)
	if err != nil {
		return nil, utils.GetStackWithError(fmt.Errorf("failed to get reservations with status %s: %w", status, err))
	}

	log.Printf("Found %d reservations with status %s", len(reservations), status)

	// 成功した予約のイベントを収集
	var events []model.ReservationEvent

	for _, reservation := range reservations {
		// トランザクション開始
		tx, err := s.reservationRepo.BeginTx()
		if err != nil {
			log.Printf("Failed to begin transaction for reservation %d: %v\nStack trace:\n%s",
				reservation.ReservationID, err, debug.Stack())
			continue
		}

		// 既存の予約をチェック
		exists, err := s.reservationRepo.CheckExistingReservation(reservation.PetID)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("Failed to rollback transaction for reservation %d: %v\nStack trace:\n%s",
					reservation.ReservationID, rollbackErr, debug.Stack())
			}
			log.Printf("Failed to check existing reservation for pet %s: %v\nStack trace:\n%s",
				reservation.PetID, err, debug.Stack())
			continue
		}

		if exists {
			// 既存の予約がある場合は、この予約をキャンセル
			if err := s.reservationRepo.UpdateStatus(tx, reservation.ReservationID, "cancelled"); err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					log.Printf("Failed to rollback transaction for reservation %d: %v\nStack trace:\n%s",
						reservation.ReservationID, rollbackErr, debug.Stack())
				}
				log.Printf("Failed to update reservation status to cancelled: %v\nStack trace:\n%s",
					err, debug.Stack())
				continue
			}
		} else {
			// 既存の予約がない場合は、予約を確定
			if err := s.reservationRepo.UpdateStatus(tx, reservation.ReservationID, "confirmed"); err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					log.Printf("Failed to rollback transaction for reservation %d: %v\nStack trace:\n%s",
						reservation.ReservationID, rollbackErr, debug.Stack())
				}
				log.Printf("Failed to update reservation status to confirmed: %v\nStack trace:\n%s",
					err, debug.Stack())
				continue
			}
		}

		// トランザクションをコミット
		if err := tx.Commit(); err != nil {
			log.Printf("Failed to commit transaction for reservation %d: %v\nStack trace:\n%s",
				reservation.ReservationID, err, debug.Stack())
			continue
		}

		// 成功した予約のイベントを収集
		events = append(events, model.ReservationEvent{
			UserID:    reservation.UserID,
			DateTime:  reservation.ReservationDateTime,
			PetID:     reservation.PetID,
			CreatedAt: reservation.CreatedAt,
		})
	}

	return events, nil
}

// sendTaskSuccess は、Step Functionsのタスク成功を通知し、イベントを返却します
func (s *ReservationBatchService) sendTaskSuccess(ctx context.Context, events []model.ReservationEvent) error {
	// ローカルの場合はStep Functionsの処理をスキップ
	if os.Getenv("ENV") == "LOCAL" {
		log.Printf("Local environment detected. Skipping Step Functions task success notification. Events: %+v", events)
		return nil
	}

	// イベントを通知形式に変換
	notifications := make([]model.Notification, len(events))
	for i, event := range events {
		notifications[i] = model.NewReservationNotification(event)
	}

	// 通知をJSONに変換
	output, err := json.Marshal(map[string]any{
		"notifications": notifications,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal notifications: %w", err)
	}

	// タスクトークンを設定から取得
	taskToken := s.cfg.SFN.TaskToken
	if taskToken == "" && os.Getenv("ENV") != "LOCAL" {
		return fmt.Errorf("SFN_TASK_TOKEN is not set in config")
	}

	// SendTaskSuccess APIを呼び出す
	input := &sfn.SendTaskSuccessInput{
		TaskToken: &taskToken,
		Output:    aws.String(string(output)),
	}

	_, err = s.sfnClient.SendTaskSuccess(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to send task success: %w", err)
	}

	log.Printf("Successfully sent task success with notifications: %s", string(output))
	return nil
}
