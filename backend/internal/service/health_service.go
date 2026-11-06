package service

import (
	"context"
	"fmt"

	"stockflow/internal/repository"
)

type HealthService struct {
	repo repository.HealthRepository
}

func NewHealthService(repo repository.HealthRepository) *HealthService {
	return &HealthService{repo: repo}
}

func (s *HealthService) CheckDatabase(ctx context.Context) error {
	if err := s.repo.Ping(ctx); err != nil {
		return fmt.Errorf("database unreachable: %w", err)
	}
	return nil
}
