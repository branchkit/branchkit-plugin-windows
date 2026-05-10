package main

import (
	"github.com/branchkit/plugin-sdk-go"
)

var plugin *shared.Plugin

func main() {
	plugin = shared.NewPlugin()

	plugin.HandleAction("windows.snap", handleWindowsSnap)
	plugin.HandleAction("windows.desk_mode", noopAction)
	plugin.HandleAction("windows.desk_mode_done", noopAction)
	plugin.HandleAction("windows.desk_switch", handleDeskSwitch)
	plugin.HandleAction("windows.move_to_space", handleWindowsMoveToSpace)
	plugin.HandleAction("windows.tab_to_space", handleWindowsTabToSpace)
	plugin.HandleAction("windows.tab_to_window", handleWindowsTabToWindow)
	shared.HandleTyped(plugin, "render_settings", handleRenderSettingsRPC)

	plugin.Run()
}
