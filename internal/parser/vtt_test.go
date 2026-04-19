package parser

import (
	"strings"
	"testing"
)

func TestParseVTT_Basic(t *testing.T) {
	input := "WEBVTT\n\n00:00:01.000 --> 00:00:04.000\nHello world\n\n00:00:05.000 --> 00:00:08.000\nSecond line\n"
	blocks, err := ParseVTT(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseVTT error: %v", err)
	}
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
	if blocks[0].Lines[0] != "Hello world" {
		t.Errorf("expected 'Hello world', got %q", blocks[0].Lines[0])
	}
}

func TestParseVTT_ShortTimestamp(t *testing.T) {
	input := `WEBVTT

00:01.000 --> 00:04.000
Short format
`
	blocks, err := ParseVTT(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseVTT error: %v", err)
	}
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0].Start.Milliseconds() != 1000 {
		t.Errorf("expected 1000ms, got %d", blocks[0].Start.Milliseconds())
	}
}

func TestParseVTT_MultiLine(t *testing.T) {
	input := `WEBVTT

00:00:01.000 --> 00:00:04.000
Line one
Line two
`
	blocks, err := ParseVTT(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseVTT error: %v", err)
	}
	if len(blocks[0].Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(blocks[0].Lines))
	}
}

func TestParseVTT_Empty(t *testing.T) {
	input := `WEBVTT
`
	blocks, err := ParseVTT(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseVTT error: %v", err)
	}
	if len(blocks) != 0 {
		t.Errorf("expected 0 blocks, got %d", len(blocks))
	}
}

func TestParseVTT_WithCueSettings(t *testing.T) {
	input := `WEBVTT

00:00:01.000 --> 00:00:04.000 position:50% align:middle
Text with settings
`
	blocks, err := ParseVTT(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseVTT error: %v", err)
	}
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0].Lines[0] != "Text with settings" {
		t.Errorf("expected 'Text with settings', got %q", blocks[0].Lines[0])
	}
}
