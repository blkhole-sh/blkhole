package repos

import (
	"context"
	"database/sql"
	"time"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/lemon3studio/blkhole/internal/model"
)

// QueryLogRepo defines the interface for query log repository operations
type QueryLogRepo interface {
	Insert(deviceHash, domain string, blocked bool) error
	FindByUser(userID int, limit int) ([]*model.QueryLog, error)
	DeleteOlderThan(days int) error
}

type queryLogRepo struct {
	db  *sql.DB
	ctx context.Context
}

// NewQueryLogRepo creates a new QueryLogRepo instance
func NewQueryLogRepo(db *sql.DB) QueryLogRepo {
	return &queryLogRepo{db: db, ctx: context.Background()}
}

func (r *queryLogRepo) Insert(deviceHash, domain string, blocked bool) error {
	blockedInt := 0
	if blocked {
		blockedInt = 1
	}
	_, err := r.db.ExecContext(r.ctx,
		"INSERT INTO query_log (device_hash, domain, blocked, timestamp) VALUES (?, ?, ?, ?)",
		deviceHash, domain, blockedInt, time.Now().Unix(),
	)
	return err
}

func (r *queryLogRepo) FindByUser(userID int, limit int) ([]*model.QueryLog, error) {
	query := `
		SELECT ql.id, ql.device_hash, ql.domain, ql.blocked, ql.timestamp
		FROM query_log ql
		JOIN device d ON d.hash = ql.device_hash
		WHERE d.user_id = ?
		ORDER BY ql.timestamp DESC
		LIMIT ?`
	var logs []*model.QueryLog
	if err := sqlscan.Select(r.ctx, r.db, &logs, query, userID, limit); err != nil {
		return nil, err
	}
	if logs == nil {
		return []*model.QueryLog{}, nil
	}
	return logs, nil
}

func (r *queryLogRepo) DeleteOlderThan(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days).Unix()
	_, err := r.db.ExecContext(r.ctx, "DELETE FROM query_log WHERE timestamp < ?", cutoff)
	return err
}
