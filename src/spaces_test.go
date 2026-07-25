package main

import "testing"

func TestApplescriptAppName(t *testing.T) {
	tests := []struct {
		bundleID string
		want     string
	}{
		{"com.google.Chrome", "Google Chrome"},
		{"com.google.Chrome.canary", "Google Chrome Canary"},
		{"org.mozilla.firefox", "Firefox"},
		{"company.thebrowser.Browser", "Arc"},
		{"com.apple.Safari", "Safari"},
		{"com.unknown.thing.MyApp", "MyApp"},
	}
	for _, tt := range tests {
		got := applescriptAppName(tt.bundleID)
		if got != tt.want {
			t.Errorf("applescriptAppName(%q) = %q, want %q", tt.bundleID, got, tt.want)
		}
	}
}

