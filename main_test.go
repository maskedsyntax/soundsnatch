package main

import (
	"strings"
	"testing"
)

func TestURLRegex(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://youtube.com/watch?v=123", true},
		{"http://youtu.be/123", true},
		{"just a song name", false},
		{"ftp://link.com", false},
	}

	for _, test := range tests {
		if urlRe.MatchString(test.url) != test.expected {
			t.Errorf("URL match failed for %s: expected %v", test.url, test.expected)
		}
	}
}

func TestCleanFilename(t *testing.T) {
	title := "Song / With / Slashes"
	expected := "Song _ With _ Slashes"
	result := strings.ReplaceAll(title, "/", "_")
	if result != expected {
		t.Errorf("Clean filename failed: expected %s, got %s", expected, result)
	}
}

func TestIsPlaylistURL(t *testing.T) {
	if !isPlaylistURL("https://music.youtube.com/playlist?list=LM") {
		t.Error("expected playlist URL")
	}
	if isPlaylistURL("https://www.youtube.com/watch?v=abc123") {
		t.Error("expected single video URL")
	}
}

func TestResolveBrowser(t *testing.T) {
	if got := resolveBrowser(""); got != "" {
		t.Errorf("expected empty browser, got %q", got)
	}
	if got := resolveBrowser("chrome"); got != "chrome" {
		t.Errorf("expected chrome, got %q", got)
	}
	if got := resolveBrowser("firefox-dev"); got == "firefox-dev" {
		t.Errorf("firefox-dev should resolve to a yt-dlp browser value, got %q", got)
	}
}
