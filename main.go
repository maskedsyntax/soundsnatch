package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

type state int

const (
	stateInputURL state = iota
	stateFetching
	stateSearching
	statePickSearchResult
	stateInfo
	statePickDir
	stateCreateDir
	stateInputFilename
	statePickFormat
	stateDownloading
	stateDone
	stateError
)

var (
	titleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	boldStyle    = lipgloss.NewStyle().Bold(true)
	progressRe   = regexp.MustCompile(`(\d+(\.\d+)?)%`)
	urlRe        = regexp.MustCompile(`^https?://`)
)

type Config struct {
	LastSaveDir   string `yaml:"last_save_dir"`
	DefaultFormat string `yaml:"default_format"`
}

type formatItem struct {
	label string
	ext   string
}

func (i formatItem) Title() string       { return i.label }
func (i formatItem) Description() string { return "Download as " + i.ext }
func (i formatItem) FilterValue() string { return i.label }

type searchResultItem struct {
	title string
	url   string
	dur   string
}

func (i searchResultItem) Title() string       { return i.title }
func (i searchResultItem) Description() string { return "Duration: " + i.dur + " | " + i.url }
func (i searchResultItem) FilterValue() string { return i.title }

type model struct {
	state         state
	urlInput      textinput.Model
	filenameInput textinput.Model
	mkdirInput    textinput.Model
	spinner       spinner.Model
	filepicker    filepicker.Model
	progress      progress.Model
	formatList    list.Model
	searchList    list.Model

	url           string
	videoTitle    string
	videoDuration float64
	saveDir       string
	saveFilename  string
	selectedFormat string

	downloadPercent float64
	err             error
	doneMessage     string
	lastWindowHeight int
	lastWindowWidth  int

	msgChan chan tea.Msg
	config  Config
}

type infoFetchedMsg struct {
	title    string
	duration float64
}

type searchResultsMsg []list.Item

type progressMsg float64

type downloadDoneMsg struct {
	message string
}

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func loadConfig() Config {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".soundsnatch.yaml")
	
	config := Config{
		LastSaveDir:   "",
		DefaultFormat: "mp3",
	}

	data, err := os.ReadFile(configPath)
	if err == nil {
		yaml.Unmarshal(data, &config)
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

func initialModel() model {
	config := loadConfig()

	uInput := textinput.New()
	uInput.Placeholder = "URL or search query..."
	uInput.Focus()
	uInput.CharLimit = 256
	uInput.Width = 60

	fInput := textinput.New()
	fInput.Placeholder = "filename (without extension)"
	fInput.CharLimit = 100
	fInput.Width = 60

	mInput := textinput.New()
	mInput.Placeholder = "folder name"
	mInput.CharLimit = 100
	mInput.Width = 60

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	fp := filepicker.New()
	fp.DirAllowed = true
	fp.FileAllowed = false
	fp.ShowHidden = false
	
	if config.LastSaveDir != "" {
		fp.CurrentDirectory = config.LastSaveDir
	} else {
		fp.CurrentDirectory, _ = os.UserHomeDir()
		if fp.CurrentDirectory == "" {
			fp.CurrentDirectory = "."
		}
	}
	
	fp.KeyMap.Select = key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "select highlighted"),
	)
	fp.Height = 10

	prog := progress.New(progress.WithDefaultGradient())

	items := []list.Item{
		formatItem{label: "MP3 (Standard)", ext: "mp3"},
		formatItem{label: "FLAC (Lossless)", ext: "flac"},
		formatItem{label: "WAV (High Quality)", ext: "wav"},
	}
	fl := list.New(items, list.NewDefaultDelegate(), 0, 0)
	fl.Title = "Select Audio Format"
	fl.SetShowStatusBar(false)
	fl.SetFilteringEnabled(false)
	fl.Styles.Title = titleStyle

	sl := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	sl.Title = "Search Results"
	sl.Styles.Title = titleStyle

	return model{
		state:         stateInputURL,
		urlInput:      uInput,
		filenameInput: fInput,
		mkdirInput:    mInput,
		spinner:       s,
		filepicker:    fp,
		progress:      prog,
		formatList:    fl,
		searchList:    sl,
		msgChan:       make(chan tea.Msg),
		config:        config,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

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

func waitForMsg(c chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <-c
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.lastWindowHeight = msg.Height
		m.lastWindowWidth = msg.Width
		m.progress.Width = msg.Width - 10
		if m.progress.Width > 80 {
			m.progress.Width = 80
		}
		m.formatList.SetSize(msg.Width-4, msg.Height-8)
		m.searchList.SetSize(msg.Width-4, msg.Height-8)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	case errMsg:
		m.err = msg.err
		m.state = stateError
		return m, nil
	case infoFetchedMsg:
		m.videoTitle = msg.title
		m.videoDuration = msg.duration
		m.state = stateInfo

		cleanTitle := strings.ReplaceAll(m.videoTitle, "/", "_")
		m.filenameInput.SetValue(cleanTitle)
		return m, nil
	case searchResultsMsg:
		m.searchList.SetItems(msg)
		m.state = statePickSearchResult
		return m, nil
	case progressMsg:
		m.downloadPercent = float64(msg)
		return m, waitForMsg(m.msgChan)
	case downloadDoneMsg:
		m.doneMessage = msg.message
		m.state = stateDone
		return m, nil
	case spinner.TickMsg:
		if m.state == stateFetching || m.state == stateDownloading || m.state == stateSearching {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if newProg, ok := newModel.(progress.Model); ok {
			m.progress = newProg
		}
		return m, cmd
	}

	if _, isKey := msg.(tea.KeyMsg); !isKey {
		m.filepicker, cmd = m.filepicker.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch m.state {
	case stateInputURL:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.Type == tea.KeyEnter {
				input := strings.TrimSpace(m.urlInput.Value())
				if input != "" {
					if urlRe.MatchString(input) {
						m.url = input
						m.state = stateFetching
						return m, tea.Batch(m.spinner.Tick, fetchInfoCmd(m.url))
					} else {
						m.state = stateSearching
						return m, tea.Batch(m.spinner.Tick, searchCmd(input))
					}
				}
			}
		}
		m.urlInput, cmd = m.urlInput.Update(msg)
		cmds = append(cmds, cmd)

	case statePickSearchResult:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.Type == tea.KeyEnter {
				i, ok := m.searchList.SelectedItem().(searchResultItem)
				if ok {
					m.url = i.url
					m.state = stateFetching
					return m, tea.Batch(m.spinner.Tick, fetchInfoCmd(m.url))
				}
			}
		}
		m.searchList, cmd = m.searchList.Update(msg)
		cmds = append(cmds, cmd)

	case stateInfo:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.Type == tea.KeyEnter {
				m.state = statePickDir
				return m, m.filepicker.Init()
			}
		}

	case statePickDir:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "S" {
				m.saveDir = m.filepicker.CurrentDirectory
				m.state = stateInputFilename
				m.filenameInput.Focus()
				
				// Save to config
				m.config.LastSaveDir = m.saveDir
				saveConfig(m.config)
				
				return m, textinput.Blink
			}
			if msg.String() == "n" {
				m.state = stateCreateDir
				m.mkdirInput.Focus()
				return m, textinput.Blink
			}

			m.filepicker, cmd = m.filepicker.Update(msg)
			cmds = append(cmds, cmd)

			if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
				m.saveDir = path
				m.state = stateInputFilename
				m.filenameInput.Focus()
				
				// Save to config
				m.config.LastSaveDir = m.saveDir
				saveConfig(m.config)
				
				return m, textinput.Blink
			}
		}

	case stateCreateDir:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.Type == tea.KeyEnter {
				newDirName := strings.TrimSpace(m.mkdirInput.Value())
				if newDirName != "" {
					newPath := filepath.Join(m.filepicker.CurrentDirectory, newDirName)
					err := os.MkdirAll(newPath, 0755)
					if err != nil {
						m.err = fmt.Errorf("failed to create directory: %v", err)
						m.state = stateError
						return m, nil
					}
					m.mkdirInput.Reset()
					m.filepicker.CurrentDirectory = newPath
					m.state = statePickDir
					return m, m.filepicker.Init()
				}
			}
			if msg.Type == tea.KeyEsc {
				m.state = statePickDir
				return m, nil
			}
		}
		m.mkdirInput, cmd = m.mkdirInput.Update(msg)
		cmds = append(cmds, cmd)

	case stateInputFilename:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.Type == tea.KeyEnter {
				m.saveFilename = strings.TrimSpace(m.filenameInput.Value())
				if m.saveFilename != "" {
					m.state = statePickFormat
					return m, nil
				}
			}
		}
		m.filenameInput, cmd = m.filenameInput.Update(msg)
		cmds = append(cmds, cmd)

	case statePickFormat:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.Type == tea.KeyEnter {
				i, ok := m.formatList.SelectedItem().(formatItem)
				if ok {
					m.selectedFormat = i.ext
					m.state = stateDownloading
					go startDownloadTask(m.msgChan, m.url, m.saveDir, m.saveFilename, m.selectedFormat)
					return m, tea.Batch(m.spinner.Tick, waitForMsg(m.msgChan))
				}
			}
		}
		m.formatList, cmd = m.formatList.Update(msg)
		cmds = append(cmds, cmd)

	case stateDone:
		switch msg.(type) {
		case tea.KeyMsg:
			return m, tea.Quit
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v\n\nPress Esc to quit.", m.err))
	}

	var sections []string
	sections = append(sections, titleStyle.Render("🎵 SoundSnatch TUI"))

	switch m.state {
	case stateInputURL:
		sections = append(sections, "Enter video URL or search query:", m.urlInput.View())
		sections = append(sections, helpStyle.Render("Press Enter to continue, Esc to quit."))

	case stateFetching:
		sections = append(sections, fmt.Sprintf("%s Fetching info...", m.spinner.View()))

	case stateSearching:
		sections = append(sections, fmt.Sprintf("%s Searching YouTube...", m.spinner.View()))

	case statePickSearchResult:
		sections = append(sections, m.searchList.View())

	case stateInfo:
		sections = append(sections, lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render("✨ Info fetched:"))
		sections = append(sections, infoStyle.Render(fmt.Sprintf("Title:    %s", m.videoTitle)))
		minutes := int(m.videoDuration) / 60
		seconds := int(m.videoDuration) % 60
		sections = append(sections, infoStyle.Render(fmt.Sprintf("Duration: %02d:%02d", minutes, seconds)))
		sections = append(sections, helpStyle.Render("Press Enter to choose destination directory, Esc to quit."))

	case statePickDir:
		sections = append(sections, fmt.Sprintf("%s %s", boldStyle.Render("Browsing:"), m.filepicker.CurrentDirectory))
		sections = append(sections, helpStyle.Render("Nav: ↑/k, ↓/j, Enter/→ (open) | Select: s (highlight), S (current) | n: New | Esc: Quit"))
		sections = append(sections, "\nWhere would you like to save your audio file?")
		
		chrome := 0
		for _, s := range sections {
			chrome += lipgloss.Height(s)
		}
		chrome += 4
		m.filepicker.Height = m.lastWindowHeight - chrome
		if m.filepicker.Height < 3 {
			m.filepicker.Height = 3
		}
		sections = append(sections, m.filepicker.View())

	case stateCreateDir:
		sections = append(sections, fmt.Sprintf("Create folder in: %s", m.filepicker.CurrentDirectory))
		sections = append(sections, m.mkdirInput.View())
		sections = append(sections, helpStyle.Render("Press Enter to create and enter, Esc to cancel."))

	case stateInputFilename:
		sections = append(sections, "📝 Name your download:", m.filenameInput.View())
		sections = append(sections, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(fmt.Sprintf("Destination: %s", m.saveDir)))
		sections = append(sections, helpStyle.Render("Press Enter to select format, Esc to quit."))

	case statePickFormat:
		sections = append(sections, m.formatList.View())

	case stateDownloading:
		sections = append(sections, fmt.Sprintf("%s Downloading and converting to %s...", m.spinner.View(), strings.ToUpper(m.selectedFormat)))
		sections = append(sections, infoStyle.Render(fmt.Sprintf("Target: %s", m.videoTitle)))
		sections = append(sections, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(fmt.Sprintf("Saving to: %s", m.saveDir)))
		
		progView := m.progress.ViewAs(m.downloadPercent)
		sections = append(sections, "\n"+progView)

	case stateDone:
		sections = append(sections, successStyle.Render(m.doneMessage))
		sections = append(sections, helpStyle.Render("Press any key to exit."))
	}

	return lipgloss.NewStyle().Margin(1, 2).Render(lipgloss.JoinVertical(lipgloss.Left, sections...))
}

func main() {
	m := initialModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
