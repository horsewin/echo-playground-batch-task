package repository

import (
	"database/sql"
	"fmt"

	"github.com/horsewin/echo-playground-batch-task/internal/model"
)

// NotificationRepository は通知の永続化を担当します
type NotificationRepository struct {
	db *sql.DB
}

// NewNotificationRepository は新しいNotificationRepositoryを作成します
func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{
		db: db,
	}
}

// CreateNotifications は複数の通知レコードを作成します
func (r *NotificationRepository) CreateNotifications(records []model.NotificationRecord) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// トランザクションのロールバックを遅延実行
	// エラーが発生した場合のみロールバックを実行
	var rollbackErr error
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				rollbackErr = fmt.Errorf("rollback failed: %v, original error: %w", rbErr, err)
			}
		}
	}()

	query := `
		INSERT INTO notifications (
			user_id, title, message, is_read, type, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`

	for _, record := range records {
		_, err = tx.Exec(
			query,
			record.UserID,
			record.Title,
			record.Message,
			record.IsRead,
			record.Type,
			record.CreatedAt,
			record.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert notification: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	if rollbackErr != nil {
		return rollbackErr
	}

	return nil
}

// Create は単一の通知レコードを作成します
func (r *NotificationRepository) Create(tx *sql.Tx, record *model.NotificationRecord) error {
	query := `
		INSERT INTO notifications (
			user_id, title, message, is_read, type, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		RETURNING id`

	return tx.QueryRow(
		query,
		record.UserID,
		record.Title,
		record.Message,
		record.IsRead,
		record.Type,
		record.CreatedAt,
		record.UpdatedAt,
	).Scan(&record.ID)
}
