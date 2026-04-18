package main

import (
	"fmt"

	"github.com/os-guy-original/lrcx/internal/beta/yt"
)

// BetaFeature executes a beta feature by name.
func BetaFeature(name string, args []string, outputPath string, offsetMs int) error {
	switch name {
	case "yt":
		if len(args) == 0 {
			return fmt.Errorf("--beta-feature=yt requires a URL argument")
		}
		return yt.Run(args[0], outputPath, offsetMs)
	default:
		return fmt.Errorf("unknown beta feature: %s", name)
	}
}
