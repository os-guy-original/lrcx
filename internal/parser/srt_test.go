package parser

import (
	"strings"
	"testing"
	"time"
)

func ms(n int) time.Duration { return time.Duration(n) * time.Millisecond }

var basicSRT = `1
00:00:01,000 --> 00:00:03,500
Hello world

2
00:00:05,000 --> 00:00:07,000
Second line`

func TestParseSRT_Basic(t *testing.T) {
	blocks, err := ParseSRT(strings.NewReader(basicSRT))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 2 {
		t.Fatalf("want 2 blocks, got %d", len(blocks))
	}
	b := blocks[0]
	if b.Sequence != 1 {
		t.Errorf("seq: want 1, got %d", b.Sequence)
	}
	if b.Start != ms(1000) {
		t.Errorf("start: want 1000ms, got %v", b.Start)
	}
	if b.End != ms(3500) {
		t.Errorf("end: want 3500ms, got %v", b.End)
	}
	if b.Lines[0] != "Hello world" {
		t.Errorf("text: want 'Hello world', got %q", b.Lines[0])
	}
}

func TestParseSRT_MultiLine(t *testing.T) {
	src := "1\n00:00:01,000 --> 00:00:03,000\nLine one\nLine two\nLine three\n"
	blocks, _ := ParseSRT(strings.NewReader(src))
	if len(blocks) != 1 {
		t.Fatalf("want 1 block, got %d", len(blocks))
	}
	if len(blocks[0].Lines) != 3 {
		t.Errorf("want 3 lines, got %d", len(blocks[0].Lines))
	}
}

func TestParseSRT_NoTrailingNewline(t *testing.T) {
	src := "1\n00:00:01,000 --> 00:00:02,000\nNo newline at end"
	blocks, err := ParseSRT(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("want 1 block, got %d", len(blocks))
	}
}

func TestParseSRT_SkipsMalformedSequence(t *testing.T) {
	src := "bad\n00:00:01,000 --> 00:00:02,000\nText\n\n1\n00:00:03,000 --> 00:00:04,000\nGood\n"
	blocks, _ := ParseSRT(strings.NewReader(src))
	if len(blocks) != 1 {
		t.Fatalf("want 1 valid block, got %d", len(blocks))
	}
	if blocks[0].Lines[0] != "Good" {
		t.Errorf("unexpected text: %q", blocks[0].Lines[0])
	}
}

func TestParseSRT_SkipsMalformedTimestamp(t *testing.T) {
	src := "1\nnot-a-timestamp\nText\n\n2\n00:00:01,000 --> 00:00:02,000\nGood\n"
	blocks, _ := ParseSRT(strings.NewReader(src))
	if len(blocks) != 1 {
		t.Fatalf("want 1 valid block, got %d", len(blocks))
	}
}

func TestParseSRT_HoursTimestamp(t *testing.T) {
	src := "1\n01:02:03,456 --> 01:02:05,000\nText\n"
	blocks, _ := ParseSRT(strings.NewReader(src))
	want := time.Hour + 2*time.Minute + 3*time.Second + 456*time.Millisecond
	if blocks[0].Start != want {
		t.Errorf("want %v, got %v", want, blocks[0].Start)
	}
}

func TestParseSRT_WindowsCRLF(t *testing.T) {
	src := "1\r\n00:00:01,000 --> 00:00:02,000\r\nHello\r\n\r\n"
	blocks, err := ParseSRT(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("want 1 block, got %d", len(blocks))
	}
	if blocks[0].Lines[0] != "Hello" {
		t.Errorf("unexpected text: %q", blocks[0].Lines[0])
	}
}

func TestParseSRT_Empty(t *testing.T) {
	blocks, err := ParseSRT(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 0 {
		t.Errorf("want 0 blocks, got %d", len(blocks))
	}
}

func TestParseSRT_MultipleBlocks(t *testing.T) {
	blocks, err := ParseSRT(strings.NewReader(basicSRT))
	if err != nil {
		t.Fatal(err)
	}
	if blocks[1].Sequence != 2 {
		t.Errorf("seq: want 2, got %d", blocks[1].Sequence)
	}
	if blocks[1].Start != ms(5000) {
		t.Errorf("start: want 5000ms, got %v", blocks[1].Start)
	}
}
