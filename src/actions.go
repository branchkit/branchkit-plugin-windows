package main

import (
	"strconv"

	"github.com/branchkit/plugin-sdk-go"
)

// --- Per-action handlers ---
//
// Param structs (SnapParams, MoveToSpaceParams, …) live in actions_gen.go,
// generated from plugin.json's action_types block. Slot captures from voice
// commands (`<number>`, `<text>`) substitute into action params via the
// matching engine's template syntax (`"space": "{number}"`) which always produces
// strings — so integer-valued params stay typed as strings in the manifest
// and are parsed inside the handler with strconv.

func handleDeskSwitch(req *shared.OnActionRequest) (any, error) {
	var p struct {
		Space string `json:"space"`
	}
	if err := req.UnmarshalParams(&p); err != nil {
		return nil, err
	}
	space, err := strconv.Atoi(p.Space)
	if err != nil || space < 1 || space > 16 {
		shared.Logf("windows", "desk_switch: invalid space: %q", p.Space)
		return nil, nil
	}
	// The actuator resolves the user's actual "Switch to Desktop N" symbolic
	// hotkey (respects remaps, auto-enables disabled shortcuts) — no
	// hardcoded Ctrl+N keycode map.
	switchToDesktop(space)
	return nil, nil
}

func handleWindowsSnap(req *shared.OnActionRequest) (any, error) {
	var p SnapParams
	if err := req.UnmarshalParams(&p); err != nil {
		return nil, err
	}
	if p.Position == nil {
		return nil, nil
	}
	handleSnap(req.ActiveWindowID, string(*p.Position))
	return nil, nil
}

func handleWindowsMoveToSpace(req *shared.OnActionRequest) (any, error) {
	var p MoveToSpaceParams
	if err := req.UnmarshalParams(&p); err != nil {
		return nil, err
	}
	space, err := strconv.Atoi(p.Space)
	if err != nil || space < 1 || space > 9 {
		shared.Logf("windows", "move_to_space: invalid space: %q", p.Space)
		return nil, nil
	}
	handleMoveToSpace(req.ActiveWindowID, space, p.Stay != nil && *p.Stay)
	return nil, nil
}

func handleWindowsTabToSpace(req *shared.OnActionRequest) (any, error) {
	var p TabToSpaceParams
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

func handleWindowsTabToWindow(req *shared.OnActionRequest) (any, error) {
	var p TabToWindowParams
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
