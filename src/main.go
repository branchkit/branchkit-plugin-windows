package main

import (
	"github.com/branchkit/plugin-sdk-go"
)

var plugin *branchkit.Plugin

func main() {
	plugin = branchkit.NewPlugin()

	// Registrars are generated from plugin.json into actions_gen.go, so the
	// action string and the params type both come from the manifest and
	// neither can drift from it.
	HandleSnap(plugin, handleWindowsSnap)
	HandleDeskSwitch(plugin, handleDeskSwitch)
	HandleMoveToSpace(plugin, handleWindowsMoveToSpace)
	branchkit.HandleTyped(plugin, "render_settings", handleRenderSettingsRPC)

	plugin.Run()
}
