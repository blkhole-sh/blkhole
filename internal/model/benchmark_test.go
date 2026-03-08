package model

import (
	"testing"
)

func BenchmarkListToDTO_Baseline(b *testing.B) {
	// Create a list with many rules
	count := 100000
	rules := make([]Rule, count)
	for i := 0; i < count; i++ {
		rules[i] = Rule{ID: i, DomainID: i, Allowed: true}
	}

	l := List{
		ID:          1,
		Name:        "Benchmark List",
		Description: "A list for benchmarking",
		Source:      "http://example.com",
		UserID:      1,
		Rules:       rules,
		ScheduleIDs: []int{1, 2, 3},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = l.ToDTO()
	}
}

func BenchmarkListToDTO_Optimized(b *testing.B) {
	// Create a list with RuleCount but empty Rules
	count := 100000

	l := List{
		ID:          1,
		Name:        "Benchmark List",
		Description: "A list for benchmarking",
		Source:      "http://example.com",
		UserID:      1,
		Rules:       nil,
		RuleCount:   count,
		ScheduleIDs: []int{1, 2, 3},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = l.ToDTO()
	}
}
