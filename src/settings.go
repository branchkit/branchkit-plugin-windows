package main

import (
	"github.com/branchkit/plugin-sdk-go"
	toolkit "github.com/branchkit/plugin-toolkit-go"
)

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
			if toolkit.MatchesSearch(search, c.Phrase, c.Description) {
				filtered = append(filtered, c)
			}
		}
		cmds = filtered
	}

	return toolkit.RenderTempl("windows", WindowsSettings(cmds))
}

func handleRenderSettingsRPC(req *branchkit.RenderSettingsRequest) (any, error) {
	html := renderSettings(req.Search)
	return branchkit.RenderSettingsResponse{HTML: html}, nil
}
