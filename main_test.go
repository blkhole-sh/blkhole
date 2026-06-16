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

func TestRevisionFromBuildInfoUsesPseudoVersionHashFallback(t *testing.T) {
	info := &debug.BuildInfo{
		Main: debug.Module{Version: "v0.0.0-20230613-abc1234f"},
	}

	if got := revisionFromBuildInfo(info); got != "abc1234" {
		t.Errorf("revisionFromBuildInfo() = %q; want %q", got, "abc1234")
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

func TestRevisionFromModuleVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "shortens example pseudo-version hash",
			version: "v0.0.0-20230613-abc1234f",
			want:    "abc1234",
		},
		{
			name:    "shortens go pseudo-version hash",
			version: "v0.0.0-20230613123456-abcdef123456",
			want:    "abcdef1",
		},
		{
			name:    "shortens prerelease pseudo-version hash",
			version: "v1.2.3-pre.0.20230613123456-abcdef123456",
			want:    "abcdef1",
		},
		{
			name:    "shortens incompatible pseudo-version hash",
			version: "v2.0.1-0.20190915032832-14c0d48ead0c+incompatible",
			want:    "14c0d48",
		},
		{
			name:    "ignores release version",
			version: "v1.2.3",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := revisionFromModuleVersion(tt.version); got != tt.want {
				t.Errorf("revisionFromModuleVersion(%q) = %q; want %q", tt.version, got, tt.want)
			}
		})
	}
}
