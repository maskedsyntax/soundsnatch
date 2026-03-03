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

func fetchInfoCmd(url string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("yt-dlp", "-j", "--no-warnings", "--quiet", url)
		if _, err := exec.LookPath("yt-dlp"); err != nil {
			cmd = exec.Command("python3", "-m", "yt_dlp", "-j", "--no-warnings", "--quiet", url)
		}

		var stderr strings.Builder
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		if err != nil {
			errMsgStr := stderr.String()
			if errMsgStr == "" {
				errMsgStr = err.Error()
			}
			return errMsg{err: fmt.Errorf("could not fetch video info: %s", errMsgStr)}
		}

		var info map[string]interface{}
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) == 0 {
			return errMsg{err: fmt.Errorf("no info returned")}
		}
		if err := json.Unmarshal([]byte(lines[0]), &info); err != nil {
			return errMsg{err: fmt.Errorf("failed to parse video info: %v", err)}
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

func startDownloadTask(c chan tea.Msg, url, saveDir, saveFilename, format string) {
	outtmpl := filepath.Join(saveDir, saveFilename+"."+format)
	isPlaylist := strings.Contains(url, "list=") || strings.Contains(url, "playlist")
	if isPlaylist {
		outtmpl = filepath.Join(saveDir, saveFilename, "%(title)s.%(ext)s")
		os.MkdirAll(filepath.Join(saveDir, saveFilename), 0755)
	}

	cmd := exec.Command("yt-dlp", "-f", "bestaudio/best", "--extract-audio", "--audio-format", format, "-o", outtmpl, "--no-warnings", "--newline", "--progress", url)
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		cmd = exec.Command("python3", "-m", "yt_dlp", "-f", "bestaudio/best", "--extract-audio", "--audio-format", format, "-o", outtmpl, "--no-warnings", "--newline", "--progress", url)
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
	for scanner.Scan() {
		line := scanner.Text()
		matches := progressRe.FindStringSubmatch(line)
		if len(matches) > 1 {
			pct, _ := strconv.ParseFloat(matches[1], 64)
			c <- progressMsg(pct / 100.0)
		}
	}

	if err := cmd.Wait(); err != nil {
		errMsgStr := stderr.String()
		if errMsgStr == "" {
			errMsgStr = err.Error()
		}
		c <- errMsg{err: fmt.Errorf("download failed: %s", errMsgStr)}
		return
	}

	c <- downloadDoneMsg{message: fmt.Sprintf("🎉 Download Complete!\nFiles saved to: %s", saveDir)}
}

func waitForMsg(c chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <-c
	}
}
