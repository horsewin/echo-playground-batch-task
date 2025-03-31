package batch

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/horsewin/echo-playground-batch-task/internal/common/config"
	"github.com/horsewin/echo-playground-batch-task/internal/common/models"
	"github.com/horsewin/echo-playground-batch-task/internal/model"
	"github.com/jmoiron/sqlx"
)

// MockReservationRepository はテスト用のモックリポジトリです
type MockReservationRepository struct {
	createReservationsCalled bool
	createReservationsError  error
	reservations             []model.Reservation
}

func (m *MockReservationRepository) CreateReservations(ctx context.Context, reservations []model.Reservation) error {
	m.createReservationsCalled = true
	m.reservations = reservations
	return m.createReservationsError
}

func (m *MockReservationRepository) BeginTx() (*sqlx.Tx, error) {
	return nil, nil
}

func (m *MockReservationRepository) CheckExistingReservation(ctx context.Context, petID string) (bool, error) {
	return false, nil
}

func (m *MockReservationRepository) UpdateStatus(ctx context.Context, tx *sqlx.Tx, reservationID int64, status string) error {
	return nil
}

func (m *MockReservationRepository) GetReservationsByStatus(ctx context.Context, status string) ([]models.Reservation, error) {
	return nil, nil
}

// newTestReservationBatchService はテスト用のReservationBatchServiceを作成します
func newTestReservationBatchService(mockReservationRepo *MockReservationRepository) *ReservationBatchService {
	return &ReservationBatchService{
		reservationRepo: mockReservationRepo,
		cfg:             &config.Config{},
	}
}

func TestReservationBatchService_Run(t *testing.T) {
	// X-Rayのセグメントを設定
	ctx, seg := xray.BeginSegment(context.Background(), "TestReservationBatchService_Run")
	defer seg.Close(nil)

	now := time.Now().UTC()
	tests := []struct {
		name         string
		reservations []model.Reservation
		mockError    error
		wantErr      bool
	}{
		{
			name:         "0件の予約を正常に処理",
			reservations: []model.Reservation{},
			mockError:    nil,
			wantErr:      false,
		},
		{
			name: "1件の予約を正常に処理",
			reservations: []model.Reservation{
				{
					ID:                  1,
					UserID:              "user1",
					UserName:            "Test User 1",
					Email:               "test1@example.com",
					PetID:               "pet1",
					ReservationDateTime: now,
					Status:              "pending",
					CreatedAt:           now,
					UpdatedAt:           now,
				},
			},
			mockError: nil,
			wantErr:   false,
		},
		{
			name: "2件の予約を正常に処理",
			reservations: []model.Reservation{
				{
					ID:                  1,
					UserID:              "user1",
					UserName:            "Test User 1",
					Email:               "test1@example.com",
					PetID:               "pet1",
					ReservationDateTime: now,
					Status:              "pending",
					CreatedAt:           now,
					UpdatedAt:           now,
				},
				{
					ID:                  2,
					UserID:              "user2",
					UserName:            "Test User 2",
					Email:               "test2@example.com",
					PetID:               "pet2",
					ReservationDateTime: now,
					Status:              "pending",
					CreatedAt:           now,
					UpdatedAt:           now,
				},
			},
			mockError: nil,
			wantErr:   false,
		},
		{
			name: "異なるステータスの予約を処理",
			reservations: []model.Reservation{
				{
					ID:                  1,
					UserID:              "user1",
					UserName:            "Test User 1",
					Email:               "test1@example.com",
					PetID:               "pet1",
					ReservationDateTime: now,
					Status:              "confirmed",
					CreatedAt:           now,
					UpdatedAt:           now,
				},
				{
					ID:                  2,
					UserID:              "user2",
					UserName:            "Test User 2",
					Email:               "test2@example.com",
					PetID:               "pet2",
					ReservationDateTime: now,
					Status:              "cancelled",
					CreatedAt:           now,
					UpdatedAt:           now,
				},
				{
					ID:                  3,
					UserID:              "user3",
					UserName:            "Test User 3",
					Email:               "test3@example.com",
					PetID:               "pet3",
					ReservationDateTime: now,
					Status:              "pending",
					CreatedAt:           now,
					UpdatedAt:           now,
				},
			},
			mockError: nil,
			wantErr:   false,
		},
		{
			name: "リポジトリからのエラーを処理",
			reservations: []model.Reservation{
				{
					ID:                  1,
					UserID:              "user1",
					UserName:            "Test User 1",
					Email:               "test1@example.com",
					PetID:               "pet1",
					ReservationDateTime: now,
					Status:              "pending",
					CreatedAt:           now,
					UpdatedAt:           now,
				},
			},
			mockError: fmt.Errorf("database error: connection failed"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReservationRepo := &MockReservationRepository{
				createReservationsError: tt.mockError,
			}

			service := newTestReservationBatchService(mockReservationRepo)
			service.SetArgs(tt.reservations)
			err := service.Run(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !mockReservationRepo.createReservationsCalled {
				t.Error("CreateReservations was not called")
			}

			if len(mockReservationRepo.reservations) != len(tt.reservations) {
				t.Errorf("Expected %d reservations, got %d", len(tt.reservations), len(mockReservationRepo.reservations))
			}
		})
	}
}
