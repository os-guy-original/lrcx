// Package yt provides YouTube subtitle fetching via yt-dlp.
package yt

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/os-guy-original/lrcx/internal/converter"
	"github.com/os-guy-original/lrcx/internal/parser"
)

// SubtitleInfo represents available subtitle info.
type SubtitleInfo struct {
	Lang string
	Name string
	Type string // "manual" or "auto"
}

// Options configures the yt-dlp subtitle fetch.
type Options struct {
	URL         string // YouTube video URL
	Output      string // Output file path (empty = stdout)
	OffsetMs    int    // Time offset in milliseconds
	SubLang     string // Subtitle language (default: en)
	Interactive bool   // Prompt user to select subtitle
}

// RunWithOpts fetches subtitles with full options.
func RunWithOpts(opts Options) error {
	if err := validateURL(opts.URL); err != nil {
		return err
	}

	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return fmt.Errorf("yt-dlp not found: install from https://github.com/yt-dlp/yt-dlp")
	}

	// If interactive, list and select
	if opts.Interactive {
		stop := spin("Getting the subtitles")
		subs, err := ListSubtitles(opts.URL)
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
		// Clean up any files yt-dlp wrote (e.g. <base>.<lang>.vtt)
		if matches, _ := filepath.Glob(tmpBase + ".*"); matches != nil {
			for _, m := range matches {
				os.Remove(m)
			}
		}
	}()

	args := []string{
		"--write-subs",
		"--skip-download",
		"--sub-lang", opts.SubLang,
		"--sub-format", "vtt",
		"-o", tmpBase,
		opts.URL,
	}

	cmd := exec.Command("yt-dlp", args...)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("yt-dlp failed: %w", err)
	}

	// yt-dlp writes <base>.<lang>.vtt, not <base>.vtt
	matches, _ := filepath.Glob(tmpBase + "*.vtt")
	if len(matches) == 0 {
		return fmt.Errorf("subtitle file not created (maybe no subs available for lang=%s)", opts.SubLang)
	}
	vttPath := matches[0]

	r, err := os.Open(vttPath)
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

// ListSubtitles fetches available subtitles for a URL.
func ListSubtitles(urlStr string) ([]SubtitleInfo, error) {
	cmd := exec.Command("yt-dlp", "--list-subs", urlStr)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp failed: %w", err)
	}
	return parseListSubs(string(out)), nil
}

func parseListSubs(output string) []SubtitleInfo {
	var subs []SubtitleInfo
	var section string

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "Available subtitles for") {
			section = "manual"
			continue
		}
		if strings.Contains(line, "Available automatic captions for") {
			section = "auto"
			continue
		}

		if section != "manual" {
			continue
		}

		// Skip header lines like "Language   Name    Formats"
		if strings.HasPrefix(line, "Language") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 1 && !strings.HasPrefix(fields[0], "[") {
			lang := fields[0]
			name := lang

			for i := 1; i < len(fields); i++ {
				f := fields[i]
				if !strings.Contains(f, "vtt") && !strings.Contains(f, "srt") && !strings.Contains(f, ",") && f != "from" {
					name = f
					break
				}
			}

			subs = append(subs, SubtitleInfo{
				Lang: lang,
				Name: name,
				Type: section,
			})
		}
	}
	return subs
}

func promptSelect(subs []SubtitleInfo) SubtitleInfo {
	fmt.Fprintln(os.Stderr, "Available subtitles:")
	for i, s := range subs {
		fmt.Fprintf(os.Stderr, "  %d) [%s] %s (%s)\n", i+1, s.Lang, s.Name, s.Type)
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

// spin starts a spinner on stderr with the given message.
// Call the returned stop func with the operation error when done.
// On success it clears the line; on error it morphs into the error text.
func spin(msg string) func(error) {
	frames := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
	done := make(chan struct{})
	go func() {
		for i := 0; ; i++ {
			select {
			case <-done:
				return
			default:
				fmt.Fprintf(os.Stderr, "\r%c %s...", frames[i%len(frames)], msg)
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
	return func(err error) {
		close(done)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\r✗ %s\n", err)
		} else {
			// Clear the spinner line
			fmt.Fprintf(os.Stderr, "\r\033[K")
		}
	}
}
