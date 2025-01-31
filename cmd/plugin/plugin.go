package main

import (
	"log"
	"runtime/debug"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/hashicorp/packer-plugin-sdk/version"
)

var (
	// Version is overridden by the linker during build, defaults to 0.0.0
	Version string

	pps *plugin.Set
)

func init() {
	var metadata string
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, bs := range bi.Settings {
			switch bs.Key {
			case "vcs.revision":
				if len(bs.Value) > 7 {
					metadata = bs.Value[0:7]
				}
			}
		}
	}

	// metadata can be used in 0.6.0+ API and beyond
	_ = metadata

	pps = plugin.NewSet()
	pps.SetVersion(version.InitializePluginVersion(Version, ""))
	pps.RegisterBuilder(plugin.DEFAULT_NAME, new(Builder))
}

func main() {
	log.SetFlags(0)

	err := pps.Run()
	if err != nil {
		log.Panicf("Failed to run plugin: %s", err)
	}
}
