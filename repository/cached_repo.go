package repository

import (
	"log"

	"qrcode-gen/model"
)

type cachedRepo struct {
	cache Repository // memory（快）
	db    Repository // sqlite（持久）
}

func NewCachedRepository(cache Repository, db Repository) Repository {
	return &cachedRepo{cache: cache, db: db}
}
func (r *cachedRepo) Create(qr *model.QRCode) error {
	// 先寫 DB（持久化最重要）
	if err := r.db.Create(qr); err != nil {
		return err
	}

	// 再寫 cache
	r.cache.Create(qr)

	return nil
}
func (r *cachedRepo) Update(qrToken string, url string) error {
	// 先更新 DB
	if err := r.db.Update(qrToken, url); err != nil {
		return err
	}
	// 寫入時同時更新 DB，並在必要時做 cache invalidation
	r.cache.Delete(qrToken)

	return nil
}

func (r *cachedRepo) GetByToken(qrToken string) (*model.QRCode, error) {
	// Step 1: 先查 cache
	qr, err := r.cache.GetByToken(qrToken)
	if err == nil {
		log.Printf("[CACHE HIT] %s", qrToken)
		return qr, nil
	}

	// Step 2: cache miss，查 DB
	log.Printf("[CACHE MISS] %s, querying DB", qrToken)
	qr, err = r.db.GetByToken(qrToken)
	if err != nil {
		return nil, err
	}

	// Step 3: 寫回 cache（下次就不用查 DB 了）
	r.cache.Create(qr)

	return qr, nil
}

func (r *cachedRepo) Delete(qrToken string) error {
	if err := r.db.Delete(qrToken); err != nil {
		return err
	}

	// 也從 cache 移除
	r.cache.Delete(qrToken)

	return nil
}

func (r *cachedRepo) TokenExists(qrToken string) (bool, error) {
	// 先查 cache
	exists, _ := r.cache.TokenExists(qrToken)
	if exists {
		return true, nil
	}
	// cache 沒有，查 DB
	return r.db.TokenExists(qrToken)
}

func (r *cachedRepo) GetByUserID(userID string) ([]*model.QRCode, error) {
	// 這個操作不常用，直接查 DB
	return r.db.GetByUserID(userID)
}
