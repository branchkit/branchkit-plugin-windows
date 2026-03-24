package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOnActionPassesNonWindowsAction(t *testing.T) {
	req := OnActionRequest{Action: "wm tile"}
	body, _ := json.Marshal(req)
	r := httptest.NewRequest("POST", "/hooks/on-action", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handleOnAction(w, r)

	var resp OnActionResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Result != "pass" {
		t.Errorf("expected 'pass' for non-windows action, got %q", resp.Result)
	}
}

func TestOnActionPassesEmpty(t *testing.T) {
	req := OnActionRequest{Action: "windows"}
	body, _ := json.Marshal(req)
	r := httptest.NewRequest("POST", "/hooks/on-action", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handleOnAction(w, r)

	var resp OnActionResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Result != "pass" {
		t.Errorf("expected 'pass' for bare 'windows' action, got %q", resp.Result)
	}
}

func TestOnActionPassesUnknownCommand(t *testing.T) {
	req := OnActionRequest{Action: "windows bogus-cmd"}
	body, _ := json.Marshal(req)
	r := httptest.NewRequest("POST", "/hooks/on-action", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handleOnAction(w, r)

	var resp OnActionResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Result != "pass" {
		t.Errorf("expected 'pass' for unknown command, got %q", resp.Result)
	}
}

func TestOnActionBadJSON(t *testing.T) {
	r := httptest.NewRequest("POST", "/hooks/on-action", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()

	handleOnAction(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for bad JSON, got %d", w.Code)
	}
}

func TestHealthEndpoint(t *testing.T) {
	r := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handleHealth(w, r)

	var resp map[string]bool
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp["ready"] {
		t.Error("expected ready=true")
	}
}
