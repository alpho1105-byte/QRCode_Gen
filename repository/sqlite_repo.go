package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"qrcode-gen/model"

	_ "modernc.org/sqlite"
)

type sqliteRepo struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string, maxOpenConns int, maxIdleConns int, connMaxLifetime time.Duration) (Repository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS qr_codes (
		id         TEXT PRIMARY KEY,
		user_id    TEXT NOT NULL,
		qr_token   TEXT NOT NULL UNIQUE,
		url        TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_qr_token ON qr_codes(qr_token);
	CREATE INDEX IF NOT EXISTS idx_user_id ON qr_codes(user_id);
	`

	if _, err := db.Exec(createTable); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	return &sqliteRepo{db: db}, nil
}

func (r *sqliteRepo) Create(ctx context.Context, qr *model.QRCode) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO qr_codes (id, user_id, qr_token, url, created_at) VALUES (?, ?, ?, ?, ?)",
		qr.ID, qr.UserID, qr.QRToken, qr.URL, qr.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert: %w", err)
	}
	return nil
}

func (r *sqliteRepo) GetByToken(ctx context.Context, qrToken string) (*model.QRCode, error) {
	qr := &model.QRCode{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, user_id, qr_token, url, created_at FROM qr_codes WHERE qr_token = ?",
		qrToken,
	).Scan(&qr.ID, &qr.UserID, &qr.QRToken, &qr.URL, &qr.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("qr code not found: %s", qrToken)
	}
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	return qr, nil
}

func (r *sqliteRepo) GetByUserID(ctx context.Context, userID string) ([]*model.QRCode, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, user_id, qr_token, url, created_at FROM qr_codes WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var results []*model.QRCode
	for rows.Next() {
		qr := &model.QRCode{}
		if err := rows.Scan(&qr.ID, &qr.UserID, &qr.QRToken, &qr.URL, &qr.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		results = append(results, qr)
	}
	return results, nil
}

func (r *sqliteRepo) Update(ctx context.Context, qrToken string, url string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE qr_codes SET url = ? WHERE qr_token = ?",
		url, qrToken,
	)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("qr code not found: %s", qrToken)
	}
	return nil
}

func (r *sqliteRepo) Delete(ctx context.Context, qrToken string) error {
	result, err := r.db.ExecContext(ctx,
		"DELETE FROM qr_codes WHERE qr_token = ?",
		qrToken,
	)
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("qr code not found: %s", qrToken)
	}
	return nil
}

func (r *sqliteRepo) TokenExists(ctx context.Context, qrToken string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM qr_codes WHERE qr_token = ?",
		qrToken,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("query failed: %w", err)
	}
	return count > 0, nil
}
