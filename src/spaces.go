package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/branchkit/plugin-sdk-go"
)

// spaceCodes maps Mission Control space numbers (1-9) to macOS virtual keycodes.
var spaceCodes = map[int]int{
	1: 18, 2: 19, 3: 20, 4: 21, 5: 23, 6: 22, 7: 26, 8: 28, 9: 25,
}

// handleMoveToSpace moves the active window to the given Mission Control space
// by holding the title bar with the mouse, pressing Ctrl+N, then releasing.
func handleMoveToSpace(activeWindowID *string, space int) {
	code, ok := spaceCodes[space]
	if !ok {
		log("move-to-space: invalid space %d", space)
		return
	}

	var wm shared.WorldModel
	if err := plugin.Call("native.world_model", nil, &wm); err != nil {
		log("move-to-space: get world model: %v", err)
		return
	}

	winID := ""
	if activeWindowID != nil {
		winID = *activeWindowID
	} else if wm.ActiveWindowID != nil {
		winID = *wm.ActiveWindowID
	}

	var winX, winY int
	found := false
	for _, w := range wm.Windows {
		if w.ID == winID {
			winX = w.X
			winY = w.Y
			found = true
			break
		}
	}

	// Fallback: AppleScript to find frontmost window position
	if !found {
		var result shared.ApplescriptResult
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
		log("move-to-space: could not find window position")
		return
	}

	// Click title bar area, hold, switch space, release
	clickX := winX + 75
	clickY := winY + 10

	// Warp cursor to title bar
	warpReq := struct {
		X int `json:"x"`
		Y int `json:"y"`
	}{X: clickX, Y: clickY}
	if err := plugin.Call("native.warp_cursor", warpReq, nil); err != nil {
		log("move-to-space: warp cursor: %v", err)
		return
	}
	time.Sleep(50 * time.Millisecond)

	// Mouse down
	executeAction(json.RawMessage(`{"type":"mouse_down","button":"left"}`))
	time.Sleep(25 * time.Millisecond)

	// Ctrl + space number
	executeAction(json.RawMessage(fmt.Sprintf(
		`{"type":"shortcut","code":%d,"modifiers":["ctrl"]}`, code)))

	time.Sleep(200 * time.Millisecond)

	// Mouse up
	executeAction(json.RawMessage(`{"type":"mouse_up","button":"left"}`))
}

// handleMoveTabToSpace moves the active browser tab to another Mission Control space.
func handleMoveTabToSpace(space int, appID string) {
	if appID == "" {
		log("tab-to-space: no app_id")
		return
	}
	code, ok := spaceCodes[space]
	if !ok {
		log("tab-to-space: invalid space %d", space)
		return
	}

	appName := applescriptAppName(appID)

	// Capture active tab ID
	captureScript := fmt.Sprintf(`tell application "%s" to get {URL, id} of active tab of window 1`, appName)
	var result shared.ApplescriptResult
	if err := plugin.Call("native.run_applescript", map[string]string{"script": captureScript}, &result); err != nil || result.ExitCode != 0 {
		log("tab-to-space: capture tab failed: %v", err)
		return
	}

	parts := strings.Split(result.Stdout, ",")
	if len(parts) < 2 {
		log("tab-to-space: unexpected capture result: %s", result.Stdout)
		return
	}
	tabID := strings.TrimSpace(parts[1])

	// Check if target space has an existing browser window
	var wm shared.WorldModel
	if err := plugin.Call("native.world_model", nil, &wm); err != nil {
		log("tab-to-space: get world model: %v", err)
		return
	}

	var existingWinID string
	targetIdx := space - 1
	if targetIdx >= 0 && targetIdx < len(wm.Displays) {
		screen := wm.Displays[targetIdx]
		for _, w := range wm.Windows {
			if w.AppID == appID && w.X >= screen.X && w.X < screen.X+screen.W {
				existingWinID = w.ID
				break
			}
		}
	}

	// Switch to target space
	executeAction(json.RawMessage(fmt.Sprintf(
		`{"type":"shortcut","code":%d,"modifiers":["ctrl"]}`, code)))

	time.Sleep(600 * time.Millisecond)

	// Deliver tab to target space
	var deliverScript string
	if existingWinID != "" {
		deliverScript = fmt.Sprintf(`tell application "%s" to move tab id %s to window id %s`, appName, tabID, existingWinID)
	} else {
		deliverScript = fmt.Sprintf(`tell application "%s"
	set newWin to make new window
	move tab id %s to newWin
	close tab 2 of newWin
end tell`, appName, tabID)
	}
	if err := plugin.Call("native.run_applescript", map[string]string{"script": deliverScript}, nil); err != nil {
		log("tab-to-space: deliver tab: %v", err)
	}
}

// handleMoveTabToWindow moves the active browser tab to another browser window by index.
func handleMoveTabToWindow(windowIndex int, appID string) {
	if appID == "" {
		log("tab-to-window: no app_id")
		return
	}

	appName := applescriptAppName(appID)

	// Capture active tab ID
	captureScript := fmt.Sprintf(`tell application "%s" to get {URL, id} of active tab of window 1`, appName)
	var result shared.ApplescriptResult
	if err := plugin.Call("native.run_applescript", map[string]string{"script": captureScript}, &result); err != nil || result.ExitCode != 0 {
		log("tab-to-window: capture tab failed: %v", err)
		return
	}

	parts := strings.Split(result.Stdout, ",")
	if len(parts) < 2 {
		log("tab-to-window: unexpected capture result: %s", result.Stdout)
		return
	}
	tabID := strings.TrimSpace(parts[1])

	// Get world model to find target window
	var wm shared.WorldModel
	if err := plugin.Call("native.world_model", nil, &wm); err != nil {
		log("tab-to-window: get world model: %v", err)
		return
	}

	// Collect browser windows
	var browserWindows []string
	for _, w := range wm.Windows {
		if w.AppID == appID {
			browserWindows = append(browserWindows, w.ID)
		}
	}

	if windowIndex < 1 || windowIndex > len(browserWindows) {
		log("tab-to-window: window index %d out of range (have %d)", windowIndex, len(browserWindows))
		return
	}

	targetWinID := browserWindows[windowIndex-1]
	deliverScript := fmt.Sprintf(`tell application "%s" to move tab id %s to window id %s`, appName, tabID, targetWinID)
	if err := plugin.Call("native.run_applescript", map[string]string{"script": deliverScript}, nil); err != nil {
		log("tab-to-window: deliver tab: %v", err)
	}
}

// executeAction calls the actuator's Execute endpoint with a raw action JSON.
func executeAction(action json.RawMessage) {
	req := shared.ExecuteActionRequest{Action: action}
	if err := plugin.Call("execute", req, nil); err != nil {
		log("execute: %v", err)
	}
}

// applescriptAppName maps bundle IDs to AppleScript application names.
func applescriptAppName(bundleID string) string {
	switch bundleID {
	case "com.google.Chrome":
		return "Google Chrome"
	case "com.google.Chrome.canary":
		return "Google Chrome Canary"
	case "org.mozilla.firefox":
		return "Firefox"
	case "company.thebrowser.Browser":
		return "Arc"
	default:
		// Fallback: last component of bundle ID
		parts := strings.Split(bundleID, ".")
		return parts[len(parts)-1]
	}
}
