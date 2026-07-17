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
	"sync"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func fetchInfoCmd(url string, browser string) tea.Cmd {
	return func() tea.Msg {
		args := []string{"-J", "--flat-playlist", "--no-warnings", "--quiet", "--ignore-errors", "--ignore-config", "--no-check-formats", "--js-runtimes", "node"}
		if browser != "" && browser != "none" {
			args = append(args, "--cookies-from-browser", resolveBrowser(browser))
		}
		args = append(args, url)

		cmd := exec.Command("yt-dlp", args...)
		if _, err := exec.LookPath("yt-dlp"); err != nil {
			cmd = exec.Command("python3", append([]string{"-m", "yt_dlp"}, args...)...)
		}

		var stdout, stderr strings.Builder
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		runErr := cmd.Run()

		outStr := strings.TrimSpace(stdout.String())

		if outStr != "" {
			var info map[string]interface{}
			if err := json.Unmarshal([]byte(outStr), &info); err == nil {
				title, _ := info["title"].(string)
				durFloat, _ := info["duration"].(float64)
				playlistCount := intFromInterface(info["playlist_count"])
				entryCount := 0
				if entries, ok := info["entries"].([]interface{}); ok {
					entryCount = len(entries)
				}
				if typ, _ := info["_type"].(string); typ == "playlist" && playlistCount > 0 && entryCount > 0 && entryCount < playlistCount {
					return errMsg{err: fmt.Errorf(
						"playlist truncated: yt-dlp returned %d of %d tracks. Upgrade yt-dlp (brew upgrade yt-dlp) and try again",
						entryCount, playlistCount,
					)}
				}
				return infoFetchedMsg{
					title:         title,
					duration:      durFloat,
					playlistCount: playlistCount,
					entryCount:    entryCount,
				}
			}
		}

		if runErr != nil {
			errMsgStr := strings.TrimSpace(stderr.String())
			if errMsgStr == "" {
				errMsgStr = runErr.Error()
			}
			return errMsg{err: fmt.Errorf("could not fetch info: %s", errMsgStr)}
		}

		return errMsg{err: fmt.Errorf("could not fetch info: no output from yt-dlp")}
	}
}

func searchCmd(query string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("yt-dlp", "ytsearch5:"+query, "-j", "--no-warnings", "--quiet", "--flat-playlist", "--ignore-config")
		if _, err := exec.LookPath("yt-dlp"); err != nil {
			cmd = exec.Command("python3", "-m", "yt_dlp", "ytsearch5:"+query, "-j", "--no-warnings", "--quiet", "--flat-playlist", "--ignore-config")
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

func startDownloadTask(c chan tea.Msg, url, saveDir, saveFilename, format, browser, archivePath string, expectedTracks int) {
	c <- statusMsg{text: "Starting download..."}
	if expectedTracks > 0 {
		c <- progressMsg{pct: 0, current: 0, total: expectedTracks}
	}

	outtmpl := filepath.Join(saveDir, saveFilename+"."+format)
	isPlaylist := isPlaylistURL(url)

	args := []string{
		"-f", "bestaudio[ext=m4a]/bestaudio[ext=webm]/bestaudio/best",
		"--extract-audio",
		"--audio-format", format,
		"--audio-quality", "0",
		"--no-check-formats",
		"--js-runtimes", "node",
		"--embed-metadata",
		"--no-warnings",
		"--newline",
		"--progress",
		"--ignore-errors",
		"--no-cache-dir",
		"--lazy-playlist",
		"--no-overwrites",
		"--ignore-config",
		"--retries", "10",
		"--fragment-retries", "10",
	}

	playlistDir := saveDir
	if isPlaylist {
		if saveFilename != "" {
			playlistDir = filepath.Join(saveDir, saveFilename)
		}
		os.MkdirAll(playlistDir, 0755)
		outtmpl = filepath.Join(playlistDir, "%(title)s.%(ext)s")
		localArchive := filepath.Join(playlistDir, ".archive.txt")
		args = append(args, "--download-archive", localArchive)
	} else {
		args = append(args, "--no-playlist")
		if archivePath != "" {
			args = append(args, "--download-archive", archivePath)
		}
	}

	args = append(args, "-o", outtmpl)

	if browser != "" && browser != "none" {
		args = append(args, "--cookies-from-browser", resolveBrowser(browser))
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
	stderr, err := cmd.StderrPipe()
	if err != nil {
		c <- errMsg{err: err}
		return
	}

	if err := cmd.Start(); err != nil {
		c <- errMsg{err: err}
		return
	}

	totItems := expectedTracks
	lastPct := 0.0
	skipCount := 0
	downloadIndex := 0
	var stderrBuf strings.Builder
	var stderrMu sync.Mutex

	sendProgress := func(pct float64, current, total int) {
		if pct < 0 {
			pct = 0
		}
		if pct > 1 {
			pct = 1
		}
		lastPct = pct
		if total > 0 {
			totItems = total
		}
		c <- progressMsg{pct: pct, current: current, total: totItems}
	}

	parseLine := func(line string) {
		if strings.Contains(line, "Downloading API JSON") || strings.Contains(line, "Downloading webpage") {
			c <- statusMsg{text: "Loading playlist from YouTube (this can take a few minutes)..."}
		}
		if m := playlistTotalRe.FindStringSubmatch(line); len(m) > 2 {
			total, _ := strconv.Atoi(m[2])
			sendProgress(float64(skipCount)/float64(total), skipCount, total)
			return
		}
		if archiveSkipRe.MatchString(line) {
			skipCount++
			c <- statusMsg{text: "Skipping already downloaded tracks..."}
			if totItems > 0 {
				sendProgress(float64(skipCount)/float64(totItems), skipCount, totItems)
			}
			return
		}
		if destinationRe.MatchString(line) {
			downloadIndex++
			current := skipCount + downloadIndex
			if totItems > 0 {
				sendProgress(float64(current-1)/float64(totItems), current, totItems)
			} else {
				sendProgress(0, current, totItems)
			}
			c <- statusMsg{text: ""}
			return
		}
		if m := itemRe.FindStringSubmatch(line); len(m) > 2 {
			current, _ := strconv.Atoi(m[1])
			total, _ := strconv.Atoi(m[2])
			sendProgress(lastPct, current, total)
			return
		}
		if m := progressRe.FindStringSubmatch(line); len(m) > 1 {
			filePct, _ := strconv.ParseFloat(m[1], 64)
			current := skipCount + downloadIndex
			if current == 0 {
				current = skipCount
			}
			overall := filePct / 100.0
			if totItems > 0 && downloadIndex > 0 {
				overall = (float64(skipCount+downloadIndex-1) + filePct/100.0) / float64(totItems)
			} else if totItems > 0 && skipCount > 0 && downloadIndex == 0 {
				overall = float64(skipCount) / float64(totItems)
			}
			sendProgress(overall, current, totItems)
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		scan := bufio.NewScanner(stdout)
		scan.Buffer(make([]byte, 64*1024), 1024*1024)
		for scan.Scan() {
			parseLine(scan.Text())
		}
	}()
	go func() {
		defer wg.Done()
		scan := bufio.NewScanner(stderr)
		scan.Buffer(make([]byte, 64*1024), 1024*1024)
		for scan.Scan() {
			line := scan.Text()
			stderrMu.Lock()
			stderrBuf.WriteString(line)
			stderrBuf.WriteByte('\n')
			stderrMu.Unlock()
			parseLine(line)
		}
	}()

	wg.Wait()
	runErr := cmd.Wait()

	if runErr != nil {
		errMsgStr := strings.TrimSpace(stderrBuf.String())
		if errMsgStr != "" && !isPlaylist {
			c <- errMsg{err: fmt.Errorf("download failed: %s", errMsgStr)}
			return
		}
	}

	doneMsg := fmt.Sprintf("🎉 Sync Complete!\nYour library at: %s is now up to date.", playlistDir)
	if isPlaylist {
		saved := countMediaFiles(playlistDir, format)
		total := expectedTracks
		if total == 0 {
			total = totItems
		}
		if total > 0 && saved < total {
			doneMsg = fmt.Sprintf(
				"⚠️ Partial sync: %d of %d tracks saved.\nRe-run the same playlist to continue — finished tracks are skipped.\nLocation: %s",
				saved, total, playlistDir,
			)
		}
	}

	c <- downloadDoneMsg{message: doneMsg}
}

func waitForMsg(c chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <-c
	}
}

func intFromInterface(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}

func countMediaFiles(dir, format string) int {
	matches, err := filepath.Glob(filepath.Join(dir, "*."+format))
	if err != nil {
		return 0
	}
	return len(matches)
}
