package main

import (
	"regexp"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	statePickBrowser
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
	Browser       string `yaml:"browser"`
}

type formatItem struct {
	label string
	ext   string
}

func (i formatItem) Title() string       { return i.label }
func (i formatItem) Description() string { return "Download as " + i.ext }
func (i formatItem) FilterValue() string { return i.label }

type browserItem string

func (i browserItem) Title() string       { return string(i) }
func (i browserItem) Description() string { 
	if i == "none" {
		return "Do not use browser cookies"
	}
	return "Extract cookies from " + string(i) 
}
func (i browserItem) FilterValue() string { return string(i) }

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
	browserList   list.Model

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
