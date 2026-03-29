package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/branchkit/plugin-sdk-go"
)

var plugin *shared.Plugin

// --- Request/Response types ---

type OnActionRequest struct {
	Action         string  `json:"action"`
	ActiveApp      *string `json:"active_app,omitempty"`
	ActiveWindowID *string `json:"active_window_id,omitempty"`
}

type OnActionResponse struct {
	Result string `json:"result"` // "handled" or "pass"
}

// rpcHandler creates a HandlerFunc that unmarshals params into the given request type,
// calls the handler, and returns the result.
func rpcHandler[Req any](fn func(*Req) (any, error)) shared.HandlerFunc {
	return func(params json.RawMessage) (any, error) {
		var req Req
		if len(params) > 0 {
			if err := json.Unmarshal(params, &req); err != nil {
				return nil, fmt.Errorf("bad params: %w", err)
			}
		}
		return fn(&req)
	}
}

func handleOnAction(req *OnActionRequest) (any, error) {
	sub, ok := strings.CutPrefix(req.Action, "windows ")
	if !ok {
		return OnActionResponse{Result: "pass"}, nil
	}

	parts := strings.Fields(sub)
	if len(parts) == 0 {
		return OnActionResponse{Result: "pass"}, nil
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
		return OnActionResponse{Result: "pass"}, nil
	}

	return OnActionResponse{Result: "handled"}, nil
}

func log(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[windows] "+format+"\n", args...)
}

func pushCommands(p *shared.Plugin) {
	pluginDir := os.Getenv("BRANCHKIT_PLUGIN_DIR")
	if pluginDir == "" {
		return
	}
	data, err := os.ReadFile(pluginDir + "/commands.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[WINDOWS] failed to read commands.json: %v\n", err)
		return
	}
	var commands []json.RawMessage
	if err := json.Unmarshal(data, &commands); err != nil {
		fmt.Fprintf(os.Stderr, "[WINDOWS] failed to parse commands.json: %v\n", err)
		return
	}
	var resp struct{ Count int `json:"count"` }
	if err := p.Call("grammar.push", map[string]any{"commands": commands}, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "[WINDOWS] grammar.push failed: %v\n", err)
		return
	}
	fmt.Fprintf(os.Stderr, "[WINDOWS] Registered %d command variants\n", resp.Count)
}

func main() {
	plugin = shared.NewPlugin()
	pushCommands(plugin)

	plugin.Handle("on_action", rpcHandler(handleOnAction))
	plugin.Handle("render_settings", rpcHandler(handleRenderSettingsRPC))

	plugin.Run()
}
