package main

import (
	"github.com/os-guy-original/lrcx/internal/lrclib"
)

func runLRCLibImpl(opts struct {
	Query       string
	Artist      string
	Track       string
	Output      string
	OffsetMs    int
	Interactive bool
	PlainOnly   bool
	Verbose     bool
}) error {
	lrclibOpts := lrclib.Options{
		Query:       opts.Query,
		Artist:      opts.Artist,
		Track:       opts.Track,
		Output:      opts.Output,
		OffsetMs:    opts.OffsetMs,
		Interactive: opts.Interactive,
		PlainOnly:   opts.PlainOnly,
		Verbose:     opts.Verbose,
	}

	return lrclib.RunWithOpts(lrclibOpts)
}
