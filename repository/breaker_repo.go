package repository

import (
	"context"
	"fmt"
	"time"

	"qrcode-gen/model"

	"github.com/sony/gobreaker"
)

type breakerRepo struct {
	repo Repository
	cb   *gobreaker.CircuitBreaker
}

func NewBreakerRepository(repo Repository, maxRequests uint32, interval time.Duration, timeout time.Duration, failThreshold uint32) Repository {
	settings := gobreaker.Settings{
		Name:        "db-breaker",
		MaxRequests: maxRequests,
		Interval:    interval,
		Timeout:     timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= failThreshold
		},
	}

	return &breakerRepo{
		repo: repo,
		cb:   gobreaker.NewCircuitBreaker(settings),
	}
}

func (r *breakerRepo) Create(ctx context.Context, qr *model.QRCode) error {
	_, err := r.cb.Execute(func() (interface{}, error) {
		return nil, r.repo.Create(ctx, qr)
	})
	return err
}

func (r *breakerRepo) GetByToken(ctx context.Context, qrToken string) (*model.QRCode, error) {
	result, err := r.cb.Execute(func() (interface{}, error) {
		return r.repo.GetByToken(ctx, qrToken)
	})
	if err != nil {
		return nil, fmt.Errorf("circuit breaker: %w", err)
	}
	return result.(*model.QRCode), nil
}

func (r *breakerRepo) GetByUserID(ctx context.Context, userID string) ([]*model.QRCode, error) {
	result, err := r.cb.Execute(func() (interface{}, error) {
		return r.repo.GetByUserID(ctx, userID)
	})
	if err != nil {
		return nil, fmt.Errorf("circuit breaker: %w", err)
	}
	return result.([]*model.QRCode), nil
}

func (r *breakerRepo) Update(ctx context.Context, qrToken string, url string) error {
	_, err := r.cb.Execute(func() (interface{}, error) {
		return nil, r.repo.Update(ctx, qrToken, url)
	})
	return err
}

func (r *breakerRepo) Delete(ctx context.Context, qrToken string) error {
	_, err := r.cb.Execute(func() (interface{}, error) {
		return nil, r.repo.Delete(ctx, qrToken)
	})
	return err
}

func (r *breakerRepo) TokenExists(ctx context.Context, qrToken string) (bool, error) {
	result, err := r.cb.Execute(func() (interface{}, error) {
		return r.repo.TokenExists(ctx, qrToken)
	})
	if err != nil {
		return false, err
	}
	return result.(bool), nil
}
