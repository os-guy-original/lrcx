package main

import (
	"fmt"
	"time"

	"github.com/os-guy-original/lrcx/internal/beta/yt"
)

func betaFeatureName(name string) string {
	switch name {
	case "yt":
		return "yt-dlp integration"
	default:
		return name
	}
}

func BetaFeature(name string, args []string, outputPath string, offsetMs int, interactive, verbose, autoCaptions bool, timeout time.Duration) error {
	switch name {
	case "yt":
		if len(args) == 0 {
			return fmt.Errorf("--beta-feature=yt requires a URL argument")
		}
		return yt.RunWithOpts(yt.Options{
			URL:          args[0],
			Output:       outputPath,
			OffsetMs:     offsetMs,
			SubLang:      "en",
			Interactive:  interactive,
			Verbose:      verbose,
			AutoCaptions: autoCaptions,
			Timeout:      timeout,
		})
	default:
		return fmt.Errorf("unknown beta feature: %s", name)
	}
}
