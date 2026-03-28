package repository

import (
	"database/sql"
	"fmt"

	"qrcode-gen/model"

	_ "modernc.org/sqlite"
)

type sqliteRepo struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (Repository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 建表：對應設計文件的 QrCodes table
	// IF NOT EXISTS 讓它可以重複執行不會報錯
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

	return &sqliteRepo{db: db}, nil
}

func (r *sqliteRepo) Create(qr *model.QRCode) error {
	_, err := r.db.Exec(
		"INSERT INTO qr_codes (id, user_id, qr_token, url, created_at) VALUES (?, ?, ?, ?, ?)",
		qr.ID, qr.UserID, qr.QRToken, qr.URL, qr.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert: %w", err)
	}
	return nil
}

func (r *sqliteRepo) GetByToken(qrToken string) (*model.QRCode, error) {
	qr := &model.QRCode{}
	err := r.db.QueryRow(
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

func (r *sqliteRepo) GetByUserID(userID string) ([]*model.QRCode, error) {
	rows, err := r.db.Query(
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

func (r *sqliteRepo) Update(qrToken string, url string) error {
	result, err := r.db.Exec(
		"UPDATE qr_codes SET url = ? WHERE qr_token = ?",
		url, qrToken,
	)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	// 檢查有沒有真的更新到資料
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("qr code not found: %s", qrToken)
	}
	return nil
}

func (r *sqliteRepo) Delete(qrToken string) error {
	result, err := r.db.Exec(
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

func (r *sqliteRepo) TokenExists(qrToken string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM qr_codes WHERE qr_token = ?",
		qrToken,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("query failed: %w", err)
	}
	return count > 0, nil
}
