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

	if len(subs) != 2 {
		t.Fatalf("expected 2 manual subtitles, got %d", len(subs))
	}
	if subs[0].Type != "manual" || subs[0].Lang != "en" || subs[0].Name != "English" {
		t.Errorf("unexpected first sub: %+v", subs[0])
	}
}

func TestRunWithOpts_InvalidURL(t *testing.T) {
	err := RunWithOpts(Options{URL: "not-a-url", SubLang: "en"})
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}
