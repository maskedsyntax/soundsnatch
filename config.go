package main

import (
	"os"
	"path/filepath"

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
