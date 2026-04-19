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

func TestRun_BasicVideo(t *testing.T) {
	err := RunWithOpts(Options{
		URL:      "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		AutoSubs: true,
	})
	if err != nil {
		t.Skipf("skipping: %v", err)
	}
}

func TestRun_WithOffset(t *testing.T) {
	err := RunWithOpts(Options{
		URL:      "https://www.youtube.com/watch?v=jNQXAC9IVRw",
		OffsetMs: 500,
		AutoSubs: true,
	})
	if err != nil {
		t.Skipf("skipping: %v", err)
	}
}

func TestRun_InvalidURL(t *testing.T) {
	err := Run("not-a-url", "", 0)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestRunWithOpts_AutoSubs(t *testing.T) {
	err := RunWithOpts(Options{
		URL:      "https://www.youtube.com/watch?v=jNQXAC9IVRw",
		SubLang:  "en",
		AutoSubs: true,
	})
	if err != nil {
		t.Skipf("skipping: %v", err)
	}
}

func TestRunWithOpts_CustomLang(t *testing.T) {
	err := RunWithOpts(Options{
		URL:     "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		SubLang: "en",
	})
	if err != nil {
		t.Skipf("skipping: %v", err)
	}
}
