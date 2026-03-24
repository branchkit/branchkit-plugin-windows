package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"branchkit.local/shared"
)

var platform *shared.PlatformClient

// --- Request/Response types ---

type OnActionRequest struct {
	Action         string  `json:"action"`
	ActiveApp      *string `json:"active_app,omitempty"`
	ActiveWindowID *string `json:"active_window_id,omitempty"`
}

type OnActionResponse struct {
	Result string `json:"result"` // "handled" or "pass"
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	shared.WriteJSON(w, map[string]bool{"ready": true})
}

func handleOnAction(w http.ResponseWriter, r *http.Request) {
	var req OnActionRequest
	if err := shared.ReadJSON(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sub, ok := strings.CutPrefix(req.Action, "windows ")
	if !ok {
		shared.WriteJSON(w, OnActionResponse{Result: "pass"})
		return
	}

	parts := strings.Fields(sub)
	if len(parts) == 0 {
		shared.WriteJSON(w, OnActionResponse{Result: "pass"})
		return
	}

	cmd := parts[0]
	args := parts[1:]
	log("action: %s (args: %v)", cmd, args)

	switch cmd {
	case "snap":
		if len(args) < 1 {
			log("snap: missing direction")
			break
		}
		direction := strings.Join(args, " ")
		handleSnap(req.ActiveWindowID, direction)

	case "move-to-space":
		if len(args) < 1 {
			log("move-to-space: missing space number")
			break
		}
		space, err := strconv.Atoi(args[0])
		if err != nil || space < 1 || space > 9 {
			log("move-to-space: invalid space: %s", args[0])
			break
		}
		handleMoveToSpace(req.ActiveWindowID, space)

	case "tab-to-space":
		if len(args) < 1 {
			log("tab-to-space: missing space number")
			break
		}
		space, err := strconv.Atoi(args[0])
		if err != nil || space < 1 || space > 9 {
			log("tab-to-space: invalid space: %s", args[0])
			break
		}
		appID := ""
		if req.ActiveApp != nil {
			appID = *req.ActiveApp
		}
		handleMoveTabToSpace(space, appID)

	case "tab-to-window":
		if len(args) < 1 {
			log("tab-to-window: missing window index")
			break
		}
		idx, err := strconv.Atoi(args[0])
		if err != nil || idx < 1 {
			log("tab-to-window: invalid index: %s", args[0])
			break
		}
		appID := ""
		if req.ActiveApp != nil {
			appID = *req.ActiveApp
		}
		handleMoveTabToWindow(idx, appID)

	default:
		log("unknown command: %s", cmd)
		shared.WriteJSON(w, OnActionResponse{Result: "pass"})
		return
	}

	shared.WriteJSON(w, OnActionResponse{Result: "handled"})
}

func log(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[windows] "+format+"\n", args...)
}

func main() {
	platform = shared.NewPlatformClient()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /hooks/on-action", handleOnAction)
	mux.HandleFunc("POST /hooks/render-settings", handleRenderSettings)

	shared.RunPlugin(mux)
}
