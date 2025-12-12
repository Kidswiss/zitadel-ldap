package main

import (
	ghandler "git.earthnet.ch/simon.beck/zitadel-ldap/handler"
	"github.com/glauth/glauth/v2/pkg/handler"
)

// NewZitadelHandler creates a new instance of the zitadel backend.
// This function is exported for use as a GLAuth plugin.
func NewZitadelHandler(opts ...handler.Option) handler.Handler {
	return ghandler.NewZitadelHandler(opts...)
}

func main() {
	// This is a GLAuth plugin - it should be loaded via GLAuth's plugin system
	// and not run directly. The NewZitadelHandler function is the plugin entry point.
}
