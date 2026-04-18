package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/srt2lrc/srt2lrc/internal/converter"
	"github.com/srt2lrc/srt2lrc/internal/parser"
)

const version = "0.1.0"

func main() {
	input := flag.String("i", "", "input SRT file (default: stdin)")
	output := flag.String("o", "", "output LRC file (default: stdout)")
	offsetMs := flag.Int("offset", 0, "time offset in milliseconds")
	ver := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *ver {
		fmt.Println("srt2lrc", version)
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
	fmt.Fprintln(os.Stderr, "srt2lrc:", err)
	os.Exit(1)
}
