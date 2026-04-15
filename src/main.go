package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/branchkit/plugin-sdk-go"
)

var plugin *shared.Plugin

// --- Request/Response types ---

type OnActionRequest struct {
	Action         string                 `json:"action"`
	ActiveApp      *string                `json:"active_app,omitempty"`
	ActiveWindowID *string                `json:"active_window_id,omitempty"`
	Params         map[string]interface{} `json:"params,omitempty"`
	Args           []string               `json:"args,omitempty"`
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
	p := req.Params
	args := req.Args
	shared.Logf("windows", "action: %s (params: %v, args: %v)", req.Action, p, args)

	switch req.Action {
	case "windows.snap":
		position, _ := p["position"].(string)
		if position == "" {
			shared.Logf("windows", "snap: missing position")
			break
		}
		handleSnap(req.ActiveWindowID, position)

	case "windows.move_to_space":
		if len(args) < 1 {
			shared.Logf("windows", "move_to_space: missing space number")
			break
		}
		space, err := strconv.Atoi(args[0])
		if err != nil || space < 1 || space > 9 {
			shared.Logf("windows", "move_to_space: invalid space: %s", args[0])
			break
		}
		handleMoveToSpace(req.ActiveWindowID, space)

	case "windows.tab_to_space":
		if len(args) < 1 {
			shared.Logf("windows", "tab_to_space: missing space number")
			break
		}
		space, err := strconv.Atoi(args[0])
		if err != nil || space < 1 || space > 9 {
			shared.Logf("windows", "tab_to_space: invalid space: %s", args[0])
			break
		}
		appID := ""
		if req.ActiveApp != nil {
			appID = *req.ActiveApp
		}
		handleMoveTabToSpace(space, appID)

	case "windows.tab_to_window":
		if len(args) < 1 {
			shared.Logf("windows", "tab_to_window: missing window index")
			break
		}
		idx, err := strconv.Atoi(args[0])
		if err != nil || idx < 1 {
			shared.Logf("windows", "tab_to_window: invalid index: %s", args[0])
			break
		}
		appID := ""
		if req.ActiveApp != nil {
			appID = *req.ActiveApp
		}
		handleMoveTabToWindow(idx, appID)

	default:
		shared.Logf("windows", "unknown action: %s", req.Action)
		return OnActionResponse{Result: "pass"}, nil
	}

	return OnActionResponse{Result: "handled"}, nil
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

	plugin.Handle("on_action", rpcHandler(handleOnAction))
	plugin.Handle("render_settings", rpcHandler(handleRenderSettingsRPC))

	plugin.Run()
}
