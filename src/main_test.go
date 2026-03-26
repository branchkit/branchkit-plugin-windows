package main

import (
	"testing"
)

func TestOnActionPassesNonWindowsAction(t *testing.T) {
	req := &OnActionRequest{Action: "wm tile"}
	resp, err := handleOnAction(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := resp.(OnActionResponse)
	if r.Result != "pass" {
		t.Errorf("expected 'pass' for non-windows action, got %q", r.Result)
	}
}

func TestOnActionPassesEmpty(t *testing.T) {
	req := &OnActionRequest{Action: "windows"}
	resp, err := handleOnAction(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := resp.(OnActionResponse)
	if r.Result != "pass" {
		t.Errorf("expected 'pass' for bare 'windows' action, got %q", r.Result)
	}
}

func TestOnActionPassesUnknownCommand(t *testing.T) {
	req := &OnActionRequest{Action: "windows bogus-cmd"}
	resp, err := handleOnAction(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := resp.(OnActionResponse)
	if r.Result != "pass" {
		t.Errorf("expected 'pass' for unknown command, got %q", r.Result)
	}
}
