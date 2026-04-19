// Package yt provides YouTube subtitle fetching via yt-dlp.
package yt

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/os-guy-original/lrcx/internal/converter"
	"github.com/os-guy-original/lrcx/internal/parser"
)

// Options configures the yt-dlp subtitle fetch.
type Options struct {
	URL       string // YouTube video URL
	Output    string // Output file path (empty = stdout)
	OffsetMs  int    // Time offset in milliseconds
	SubLang   string // Subtitle language (default: en)
	AutoSubs  bool   // Use auto-generated subtitles if manual not available
}

// Run fetches subtitles from a YouTube URL and converts to LRC.
func Run(urlStr, outputPath string, offsetMs int) error {
	return RunWithOpts(Options{
		URL:      urlStr,
		Output:   outputPath,
		OffsetMs: offsetMs,
		SubLang:  "en",
	})
}

// RunWithOpts fetches subtitles with full options.
func RunWithOpts(opts Options) error {
	if err := validateURL(opts.URL); err != nil {
		return err
	}

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

	args := []string{
		"--write-subs",
		"--skip-download",
		"--sub-lang", opts.SubLang,
		"-o", strings.TrimSuffix(tmpPath, ".srt"),
	}
	if opts.AutoSubs {
		args = append(args, "--write-auto-subs")
	}
	args = append(args, opts.URL)

	cmd := exec.Command("yt-dlp", args...)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("yt-dlp failed: %w", err)
	}

	r, err := os.Open(tmpPath)
	if err != nil {
		return fmt.Errorf("subtitle file not created (maybe no subs available for lang=%s): %w", opts.SubLang, err)
	}
	defer r.Close()

	blocks, err := parser.ParseSRT(r)
	if err != nil {
		return err
	}

	if len(blocks) == 0 {
		return fmt.Errorf("no subtitle blocks found")
	}

	var w *os.File
	if opts.Output == "" {
		w = os.Stdout
	} else {
		w, err = os.Create(opts.Output)
		if err != nil {
			return err
		}
		defer w.Close()
	}

	lines := converter.ToLRC(blocks, time.Duration(opts.OffsetMs)*time.Millisecond)
	fmt.Fprintln(w, strings.Join(lines, "\n"))
	return nil
}

func validateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL must be http or https")
	}
	if u.Host == "" {
		return fmt.Errorf("URL missing host")
	}
	return nil
}
