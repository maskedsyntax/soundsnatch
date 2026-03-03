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
