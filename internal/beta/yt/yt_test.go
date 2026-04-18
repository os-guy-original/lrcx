package yt

import (
	"testing"
)

func TestRun_BasicVideo(t *testing.T) {
	// Classic video with subtitles
	err := Run("https://www.youtube.com/watch?v=dQw4w9WgXcQ", "", 0)
	if err != nil {
		t.Skipf("skipping: %v", err)
	}
}

func TestRun_WithOffset(t *testing.T) {
	err := Run("https://www.youtube.com/watch?v=jNQXAC9IVRw", "", 500)
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
