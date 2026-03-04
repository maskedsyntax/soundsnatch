package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func fetchInfoCmd(url string, browser string) tea.Cmd {
	return func() tea.Msg {
		args := []string{"-J", "--flat-playlist", "--no-warnings", "--quiet", "--ignore-errors"}
		if browser != "" && browser != "none" {
			args = append(args, "--cookies-from-browser", browser)
		}
		args = append(args, url)

		cmd := exec.Command("yt-dlp", args...)
		if _, err := exec.LookPath("yt-dlp"); err != nil {
			cmd = exec.Command("python3", append([]string{"-m", "yt_dlp"}, args...)...)
		}

		var stderr strings.Builder
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		if err != nil {
			errMsgStr := stderr.String()
			if errMsgStr == "" {
				errMsgStr = err.Error()
			}
			return errMsg{err: fmt.Errorf("could not fetch info: %s", errMsgStr)}
		}

		var info map[string]interface{}
		if err := json.Unmarshal(out, &info); err != nil {
			return errMsg{err: fmt.Errorf("failed to parse info: %v", err)}
		}

		title, _ := info["title"].(string)
		durFloat, _ := info["duration"].(float64)

		return infoFetchedMsg{
			title:    title,
			duration: durFloat,
		}
	}
}

func searchCmd(query string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("yt-dlp", "ytsearch5:"+query, "-j", "--no-warnings", "--quiet", "--flat-playlist")
		if _, err := exec.LookPath("yt-dlp"); err != nil {
			cmd = exec.Command("python3", "-m", "yt_dlp", "ytsearch5:"+query, "-j", "--no-warnings", "--quiet", "--flat-playlist")
		}

		out, err := cmd.Output()
		if err != nil {
			return errMsg{err: fmt.Errorf("search failed: %v", err)}
		}

		var results []list.Item
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			var info map[string]interface{}
			if err := json.Unmarshal([]byte(line), &info); err != nil {
				continue
			}
			title, _ := info["title"].(string)
			id, _ := info["id"].(string)
			url := "https://www.youtube.com/watch?v=" + id
			durFloat, _ := info["duration"].(float64)
			dur := fmt.Sprintf("%02d:%02d", int(durFloat)/60, int(durFloat)%60)
			
			results = append(results, searchResultItem{
				title: title,
				url:   url,
				dur:   dur,
			})
		}

		if len(results) == 0 {
			return errMsg{err: fmt.Errorf("no results found")}
		}

		return searchResultsMsg(results)
	}
}

func startDownloadTask(c chan tea.Msg, url, saveDir, saveFilename, format, browser string) {
	outtmpl := filepath.Join(saveDir, saveFilename+"."+format)
	isPlaylist := strings.Contains(url, "list=") || strings.Contains(url, "playlist")
	
	args := []string{
		"-f", "bestaudio/best",
		"--extract-audio",
		"--audio-format", format,
		"--no-warnings",
		"--newline",
		"--progress",
		"--ignore-errors",
		"--no-cache-dir",
		"--lazy-playlist",
		"--no-overwrites",
	}

	if isPlaylist {
		playlistDir := filepath.Join(saveDir, saveFilename)
		os.MkdirAll(playlistDir, 0755)
		
		// LOCAL ARCHIVE: Keep track of downloads INSIDE the playlist folder
		// This prevents duplicates within the folder but allows downloading 
		// the same song to a DIFFERENT folder.
		archivePath := filepath.Join(playlistDir, ".soundsnatch_archive.txt")
		args = append(args, "--download-archive", archivePath)
		
		outtmpl = filepath.Join(playlistDir, "%(title)s [%(id)s].%(ext)s")
	} else {
		args = append(args, "--no-playlist")
	}

	args = append(args, "-o", outtmpl)

	if browser != "" && browser != "none" {
		args = append(args, "--cookies-from-browser", browser)
	}
	args = append(args, url)

	cmd := exec.Command("yt-dlp", args...)
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		cmd = exec.Command("python3", append([]string{"-m", "yt_dlp"}, args...)...)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		c <- errMsg{err: err}
		return
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		c <- errMsg{err: err}
		return
	}

	scanner := bufio.NewScanner(stdout)
	curItem := 0
	totItems := 0
	for scanner.Scan() {
		line := scanner.Text()
		
		// Parse playlist status
		if m := itemRe.FindStringSubmatch(line); len(m) > 2 {
			curItem, _ = strconv.Atoi(m[1])
			totItems, _ = strconv.Atoi(m[2])
		}

		// Parse progress
		if m := progressRe.FindStringSubmatch(line); len(m) > 1 {
			pct, _ := strconv.ParseFloat(m[1], 64)
			c <- progressMsg{
				pct:     pct / 100.0,
				current: curItem,
				total:   totItems,
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		errMsgStr := stderr.String()
		if errMsgStr != "" && !isPlaylist {
			c <- errMsg{err: fmt.Errorf("download failed: %s", errMsgStr)}
			return
		}
	}

	c <- downloadDoneMsg{message: fmt.Sprintf("🎉 Sync Complete!\nFiles are in: %s", saveDir)}
}

func waitForMsg(c chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <-c
	}
}
