package main

import (
	"runtime/debug"
	"testing"
)

func TestRevisionFromBuildInfoIgnoresModuleVersion(t *testing.T) {
	info := &debug.BuildInfo{
		Main: debug.Module{Version: "v0.0.0-20230613-abc1234f"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "abcdef1234567890"},
		},
	}

	if got := revisionFromBuildInfo(info); got != "abcdef1" {
		t.Errorf("revisionFromBuildInfo() = %q; want %q", got, "abcdef1")
	}
}

func TestRevisionFromSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings []debug.BuildSetting
		want     string
	}{
		{
			name: "shortens full vcs revision",
			settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "abcdef1234567890"},
			},
			want: "abcdef1",
		},
		{
			name: "keeps short vcs revision",
			settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "abc1234"},
			},
			want: "abc1234",
		},
		{
			name: "empty without vcs revision",
			settings: []debug.BuildSetting{
				{Key: "other", Value: "v0.0.0-20230613-abc1234f"},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := revisionFromSettings(tt.settings); got != tt.want {
				t.Errorf("revisionFromSettings() = %q; want %q", got, tt.want)
			}
		})
	}
}
