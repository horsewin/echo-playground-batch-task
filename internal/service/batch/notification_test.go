package batch

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/horsewin/echo-playground-batch-task/internal/common/config"
	"github.com/horsewin/echo-playground-batch-task/internal/model"
)

// MockNotificationRepository はテスト用のモックリポジトリです
type MockNotificationRepository struct {
	createNotificationsCalled bool
	createNotificationsError  error
	notifications             []model.NotificationRecord
}

func (m *MockNotificationRepository) CreateNotifications(records []model.NotificationRecord) error {
	m.createNotificationsCalled = true
	m.notifications = records
	return m.createNotificationsError
}

func (m *MockNotificationRepository) Create(tx *sql.Tx, record *model.NotificationRecord) error {
	return nil
}

func TestNotificationBatchService_Run(t *testing.T) {
	tests := []struct {
		name          string
		notifications []model.Notification
		mockError     error
		wantErr       bool
	}{
		{
			name:          "0件の通知を正常に処理",
			notifications: []model.Notification{},
			mockError:     nil,
			wantErr:       false,
		},
		{
			name: "1件の通知を正常に処理",
			notifications: []model.Notification{
				{
					UserID:    "user1",
					Type:      "test",
					Data:      map[string]interface{}{"key": "value"},
					CreatedAt: time.Now(),
				},
			},
			mockError: nil,
			wantErr:   false,
		},
		{
			name: "2件の通知を正常に処理",
			notifications: []model.Notification{
				{
					UserID:    "user1",
					Type:      "test1",
					Data:      map[string]interface{}{"key": "value1"},
					CreatedAt: time.Now(),
				},
				{
					UserID:    "user2",
					Type:      "test2",
					Data:      map[string]interface{}{"key": "value2"},
					CreatedAt: time.Now(),
				},
			},
			mockError: nil,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockNotificationRepository{
				createNotificationsError: tt.mockError,
			}

			service := &NotificationBatchService{
				notificationRepo: mockRepo,
				cfg:              &config.Config{},
			}

			err := service.Run(context.Background(), tt.notifications)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !mockRepo.createNotificationsCalled {
				t.Error("CreateNotifications was not called")
			}

			if len(mockRepo.notifications) != len(tt.notifications) {
				t.Errorf("Expected %d notifications, got %d", len(tt.notifications), len(mockRepo.notifications))
			}
		})
	}
}
