package live

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	textStyle   = lipgloss.NewStyle().Align(lipgloss.Left).Foreground(lipgloss.Color("#EFEDFF"))
	activeStyle = lipgloss.
			NewStyle().
			Align(lipgloss.Left).
			Background(lipgloss.Color("#EFEDFF")).
			Foreground(lipgloss.Color("#595A8B")).
			Bold(true)
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#EFEDFF")).
			Background(lipgloss.Color("#595A8B")).
			MarginTop(1).
			MarginBottom(2)
)

// keymap defines a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type keymap struct {
	Up     key.Binding
	Down   key.Binding
	Help   key.Binding
	Quit   key.Binding
	Back   key.Binding
	Select key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},   // first column
		{k.Help, k.Quit}, // second column
	}
}

var keys = keymap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "to go back"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter", " ", "space"),
		key.WithHelp("⏎/⌴", "to confirm selection"),
	),
}

type LiveModel struct {
	keys     keymap
	choices  []string
	cursor   int
	selected map[int]struct{}
	help     help.Model
}

func NewModel() LiveModel {
	return LiveModel{
		keys: keys,
		choices: []string{
			"To Staging",
			"To Production",
			"Back",
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
	s = append(s, titleStyle.Render("We push to ..."))

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
