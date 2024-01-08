package live

import (
	"fmt"
	"go-live/internal/common"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const logo = `
    ____             __
   / __ \___  ____  / /___  __  __
  / / / / _ \/ __ \/ / __ \/ / / /
 / /_/ /  __/ /_/ / / /_/ / /_/ /
/_____/\___/ .___/_/\____/\__, /
          /_/            /____/

`

var (
	logoStyle = lipgloss.
			NewStyle().
			PaddingTop(2).
			Foreground(lipgloss.Color("#01FAC6"))
	textStyle   = lipgloss.NewStyle().Align(lipgloss.Left).Foreground(lipgloss.Color("#EFEDFF"))
	activeStyle = lipgloss.
			NewStyle().
			Foreground(lipgloss.Color("#FF6E81")).
			Bold(true)
	titleStyle = lipgloss.
			NewStyle().
			MarginTop(1).
			MarginBottom(2).
			Bold(true)
)

type LiveModel struct {
	keys     common.Keymap
	choices  []string
	cursor   int
	selected map[int]struct{}
	help     help.Model
}

func NewModel() LiveModel {
	return LiveModel{
		keys: common.Keys,
		choices: []string{
			"To Staging",
			"To Production",
		},
		selected: make(map[int]struct{}),
		help:     help.New(),
	}
}

func (m LiveModel) Init() tea.Cmd {
	return nil
}

func (m LiveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// log.Println("live.Update msg:", msg)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
	// Is it a key press?
	case tea.KeyMsg:
		// Cool, what was the actual key pressed?
		switch {
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll

		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case key.Matches(msg, m.keys.Select):
			if _, ok := m.selected[m.cursor]; ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	return m, nil
}

func (m LiveModel) View() string {
	s := []string{}
	s = append(s, logoStyle.Render(logo), titleStyle.Render("Where are you deploying to?"))

	for i, choice := range m.choices {
		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = "->" // cursor!
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "✓" // selected!
		}

		// Render the row
		if i == m.cursor {
			s = append(s, activeStyle.Render(fmt.Sprintf("%s [%s] %s", cursor, checked, choice)))
		} else {
			s = append(s, textStyle.Render(fmt.Sprintf(" %s [%s] %s", cursor, checked, choice)))
		}
	}

	// The footer
	helpView := m.help.View(m.keys)
	spacerView := strings.Repeat("\n", 2)
	s = append(s, spacerView, helpView)

	// Send the string back to BubbleTea for rendering
	return lipgloss.JoinVertical(lipgloss.Top, s...)
}
