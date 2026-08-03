package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/branchkit/plugin-sdk-go"
)

// Timing constants for Mission Control space transitions.
// These values account for macOS animation latency between operations.
const (
	cursorSettleDelay  = 50 * time.Millisecond  // wait for cursor warp to register
	mouseDownHoldDelay = 25 * time.Millisecond  // hold before space switch keystroke
	spaceTransitDelay  = 200 * time.Millisecond // wait for Mission Control space animation
)

// switchToDesktop switches to a Mission Control desktop by number via the
// actuator, which resolves the user's own "Switch to Desktop N" symbolic
// hotkey (respecting remaps and auto-enabling disabled shortcuts) instead of
// assuming Ctrl+N. Desktops 1-16.
func switchToDesktop(desktop int) {
	shared.Logf("windows", "switch_space → desktop %d", desktop)
	if err := plugin.Call("native.switch_space", map[string]any{"space_id": desktop}, nil); err != nil {
		shared.Logf("windows", "switch to desktop %d: %v", desktop, err)
	}
}

// cursorPosition returns the current cursor location, or ok=false.
func cursorPosition() (x, y int, ok bool) {
	var info shared.NativeCursorInfoResponse
	if err := plugin.Call("native.cursor_info", nil, &info); err != nil {
		return 0, 0, false
	}
	return info.X, info.Y, true
}

// originDesktopOrdinal returns the Mission Control desktop number (the Ctrl+N
// index, counted across displays in managed-display order, matching the
// desk_switch convention) the user is looking at on the display containing
// the given point — the desk to hop back to after a stay move.
//
// Derived from the display's ACTIVE space, deliberately NOT from
// window↔space membership: the window being moved is by construction on the
// user's current space (the drag grab requires it under the cursor), and a
// freshly created window (the browser plugin's tab-to-desk pop path) can lag
// CGS membership queries — the old membership-based derivation hopped the
// user to the wrong desk (2026-07-25). Falls back to the first active user
// space when no display contains the point. Returns 0 only when spaces can't
// be listed.
func originDesktopOrdinal(displays []shared.DisplayInfo, pointX, pointY int) int {
	var spaces shared.NativeListSpacesResponse
	if err := plugin.Call("native.list_spaces", nil, &spaces); err != nil {
		shared.Logf("windows", "move-to-space: list spaces: %v", err)
		return 0
	}
	displayID := 0
	for _, d := range displays {
		if pointX >= d.X && pointX < d.X+d.W && pointY >= d.Y && pointY < d.Y+d.H {
			displayID = d.ID
			break
		}
	}
	ordinal, firstActive, matched := 0, 0, 0
	var order []string
	for _, s := range spaces.Spaces {
		if s.SpaceType != "user" {
			continue
		}
		ordinal++
		order = append(order, fmt.Sprintf("%d:d%d:%v", s.SpaceID, s.DisplayID, s.IsActive))
		if !s.IsActive {
			continue
		}
		if firstActive == 0 {
			firstActive = ordinal
		}
		if matched == 0 && s.DisplayID == displayID {
			matched = ordinal
		}
	}
	result := matched
	if result == 0 {
		result = firstActive
	}
	shared.Logf("windows", "move-to-space: origin desk=%d (window display=%d, spaces=%s)",
		result, displayID, strings.Join(order, " "))
	return result
}

// handleMoveToSpace moves the active window to the given Mission Control space
// by holding the title bar with the mouse, pressing Ctrl+N, then releasing —
// which inherently navigates to the target space with the window. With `stay`,
// it hops back to the origin desktop after delivery (the private CGS
// move-without-switching APIs are dead on modern macOS — verified silent no-op
// on Sequoia 2026-07-25 — so a visible round trip is the only non-SIP path).
func handleMoveToSpace(activeWindowID *string, space int, stay bool) {
	if space < 1 || space > 16 {
		shared.Logf("windows", "move-to-space: invalid space %d", space)
		return
	}

	var wm shared.WorldModel
	if err := plugin.Call("native.world_model", nil, &wm); err != nil {
		shared.Logf("windows", "move-to-space: get world model: %v", err)
		return
	}

	winID := ""
	if activeWindowID != nil {
		winID = *activeWindowID
	} else if wm.ActiveWindowID != nil {
		winID = *wm.ActiveWindowID
	}

	var winX, winY, winW, winH int
	found := false
	for _, w := range wm.Windows {
		if w.ID == winID {
			winX = w.X
			winY = w.Y
			winW = w.W
			winH = w.H
			found = true
			break
		}
	}

	// Fallback: AppleScript to find frontmost window position
	if !found {
		var result shared.NativeRunApplescriptResponse
		err := plugin.Call("native.run_applescript", map[string]string{
			"script": `tell application "System Events" to tell (first process whose frontmost is true) to get position of window 1`,
		}, &result)
		if err == nil && result.ExitCode == 0 {
			parts := strings.Split(result.Stdout, ",")
			if len(parts) == 2 {
				x, e1 := strconv.Atoi(strings.TrimSpace(parts[0]))
				y, e2 := strconv.Atoi(strings.TrimSpace(parts[1]))
				if e1 == nil && e2 == nil {
					winX = x
					winY = y
					found = true
				}
			}
		}
	}

	if !found {
		shared.Logf("windows", "move-to-space: could not find window position")
		return
	}

	// Resolve the return desktop BEFORE the move — afterwards the window (and
	// we, riding along with it) are on the target space.
	returnOrdinal := 0
	if stay {
		returnOrdinal = originDesktopOrdinal(wm.Displays, winX+winW/2, winY+winH/2)
		if returnOrdinal == 0 {
			shared.Logf("windows", "move-to-space: stay requested but origin desktop unknown — will follow instead")
		} else if returnOrdinal == space {
			returnOrdinal = 0 // already there; nothing to hop back to
		}
	}

	// Remember where the cursor was so the whole operation doesn't strand it
	// on the moved window's title bar.
	origCursorX, origCursorY, restoreCursor := cursorPosition()

	// Click title bar area, hold, switch space, release
	clickX := winX + 75
	clickY := winY + 10

	// Warp cursor to title bar
	warpReq := struct {
		X int `json:"x"`
		Y int `json:"y"`
	}{X: clickX, Y: clickY}
	if err := plugin.Call("native.warp_cursor", warpReq, nil); err != nil {
		shared.Logf("windows", "move-to-space: warp cursor: %v", err)
		return
	}
	time.Sleep(cursorSettleDelay)

	// Mouse down, then a zero-distance drag to latch the window onto the
	// cursor — macOS only treats the window as grabbed once a dragged event
	// follows the press (the Amethyst/Silica latch).
	mouseButton("press")
	mouseButton("drag")
	time.Sleep(mouseDownHoldDelay)

	// Switch to the target desktop from under the held window (symbolic
	// hotkey — respects the user's actual shortcut config)
	switchToDesktop(space)

	time.Sleep(spaceTransitDelay)

	// Mouse up
	mouseButton("release")

	// Stay variant: hop back to the origin desktop once the drop has landed.
	if returnOrdinal != 0 {
		time.Sleep(spaceTransitDelay)
		switchToDesktop(returnOrdinal)
	}

	if restoreCursor {
		restoreReq := struct {
			X int `json:"x"`
			Y int `json:"y"`
		}{X: origCursorX, Y: origCursorY}
		if err := plugin.Call("native.warp_cursor", restoreReq, nil); err != nil {
			shared.Logf("windows", "cursor restore: %v", err)
		}
	}
}

// mouseButton presses, releases, or drag-latches the left mouse button via
// the input.mouse_button RPC. (The old raw `dispatch` route is denied to
// plugin callers by the operation auth layer — the grab half of the drag
// trick had been failing silently through it.)
func mouseButton(direction string) {
	if err := plugin.Call("input.mouse_button", map[string]any{"button": "left", "direction": direction}, nil); err != nil {
		shared.Logf("windows", "mouse_button %s: %v", direction, err)
	}
}
