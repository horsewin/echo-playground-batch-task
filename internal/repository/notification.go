package repository

import (
	"database/sql"

	"github.com/horsewin/echo-playground-batch-task/internal/model"
)

type NotificationRepository struct {
	db *DB
}

func NewNotificationRepository(db *DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create は新しい通知を作成します
func (r *NotificationRepository) Create(tx *sql.Tx, notification *model.Notification) error {
	query := `
		INSERT INTO notifications (user_id, title, message, is_read, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`
	return tx.QueryRow(
		query,
		notification.UserID,
		notification.Title,
		notification.Message,
		notification.IsRead,
	).Scan(&notification.ID)
}
