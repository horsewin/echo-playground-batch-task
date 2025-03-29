package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/horsewin/echo-playground-batch-task/internal/common/models"
)

type ReservationRepository struct {
	db *sql.DB
}

func NewReservationRepository(db *sql.DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

// BeginTx starts a new transaction
func (r *ReservationRepository) BeginTx() (*sql.Tx, error) {
	return r.db.Begin()
}

// GetPendingReservations は、ステータスがpendingの予約を取得します
func (r *ReservationRepository) GetPendingReservations() ([]models.Reservation, error) {
	query := `
		SELECT 
			id,
			user_id,
			user_name,
			email,
			reservation_date_time,
			pet_id,
			created_at,
			updated_at,
			status
		FROM reservations
		WHERE status = 'pending'
		ORDER BY reservation_date_time ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending reservations: %w", err)
	}
	defer rows.Close()

	var reservations []models.Reservation
	for rows.Next() {
		var r models.Reservation
		err := rows.Scan(
			&r.ReservationID,
			&r.UserID,
			&r.UserName,
			&r.Email,
			&r.ReservationDateTime,
			&r.PetID,
			&r.CreatedAt,
			&r.UpdatedAt,
			&r.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reservation row: %w", err)
		}
		reservations = append(reservations, r)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reservation rows: %w", err)
	}

	return reservations, nil
}

// UpdateStatus は予約のステータスを更新します
func (r *ReservationRepository) UpdateStatus(tx *sql.Tx, reservationID int64, status string) error {
	query := `
		UPDATE reservations
		SET status = $1,
			updated_at = $2
		WHERE id = $3
	`

	result, err := tx.Exec(query, status, time.Now(), reservationID)
	if err != nil {
		return fmt.Errorf("failed to update reservation status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no reservation found with ID %d", reservationID)
	}

	return nil
}

// CheckExistingReservation は、指定されたペットIDに対して予約が存在するかチェックします
func (r *ReservationRepository) CheckExistingReservation(petID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM reservations
			WHERE pet_id = $1
			AND status = 'confirmed'
			AND reservation_date_time > NOW()
		)
	`

	var exists bool
	err := r.db.QueryRow(query, petID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check existing reservation: %w", err)
	}

	return exists, nil
}
