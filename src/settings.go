package main

import (
	"bytes"
	_ "embed"
	"html/template"

	"branchkit.local/shared"
)

//go:embed templates/settings.html
var settingsHTML string

var settingsTmpl = template.Must(template.New("windows-settings").Parse(settingsHTML))

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
			if containsFold(c.Phrase, search) || containsFold(c.Description, search) {
				filtered = append(filtered, c)
			}
		}
		cmds = filtered
	}

	data := struct {
		Commands []commandRow
	}{Commands: cmds}

	var buf bytes.Buffer
	if err := settingsTmpl.Execute(&buf, data); err != nil {
		log("settings template error: %v", err)
		return ""
	}
	return buf.String()
}

type RenderSettingsRequest struct {
	TabKey string `json:"tab_key"`
	Search string `json:"search"`
}

func handleRenderSettingsRPC(req *RenderSettingsRequest) (any, error) {
	html := renderSettings(req.Search)
	return shared.SettingsResponse{HTML: html}, nil
}

// containsFold does a case-insensitive substring match.
func containsFold(s, substr string) bool {
	return len(s) >= len(substr) && func() bool {
		sl := bytes.ToLower([]byte(s))
		sub := bytes.ToLower([]byte(substr))
		return bytes.Contains(sl, sub)
	}()
}
