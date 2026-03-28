package service

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"qrcode-gen/model"
	"qrcode-gen/repository"
	"qrcode-gen/token"
)

type Service interface {
	CreateQRCode(ctx context.Context, userID string, rawURL string) (*model.CreateResponse, error)
	GetOriginalURL(ctx context.Context, qrToken string) (string, error)
	UpdateQRCode(ctx context.Context, qrToken string, newURL string) error
	DeleteQRCode(ctx context.Context, qrToken string) error
}

type qrCodeService struct {
	repo       repository.Repository
	tokenGen   token.Generator
	baseURL    string
	maxRetries int
	maxURLLen  int
}

// NewService 接收 maxRetries 和 maxURLLen，不再寫死
func NewService(repo repository.Repository, tokenGen token.Generator, baseURL string, maxRetries int, maxURLLen int) Service {
	return &qrCodeService{
		repo:       repo,
		tokenGen:   tokenGen,
		baseURL:    baseURL,
		maxRetries: maxRetries,
		maxURLLen:  maxURLLen,
	}
}

func (s *qrCodeService) CreateQRCode(ctx context.Context, userID string, rawURL string) (*model.CreateResponse, error) {
	if err := s.validateURL(rawURL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	var qrToken string
	for i := 0; i < s.maxRetries; i++ {
		candidate, err := s.tokenGen.Generate(rawURL)
		if err != nil {
			return nil, fmt.Errorf("failed to generate token: %w", err)
		}

		exists, err := s.repo.TokenExists(ctx, candidate)
		if err != nil {
			return nil, fmt.Errorf("failed to check token: %w", err)
		}
		if !exists {
			qrToken = candidate
			break
		}
	}

	if qrToken == "" {
		return nil, fmt.Errorf("failed to generate unique token after %d retries", s.maxRetries)
	}

	qr := &model.QRCode{
		ID:        qrToken,
		UserID:    userID,
		QRToken:   qrToken,
		URL:       rawURL,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, qr); err != nil {
		return nil, fmt.Errorf("failed to create qr code: %w", err)
	}

	return &model.CreateResponse{QRToken: qrToken}, nil
}

func (s *qrCodeService) GetOriginalURL(ctx context.Context, qrToken string) (string, error) {
	qr, err := s.repo.GetByToken(ctx, qrToken)
	if err != nil {
		return "", err
	}
	return qr.URL, nil
}

func (s *qrCodeService) UpdateQRCode(ctx context.Context, qrToken string, newURL string) error {
	if err := s.validateURL(newURL); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	return s.repo.Update(ctx, qrToken, newURL)
}

func (s *qrCodeService) DeleteQRCode(ctx context.Context, qrToken string) error {
	return s.repo.Delete(ctx, qrToken)
}

func (s *qrCodeService) validateURL(rawURL string) error {
	if len(rawURL) == 0 {
		return fmt.Errorf("URL cannot be empty")
	}
	if len(rawURL) > s.maxURLLen {
		return fmt.Errorf("URL exceeds maximum length of %d characters", s.maxURLLen)
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
