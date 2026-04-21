// Package yt provides YouTube subtitle fetching via yt-dlp.
package yt

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/os-guy-original/lrcx/internal/converter"
	"github.com/os-guy-original/lrcx/internal/parser"
	"github.com/os-guy-original/lrcx/internal/ui"
)

// Default timeout for yt-dlp operations.
const defaultTimeout = 60 * time.Second

// SubtitleInfo represents available subtitle info.
type SubtitleInfo struct {
	Lang string
	Name string
	Type string // "manual" or "auto"
}

// Options configures the yt-dlp subtitle fetch.
type Options struct {
	URL          string
	Output       string
	OffsetMs     int
	SubLang      string
	Interactive  bool
	Verbose      bool
	AutoCaptions bool   // Include auto-generated captions
	Timeout      time.Duration // Operation timeout (0 = default)
}

// RunWithOpts fetches subtitles with full options.
func RunWithOpts(opts Options) error {
	// Set default timeout
	if opts.Timeout == 0 {
		opts.Timeout = defaultTimeout
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	if err := validateURL(opts.URL); err != nil {
		return err
	}
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return fmt.Errorf("yt-dlp not found: install from https://github.com/yt-dlp/yt-dlp")
	}

	if opts.Interactive {
		stop := ui.Spin("Getting available subtitles", opts.Verbose)
		subs, err := listSubtitles(ctx, opts.URL, opts.AutoCaptions, opts.Verbose)
		stop(err)
		if err != nil {
			return fmt.Errorf("failed to list subtitles: %w", err)
		}
		if len(subs) == 0 {
			return fmt.Errorf("no subtitles available for this video")
		}
		selected := promptSelect(subs)
		opts.SubLang = selected.Lang
		if selected.Type == "auto" {
			opts.AutoCaptions = true
		}
	}

	tmp, err := os.CreateTemp("", "lrcx-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpBase := tmp.Name()
	tmp.Close()
	os.Remove(tmpBase)
	defer func() {
		if matches, _ := filepath.Glob(tmpBase + ".*"); matches != nil {
			for _, m := range matches {
				os.Remove(m)
			}
		}
	}()

	stop := ui.Spin("Downloading subtitles", opts.Verbose)
	subFlag := "--write-subs"
	if opts.AutoCaptions {
		subFlag = "--write-auto-subs"
	}
	err = ytdlp(ctx, opts.Verbose, subFlag, "--skip-download",
		"--sub-lang", opts.SubLang, "--sub-format", "vtt",
		"-o", tmpBase, opts.URL)
	stop(err)
	if err != nil {
		return fmt.Errorf("failed to download subtitles: %w", err)
	}

	matches, _ := filepath.Glob(tmpBase + "*.vtt")
	if len(matches) == 0 {
		langHint := fmt.Sprintf("lang=%s", opts.SubLang)
		if opts.AutoCaptions {
			langHint += " (auto-generated)"
		}
		return fmt.Errorf("subtitle file not created (no subtitles for %s)", langHint)
	}

	r, err := os.Open(matches[0])
	if err != nil {
		return fmt.Errorf("cannot open subtitle file: %w", err)
	}
	defer r.Close()

	blocks, err := parser.ParseVTT(r)
	if err != nil {
		return fmt.Errorf("failed to parse subtitle file: %w", err)
	}
	if len(blocks) == 0 {
		return fmt.Errorf("no subtitle blocks found in downloaded file")
	}

	var w *os.File
	if opts.Output == "" {
		w = os.Stdout
	} else {
		w, err = os.Create(opts.Output)
		if err != nil {
			return fmt.Errorf("cannot create output file: %w", err)
		}
		defer w.Close()
	}

	lines := converter.ToLRC(blocks, time.Duration(opts.OffsetMs)*time.Millisecond)
	fmt.Fprintln(w, strings.Join(lines, "\n"))
	return nil
}

// ytdlp runs yt-dlp with the given args, routing stderr to os.Stderr only when verbose.
func ytdlp(ctx context.Context, verbose bool, args ...string) error {
	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	if verbose {
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stderr = io.Discard
	}
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("yt-dlp timed out after %v", defaultTimeout)
		}
		return fmt.Errorf("yt-dlp failed: %w", err)
	}
	return nil
}

// listSubtitles fetches available subtitles for a URL.
func listSubtitles(ctx context.Context, urlStr string, includeAuto, verbose bool) ([]SubtitleInfo, error) {
	cmd := exec.CommandContext(ctx, "yt-dlp", "--list-subs", urlStr)
	var stderr io.Writer = io.Discard
	if verbose {
		stderr = os.Stderr
	}
	cmd.Stderr = stderr
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("timed out while listing subtitles")
		}
		return nil, fmt.Errorf("yt-dlp failed: %w", err)
	}
	return parseListSubs(string(out), includeAuto), nil
}

func parseListSubs(output string, includeAuto bool) []SubtitleInfo {
	var subs []SubtitleInfo
	var inAuto bool

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Available automatic captions for") {
			inAuto = true
			continue
		}
		if strings.Contains(line, "Available subtitles for") {
			inAuto = false
			continue
		}
		if strings.HasPrefix(line, "Language") {
			continue
		}
		if inAuto && !includeAuto {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 1 || strings.HasPrefix(fields[0], "[") {
			continue
		}
		lang, name := fields[0], fields[0]
		for _, f := range fields[1:] {
			if !strings.Contains(f, "vtt") && !strings.Contains(f, "srt") && !strings.Contains(f, ",") && f != "from" {
				name = f
				break
			}
		}
		subType := "manual"
		if inAuto {
			subType = "auto"
		}
		subs = append(subs, SubtitleInfo{Lang: lang, Name: name, Type: subType})
	}
	return subs
}

func promptSelect(subs []SubtitleInfo) SubtitleInfo {
	fmt.Fprintln(os.Stderr, "Available subtitles:")
	for i, s := range subs {
		typeMarker := ""
		if s.Type == "auto" {
			typeMarker = " (auto)"
		}
		fmt.Fprintf(os.Stderr, "  %d) [%s] %s%s\n", i+1, s.Lang, s.Name, typeMarker)
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, "Select subtitle [1]: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			return subs[0]
		}
		var idx int
		if _, err := fmt.Sscanf(input, "%d", &idx); err == nil && idx >= 1 && idx <= len(subs) {
			return subs[idx-1]
		}
		fmt.Fprintln(os.Stderr, "Invalid selection")
	}
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
