package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/horsewin/echo-playground-batch-task/internal/config"
	"github.com/horsewin/echo-playground-batch-task/internal/repository"
)

// BaseService は共通のサービス機能を提供します
type BaseService struct {
	db *sql.DB
}

// NewBaseService は新しいBaseServiceを作成します
func NewBaseService(cfg *config.Config) (*BaseService, error) {
	db, err := repository.NewDB(cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	return &BaseService{
		db: db,
	}, nil
}

// Close はデータベース接続を閉じます
func (s *BaseService) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// RunWithTimeout は指定された時間でタイムアウトするコンテキストでバッチ処理を実行します
func (s *BaseService) RunWithTimeout(ctx context.Context, timeout time.Duration, fn func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan error)
	go func() {
		done <- fn(ctx)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation timed out after %v", timeout)
	}
}
