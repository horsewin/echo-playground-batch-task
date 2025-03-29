package repository

import (
	"fmt"

	"github.com/horsewin/echo-playground-batch-task/internal/model"
	"github.com/jmoiron/sqlx"
)

// NotificationRepository は通知の永続化を担当します
type NotificationRepository struct {
	db *sqlx.DB
}

// NewNotificationRepository は新しいNotificationRepositoryを作成します
func NewNotificationRepository(db *sqlx.DB) *NotificationRepository {
	return &NotificationRepository{
		db: db,
	}
}

// CreateNotifications は複数の通知レコードを作成します
func (r *NotificationRepository) CreateNotifications(records []model.NotificationRecord) error {
	tx, err := r.db.Beginx()
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

	for _, record := range records {
		if err := r.Create(tx, &record); err != nil {
			return fmt.Errorf("failed to create notification: %w", err)
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
func (r *NotificationRepository) Create(tx *sqlx.Tx, record *model.NotificationRecord) error {
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

// BeginTx は新しいトランザクションを開始します
func (r *NotificationRepository) BeginTx() (*sqlx.Tx, error) {
	return r.db.Beginx()
}

// GetByUserID は指定されたユーザーIDの通知を取得します
func (r *NotificationRepository) GetByUserID(userID string) ([]model.NotificationRecord, error) {
	query := `
		SELECT id, user_id, title, message, is_read, type, created_at, updated_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	var records []model.NotificationRecord
	for rows.Next() {
		var record model.NotificationRecord
		err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.Title,
			&record.Message,
			&record.IsRead,
			&record.Type,
			&record.CreatedAt,
			&record.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		records = append(records, record)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notifications: %w", err)
	}

	return records, nil
}

// UpdateIsRead は通知の既読状態を更新します
func (r *NotificationRepository) UpdateIsRead(tx *sqlx.Tx, id int, isRead bool) error {
	query := `
		UPDATE notifications
		SET is_read = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2`

	result, err := tx.Exec(query, isRead, id)
	if err != nil {
		return fmt.Errorf("failed to update notification is_read: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification with id %d not found", id)
	}

	return nil
}
