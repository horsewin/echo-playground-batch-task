package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// PetRepository はペット情報の永続化を担当するインターフェースです
type PetRepository interface {
	GetNameByID(petID string) (string, error)
}

// PetRepositoryImpl はPetRepositoryの実装です
type PetRepositoryImpl struct {
	db *sqlx.DB
}

// NewPetRepository は新しいPetRepositoryを作成します
func NewPetRepository(db *sqlx.DB) PetRepository {
	return &PetRepositoryImpl{
		db: db,
	}
}

// GetNameByID は指定されたペットIDからペット名を取得します
func (r *PetRepositoryImpl) GetNameByID(petID string) (string, error) {
	query := `
		SELECT name
		FROM pets
		WHERE id = $1`

	var name string
	err := r.db.QueryRow(query, petID).Scan(&name)
	if err != nil {
		return "", fmt.Errorf("failed to get pet name: %w", err)
	}

	return name, nil
}
