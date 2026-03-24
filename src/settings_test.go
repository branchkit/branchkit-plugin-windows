package main

import "testing"

func TestContainsFold(t *testing.T) {
	tests := []struct {
		s, substr string
		want      bool
	}{
		{"Snap window to left", "snap", true},
		{"Snap window to left", "SNAP", true},
		{"Snap window to left", "window", true},
		{"Snap window to left", "right", false},
		{"", "anything", false},
		{"anything", "", true},
	}
	for _, tt := range tests {
		got := containsFold(tt.s, tt.substr)
		if got != tt.want {
			t.Errorf("containsFold(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
		}
	}
}

func TestRenderSettingsNoFilter(t *testing.T) {
	html := renderSettings("")
	if html == "" {
		t.Fatal("expected non-empty HTML")
	}
	// Should contain some known commands
	if !containsFold(html, "snap left") {
		t.Error("expected 'snap left' in output")
	}
	if !containsFold(html, "Mission Control") {
		t.Error("expected 'Mission Control' in output")
	}
}

func TestRenderSettingsWithFilter(t *testing.T) {
	html := renderSettings("tab")
	if html == "" {
		t.Fatal("expected non-empty HTML")
	}
	// Should contain tab commands
	if !containsFold(html, "tab") {
		t.Error("expected 'tab' in filtered output")
	}
	// Should NOT contain snap left (doesn't match "tab")
	if containsFold(html, "snap left") {
		t.Error("'snap left' should be filtered out by 'tab' search")
	}
}

func TestRenderSettingsNoMatch(t *testing.T) {
	html := renderSettings("zzzznonexistent")
	// Should still render (empty table), not crash
	if html == "" {
		t.Fatal("expected non-empty HTML even with no matches")
	}
}
