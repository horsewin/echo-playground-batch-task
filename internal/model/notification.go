package model

import "time"

// NotificationType は通知の種類を表します
type NotificationType string

const (
	// NotificationTypeReservation は予約関連の通知を表します
	NotificationTypeReservation NotificationType = "reservation"
)

// Notification は共通的な通知定義です（イベント用）
type Notification struct {
	Type      NotificationType `json:"type"`
	UserID    string           `json:"user_id"`
	DateTime  time.Time        `json:"date_time"`
	CreatedAt time.Time        `json:"created_at"`
	Data      interface{}      `json:"data"`
}

// NotificationRecord はデータベースに永続化される通知レコードです
type NotificationRecord struct {
	ID        int              `db:"id"`
	UserID    string           `db:"user_id"`
	Title     string           `db:"title"`
	Message   string           `db:"message"`
	IsRead    bool             `db:"is_read"`
	Type      NotificationType `db:"type"`
	CreatedAt time.Time        `db:"created_at"`
	UpdatedAt time.Time        `db:"updated_at"`
}

// NewReservationNotification は予約イベントから通知を作成します
func NewReservationNotification(event ReservationEvent) Notification {
	return Notification{
		Type:      NotificationTypeReservation,
		CreatedAt: event.CreatedAt,
		Data: map[string]interface{}{
			"user_id":   event.UserID,
			"pet_id":    event.PetID,
			"date_time": event.DateTime,
		},
	}
}

// NewReservationNotificationRecord は予約イベントから通知レコードを作成します
func NewReservationNotificationRecord(event ReservationEvent) NotificationRecord {
	now := time.Now()
	return NotificationRecord{
		UserID:    event.UserID,
		Title:     "予約の更新",
		Message:   "予約のステータスが更新されました",
		IsRead:    false,
		Type:      NotificationTypeReservation,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// ToNotification は通知レコードをイベント用の通知に変換します
func (r NotificationRecord) ToNotification() Notification {
	return Notification{
		Type:      r.Type,
		UserID:    r.UserID,
		DateTime:  r.CreatedAt,
		CreatedAt: r.CreatedAt,
		Data: map[string]interface{}{
			"title":   r.Title,
			"message": r.Message,
			"is_read": r.IsRead,
		},
	}
}
