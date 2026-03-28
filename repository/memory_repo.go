package repository

import (
	"fmt"
	"sync"

	"qrcode-gen/model"
)

type memoryRepo struct {
	mu      sync.RWMutex
	byToken map[string]*model.QRCode
}

func NewMemoryRepository() Repository {
	return &memoryRepo{
		byToken: make(map[string]*model.QRCode),
	}
}

func (r *memoryRepo) Create(qr *model.QRCode) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byToken[qr.QRToken]; exists {
		return fmt.Errorf("token already exists: %s", qr.QRToken)
	}

	r.byToken[qr.QRToken] = qr
	return nil
}

func (r *memoryRepo) GetByToken(qrToken string) (*model.QRCode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	qr, exists := r.byToken[qrToken]
	if !exists {
		return nil, fmt.Errorf("qr code not found: %s", qrToken)
	}
	return qr, nil
}

func (r *memoryRepo) GetByUserID(userID string) ([]*model.QRCode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*model.QRCode
	for _, qr := range r.byToken {
		if qr.UserID == userID {
			results = append(results, qr)
		}
	}
	return results, nil
}

func (r *memoryRepo) Update(qrToken string, url string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	qr, exists := r.byToken[qrToken]
	if !exists {
		return fmt.Errorf("qr code not found: %s", qrToken)
	}
	qr.URL = url
	return nil
}

func (r *memoryRepo) Delete(qrToken string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byToken[qrToken]; !exists {
		return fmt.Errorf("qr code not found: %s", qrToken)
	}
	delete(r.byToken, qrToken)
	return nil
}

func (r *memoryRepo) TokenExists(qrToken string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.byToken[qrToken]
	return exists, nil
}
