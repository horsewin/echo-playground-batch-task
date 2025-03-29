package batch

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/horsewin/echo-playground-batch-task/internal/common/config"
	"github.com/horsewin/echo-playground-batch-task/internal/common/database"
	"github.com/horsewin/echo-playground-batch-task/internal/common/utils"
	"github.com/horsewin/echo-playground-batch-task/internal/repository"
)

type ReservationBatchService struct {
	db              *database.DB
	reservationRepo *repository.ReservationRepository
}

// NewReservationBatchService ... 予約バッチサービスを作成する
func NewReservationBatchService(cfg *config.Config) (*ReservationBatchService, error) {
	db, err := database.NewDB(cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	return &ReservationBatchService{
		db:              db,
		reservationRepo: repository.NewReservationRepository(db.DB),
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
	if err := s.processReservationsByStatus("pending"); err != nil {
		return utils.GetStackWithError(fmt.Errorf("failed to process pending reservations: %w", err))
	}

	// 処理終了時刻を記録し、実行時間を計算
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	log.Printf("Reservation batch process completed successfully. Duration: %v", duration)
	return nil
}

// processReservationsByStatus は、指定されたステータスの予約を処理します
func (s *ReservationBatchService) processReservationsByStatus(status string) error {
	// 指定されたステータスの予約を取得
	reservations, err := s.reservationRepo.GetReservationsByStatus(status)
	if err != nil {
		return utils.GetStackWithError(fmt.Errorf("failed to get reservations with status %s: %w", status, err))
	}

	log.Printf("Found %d reservations with status %s", len(reservations), status)

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

		log.Printf("Successfully processed reservation %d", reservation.ReservationID)
	}

	return nil
}
