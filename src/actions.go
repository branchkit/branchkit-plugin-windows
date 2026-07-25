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
	// Explicit window_id wins over the envelope's active window — a
	// dispatching plugin (browser tab-to-desk) targets a window it just
	// created, which may not have claimed focus by the time this lands.
	windowID := req.ActiveWindowID
	if p.WindowID != nil && *p.WindowID != "" {
		windowID = p.WindowID
	}
	handleMoveToSpace(windowID, space, p.Stay != nil && *p.Stay)
	return nil, nil
}
