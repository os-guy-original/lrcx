// Package yt provides YouTube subtitle fetching via yt-dlp.
package yt

import (
	"bufio"
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

// SubtitleInfo represents available subtitle info.
type SubtitleInfo struct {
	Lang string
	Name string
	Type string // "manual" or "auto"
}

// Options configures the yt-dlp subtitle fetch.
type Options struct {
	URL         string
	Output      string
	OffsetMs    int
	SubLang     string
	Interactive bool
	Verbose     bool
}

// RunWithOpts fetches subtitles with full options.
func RunWithOpts(opts Options) error {
	if err := validateURL(opts.URL); err != nil {
		return err
	}
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return fmt.Errorf("yt-dlp not found: install from https://github.com/yt-dlp/yt-dlp")
	}

	if opts.Interactive {
		stop := ui.Spin("Getting the subtitles", opts.Verbose)
		subs, err := listSubtitles(opts.URL, opts.Verbose)
		stop(err)
		if err != nil {
			return err
		}
		if len(subs) == 0 {
			return fmt.Errorf("no subtitles available")
		}
		selected := promptSelect(subs)
		opts.SubLang = selected.Lang
	}

	tmp, err := os.CreateTemp("", "lrcx-*")
	if err != nil {
		return err
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
	err = ytdlp(opts.Verbose, "--write-subs", "--skip-download",
		"--sub-lang", opts.SubLang, "--sub-format", "vtt",
		"-o", tmpBase, opts.URL)
	stop(err)
	if err != nil {
		return err
	}

	matches, _ := filepath.Glob(tmpBase + "*.vtt")
	if len(matches) == 0 {
		return fmt.Errorf("subtitle file not created (maybe no subs available for lang=%s)", opts.SubLang)
	}

	r, err := os.Open(matches[0])
	if err != nil {
		return fmt.Errorf("cannot open subtitle file: %w", err)
	}
	defer r.Close()

	blocks, err := parser.ParseVTT(r)
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

// ytdlp runs yt-dlp with the given args, routing stderr to os.Stderr only when verbose.
func ytdlp(verbose bool, args ...string) error {
	cmd := exec.Command("yt-dlp", args...)
	if verbose {
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stderr = io.Discard
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("yt-dlp failed: %w", err)
	}
	return nil
}

// listSubtitles fetches available manual subtitles for a URL.
func listSubtitles(urlStr string, verbose bool) ([]SubtitleInfo, error) {
	cmd := exec.Command("yt-dlp", "--list-subs", urlStr)
	var stderr io.Writer = io.Discard
	if verbose {
		stderr = os.Stderr
	}
	cmd.Stderr = stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp failed: %w", err)
	}
	return parseListSubs(string(out)), nil
}

func parseListSubs(output string) []SubtitleInfo {
	var subs []SubtitleInfo
	var inManual bool

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Available subtitles for") {
			inManual = true
			continue
		}
		if strings.Contains(line, "Available automatic captions for") {
			inManual = false
			continue
		}
		if !inManual || strings.HasPrefix(line, "Language") {
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
		subs = append(subs, SubtitleInfo{Lang: lang, Name: name, Type: "manual"})
	}
	return subs
}

func promptSelect(subs []SubtitleInfo) SubtitleInfo {
	fmt.Fprintln(os.Stderr, "Available subtitles:")
	for i, s := range subs {
		fmt.Fprintf(os.Stderr, "  %d) [%s] %s\n", i+1, s.Lang, s.Name)
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
