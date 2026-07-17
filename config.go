package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func loadConfig() Config {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".soundsnatch.yaml")
	archivePath := filepath.Join(home, ".soundsnatch_archive.txt")
	
	config := Config{
		LastSaveDir:   "",
		DefaultFormat: "mp3",
		Browser:       "",
		ArchivePath:   archivePath,
	}

	data, err := os.ReadFile(configPath)
	if err == nil {
		yaml.Unmarshal(data, &config)
	}

	// Always ensure ArchivePath is set to default if missing from config
	if config.ArchivePath == "" {
		config.ArchivePath = archivePath
	}

	return config
}

func saveConfig(config Config) {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".soundsnatch.yaml")
	data, err := yaml.Marshal(config)
	if err == nil {
		os.WriteFile(configPath, data, 0644)
	}
}

// resolveBrowser maps UI/config aliases to yt-dlp --cookies-from-browser values.
func resolveBrowser(browser string) string {
	if browser == "" || browser == "none" {
		return browser
	}
	if browser != "firefox-dev" {
		return browser
	}
	if profile := findFirefoxDevProfile(); profile != "" {
		return "firefox:" + profile
	}
	return "firefox"
}

func findFirefoxDevProfile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	profilesIni := filepath.Join(home, "Library", "Application Support", "Firefox", "profiles.ini")
	data, err := os.ReadFile(profilesIni)
	if err != nil {
		return ""
	}

	var currentName, currentPath string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") {
			if currentName == "dev-edition-default" && currentPath != "" {
				return currentPath
			}
			currentName = ""
			currentPath = ""
			continue
		}
		if strings.HasPrefix(line, "Name=") {
			currentName = strings.TrimPrefix(line, "Name=")
		}
		if strings.HasPrefix(line, "Path=") {
			currentPath = strings.TrimPrefix(line, "Path=")
			if strings.HasPrefix(currentPath, "Profiles/") {
				currentPath = strings.TrimPrefix(currentPath, "Profiles/")
			}
		}
	}
	if currentName == "dev-edition-default" && currentPath != "" {
		return currentPath
	}
	return ""
}
