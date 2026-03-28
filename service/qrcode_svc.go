package service

import (
	"fmt"
	"log"
	"net/url"
	"time"

	//local
	"qrcode-gen/model"
	"qrcode-gen/repository"
	"qrcode-gen/token"
)

const (
	maxURLLength = 20 // 設計文件：ASCII，最長 20 字元
	maxRetries   = 5  // Token 碰撞時的最大重試次數
)

type Service interface {
	CreateQRCode(userID string, rawURL string) (*model.CreateResponse, error)
	GetOriginalURL(qrToken string) (string, error)
	UpdateQRCode(qrToken string, newURL string) error
	DeleteQRCode(qrToken string) error
}

type qrCodeService struct {
	repo     repository.Repository
	tokenGen token.Generator
	baseURL  string
}

func NewService(repo repository.Repository, tokenGen token.Generator, baseURL string) Service {
	return &qrCodeService{
		repo:     repo,
		tokenGen: tokenGen,
		baseURL:  baseURL,
	}
}

func (s *qrCodeService) CreateQRCode(userID string, rawURL string) (*model.CreateResponse, error) {
	if err := validateURL(rawURL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	var qrToken string
	for i := 0; i < maxRetries; i++ {

		candidate, err := s.tokenGen.Generate(rawURL)
		if err != nil {
			return nil, fmt.Errorf("failed to generate token: %w", err)
		}

		log.Printf("retry %d, candidate token: [%s]", i, candidate) // ← 加這行

		exists, err := s.repo.TokenExists(candidate)
		if err != nil {
			return nil, fmt.Errorf("failed to check token: %w", err)
		}
		if !exists {
			qrToken = candidate
			break
		}
		// collision，繼續重試
	}

	if qrToken == "" {
		return nil, fmt.Errorf("failed to generate unique token after %d retries", maxRetries)
	}

	// 取得token成功後，建立物件
	qr := &model.QRCode{
		ID:        qrToken,
		UserID:    userID,
		QRToken:   qrToken,
		URL:       rawURL,
		CreatedAt: time.Now(),
	}

	// 儲存
	if err := s.repo.Create(qr); err != nil {
		return nil, fmt.Errorf("failed to create qr code: %w", err)
	}

	// 儲存成功，回傳 token
	return &model.CreateResponse{QRToken: qrToken}, nil
}

func (s *qrCodeService) GetOriginalURL(qrToken string) (string, error) {
	qr, err := s.repo.GetByToken(qrToken)
	if err != nil {
		return "", err
	}
	return qr.URL, nil
}

func (s *qrCodeService) UpdateQRCode(qrToken string, newURL string) error {
	if err := validateURL(newURL); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	return s.repo.Update(qrToken, newURL)
}

func (s *qrCodeService) DeleteQRCode(qrToken string) error {
	return s.repo.Delete(qrToken)
}

func validateURL(rawURL string) error {
	if len(rawURL) == 0 {
		return fmt.Errorf("URL cannot be empty")
	}
	if len(rawURL) > maxURLLength {
		return fmt.Errorf("URL exceeds maximum length of %d characters", maxURLLength)
	}

	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("malformed URL: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("URL must start with http:// or https://")
	}

	return nil
}
