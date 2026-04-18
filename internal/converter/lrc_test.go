package converter

import (
	"testing"
	"time"

	"github.com/os-guy-original/lrcx/internal/parser"
)

func dur(h, m, s, ms int) time.Duration {
	return time.Duration(h)*time.Hour +
		time.Duration(m)*time.Minute +
		time.Duration(s)*time.Second +
		time.Duration(ms)*time.Millisecond
}

func TestFormatLRCTime(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{dur(0, 0, 0, 0), "[00:00.00]"},
		{dur(0, 0, 1, 0), "[00:01.00]"},
		{dur(0, 1, 23, 450), "[01:23.45]"},
		{dur(0, 0, 0, 999), "[00:00.99]"},
		{dur(0, 0, 0, 10), "[00:00.01]"},
		{dur(1, 2, 3, 400), "[62:03.40]"}, // hours roll into minutes
		{dur(0, 59, 59, 990), "[59:59.99]"},
	}
	for _, c := range cases {
		got := formatLRCTime(c.d)
		if got != c.want {
			t.Errorf("formatLRCTime(%v) = %q, want %q", c.d, got, c.want)
		}
	}
}

func TestToLRC_Basic(t *testing.T) {
	blocks := []parser.Block{
		{Start: dur(0, 0, 1, 0), Lines: []string{"Hello world"}},
	}
	got := ToLRC(blocks, 0)
	if len(got) != 1 || got[0] != "[00:01.00]Hello world" {
		t.Errorf("unexpected output: %v", got)
	}
}

func TestToLRC_MultiLine(t *testing.T) {
	blocks := []parser.Block{
		{Start: dur(0, 0, 5, 0), Lines: []string{"Line one", "line two"}},
	}
	got := ToLRC(blocks, 0)
	if got[0] != "[00:05.00]Line one line two" {
		t.Errorf("unexpected: %q", got[0])
	}
}

func TestToLRC_SkipsEmptyAfterTagStrip(t *testing.T) {
	blocks := []parser.Block{
		{Start: dur(0, 0, 1, 0), Lines: []string{"<b></b>"}},
		{Start: dur(0, 0, 2, 0), Lines: []string{"Keep me"}},
	}
	got := ToLRC(blocks, 0)
	if len(got) != 1 || got[0] != "[00:02.00]Keep me" {
		t.Errorf("unexpected output: %v", got)
	}
}

func TestToLRC_PositiveOffset(t *testing.T) {
	blocks := []parser.Block{
		{Start: dur(0, 0, 1, 0), Lines: []string{"Text"}},
	}
	got := ToLRC(blocks, 500*time.Millisecond)
	if got[0] != "[00:01.50]Text" {
		t.Errorf("unexpected: %q", got[0])
	}
}

func TestToLRC_NegativeOffsetClampsToZero(t *testing.T) {
	blocks := []parser.Block{
		{Start: dur(0, 0, 0, 200), Lines: []string{"Early"}},
	}
	got := ToLRC(blocks, -500*time.Millisecond)
	if got[0] != "[00:00.00]Early" {
		t.Errorf("unexpected: %q", got[0])
	}
}

func TestToLRC_ItalicTag(t *testing.T) {
	blocks := []parser.Block{
		{Start: dur(0, 0, 3, 0), Lines: []string{"<i>styled</i>"}},
	}
	got := ToLRC(blocks, 0)
	if got[0] != "[00:03.00]_styled_" {
		t.Errorf("unexpected: %q", got[0])
	}
}

func TestToLRC_Empty(t *testing.T) {
	got := ToLRC(nil, 0)
	if len(got) != 0 {
		t.Errorf("want empty, got %v", got)
	}
}
