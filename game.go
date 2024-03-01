package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	styleCenter   = lipgloss.NewStyle().Align(lipgloss.Center)
	styleUnopened = lipgloss.NewStyle().Foreground(lipgloss.Color("#6311cf"))
	styleSelected = lipgloss.NewStyle().Foreground(lipgloss.Color("#00db25"))
	styleWon      = lipgloss.NewStyle().Foreground(lipgloss.Color("#00db25"))
	styleLost     = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	styleFlagged  = lipgloss.NewStyle().Foreground(lipgloss.Color("#5c1111"))
)

type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding

	Dig  key.Binding
	Flag key.Binding

	Help key.Binding
	Quit key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "w"),
		key.WithHelp("↑/w", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "s"),
		key.WithHelp("↓/s", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "a"),
		key.WithHelp("←/a", "move left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "d"),
		key.WithHelp("→/d", "move right"),
	),
	Dig: key.NewBinding(
		key.WithKeys("j", "z"),
		key.WithHelp("z/j", "dig"),
	),
	Flag: key.NewBinding(
		key.WithKeys("k", "x"),
		key.WithHelp("x/k", "flag"),
	),

	Help: key.NewBinding(key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

func (k *keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Dig, k.Flag, k.Help, k.Quit}
}

func (k *keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Dig, k.Flag},
		{k.Help, k.Quit},
	}
}

type game struct {
	board board

	keys keyMap
	help help.Model

	tilesRemaining int

	lost    bool
	won     bool
	started bool

	selectedRow int
	selectedCol int
}

func NewGame(rows int, cols int, mineCount int) game {

	board := generateBoard(rows, cols, mineCount)

	return game{
		board: board,

		keys: keys,
		help: help.New(),

		tilesRemaining: rows*cols - mineCount,

		lost:    false,
		won:     false,
		started: false,

		selectedRow: 0,
		selectedCol: 0,
	}
}

func (g *game) Init() tea.Cmd {
	return nil
}

func (g *game) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		// If we set a width on the help menu it can gracefully truncate
		// its view as needed.
		g.help.Width = msg.Width

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch {
		case key.Matches(msg, g.keys.Down):
			if g.selectedRow < g.board.rowCount-1 {
				g.selectedRow++
			}

		case key.Matches(msg, g.keys.Up):
			if g.selectedRow > 0 {
				g.selectedRow--
			}

		case key.Matches(msg, g.keys.Right):
			if g.selectedCol < g.board.columnCount-1 {
				g.selectedCol++
			}

		case key.Matches(msg, g.keys.Left):
			if g.selectedCol > 0 {
				g.selectedCol--
			}

		case key.Matches(msg, g.keys.Flag):
			if g.board.mines[g.selectedRow][g.selectedCol].IsOpened() {
				g.Chord()
			} else {
				g.Flag()
			}

		case key.Matches(msg, g.keys.Dig):
			if g.board.mines[g.selectedRow][g.selectedCol].IsOpened() {
				g.Chord()
			} else {
				g.Open()
			}
		case key.Matches(msg, g.keys.Quit):
			return g, tea.Quit

		case key.Matches(msg, g.keys.Help):
			g.help.ShowAll = !g.help.ShowAll
		}

	}

	return g, nil
}

func (g *game) View() string {
	var s strings.Builder

	g.drawTop(&s)

	for i := 0; i < g.board.rowCount; i++ {
		g.drawItems(&s, i, g.won || g.lost)

		if i < g.board.rowCount-1 {
			g.drawMiddle(&s, i)
		}
	}

	g.drawBottom(&s)

	s.WriteString(g.help.View(&g.keys))

	return styleCenter.Render(s.String())
}

func (g *game) Chord() {
	t := &g.board.mines[g.selectedRow][g.selectedCol]

	if !t.IsOpened() || t.IsEmpty() {
		return
	}

	totalFlags := 0
	g.surrounding(g.selectedRow, g.selectedCol, func(i, j int) {
		if g.board.mines[i][j].IsFlagged() {
			totalFlags++
		}
	})

	if totalFlags != t.value {
		return
	}

	g.surrounding(g.selectedRow, g.selectedCol, func(i, j int) {
		g.openAll(i, j)
	})

}

func (g *game) surrounding(row int, col int, call func(int, int)) {
	canSubRow := row-1 >= 0
	canAddRow := row+1 < g.board.rowCount
	canSubCol := col-1 >= 0
	canAddCol := col+1 < g.board.columnCount

	if canSubRow {
		call(row-1, col)

		if canAddCol {
			call(row-1, col+1)
		}

		if canSubCol {
			call(row-1, col-1)
		}
	}

	if canAddRow {
		call(row+1, col)

		if canAddCol {
			call(row+1, col+1)
		}

		if canSubCol {
			call(row+1, col-1)
		}
	}

	if canSubCol {
		call(row, col-1)
	}

	if canAddCol {
		call(row, col+1)

	}

}

func (g *game) Flag() {
	t := &g.board.mines[g.selectedRow][g.selectedCol]
	// can't flag an opened mine
	if t.IsOpened() {
		return
	}

	if t.IsFlagged() {
		t.state = closed
	} else {
		t.state = flagged
	}
}

func (g *game) Open() {
	// moves the first mine
	if !g.started {
		g.board.updateStartBoard(g.selectedRow, g.selectedCol)
		g.board.calculateSurrounding()

		g.started = true
	}

	t := &g.board.mines[g.selectedRow][g.selectedCol]
	// can't flag an opened mine
	if t.IsOpened() {
		return
	}

	g.openAll(g.selectedRow, g.selectedCol)

}

func (g *game) openAll(row int, col int) {
	if row < 0 || row >= g.board.rowCount ||
		col < 0 || col >= g.board.columnCount ||
		g.board.mines[row][col].IsOpened() {
		return
	}

	success, lost := g.board.mines[row][col].open()
	if lost {
		g.lost = true
	}

	if success {
		g.tilesRemaining--
		if g.tilesRemaining == 0 && !g.lost {
			g.won = true
		}
	}

	if g.board.mines[row][col].value > 0 {
		return
	}

	if g.board.mines[row][col].IsEmpty() {
		g.surrounding(row, col, func(i int, j int) {
			g.openAll(i, j)
		})
	}

}
