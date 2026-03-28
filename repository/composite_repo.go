package repository

import (
	"context"
	"log"

	"qrcode-gen/model"

	"golang.org/x/sync/singleflight"
)

type compositeRepo struct {
	cache Repository
	db    Repository
	sf    singleflight.Group
}

func NewCompositeRepository(cache Repository, db Repository) Repository {
	return &compositeRepo{
		cache: cache,
		db:    db,
	}
}

func (r *compositeRepo) Create(ctx context.Context, qr *model.QRCode) error {
	if err := r.db.Create(ctx, qr); err != nil {
		return err
	}

	if err := r.cache.Create(ctx, qr); err != nil {
		log.Printf("[CACHE WRITE FAILED] %s: %v", qr.QRToken, err)
	}

	return nil
}

func (r *compositeRepo) GetByToken(ctx context.Context, qrToken string) (*model.QRCode, error) {
	qr, err := r.cache.GetByToken(ctx, qrToken)
	if err == nil {
		log.Printf("[CACHE HIT] %s", qrToken)
		return qr, nil
	}

	// cache miss，用 singleflight 合併相同 token 的請求
	result, err, shared := r.sf.Do(qrToken, func() (interface{}, error) {
		log.Printf("[CACHE MISS] %s, querying DB", qrToken)

		qr, err := r.db.GetByToken(ctx, qrToken)
		if err != nil {
			return nil, err
		}

		if writeErr := r.cache.Create(ctx, qr); writeErr != nil {
			log.Printf("[CACHE WRITE FAILED] %s: %v", qrToken, writeErr)
		}

		return qr, nil
	})

	if err != nil {
		return nil, err
	}

	if shared {
		log.Printf("[SINGLEFLIGHT] %s result shared with other requests", qrToken)
	}

	return result.(*model.QRCode), nil
}

func (r *compositeRepo) GetByUserID(ctx context.Context, userID string) ([]*model.QRCode, error) {
	return r.db.GetByUserID(ctx, userID)
}

func (r *compositeRepo) Update(ctx context.Context, qrToken string, url string) error {
	if err := r.db.Update(ctx, qrToken, url); err != nil {
		return err
	}

	if err := r.cache.Delete(ctx, qrToken); err != nil {
		log.Printf("[CACHE INVALIDATE FAILED] %s: %v", qrToken, err)
	}

	return nil
}

func (r *compositeRepo) Delete(ctx context.Context, qrToken string) error {
	if err := r.db.Delete(ctx, qrToken); err != nil {
		return err
	}

	if err := r.cache.Delete(ctx, qrToken); err != nil {
		log.Printf("[CACHE DELETE FAILED] %s: %v", qrToken, err)
	}

	return nil
}

func (r *compositeRepo) TokenExists(ctx context.Context, qrToken string) (bool, error) {
	exists, err := r.cache.TokenExists(ctx, qrToken)
	if err == nil && exists {
		return true, nil
	}
	return r.db.TokenExists(ctx, qrToken)
}
