package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upScheduleIntegrity, nil)
}

func upScheduleIntegrity(ctx context.Context, tx *sql.Tx) error {
	for _, column := range []struct {
		table string
		name  string
	}{
		{table: "list", name: "is_default"},
		{table: "schedule", name: "is_default"},
	} {
		exists, err := columnExists(ctx, tx, column.table, column.name)
		if err != nil {
			return err
		}
		if !exists {
			if _, err := tx.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s INTEGER NOT NULL DEFAULT 0", column.table, column.name)); err != nil {
				return err
			}
		}
	}

	cleanup := []string{
		// Builds before foreign keys were enabled left these behind when a user was deleted.
		`DELETE FROM device WHERE NOT EXISTS (SELECT 1 FROM user WHERE user.id = device.user_id)`,
		`DELETE FROM list WHERE NOT EXISTS (SELECT 1 FROM user WHERE user.id = list.user_id)`,
		`DELETE FROM schedule WHERE NOT EXISTS (SELECT 1 FROM user WHERE user.id = schedule.user_id)`,
		`DELETE FROM device_schedule WHERE NOT EXISTS (SELECT 1 FROM device WHERE device.id = device_schedule.device_id) OR NOT EXISTS (SELECT 1 FROM schedule WHERE schedule.id = device_schedule.schedule_id)`,
		`DELETE FROM list_schedule WHERE NOT EXISTS (SELECT 1 FROM list WHERE list.id = list_schedule.list_id) OR NOT EXISTS (SELECT 1 FROM schedule WHERE schedule.id = list_schedule.schedule_id)`,
		`DELETE FROM list_rule WHERE NOT EXISTS (SELECT 1 FROM list WHERE list.id = list_rule.list_id) OR NOT EXISTS (SELECT 1 FROM rule WHERE rule.id = list_rule.rule_id)`,
		`DELETE FROM schedule_rule WHERE NOT EXISTS (SELECT 1 FROM schedule WHERE schedule.id = schedule_rule.schedule_id) OR NOT EXISTS (SELECT 1 FROM rule WHERE rule.id = schedule_rule.rule_id)`,
	}
	if err := execStatements(ctx, tx, cleanup); err != nil {
		return err
	}

	var scheduleSQL string
	if err := tx.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type='table' AND name='schedule'`).Scan(&scheduleSQL); err != nil {
		return err
	}
	if strings.Contains(scheduleSQL, "start_time < end_time") {
		if err := rebuildLegacySchedule(ctx, tx); err != nil {
			return err
		}
	}

	rows, err := tx.QueryContext(ctx, "PRAGMA foreign_key_check")
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		var table, parent string
		var rowID sql.NullInt64
		var foreignKeyID int
		if err := rows.Scan(&table, &rowID, &parent, &foreignKeyID); err != nil {
			return err
		}
		return fmt.Errorf("foreign key violation in %s row %d referencing %s", table, rowID.Int64, parent)
	}
	return rows.Err()
}

func columnExists(ctx context.Context, tx *sql.Tx, table, column string) (bool, error) {
	rows, err := tx.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, notNull, primaryKey int
		var name, columnType string
		var defaultValue sql.NullString
		if err := rows.Scan(&id, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}

func rebuildLegacySchedule(ctx context.Context, tx *sql.Tx) error {
	statements := []string{
		`CREATE TEMP TABLE migration_device_schedule AS SELECT device_id, schedule_id FROM device_schedule`,
		`CREATE TEMP TABLE migration_list_schedule AS SELECT list_id, schedule_id FROM list_schedule`,
		`CREATE TEMP TABLE migration_schedule_rule AS SELECT schedule_id, rule_id FROM schedule_rule`,
		`CREATE TABLE schedule_new (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			start_time TEXT NOT NULL,
			end_time TEXT NOT NULL,
			days INTEGER NOT NULL CHECK (days >= 0 AND days < 128),
			active INTEGER NOT NULL DEFAULT 1,
			is_default INTEGER NOT NULL DEFAULT 0,
			user_id INTEGER NOT NULL REFERENCES user (id) ON DELETE CASCADE,
			CHECK (start_time GLOB '[0-2][0-9]:[0-5][0-9]'),
			CHECK (end_time GLOB '[0-2][0-9]:[0-5][0-9]'),
			CHECK (CAST(SUBSTR(start_time, 4, 2) AS INTEGER) % 5 = 0),
			CHECK (CAST(SUBSTR(end_time, 4, 2) AS INTEGER) % 5 = 0),
			CHECK (CAST(SUBSTR(start_time, 1, 2) AS INTEGER) >= 0 AND CAST(SUBSTR(start_time, 1, 2) AS INTEGER) <= 23),
			CHECK (CAST(SUBSTR(end_time, 1, 2) AS INTEGER) >= 0 AND CAST(SUBSTR(end_time, 1, 2) AS INTEGER) <= 23)
		)`,
		`INSERT INTO schedule_new (id, name, start_time, end_time, days, active, is_default, user_id) SELECT id, name, start_time, end_time, days, active, is_default, user_id FROM schedule`,
		`DROP TABLE device_schedule`,
		`DROP TABLE list_schedule`,
		`DROP TABLE schedule_rule`,
		`DROP TABLE schedule`,
		`ALTER TABLE schedule_new RENAME TO schedule`,
		`CREATE TABLE device_schedule (
			device_id INTEGER NOT NULL REFERENCES device (id) ON DELETE CASCADE,
			schedule_id INTEGER NOT NULL REFERENCES schedule (id) ON DELETE CASCADE,
			PRIMARY KEY (device_id, schedule_id)
		)`,
		`CREATE TABLE list_schedule (
			list_id INTEGER NOT NULL REFERENCES list (id) ON DELETE CASCADE,
			schedule_id INTEGER NOT NULL REFERENCES schedule (id) ON DELETE CASCADE,
			PRIMARY KEY (list_id, schedule_id)
		)`,
		`CREATE TABLE schedule_rule (
			schedule_id INTEGER NOT NULL REFERENCES schedule (id) ON DELETE CASCADE,
			rule_id INTEGER NOT NULL REFERENCES rule (id),
			PRIMARY KEY (schedule_id, rule_id)
		)`,
		`INSERT INTO device_schedule SELECT device_id, schedule_id FROM migration_device_schedule`,
		`INSERT INTO list_schedule SELECT list_id, schedule_id FROM migration_list_schedule`,
		`INSERT INTO schedule_rule SELECT schedule_id, rule_id FROM migration_schedule_rule`,
		`DROP TABLE migration_device_schedule`,
		`DROP TABLE migration_list_schedule`,
		`DROP TABLE migration_schedule_rule`,
		`CREATE INDEX IF NOT EXISTS idx_schedule_user_id ON schedule(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_device_schedule_device_id ON device_schedule(device_id)`,
		`CREATE INDEX IF NOT EXISTS idx_list_schedule_list_id ON list_schedule(list_id)`,
		`CREATE INDEX IF NOT EXISTS idx_list_schedule_schedule_id ON list_schedule(schedule_id)`,
		`CREATE INDEX IF NOT EXISTS idx_schedule_rule_rule_id ON schedule_rule(rule_id)`,
	}
	return execStatements(ctx, tx, statements)
}

func execStatements(ctx context.Context, tx *sql.Tx, statements []string) error {
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}
