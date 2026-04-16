package main

import (
	"github.com/branchkit/plugin-sdk-go"
)

var plugin *shared.Plugin

// --- Per-action handlers ---

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
	// JSON number, but template substitution can produce a string ("{0}" → "5").
	// Accept both with json.Number on the receiving side.
	Space int `json:"space"`
}

func handleWindowsMoveToSpace(req *shared.OnActionRequest) (any, error) {
	var p spaceParams
	if err := req.UnmarshalParams(&p); err != nil {
		// Fallback: template substitution can leave space as a string
		var s struct{ Space string `json:"space"` }
		if err2 := req.UnmarshalParams(&s); err2 == nil {
			p.Space = atoiOr(s.Space, 0)
		} else {
			return nil, err
		}
	}
	if p.Space < 1 || p.Space > 9 {
		shared.Logf("windows", "move_to_space: invalid space: %d", p.Space)
		return nil, nil
	}
	handleMoveToSpace(req.ActiveWindowID, p.Space)
	return nil, nil
}

func handleWindowsTabToSpace(req *shared.OnActionRequest) (any, error) {
	var p spaceParams
	if err := req.UnmarshalParams(&p); err != nil {
		var s struct{ Space string `json:"space"` }
		if err2 := req.UnmarshalParams(&s); err2 == nil {
			p.Space = atoiOr(s.Space, 0)
		} else {
			return nil, err
		}
	}
	if p.Space < 1 || p.Space > 9 {
		shared.Logf("windows", "tab_to_space: invalid space: %d", p.Space)
		return nil, nil
	}
	appID := ""
	if req.ActiveApp != nil {
		appID = *req.ActiveApp
	}
	handleMoveTabToSpace(p.Space, appID)
	return nil, nil
}

type tabToWindowParams struct {
	Index int `json:"index"`
}

func handleWindowsTabToWindow(req *shared.OnActionRequest) (any, error) {
	var p tabToWindowParams
	if err := req.UnmarshalParams(&p); err != nil {
		var s struct{ Index string `json:"index"` }
		if err2 := req.UnmarshalParams(&s); err2 == nil {
			p.Index = atoiOr(s.Index, 0)
		} else {
			return nil, err
		}
	}
	if p.Index < 1 {
		shared.Logf("windows", "tab_to_window: invalid index: %d", p.Index)
		return nil, nil
	}
	appID := ""
	if req.ActiveApp != nil {
		appID = *req.ActiveApp
	}
	handleMoveTabToWindow(p.Index, appID)
	return nil, nil
}

func atoiOr(s string, def int) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	if s == "" {
		return def
	}
	return n
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
