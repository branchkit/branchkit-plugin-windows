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

func TestSpaceCodes(t *testing.T) {
	// Verify all 9 spaces have codes
	for i := 1; i <= 9; i++ {
		if _, ok := spaceCodes[i]; !ok {
			t.Errorf("missing space code for space %d", i)
		}
	}
	// Space 0 and 10 should not exist
	if _, ok := spaceCodes[0]; ok {
		t.Error("space 0 should not exist")
	}
	if _, ok := spaceCodes[10]; ok {
		t.Error("space 10 should not exist")
	}
}
