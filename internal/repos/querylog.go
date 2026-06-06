package repos

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/lemon3studio/blkhole/internal/model"
)

// QueryLogRepo defines the interface for query log repository operations
type QueryLogRepo interface {
	Insert(deviceHash, domain string, blocked bool) error
	FindByUser(userID int, limit int) ([]*model.QueryLog, error)
	DeleteOlderThan(days int) error
	// GetAggregatedStats returns total and blocked counts per time bucket for the
	// given devices. Buckets are aligned to stepSeconds boundaries (Unix epoch).
	GetAggregatedStats(deviceHashes []string, from, to time.Time, stepSeconds int64) (total map[time.Time]int, blocked map[time.Time]int, err error)
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

func (r *queryLogRepo) GetAggregatedStats(deviceHashes []string, from, to time.Time, stepSeconds int64) (map[time.Time]int, map[time.Time]int, error) {
	if len(deviceHashes) == 0 {
		return nil, nil, nil
	}

	ph := make([]string, len(deviceHashes))
	for i := range ph {
		ph[i] = "?"
	}
	placeholders := strings.Join(ph, ",")
	query := fmt.Sprintf(`
		SELECT (timestamp / ?) * ? AS bucket, COUNT(*) AS total, SUM(blocked) AS blocked_count
		FROM query_log
		WHERE device_hash IN (%s) AND timestamp >= ? AND timestamp < ?
		GROUP BY bucket`, placeholders)

	args := make([]any, 0, 2+len(deviceHashes)+2)
	args = append(args, stepSeconds, stepSeconds)
	for _, h := range deviceHashes {
		args = append(args, h)
	}
	args = append(args, from.Unix(), to.Unix())

	rows, err := r.db.QueryContext(r.ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	totalMap := make(map[time.Time]int)
	blockedMap := make(map[time.Time]int)
	for rows.Next() {
		var bucketUnix int64
		var total, blocked int
		if err := rows.Scan(&bucketUnix, &total, &blocked); err != nil {
			return nil, nil, err
		}
		t := time.Unix(bucketUnix, 0).UTC()
		totalMap[t] = total
		blockedMap[t] = blocked
	}
	return totalMap, blockedMap, rows.Err()
}
