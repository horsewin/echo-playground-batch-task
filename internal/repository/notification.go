package repository

import (
	"database/sql"

	"github.com/horsewin/echo-playground-batch-task/internal/model"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(tx *sql.Tx, notification *model.Notification) error {
	query := `
		INSERT INTO notifications (user_id, title, message, is_read)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	return tx.QueryRow(
		query,
		notification.UserID,
		notification.Title,
		notification.Message,
		notification.IsRead,
	).Scan(&notification.ID, &notification.CreatedAt, &notification.UpdatedAt)
}
