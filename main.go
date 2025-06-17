package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// expandPath expands the user home directory in the path (e.g., ~/Music -> /home/user/Music)
func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %v", err)
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

// getVideoInfo fetches the video title using yt-dlp
func getVideoInfo(url string) (string, error) {
	// Run yt-dlp to get the video title
	cmd := exec.Command("yt-dlp", "--get-title", "--no-playlist", "--quiet", "--no-warnings", url)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("yt-dlp failed: %s", exitErr.Stderr)
		}
		return "", fmt.Errorf("failed to fetch video info: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// downloadAudio downloads the audio as MP3 using yt-dlp
func downloadAudio(url, outputPath string) error {
	// Run yt-dlp to download the audio
	cmd := exec.Command(
		"yt-dlp",
		"--format", "bestaudio/best",
		"--no-keep-video",
		"--output", outputPath,
		"--quiet",
		"--no-warnings",
		"--no-playlist",
		url,
	)
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("yt-dlp download failed: %s", exitErr.Stderr)
		}
		return fmt.Errorf("failed to download audio: %v", err)
	}
	return nil
}

func main() {
	// Create a scanner for user input
	scanner := bufio.NewScanner(os.Stdin)

	// Prompt for YouTube URL
	fmt.Print("Enter video URL: ")
	scanner.Scan()
	url := strings.TrimSpace(scanner.Text())
	if url == "" {
		fmt.Println("Error: URL cannot be empty")
		return
	}

	// Fetch video info (title)
	fmt.Println("Fetching video info...")
	videoTitle, err := getVideoInfo(url)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Prompt for destination path
	fmt.Print("Select Destination: ")
	scanner.Scan()
	path := strings.TrimSpace(scanner.Text())
	if path == "" {
		path = "~/Music"
	}
	fmt.Printf("Selected Path: %s\n", path)

	// Expand the path (e.g., ~/Music -> /home/user/Music)
	expandedPath, err := expandPath(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Prompt for filename
	fmt.Printf("Rename File? %s: ", videoTitle)
	scanner.Scan()
	filename := strings.TrimSpace(scanner.Text())
	if filename != "" {
		expandedPath = filepath.Join(expandedPath, filename+".mp3")
	} else {
		expandedPath = filepath.Join(expandedPath, videoTitle+".mp3")
	}

	// Ensure the directory exists
	dir := filepath.Dir(expandedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("Error: failed to create directory %s: %v\n", dir, err)
		return
	}

	// Download the audio
	fmt.Printf("Final audio file path: %s\n", expandedPath)
	fmt.Println("Downloading started...")
	if err := downloadAudio(url, expandedPath); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Download completed!")
}
