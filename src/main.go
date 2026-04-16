package main

import (
	"strconv"

	"github.com/branchkit/plugin-sdk-go"
)

var plugin *shared.Plugin

// --- Per-action handlers ---
//
// Slot captures from voice commands (`<number>`, `<text>`) substitute into
// action params via the matching engine's template syntax (`"space": "{0}"`),
// which always produces strings. So integer-valued params are declared as
// strings and parsed inside the handler with strconv.

type snapParams struct {
	Position string `json:"position"`
}

func handleWindowsSnap(req *shared.OnActionRequest) (any, error) {
	var p snapParams
	if err := req.UnmarshalParams(&p); err != nil {
		return nil, err
	}
	if p.Position == "" {
		shared.Logf("windows", "snap: missing position")
		return nil, nil
	}
	handleSnap(req.ActiveWindowID, p.Position)
	return nil, nil
}

type spaceParams struct {
	Space string `json:"space"`
}

func handleWindowsMoveToSpace(req *shared.OnActionRequest) (any, error) {
	var p spaceParams
	if err := req.UnmarshalParams(&p); err != nil {
		return nil, err
	}
	space, err := strconv.Atoi(p.Space)
	if err != nil || space < 1 || space > 9 {
		shared.Logf("windows", "move_to_space: invalid space: %q", p.Space)
		return nil, nil
	}
	handleMoveToSpace(req.ActiveWindowID, space)
	return nil, nil
}

func handleWindowsTabToSpace(req *shared.OnActionRequest) (any, error) {
	var p spaceParams
	if err := req.UnmarshalParams(&p); err != nil {
		return nil, err
	}
	space, err := strconv.Atoi(p.Space)
	if err != nil || space < 1 || space > 9 {
		shared.Logf("windows", "tab_to_space: invalid space: %q", p.Space)
		return nil, nil
	}
	appID := ""
	if req.ActiveApp != nil {
		appID = *req.ActiveApp
	}
	handleMoveTabToSpace(space, appID)
	return nil, nil
}

type tabToWindowParams struct {
	Index string `json:"index"`
}

func handleWindowsTabToWindow(req *shared.OnActionRequest) (any, error) {
	var p tabToWindowParams
	if err := req.UnmarshalParams(&p); err != nil {
		return nil, err
	}
	index, err := strconv.Atoi(p.Index)
	if err != nil || index < 1 {
		shared.Logf("windows", "tab_to_window: invalid index: %q", p.Index)
		return nil, nil
	}
	appID := ""
	if req.ActiveApp != nil {
		appID = *req.ActiveApp
	}
	handleMoveTabToWindow(index, appID)
	return nil, nil
}

func pushCommands(p *shared.Plugin) {
	count, err := shared.PushCommands(p)
	if err != nil {
		shared.Logf("windows", "%v", err)
		return
	}
	shared.Logf("windows", "Registered %d command variants", count)
}

func main() {
	plugin = shared.NewPlugin()
	pushCommands(plugin)

	plugin.HandleAction("windows.snap", handleWindowsSnap)
	plugin.HandleAction("windows.move_to_space", handleWindowsMoveToSpace)
	plugin.HandleAction("windows.tab_to_space", handleWindowsTabToSpace)
	plugin.HandleAction("windows.tab_to_window", handleWindowsTabToWindow)
	shared.HandleTyped(plugin, "render_settings", handleRenderSettingsRPC)

	plugin.Run()
}
