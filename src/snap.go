package main

import (
	"math"

	"branchkit.local/shared"
)

const menuBarHeight = 25

// handleSnap calculates snap geometry and applies it via batch-set-frames.
func handleSnap(activeWindowID *string, direction string) {
	var wm shared.WorldModel
	if err := plugin.Call("native.world_model", nil, &wm); err != nil {
		log("snap: failed to get world model: %v", err)
		return
	}

	winID := ""
	if activeWindowID != nil {
		winID = *activeWindowID
	} else if wm.ActiveWindowID != nil {
		winID = *wm.ActiveWindowID
	}
	if winID == "" {
		log("snap: no active window")
		return
	}

	var win *shared.WindowInfo
	for i := range wm.Windows {
		if wm.Windows[i].ID == winID {
			win = &wm.Windows[i]
			break
		}
	}
	if win == nil {
		log("snap: window %s not found", winID)
		return
	}

	if len(wm.Displays) == 0 {
		log("snap: no displays")
		return
	}

	// Find which display the window center is on
	centerX := win.X + win.W/2
	centerY := win.Y + win.H/2
	screenIdx := 0
	for i, d := range wm.Displays {
		if centerX >= d.X && centerX < d.X+d.W && centerY >= d.Y && centerY < d.Y+d.H {
			screenIdx = i
			break
		}
	}
	screen := wm.Displays[screenIdx]

	frame := calculateSnapGeometry(win, screen, screenIdx, wm.Displays, direction)
	if frame == nil {
		log("snap: no geometry for direction %q", direction)
		return
	}

	log("snap: window=%s direction=%s → x=%d y=%d w=%d h=%d (screen %d: %dx%d)",
		winID, direction, frame.X, frame.Y, frame.W, frame.H,
		screenIdx, screen.W, screen.H)

	batchReq := struct {
		Frames   []shared.WindowFrame `json:"frames"`
		Readback bool                 `json:"readback"`
	}{
		Frames: []shared.WindowFrame{
			{WindowID: winID, X: frame.X, Y: frame.Y, W: frame.W, H: frame.H},
		},
		Readback: false,
	}
	if err := plugin.Call("native.batch_set_frames", batchReq, nil); err != nil {
		log("snap: batch-set-frames error: %v", err)
	} else {
		log("snap: batch-set-frames succeeded")
	}
}

func calculateSnapGeometry(win *shared.WindowInfo, screen shared.DisplayInfo, screenIdx int, displays []shared.DisplayInfo, direction string) *shared.Rect {
	switch direction {
	case "left":
		return &shared.Rect{
			X: screen.X,
			Y: screen.Y + menuBarHeight,
			W: screen.W / 2,
			H: screen.H - menuBarHeight,
		}
	case "right":
		return &shared.Rect{
			X: screen.X + screen.W/2,
			Y: screen.Y + menuBarHeight,
			W: screen.W / 2,
			H: screen.H - menuBarHeight,
		}
	case "top", "up":
		return &shared.Rect{
			X: screen.X,
			Y: screen.Y + menuBarHeight,
			W: screen.W,
			H: (screen.H - menuBarHeight) / 2,
		}
	case "bottom", "down":
		halfH := (screen.H - menuBarHeight) / 2
		return &shared.Rect{
			X: screen.X,
			Y: screen.Y + menuBarHeight + halfH,
			W: screen.W,
			H: halfH,
		}
	case "maximize", "full":
		return &shared.Rect{
			X: screen.X,
			Y: screen.Y + menuBarHeight,
			W: screen.W,
			H: screen.H - menuBarHeight,
		}
	case "center":
		return &shared.Rect{
			X: screen.X + screen.W/4,
			Y: screen.Y + screen.H/4,
			W: screen.W / 2,
			H: screen.H / 2,
		}
	case "next", "next monitor", "other screen", "move next",
		"prev", "previous monitor", "move back":
		if len(displays) < 2 {
			return nil
		}
		var nextIdx int
		if direction == "prev" || direction == "previous monitor" || direction == "move back" {
			nextIdx = (screenIdx + len(displays) - 1) % len(displays)
		} else {
			nextIdx = (screenIdx + 1) % len(displays)
		}
		target := displays[nextIdx]

		// Proportional mapping
		relX := float64(win.X-screen.X) / float64(screen.W)
		relY := float64(win.Y-screen.Y) / float64(screen.H)
		relW := float64(win.W) / float64(screen.W)
		relH := float64(win.H) / float64(screen.H)

		return &shared.Rect{
			X: target.X + int(math.Round(relX*float64(target.W))),
			Y: target.Y + int(math.Round(relY*float64(target.H))),
			W: int(math.Round(relW * float64(target.W))),
			H: int(math.Round(relH * float64(target.H))),
		}
	default:
		return nil
	}
}
