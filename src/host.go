package main

import "github.com/branchkit/plugin-sdk-go"

// Host is what every handler in this plugin needs: the platform handle.
// Handlers are methods on it, so their dependency is visible in the
// signature and cannot be read before it exists.
type Host struct {
	plugin *branchkit.Plugin
}

func newHost(p *branchkit.Plugin) *Host { return &Host{plugin: p} }
