package repos

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/blkhole-sh/blkhole/internal/model"
	"github.com/georgysavva/scany/v2/sqlscan"
)

// QueryLogRepo defines the interface for query log repository operations
type QueryLogRepo interface {
	Insert(deviceHash, domain string, blocked bool) error
	FindByUser(userID int, limit int) ([]*model.QueryLog, error)
	FindFilteredByUser(userID int, filter QueryLogFilter) ([]*model.QueryLogDTO, int, error)
	DeleteOlderThan(days int) error
	GetDomainStats(deviceHashes []string, from, to time.Time, blockedOnly bool, limit int) ([]model.DomainStat, error)
	GetHourlyActivity(deviceHashes []string, from, to time.Time) (map[string][]int, error)
	GetLastQueries(deviceHashes []string) (map[string]time.Time, error)
	// GetAggregatedStats returns total and blocked counts per time bucket for the
	// given devices. Buckets are aligned to stepSeconds boundaries (Unix epoch).
	GetAggregatedStats(deviceHashes []string, from, to time.Time, stepSeconds int64) (total map[time.Time]int, blocked map[time.Time]int, err error)
}

// QueryLogFilter scopes and paginates dashboard query log reads.
type QueryLogFilter struct {
	DeviceIDs []int
	From      time.Time
	To        time.Time
	Limit     int
	Offset    int
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

func (r *queryLogRepo) FindFilteredByUser(userID int, filter QueryLogFilter) ([]*model.QueryLogDTO, int, error) {
	where := []string{"d.user_id = ?"}
	args := []any{userID}
	if len(filter.DeviceIDs) > 0 {
		where = append(where, "d.id IN ("+strings.TrimSuffix(strings.Repeat("?,", len(filter.DeviceIDs)), ",")+")")
		for _, id := range filter.DeviceIDs {
			args = append(args, id)
		}
	}
	if !filter.From.IsZero() {
		where = append(where, "ql.timestamp >= ?")
		args = append(args, filter.From.Unix())
	}
	if !filter.To.IsZero() {
		where = append(where, "ql.timestamp < ?")
		args = append(args, filter.To.Unix())
	}

	clause := strings.Join(where, " AND ")
	var total int
	if err := r.db.QueryRowContext(r.ctx, `SELECT COUNT(*) FROM query_log ql JOIN device d ON d.hash = ql.device_hash WHERE `+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}
	query := `
		SELECT ql.id, d.id AS device_id, d.name AS device_name, ql.domain, ql.blocked, ql.timestamp
		FROM query_log ql
		JOIN device d ON d.hash = ql.device_hash
		WHERE ` + clause + `
		ORDER BY ql.timestamp DESC, ql.id DESC
		LIMIT ? OFFSET ?`
	queryArgs := append(append([]any{}, args...), limit, filter.Offset)
	var logs []*model.QueryLogDTO
	if err := sqlscan.Select(r.ctx, r.db, &logs, query, queryArgs...); err != nil {
		return nil, 0, err
	}
	if logs == nil {
		logs = []*model.QueryLogDTO{}
	}
	return logs, total, nil
}

func (r *queryLogRepo) DeleteOlderThan(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days).Unix()
	_, err := r.db.ExecContext(r.ctx, "DELETE FROM query_log WHERE timestamp < ?", cutoff)
	return err
}

func (r *queryLogRepo) GetDomainStats(deviceHashes []string, from, to time.Time, blockedOnly bool, limit int) ([]model.DomainStat, error) {
	if len(deviceHashes) == 0 {
		return []model.DomainStat{}, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(deviceHashes)), ",")
	query := `SELECT domain, COUNT(*) AS count, SUM(blocked) AS blocked
		FROM query_log
		WHERE device_hash IN (` + placeholders + `) AND timestamp >= ? AND timestamp < ?`
	if blockedOnly {
		query += " AND blocked = 1"
	}
	query += " GROUP BY domain ORDER BY count DESC, domain LIMIT ?"
	args := make([]any, 0, len(deviceHashes)+3)
	for _, hash := range deviceHashes {
		args = append(args, hash)
	}
	args = append(args, from.Unix(), to.Unix(), limit)
	var stats []model.DomainStat
	if err := sqlscan.Select(r.ctx, r.db, &stats, query, args...); err != nil {
		return nil, err
	}
	if stats == nil {
		stats = []model.DomainStat{}
	}
	return stats, nil
}

func (r *queryLogRepo) GetHourlyActivity(deviceHashes []string, from, to time.Time) (map[string][]int, error) {
	activity := make(map[string][]int, len(deviceHashes))
	for _, hash := range deviceHashes {
		activity[hash] = make([]int, 24)
	}
	if len(deviceHashes) == 0 {
		return activity, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(deviceHashes)), ",")
	query := `SELECT device_hash, CAST(strftime('%H', timestamp, 'unixepoch', 'localtime') AS INTEGER) AS hour, COUNT(*) AS count
		FROM query_log
		WHERE device_hash IN (` + placeholders + `) AND timestamp >= ? AND timestamp < ?
		GROUP BY device_hash, hour`
	args := make([]any, 0, len(deviceHashes)+2)
	for _, hash := range deviceHashes {
		args = append(args, hash)
	}
	args = append(args, from.Unix(), to.Unix())
	rows, err := r.db.QueryContext(r.ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var hash string
		var hour, count int
		if err := rows.Scan(&hash, &hour, &count); err != nil {
			return nil, err
		}
		if hour >= 0 && hour < 24 {
			activity[hash][hour] = count
		}
	}
	return activity, rows.Err()
}

func (r *queryLogRepo) GetLastQueries(deviceHashes []string) (map[string]time.Time, error) {
	lastSeen := make(map[string]time.Time, len(deviceHashes))
	if len(deviceHashes) == 0 {
		return lastSeen, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(deviceHashes)), ",")
	args := make([]any, 0, len(deviceHashes))
	for _, hash := range deviceHashes {
		args = append(args, hash)
	}
	rows, err := r.db.QueryContext(r.ctx, `SELECT device_hash, MAX(timestamp) FROM query_log WHERE device_hash IN (`+placeholders+`) GROUP BY device_hash`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var hash string
		var timestamp int64
		if err := rows.Scan(&hash, &timestamp); err != nil {
			return nil, err
		}
		lastSeen[hash] = time.Unix(timestamp, 0)
	}
	return lastSeen, rows.Err()
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
