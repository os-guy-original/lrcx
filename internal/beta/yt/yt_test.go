package yt

import (
	"testing"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		url    string
		wantOk bool
	}{
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", true},
		{"http://youtube.com/watch?v=abc123", true},
		{"https://youtu.be/dQw4w9WgXcQ", true},
		{"not-a-url", false},
		{"ftp://youtube.com/video", false},
		{"", false},
		{"https://", false},
	}

	for _, tt := range tests {
		err := validateURL(tt.url)
		if tt.wantOk && err != nil {
			t.Errorf("validateURL(%q): unexpected error: %v", tt.url, err)
		}
		if !tt.wantOk && err == nil {
			t.Errorf("validateURL(%q): expected error, got nil", tt.url)
		}
	}
}

func TestParseListSubs(t *testing.T) {
	output := `[info] Available automatic captions for jNQXAC9IVRw:
Language   Name                               Formats
en                                            vtt
de                                            vtt
ab-en      Abkhazian from English             vtt, srt
[info] Available subtitles for jNQXAC9IVRw:
Language Name    Formats
en       English vtt, srt
de       German  vtt, srt
`
	subs := parseListSubs(output)

	if len(subs) < 4 {
		t.Fatalf("expected at least 4 subtitles, got %d", len(subs))
	}

	// Check auto captions
	if subs[0].Type != "auto" || subs[0].Lang != "en" {
		t.Errorf("expected auto en, got %+v", subs[0])
	}

	// Check manual subtitles
	var manualEn *SubtitleInfo
	for i := range subs {
		if subs[i].Type == "manual" && subs[i].Lang == "en" {
			manualEn = &subs[i]
			break
		}
	}
	if manualEn == nil {
		t.Error("expected to find manual English subtitle")
	} else if manualEn.Name != "English" {
		t.Errorf("expected name 'English', got %q", manualEn.Name)
	}
}

func TestRunWithOpts_InvalidURL(t *testing.T) {
	err := RunWithOpts(Options{URL: "not-a-url", SubLang: "en"})
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}
