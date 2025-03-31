package batch

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/horsewin/echo-playground-batch-task/internal/common/config"
	"github.com/horsewin/echo-playground-batch-task/internal/common/database"
	"github.com/horsewin/echo-playground-batch-task/internal/model"
	"github.com/horsewin/echo-playground-batch-task/internal/repository"
)

// ReservationBatchService は予約バッチ処理を担当します
type ReservationBatchService struct {
	args            []model.Reservation
	db              *database.DB
	reservationRepo repository.ReservationRepository
	cfg             *config.Config
}

// NewReservationBatchService は新しいReservationBatchServiceを作成します
func NewReservationBatchService(cfg *config.Config) (*ReservationBatchService, error) {
	db, err := database.NewDB(cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	// database.DBをrepository.DBに変換
	repoDb := &repository.DB{DB: db.DB}

	return &ReservationBatchService{
		db:              db,
		reservationRepo: repository.NewReservationRepository(repoDb),
		cfg:             cfg,
	}, nil
}

// Close は終了処理を行います
func (s *ReservationBatchService) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// SetArgs は予約バッチ処理の引数を設定します
func (s *ReservationBatchService) SetArgs(args []model.Reservation) {
	s.args = args
}

// Run は予約バッチ処理を実行します
func (s *ReservationBatchService) Run(ctx context.Context) error {
	// X-Rayセグメントの作成
	ctx, seg := xray.BeginSubsegment(ctx, "ReservationBatchService.Run")
	defer seg.Close(nil)

	reservations := s.args
	log.Printf("Starting reservation batch process for %d reservations...", len(reservations))

	// セグメントにメタデータを追加
	if err := seg.AddMetadata("reservation_count", len(reservations)); err != nil {
		log.Printf("Failed to add reservation_count metadata: %v", err)
	}

	// 処理開始時刻を記録
	startTime := time.Now()

	// 予約レコードを作成
	if err := s.reservationRepo.CreateReservations(ctx, reservations); err != nil {
		seg.Close(err)
		return fmt.Errorf("failed to create reservations: %w", err)
	}

	// 処理終了時刻を記録し、実行時間を計算
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	// セグメントにメタデータを追加
	if err := seg.AddMetadata("duration", duration.String()); err != nil {
		log.Printf("Failed to add duration metadata: %v", err)
	}

	log.Printf("Reservation batch process completed successfully. Duration: %v", duration)
	return nil
}

// processReservationsByStatus は、指定されたステータスの予約を処理します
// 現在は未使用ですが、将来的に使用される可能性があるため、コメントアウトして保持します
/*
func (s *ReservationBatchService) processReservationsByStatus(ctx context.Context, status string) ([]model.ReservationEvent, error) {
	// 指定されたステータスの予約を取得
	reservations, err := s.reservationRepo.GetReservationsByStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get reservations with status %s: %w", status, err)
	}

	log.Printf("Found %d reservations with status %s", len(reservations), status)

	// 成功した予約のイベントを収集
	var events []model.ReservationEvent

	for _, reservation := range reservations {
		// トランザクション開始
		tx, err := s.reservationRepo.BeginTx()
		if err != nil {
			log.Printf("Failed to begin transaction for reservation %d: %v",
				reservation.ReservationID, err)
			continue
		}

		// 既存の予約をチェック
		exists, err := s.reservationRepo.CheckExistingReservation(ctx, reservation.PetID)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("Failed to rollback transaction for reservation %d: %v",
					reservation.ReservationID, rollbackErr)
			}
			log.Printf("Failed to check existing reservation for pet %s: %v",
				reservation.PetID, err)
			continue
		}

		if exists {
			// 既存の予約がある場合は、この予約をキャンセル
			if err := s.reservationRepo.UpdateStatus(ctx, tx, reservation.ReservationID, "cancelled"); err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					log.Printf("Failed to rollback transaction for reservation %d: %v",
						reservation.ReservationID, rollbackErr)
				}
				log.Printf("Failed to update reservation status to cancelled: %v", err)
				continue
			}
		} else {
			// 既存の予約がない場合は、予約を確定
			if err := s.reservationRepo.UpdateStatus(ctx, tx, reservation.ReservationID, "confirmed"); err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					log.Printf("Failed to rollback transaction for reservation %d: %v",
						reservation.ReservationID, rollbackErr)
				}
				log.Printf("Failed to update reservation status to confirmed: %v", err)
				continue
			}
		}

		// トランザクションをコミット
		if err := tx.Commit(); err != nil {
			log.Printf("Failed to commit transaction for reservation %d: %v",
				reservation.ReservationID, err)
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
*/
