package repos

import (
	"context"
	"database/sql"
)

const upstreamDNSKey = "upstream_dns"

// SettingsRepo stores server settings that can be changed at runtime.
type SettingsRepo interface {
	GetUpstreamDNS(fallback string) (string, error)
	UpdateUpstreamDNS(value string) error
}

type settingsRepo struct {
	db  *sql.DB
	ctx context.Context
}

func NewSettingsRepo(db *sql.DB) SettingsRepo {
	return &settingsRepo{db: db, ctx: context.Background()}
}

func (r *settingsRepo) GetUpstreamDNS(fallback string) (string, error) {
	_, err := r.db.ExecContext(r.ctx, `
		INSERT INTO server_setting (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO NOTHING`, upstreamDNSKey, fallback)
	if err != nil {
		return "", err
	}

	var value string
	err = r.db.QueryRowContext(r.ctx,
		"SELECT value FROM server_setting WHERE key = ?", upstreamDNSKey,
	).Scan(&value)
	return value, err
}

func (r *settingsRepo) UpdateUpstreamDNS(value string) error {
	_, err := r.db.ExecContext(r.ctx, `
		INSERT INTO server_setting (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`, upstreamDNSKey, value)
	return err
}
