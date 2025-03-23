package repository

import (
	"database/sql"
	"fmt"

	"github.com/horsewin/echo-playground-batch-task/internal/model"
)

type ReservationRepository struct {
	db *DB
}

func NewReservationRepository(db *DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

// BeginTx starts a new transaction
func (r *ReservationRepository) BeginTx() (*sql.Tx, error) {
	return r.db.BeginTx()
}

// GetPendingReservations は、ステータスがpendingの予約を取得します
func (r *ReservationRepository) GetPendingReservations() ([]model.Reservation, error) {
	var reservations []model.Reservation
	query := `
		SELECT reservation_id, user_id, user_name, email, reservation_datetime, pet_id, created_at, updated_at, status
		FROM reservations
		WHERE status = 'pending'
		ORDER BY created_at ASC
	`
	err := r.db.Select(&reservations, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending reservations: %w", err)
	}
	return reservations, nil
}

// UpdateStatus は予約のステータスを更新します
func (r *ReservationRepository) UpdateStatus(tx *sql.Tx, reservationID int, status string) error {
	query := `
		UPDATE reservations
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE reservation_id = $2
	`
	result, err := tx.Exec(query, status, reservationID)
	if err != nil {
		return fmt.Errorf("failed to update reservation status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("no reservation found with ID: %d", reservationID)
	}

	return nil
}

// CheckExistingReservation は、指定されたペットIDに対して予約が存在するかチェックします
func (r *ReservationRepository) CheckExistingReservation(petID string) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM reservations
		WHERE pet_id = $1 AND status = 'confirmed'
	`
	err := r.db.Get(&count, query, petID)
	if err != nil {
		return false, fmt.Errorf("failed to check existing reservation: %w", err)
	}
	return count > 0, nil
}
