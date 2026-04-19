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

const version = "0.1.0"

func main() {
	input := flag.String("i", "", "input file (default: stdin)")
	output := flag.String("o", "", "output LRC file (default: stdout)")
	offsetMs := flag.Int("offset", 0, "time offset in milliseconds")
	ver := flag.Bool("version", false, "print version and exit")
	betaFeature := flag.String("beta-feature", "", "enable beta feature (e.g., yt)")
	interactive := flag.Bool("interactive", false, "prompt for subtitle selection")
	verbose := flag.Bool("verbose", false, "show output from third-party tools")
	flag.Parse()

	// flag.Parse stops at the first non-flag argument, so --verbose placed after
	// a URL would be silently ignored. Re-check os.Args directly.
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

	if *betaFeature != "" {
		fmt.Fprintf(os.Stderr, "warning: %s is experimental and may change or be removed\n", betaFeatureName(*betaFeature))
		if err := BetaFeature(*betaFeature, flag.Args(), *output, *offsetMs, *interactive, *verbose); err != nil {
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
