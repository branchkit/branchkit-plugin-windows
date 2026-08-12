package main

import (
	"strconv"

	"github.com/branchkit/plugin-sdk-go"
)

// --- Per-action handlers ---
//
// Param structs (SnapParams, MoveToSpaceParams, …) AND their registrars
// (HandleSnap, HandleMoveToSpace, …) live in actions_gen.go, generated from
// plugin.json's action_types block — so a handler's params are typed without
// unmarshaling by hand, and the action string is never spelled here at all.
// Register in main.go with the generated Handle<Action>, not the untyped
// plugin.HandleAction. Slot captures from voice
// commands (`<number>`, `<text>`) substitute into action params via the
// matching engine's template syntax (`"space": "{number}"`) which always produces
// strings — so integer-valued params stay typed as strings in the manifest
// and are parsed inside the handler with strconv.

func handleDeskSwitch(p DeskSwitchParams, _ *shared.OnActionRequest) (any, error) {
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

func handleWindowsSnap(p SnapParams, req *shared.OnActionRequest) (any, error) {
	if p.Position == nil {
		return nil, nil
	}
	handleSnap(req.ActiveWindowID, string(*p.Position))
	return nil, nil
}

func handleWindowsMoveToSpace(p MoveToSpaceParams, req *shared.OnActionRequest) (any, error) {
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
