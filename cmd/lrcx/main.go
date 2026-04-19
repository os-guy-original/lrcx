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
	input := flag.String("i", "", "input SRT file (default: stdin)")
	output := flag.String("o", "", "output LRC file (default: stdout)")
	offsetMs := flag.Int("offset", 0, "time offset in milliseconds")
	ver := flag.Bool("version", false, "print version and exit")
	betaFeature := flag.String("beta-feature", "", "enable beta feature (e.g., yt)")
	interactive := flag.Bool("interactive", false, "prompt for subtitle selection")
	flag.Parse()

	if *ver {
		fmt.Println("lrcx", version)
		return
	}

	if *betaFeature != "" {
		fmt.Fprintln(os.Stderr, "warning: --beta-feature is experimental and may change or be removed")
		if err := BetaFeature(*betaFeature, flag.Args(), *output, *offsetMs, *interactive); err != nil {
			fatal(err)
		}
		return
	}

	r, err := openInput(*input)
	if err != nil {
		fatal(err)
	}
	defer r.Close()

	w, err := openOutput(*output)
	if err != nil {
		fatal(err)
	}
	defer w.Close()

	blocks, err := parser.ParseSRT(r)
	if err != nil {
		fatal(err)
	}

	lines := converter.ToLRC(blocks, time.Duration(*offsetMs)*time.Millisecond)
	fmt.Fprintln(w, strings.Join(lines, "\n"))
}

func openInput(path string) (io.ReadCloser, error) {
	if path == "" {
		return io.NopCloser(os.Stdin), nil
	}
	return os.Open(path)
}

func openOutput(path string) (io.WriteCloser, error) {
	if path == "" {
		return os.Stdout, nil
	}
	return os.Create(path)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "lrcx:", err)
	os.Exit(1)
}
