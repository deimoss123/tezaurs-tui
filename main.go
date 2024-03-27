package main

import (
	"flag"
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

var tableBorder = lipgloss.Border{
	Top:         "─",
	Bottom:      "─",
	Left:        "│",
	Right:       "│",
	TopLeft:     "│",
	TopRight:    "│",
	BottomLeft:  "│",
	BottomRight: "│",
}

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

	testText string
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

	if m.parsedHtml.Verbalisation != "" {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(m.parsedHtml.Verbalisation) + "\n\n"
	}

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

	// renderējam tabulu
	if m.parsedHtml.ConjTable.ColumnCount > 0 {
		var rowStrings []string

		// renderējam rindas
		for rowIdx := range m.parsedHtml.ConjTable.RowItems {
			var colStrings []string

			for colIdx := range m.parsedHtml.ConjTable.RowItems[rowIdx] {
				rowItem := m.parsedHtml.ConjTable.RowItems[rowIdx][colIdx]

				if rowItem.ColSpan == 0 {
					continue
				}

				width := rowItem.ColSpan - 1

				for i := 0; i < rowItem.ColSpan; i++ {
					width += m.parsedHtml.ConjTable.ColItems[colIdx+i].Width + 1
				}

				alignment := lipgloss.Left
				if rowItem.ColSpan > 1 {
					alignment = lipgloss.Center
				}

				textColor := lipgloss.Color("15")

				if rowItem.IsHeading {
					textColor = lipgloss.Color("12")
				}

				style := lipgloss.NewStyle().
					BorderStyle(tableBorder).
					BorderForeground(lipgloss.Color("8")).
					BorderRight(true).
					BorderLeft(colIdx == 0).
					BorderBottom(rowItem.IsThead).
					Foreground(textColor).
					Align(alignment).
					Width(width)

				colStrings = append(colStrings, style.Render(rowItem.Text))
			}

			rowStrings = append(rowStrings, lipgloss.JoinHorizontal(0, colStrings...))
		}

		tableStr := lipgloss.JoinVertical(0, rowStrings...)

		s += lipgloss.NewStyle().Underline(true).Render("Locīšana:") + "\n\n"

		s += lipgloss.NewStyle().
			// BorderStyle(lipgloss.NormalBorder()).
			// BorderForeground(lipgloss.Color("241")).
			// BorderTop(true).
			// BorderBottom(true).
			Render(tableStr) + "\n\n"

	}

	// if m.parsedHtml.TestText != "" {
	// 	s += lipgloss.NewStyle().Width(m.width).Render(m.parsedHtml.TestText) + "\n"
	// }

	// s += m.testText + "\n"

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

	return ""
}

func main() {
	textOnly := flag.Bool("t", false, "Printēt tikai tekstu, bez saskarnes")
	help := flag.Bool("h", false, "Palīdzība")

	flag.Parse()

	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *textOnly {
		if len(flag.Args()) == 0 {
			fmt.Print("Kļūda: izmantojot -t karogu ir jānorāda vārds kuru meklēt\n\n")
			fmt.Print("Piemērs:\n\n")
			fmt.Print("tezaurs -t biezpiens\n")
			os.Exit(0)
		}

		word := flag.Args()[0]

		res, err := util.FetchTezaurs("https://tezaurs.lv/" + url.PathEscape(word))

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		parsed, err := parser.ParseHtml(res.Body)

		if len(parsed.Entries) == 0 {
			fmt.Printf("Vārds \"%s\" netika atrasts\n", word)
			os.Exit(0)
		}

		for _, v := range parsed.Entries {
			if v.NumStr == "" {
				fmt.Println(v.Content)
			} else {
				fmt.Printf("%s %s\n", v.NumStr, v.Content)
			}
		}

		os.Exit(0)
	}

	m := initialModel()

	p := tea.NewProgram(m, tea.WithMouseCellMotion(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("kļūda: ", err)
		os.Exit(1)
	}
}
