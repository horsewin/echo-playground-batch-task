package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/horsewin/echo-playground-batch-task/internal/config"
	"github.com/horsewin/echo-playground-batch-task/internal/model"
	"github.com/horsewin/echo-playground-batch-task/internal/repository"
)

type BatchService struct {
	config           *config.Config
	reservationRepo  *repository.ReservationRepository
	notificationRepo *repository.NotificationRepository
}

func NewBatchService(cfg *config.Config) *BatchService {
	db, err := repository.NewDB(cfg.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	return &BatchService{
		config:           cfg,
		reservationRepo:  repository.NewReservationRepository(db),
		notificationRepo: repository.NewNotificationRepository(db),
	}
}

func (s *BatchService) Run(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := s.processPendingReservations(); err != nil {
				log.Printf("Error processing pending reservations: %v", err)
			}
		}
	}
}

func (s *BatchService) processPendingReservations() error {
	// 保留中の予約を取得
	reservations, err := s.reservationRepo.GetPendingReservations()
	if err != nil {
		return fmt.Errorf("failed to get pending reservations: %w", err)
	}

	for _, reservation := range reservations {
		// トランザクション開始
		tx, err := s.reservationRepo.DB.BeginTx()
		if err != nil {
			log.Printf("Failed to begin transaction for reservation %d: %v", reservation.ReservationID, err)
			continue
		}

		// 既存の予約をチェック
		exists, err := s.reservationRepo.CheckExistingReservation(reservation.PetID)
		if err != nil {
			tx.Rollback()
			log.Printf("Failed to check existing reservation for pet %s: %v", reservation.PetID, err)
			continue
		}

		if exists {
			// 既存の予約がある場合は、この予約をキャンセル
			if err := s.reservationRepo.UpdateStatus(tx, reservation.ReservationID, "cancelled"); err != nil {
				tx.Rollback()
				log.Printf("Failed to update reservation status to cancelled: %v", err)
				continue
			}

			// キャンセル通知を作成
			notification := &model.Notification{
				UserID:  reservation.UserID,
				Title:   "予約キャンセル",
				Message: fmt.Sprintf("申し訳ありませんが、ペットID %s は既に予約が入っているため、予約をキャンセルさせていただきました。", reservation.PetID),
				IsRead:  false,
			}

			if err := s.notificationRepo.Create(tx, notification); err != nil {
				tx.Rollback()
				log.Printf("Failed to create cancellation notification: %v", err)
				continue
			}
		} else {
			// 既存の予約がない場合は、予約を確定
			if err := s.reservationRepo.UpdateStatus(tx, reservation.ReservationID, "confirmed"); err != nil {
				tx.Rollback()
				log.Printf("Failed to update reservation status to confirmed: %v", err)
				continue
			}

			// 予約確定通知を作成
			notification := &model.Notification{
				UserID:  reservation.UserID,
				Title:   "予約確定",
				Message: fmt.Sprintf("ペットID %s の予約が確定しました。予約日時: %s", reservation.PetID, reservation.ReservationDateTime.Format("2006-01-02 15:04:05")),
				IsRead:  false,
			}

			if err := s.notificationRepo.Create(tx, notification); err != nil {
				tx.Rollback()
				log.Printf("Failed to create confirmation notification: %v", err)
				continue
			}
		}

		// トランザクションをコミット
		if err := tx.Commit(); err != nil {
			log.Printf("Failed to commit transaction for reservation %d: %v", reservation.ReservationID, err)
			continue
		}

		log.Printf("Successfully processed reservation %d", reservation.ReservationID)
	}

	return nil
}
