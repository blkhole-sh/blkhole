package services

import (
	"context"
	"log"
	"time"

	"github.com/blkhole-sh/blkhole/internal/repos"
)

type queryLogEntry struct {
	deviceHash string
	domain     string
	blocked    bool
}

// QueryLogBuffer accepts DNS query log entries and flushes them to the repo in batches.
type QueryLogBuffer struct {
	ch   chan queryLogEntry
	repo repos.QueryLogRepo
}

// NewQueryLogBuffer creates a QueryLogBuffer backed by the given repo.
func NewQueryLogBuffer(repo repos.QueryLogRepo) *QueryLogBuffer {
	return &QueryLogBuffer{
		ch:   make(chan queryLogEntry, 1000),
		repo: repo,
	}
}

// Enqueue adds an entry to the buffer. It is non-blocking: if the channel is full, the entry is dropped.
func (b *QueryLogBuffer) Enqueue(deviceHash, domain string, blocked bool) {
	select {
	case b.ch <- queryLogEntry{deviceHash, domain, blocked}:
	default:
	}
}

// Start launches the background flush goroutine. It flushes when 50 entries accumulate or every 2 seconds.
func (b *QueryLogBuffer) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		batch := make([]queryLogEntry, 0, 50)

		flush := func() {
			for _, e := range batch {
				if err := b.repo.Insert(e.deviceHash, e.domain, e.blocked); err != nil {
					log.Printf("querylog: failed to insert entry: %v", err)
				}
			}
			batch = batch[:0]
		}

		for {
			select {
			case <-ctx.Done():
				for len(b.ch) > 0 {
					batch = append(batch, <-b.ch)
				}
				flush()
				return
			case e := <-b.ch:
				batch = append(batch, e)
				if len(batch) >= 50 {
					flush()
				}
			case <-ticker.C:
				if len(batch) > 0 {
					flush()
				}
			}
		}
	}()
}
