package main

import (
	"fmt"

	"github.com/os-guy-original/lrcx/internal/beta/yt"
)

// betaFeatureName returns a human-readable name for a beta feature key.
func betaFeatureName(name string) string {
	switch name {
	case "yt":
		return "yt-dlp integration"
	default:
		return name
	}
}

// BetaFeature executes a beta feature by name.
func BetaFeature(name string, args []string, outputPath string, offsetMs int, interactive bool) error {
	switch name {
	case "yt":
		if len(args) == 0 {
			return fmt.Errorf("--beta-feature=yt requires a URL argument")
		}
		return yt.RunWithOpts(yt.Options{
			URL:         args[0],
			Output:      outputPath,
			OffsetMs:    offsetMs,
			SubLang:     "en",
			Interactive: interactive,
		})
	default:
		return fmt.Errorf("unknown beta feature: %s", name)
	}
}
