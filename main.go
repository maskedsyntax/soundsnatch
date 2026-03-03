package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateInputURL state = iota
	stateFetching
	stateInfo
	statePickDir
	stateCreateDir
	stateInputFilename
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
)

type model struct {
	state         state
	urlInput      textinput.Model
	filenameInput textinput.Model
	mkdirInput    textinput.Model
	spinner       spinner.Model
	filepicker    filepicker.Model

	url           string
	videoTitle    string
	videoDuration float64
	saveDir       string
	saveFilename  string

	err         error
	doneMessage string
	lastWindowHeight int
}

type infoFetchedMsg struct {
	title    string
	duration float64
}

type downloadDoneMsg struct {
	message string
}

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func initialModel() model {
	uInput := textinput.New()
	uInput.Placeholder = "https://youtube.com/..."
	uInput.Focus()
	uInput.CharLimit = 256
	uInput.Width = 60

	fInput := textinput.New()
	fInput.Placeholder = "filename (without .mp3)"
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
	fp.CurrentDirectory, _ = os.UserHomeDir()
	if fp.CurrentDirectory == "" {
		fp.CurrentDirectory = "."
	}
	fp.KeyMap.Select = key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "select highlighted"),
	)
	fp.Height = 10

	return model{
		state:         stateInputURL,
		urlInput:      uInput,
		filenameInput: fInput,
		mkdirInput:    mInput,
		spinner:       s,
		filepicker:    fp,
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

func downloadCmd(url, saveDir, saveFilename string) tea.Cmd {
	return func() tea.Msg {
		if saveFilename == "" {
			return errMsg{err: fmt.Errorf("filename cannot be empty")}
		}

		outtmpl := filepath.Join(saveDir, saveFilename+".mp3")
		if strings.Contains(url, "list=") || strings.Contains(url, "playlist") {
			outtmpl = filepath.Join(saveDir, saveFilename, "%(title)s.%(ext)s")
			os.MkdirAll(filepath.Join(saveDir, saveFilename), 0755)
		}

		cmd := exec.Command("yt-dlp", "-f", "bestaudio/best", "--extract-audio", "--audio-format", "mp3", "-o", outtmpl, "--no-warnings", "--quiet", url)
		if _, err := exec.LookPath("yt-dlp"); err != nil {
			cmd = exec.Command("python3", "-m", "yt_dlp", "-f", "bestaudio/best", "--extract-audio", "--audio-format", "mp3", "-o", outtmpl, "--no-warnings", "--quiet", url)
		}

		var stderr strings.Builder
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			errMsgStr := stderr.String()
			if errMsgStr == "" {
				errMsgStr = err.Error()
			}
			return errMsg{err: fmt.Errorf("download failed: %s", errMsgStr)}
		}

		return downloadDoneMsg{message: fmt.Sprintf("🎉 Download Complete!\nFiles saved to: %s", saveDir)}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.lastWindowHeight = msg.Height
		// Re-calculate height in state transitions as well, but this is the base.
		m.filepicker.Height = msg.Height - 12
		if m.filepicker.Height < 3 {
			m.filepicker.Height = 3
		}

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
	case downloadDoneMsg:
		m.doneMessage = msg.message
		m.state = stateDone
		return m, nil
	case spinner.TickMsg:
		if m.state == stateFetching || m.state == stateDownloading {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
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
				m.url = strings.TrimSpace(m.urlInput.Value())
				if m.url != "" {
					m.state = stateFetching
					return m, tea.Batch(m.spinner.Tick, fetchInfoCmd(m.url))
				}
			}
		}
		m.urlInput, cmd = m.urlInput.Update(msg)
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
					m.state = stateDownloading
					return m, tea.Batch(m.spinner.Tick, downloadCmd(m.url, m.saveDir, m.saveFilename))
				}
			}
		}
		m.filenameInput, cmd = m.filenameInput.Update(msg)
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
		sections = append(sections, "Enter video URL:", m.urlInput.View())
		sections = append(sections, helpStyle.Render("Press Enter to continue, Esc to quit."))

	case stateFetching:
		sections = append(sections, fmt.Sprintf("%s Fetching info...", m.spinner.View()))

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
		
		// Re-calculate the exact height needed for the filepicker based on the content above it
		chrome := 0
		for _, s := range sections {
			chrome += lipgloss.Height(s)
		}
		chrome += 4 // Safety for margins and padding
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
		sections = append(sections, helpStyle.Render("Press Enter to download, Esc to quit."))

	case stateDownloading:
		sections = append(sections, fmt.Sprintf("%s Downloading and converting...", m.spinner.View()))
		sections = append(sections, infoStyle.Render(fmt.Sprintf("Target: %s", m.videoTitle)))
		sections = append(sections, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(fmt.Sprintf("Saving to: %s", m.saveDir)))

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
