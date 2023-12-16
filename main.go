package main

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"tezaurs/parser"
	"tezaurs/util"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// styles
var (
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	titleStyle  = lipgloss.NewStyle().Bold(true).Render
	numStrStyle = lipgloss.NewStyle().Bold(true).Render
)

type currentView uint

const (
	VIEW_MAIN currentView = iota
)

type model struct {
	// the current view of the program
	view currentView
	err  error

	// terminal size
	width  int
	height int

	// VIEW_MAIN
	word        string
	url         string
	mainContent string
	parsedHtml  parser.ParsedHtml

	viewport viewport.Model

	// http
	status     int
	isFetching bool
	httpErr    error

	exitErr error
}

func initialModel() model {
	return model{}
}

func (m model) headerView() string {
	var s string

	s += fmt.Sprintf(" %s\n", titleStyle(m.word))
	s += strings.Repeat("─", m.width)

	return s
}

func (m model) footerView() string {
	var s string

	// s += strings.Repeat("─", m.width) + "\n"
	s += "\n"
	s += helpStyle(" ↑/k: Uz augšu • ↓/j: Uz leju • m: Meklēt vardu • q: Iziet")

	return s
}

func (m model) renderContent() string {
	var s string

	for _, v := range m.parsedHtml.Entries {
		re := regexp.MustCompile("\\d+")

		matches := re.FindAllString(v.NumStr, -1)

		var indents int

		if len(matches) <= 1 {
			indents = 0
		} else {
			indents = len(matches) - 1
		}

		indentSize := indents * 4

		style := lipgloss.NewStyle().MarginLeft(indentSize).Width(m.width - indentSize)

		numStr := v.NumStr

		if numStr != "" {
			numStr = numStrStyle(numStr)
		}

		s += style.Render(strings.TrimSpace(fmt.Sprintf("%s %s", numStr, v.Content))) + "\n\n"
	}

	return s
}

func (m model) Init() tea.Cmd {
	return util.OpenFzfCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		case "m":
			return m, tea.Batch(util.OpenFzfCmd())
		}

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())

		verticalMarginHeight := headerHeight + footerHeight

		m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
		m.viewport.YPosition = headerHeight + 1
		// m.viewport.HighPerformanceRendering = useHighPerformanceRenderer
		m.mainContent = m.renderContent()

		m.viewport.SetContent(m.mainContent)

	case util.FzfFinished:
		switch msg.ExitCode {
		case 127:
			return m, tea.Quit
		case 130:
			if m.word == "" {
				return m, tea.Quit
			}
		}

		if msg.ExitCode >= 2 && msg.Err != nil {
			m.exitErr = msg.Err
			return m, tea.Quit
		}

		m.word = msg.Word
		m.url = "https://tezaurs.lv/" + url.PathEscape(msg.Word)

		m.isFetching = true
		return m, util.FetchTezaursCmd(m.url)

	case util.TezaursResponse:
		m.isFetching = false
		m.status = msg.Code

		parsed, err := parser.ParseHtml(msg.Body)

		if err != nil {
			return m, nil
		}

		m.parsedHtml = parsed
		m.mainContent = m.renderContent()

		m.viewport.SetContent(m.mainContent)
		return m, nil

	case util.FileErr:
		m.exitErr = msg.Err
		return m, tea.Quit

	case util.TezErr:
		m.isFetching = false
		m.httpErr = msg.Err
		return m, nil
	}

	if m.view == VIEW_MAIN {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.exitErr != nil {
		return m.exitErr.Error()
	}

	switch m.view {

	case VIEW_MAIN:
		return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())

	}

	// just so the compiler isn't mad
	return ""
}

func main() {
	m := initialModel()

	p := tea.NewProgram(m, tea.WithMouseCellMotion(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("kļūda: ", err)
		os.Exit(1)
	}
}
