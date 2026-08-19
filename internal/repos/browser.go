package repos

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/blkhole-sh/blkhole/internal/model"
	"github.com/georgysavva/scany/v2/sqlscan"
)

var (
	ErrInvalidBrowserPairing = errors.New("invalid browser pairing")
	ErrBrowserClientNotFound = errors.New("browser client not found")
)

// BrowserRepo stores short-lived pairing grants and long-lived browser clients.
type BrowserRepo interface {
	CreatePairing(deviceID int, tokenHash string, createdAt, expiresAt time.Time) error
	ConsumePairing(tokenHash, clientName, browser, accessTokenHash string, now time.Time) (*model.BrowserClient, error)
	FindActiveClientByTokenHash(tokenHash string, now time.Time) (*model.BrowserClient, error)
	FindClientsByDevice(deviceID int) ([]*model.BrowserClient, error)
	RevokeClient(deviceID, clientID int, now time.Time) error
}

type browserRepo struct {
	db  *sql.DB
	ctx context.Context
}

func NewBrowserRepo(db *sql.DB) BrowserRepo {
	return &browserRepo{db: db, ctx: context.Background()}
}

func (br *browserRepo) CreatePairing(deviceID int, tokenHash string, createdAt, expiresAt time.Time) error {
	if _, err := br.db.ExecContext(br.ctx,
		"DELETE FROM browser_pairing WHERE consumed_at IS NOT NULL OR expires_at <= ?", createdAt.Unix()); err != nil {
		return err
	}
	_, err := br.db.ExecContext(br.ctx,
		"INSERT INTO browser_pairing (device_id, token_hash, created_at, expires_at) VALUES (?, ?, ?, ?)",
		deviceID, tokenHash, createdAt.Unix(), expiresAt.Unix())
	return err
}

func (br *browserRepo) ConsumePairing(tokenHash, clientName, browser, accessTokenHash string, now time.Time) (*model.BrowserClient, error) {
	tx, err := br.db.BeginTx(br.ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var deviceID int
	err = tx.QueryRowContext(br.ctx, `
		UPDATE browser_pairing
		SET consumed_at = ?
		WHERE token_hash = ? AND consumed_at IS NULL AND expires_at > ?
		RETURNING device_id`, now.Unix(), tokenHash, now.Unix()).Scan(&deviceID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalidBrowserPairing
	}
	if err != nil {
		return nil, err
	}

	client := &model.BrowserClient{
		DeviceID:   deviceID,
		Name:       clientName,
		Browser:    browser,
		TokenHash:  accessTokenHash,
		CreatedAt:  now.Unix(),
		LastUsedAt: now.Unix(),
	}
	err = tx.QueryRowContext(br.ctx, `
		INSERT INTO browser_client (device_id, name, browser, token_hash, created_at, last_used_at)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id`, client.DeviceID, client.Name, client.Browser, client.TokenHash, client.CreatedAt, client.LastUsedAt).Scan(&client.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return client, nil
}

func (br *browserRepo) FindActiveClientByTokenHash(tokenHash string, now time.Time) (*model.BrowserClient, error) {
	client := &model.BrowserClient{}
	err := br.db.QueryRowContext(br.ctx, `
		UPDATE browser_client
		SET last_used_at = ?
		WHERE token_hash = ? AND revoked_at IS NULL
		RETURNING id, device_id, name, browser, token_hash, created_at, last_used_at`, now.Unix(), tokenHash).Scan(
		&client.ID, &client.DeviceID, &client.Name, &client.Browser, &client.TokenHash, &client.CreatedAt, &client.LastUsedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrBrowserClientNotFound
	}
	return client, err
}

func (br *browserRepo) FindClientsByDevice(deviceID int) ([]*model.BrowserClient, error) {
	var clients []*model.BrowserClient
	err := sqlscan.Select(br.ctx, br.db, &clients, `
		SELECT id, device_id, name, browser, created_at, last_used_at, revoked_at
		FROM browser_client
		WHERE device_id = ? AND revoked_at IS NULL
		ORDER BY created_at DESC, id DESC`, deviceID)
	if err != nil {
		return nil, err
	}
	if clients == nil {
		clients = []*model.BrowserClient{}
	}
	return clients, nil
}

func (br *browserRepo) RevokeClient(deviceID, clientID int, now time.Time) error {
	result, err := br.db.ExecContext(br.ctx, `
		UPDATE browser_client SET revoked_at = ?
		WHERE id = ? AND device_id = ? AND revoked_at IS NULL`, now.Unix(), clientID, deviceID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrBrowserClientNotFound
	}
	return nil
}
