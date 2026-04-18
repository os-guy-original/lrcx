// Package yt provides YouTube subtitle fetching via yt-dlp.
package yt

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/os-guy-original/lrcx/internal/converter"
	"github.com/os-guy-original/lrcx/internal/parser"
)

// Run fetches subtitles from a YouTube URL and converts to LRC.
func Run(url, outputPath string, offsetMs int) error {
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return fmt.Errorf("yt-dlp not found: install from https://github.com/yt-dlp/yt-dlp")
	}

	tmp, err := os.CreateTemp("", "lrcx-*.srt")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpPath)

	cmd := exec.Command("yt-dlp", "--write-subs", "--skip-download", "--sub-lang", "en", "-o", strings.TrimSuffix(tmpPath, ".srt"), url)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("yt-dlp failed: %w", err)
	}

	r, err := os.Open(tmpPath)
	if err != nil {
		return fmt.Errorf("subtitle file not created (maybe no subs available): %w", err)
	}
	defer r.Close()

	blocks, err := parser.ParseSRT(r)
	if err != nil {
		return err
	}

	var w *os.File
	if outputPath == "" {
		w = os.Stdout
	} else {
		w, err = os.Create(outputPath)
		if err != nil {
			return err
		}
		defer w.Close()
	}

	lines := converter.ToLRC(blocks, time.Duration(offsetMs)*time.Millisecond)
	fmt.Fprintln(w, strings.Join(lines, "\n"))
	return nil
}
