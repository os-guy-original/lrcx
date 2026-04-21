package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/os-guy-original/lrcx/internal/converter"
	"github.com/os-guy-original/lrcx/internal/parser"
)

const version = "0.2.0"

func main() {
	// Check for subcommands first
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "lrclib":
			runLRCLibCommand(os.Args[2:])
			return
		case "yt":
			runYTCommand(os.Args[2:])
			return
		}
	}

	// Standard flag parsing for non-subcommand usage
	input := flag.String("i", "", "input file (default: stdin)")
	output := flag.String("o", "", "output LRC file (default: stdout)")
	offsetMs := flag.Int("offset", 0, "time offset in milliseconds")
	ver := flag.Bool("version", false, "print version and exit")
	betaFeature := flag.String("beta-feature", "", "enable beta feature (e.g., yt)")
	interactive := flag.Bool("interactive", false, "prompt for selection")
	verbose := flag.Bool("verbose", false, "show output from third-party tools")
	autoCaptions := flag.Bool("auto-captions", false, "include auto-generated captions (yt)")
	timeout := flag.Duration("timeout", 60*time.Second, "timeout for network operations")

	flag.Parse()

	// Re-check for verbose flag after URL
	if !*verbose {
		for _, a := range os.Args[1:] {
			if a == "--verbose" || a == "-verbose" {
				*verbose = true
				break
			}
		}
	}

	if *ver {
		fmt.Println("lrcx", version)
		return
	}

	args := flag.Args()

	if *betaFeature != "" {
		fmt.Fprintf(os.Stderr, "warning: %s is experimental and may change or be removed\n", betaFeatureName(*betaFeature))
		if err := BetaFeature(*betaFeature, args, *output, *offsetMs, *interactive, *verbose, *autoCaptions, *timeout); err != nil {
			fatal(err)
		}
		return
	}

	var r io.Reader = os.Stdin
	if *input != "" {
		f, err := os.Open(*input)
		if err != nil {
			fatal(err)
		}
		defer f.Close()
		r = f
	}

	var w io.Writer = os.Stdout
	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			fatal(err)
		}
		defer f.Close()
		w = f
	}

	blocks, err := parseAuto(r)
	if err != nil {
		fatal(err)
	}

	lines := converter.ToLRC(blocks, time.Duration(*offsetMs)*time.Millisecond)
	fmt.Fprintln(w, strings.Join(lines, "\n"))
}

func runLRCLibCommand(args []string) {
	fs := flag.NewFlagSet("lrclib", flag.ExitOnError)
	output := fs.String("o", "", "output LRC file (default: stdout)")
	offsetMs := fs.Int("offset", 0, "time offset in milliseconds")
	interactive := fs.Bool("interactive", false, "prompt for selection")
	verbose := fs.Bool("verbose", false, "show debug output")
	artist := fs.String("artist", "", "artist name")
	track := fs.String("track", "", "track name")
	plainOnly := fs.Bool("plain", false, "get plain lyrics only")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: lrcx lrclib [options] [query]\n\n")
		fmt.Fprintf(os.Stderr, "Search and download lyrics from LRCLib.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  lrcx lrclib \"never gonna give you up\"\n")
		fmt.Fprintf(os.Stderr, "  lrcx lrclib --artist \"Rick Astley\" --track \"Never Gonna Give You Up\"\n")
		fmt.Fprintf(os.Stderr, "  lrcx lrclib --interactive \"bohemian rhapsody\"\n")
		fmt.Fprintf(os.Stderr, "  lrcx lrclib --interactive --artist \"Queen\" --track \"Bohemian Rhapsody\"\n")
	}

	fs.Parse(args)
	queryArgs := fs.Args()

	var query string
	if len(queryArgs) > 0 {
		query = queryArgs[0]
	}

	opts := struct {
		Query       string
		Artist      string
		Track       string
		Output      string
		OffsetMs    int
		Interactive bool
		PlainOnly   bool
		Verbose     bool
	}{
		Query:       query,
		Artist:      *artist,
		Track:       *track,
		Output:      *output,
		OffsetMs:    *offsetMs,
		Interactive: *interactive,
		PlainOnly:   *plainOnly,
		Verbose:     *verbose,
	}

	if err := runLRCLibWithOpts(opts); err != nil {
		fatal(err)
	}
}

func runYTCommand(args []string) {
	fs := flag.NewFlagSet("yt", flag.ExitOnError)
	output := fs.String("o", "", "output LRC file (default: stdout)")
	offsetMs := fs.Int("offset", 0, "time offset in milliseconds")
	interactive := fs.Bool("interactive", false, "prompt for subtitle selection")
	verbose := fs.Bool("verbose", false, "show output from yt-dlp")
	autoCaptions := fs.Bool("auto-captions", false, "include auto-generated captions")
	timeout := fs.Duration("timeout", 60*time.Second, "timeout for network operations")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: lrcx yt [options] <url>\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  lrcx yt https://youtube.com/watch?v=...\n")
		fmt.Fprintf(os.Stderr, "  lrcx yt --interactive https://youtube.com/watch?v=...\n")
		fmt.Fprintf(os.Stderr, "  lrcx yt --auto-captions https://youtube.com/watch?v=...\n")
	}

	fs.Parse(args)
	urlArgs := fs.Args()

	var url string
	if len(urlArgs) > 0 {
		url = urlArgs[0]
	}

	fmt.Fprintf(os.Stderr, "warning: yt-dlp integration is experimental and may change or be removed\n")

	if err := BetaFeature("yt", []string{url}, *output, *offsetMs, *interactive, *verbose, *autoCaptions, *timeout); err != nil {
		fatal(err)
	}
}

func runLRCLibWithOpts(opts struct {
	Query       string
	Artist      string
	Track       string
	Output      string
	OffsetMs    int
	Interactive bool
	PlainOnly   bool
	Verbose     bool
}) error {
	// Import and use the lrclib package
	return runLRCLibImpl(opts)
}

func parseAuto(r io.Reader) ([]parser.Block, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	content := string(data)
	if strings.HasPrefix(strings.TrimSpace(content), "WEBVTT") {
		return parser.ParseVTT(strings.NewReader(content))
	}
	return parser.ParseSRT(strings.NewReader(content))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "lrcx:", err)
	os.Exit(1)
}
