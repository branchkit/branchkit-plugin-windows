package main

import (
	"bytes"
	"context"
	"strings"

	"github.com/a-h/templ"
	"github.com/branchkit/plugin-sdk-go"
)

// matchesSearch reports whether any field contains the search string
// (case-insensitive). An empty search matches everything.
func matchesSearch(search string, fields ...string) bool {
	if search == "" {
		return true
	}
	lower := strings.ToLower(search)
	for _, f := range fields {
		if strings.Contains(strings.ToLower(f), lower) {
			return true
		}
	}
	return false
}

// renderTempl renders a templ component to an HTML string.
func renderTempl(c templ.Component) string {
	var buf bytes.Buffer
	if err := c.Render(context.Background(), &buf); err != nil {
		branchkit.Logf("windows", "templ render error: %v", err)
		return ""
	}
	return buf.String()
}

type commandRow struct {
	Phrase      string
	Description string
}

var staticCommands = []commandRow{
	{"snap left", "Snap window to left half"},
	{"snap right", "Snap window to right half"},
	{"snap maximize / full", "Maximize window"},
	{"snap center", "Center window at half size"},
	{"snap next", "Move to next monitor"},
	{"snap prev", "Move to previous monitor"},
	{"send/throw space N", "Move window to space N"},
	{"send/throw tab space N", "Move browser tab to space N"},
	{"send/throw tab window N", "Move browser tab to window N"},
	{"space/desktop 1-9", "Switch to space (shortcut)"},
	{"mission / overview", "Mission Control"},
	{"next window", "Cycle windows (Cmd+`)"},
}

func renderSettings(search string) string {
	cmds := staticCommands
	if search != "" {
		var filtered []commandRow
		for _, c := range cmds {
			if matchesSearch(search, c.Phrase, c.Description) {
				filtered = append(filtered, c)
			}
		}
		cmds = filtered
	}

	return renderTempl(WindowsSettings(cmds))
}

func handleRenderSettingsRPC(req *branchkit.RenderSettingsRequest) (any, error) {
	html := renderSettings(req.Search)
	return branchkit.RenderSettingsResponse{HTML: html}, nil
}
