package postgres

import (
	"context"

	"github.com/google/uuid"
)

// OwnershipAdapter wraps ScanRepository to expose only the ownership check.
// This satisfies middleware.OwnershipRepo without importing the full domain package.
type OwnershipAdapter struct {
	repo *ScanRepository
}

func NewOwnershipAdapter(repo *ScanRepository) *OwnershipAdapter {
	return &OwnershipAdapter{repo: repo}
}

// GetByID returns the owner user_id of the scan, or ErrNotFound.
func (a *OwnershipAdapter) GetByID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	scan, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return uuid.Nil, err
	}
	return scan.UserID, nil
}
