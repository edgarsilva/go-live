package root

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"go-live/internal/live"
	"go-live/internal/utils"
)

const logo = `
   ______      __    _
  / ____/___  / /   (_)   _____
 / / __/ __ \/ /   / / | / / _ \
/ /_/ / /_/ / /___/ /| |/ /  __/
\____/\____/_____/_/ |___/\___/
              ┓     ┓         •
              ┣┓┓┏  ┃┓┏┏┓┏┓ ┏┓┓
              ┗┛┗┫  ┗┗┛┛ ┗┫•┗┻┗
                 ┛        ┛
`

var (
	logoStyle = lipgloss.NewStyle().
			PaddingTop(2).
			Foreground(lipgloss.Color("#01FAC6"))
	titleStyle = lipgloss.NewStyle().
			MarginTop(1).
			MarginBottom(2).
			PaddingLeft(1).
			Bold(true)
	textStyle = lipgloss.NewStyle().
			PaddingLeft(1)
	activeStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			Foreground(lipgloss.Color("#FF6E81")).
			Bold(true)
)

type optID int

const (
	idRoot optID = iota
	idLive
	idSearch
	idUtils
)

// KeyMap defines a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type KeyMap struct {
	Up   key.Binding
	Down key.Binding
	// Right  key.Binding
	// Left   key.Binding
	Help   key.Binding
	Quit   key.Binding
	Back   key.Binding
	Select key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},   // first column
		{k.Help, k.Quit}, // second column
	}
}

var keys = KeyMap{
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

type RootModel struct {
	keys    KeyMap
	state   optID
	models  map[string]tea.Model
	current optID
	choices []string
	cursor  int
	help    help.Model
}

func NewModel() RootModel {
	return RootModel{
		keys:  keys,
		state: idRoot,
		models: map[string]tea.Model{
			"live":  live.NewModel(),
			"utils": utils.NewModel(),
		},
		choices: []string{
			"Go Live",
			"Search DB",
			"Utils",
		},
		help: help.New(),
	}
}

func (m RootModel) Init() tea.Cmd {
	return nil
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// var cmd tea.Cmd
	cmds := []tea.Cmd{}

	// log.Printf("root.Update() msg: %v", msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// If we set a width on the help menu it can gracefully truncate
		// its view as needed.
		m.help.Width = msg.Width

	// Is it a key press?
	case tea.KeyMsg:
		// Cool, what was the actual key pressed?
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Back):
			m.current = idRoot
		}
	}

	switch m.current {
	case idRoot:
		// log.Println("m.current == sRoot")
		switch msg := msg.(type) {
		// Is it a key press?
		case tea.KeyMsg:
			// Cool, what was the actual key pressed?
			switch {
			case key.Matches(msg, m.keys.Up):
				if m.cursor > 0 {
					m.cursor--
				}

			case key.Matches(msg, m.keys.Down):
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}

			case key.Matches(msg, m.keys.Select):
				m = m.setCurrent()

			case key.Matches(msg, m.keys.Help):
				m.help.ShowAll = !m.help.ShowAll
			}
		}
	default:
		// Get the new nested model and command from the current model
		nm, nCmd := m.currentModel().Update(msg)
		// Set the nested model back into the root model to keep states up to date
		m = m.setCurrentModel(nm)

		cmds = append(cmds, nCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m RootModel) View() string {
	s := []string{}
	s = append(s, logoStyle.Render(logo))
	s = append(s, titleStyle.Render("What would you like to do?"))

	switch m.current {
	case idRoot:
		for i, choice := range m.choices {
			// Is the cursor pointing at this choice?
			cursor := " " // no cursor
			if m.cursor == i {
				cursor = "->" // cursor!
			}

			// Render the row
			if i == m.cursor {
				s = append(s, activeStyle.Render(fmt.Sprintf("%s %s", cursor, choice)))
			} else {
				s = append(s, textStyle.Render(fmt.Sprintf(" %s %s", cursor, choice)))
			}
		}
	default:
		return m.currentModel().View()
	}

	// The footer

	s = append(s, titleStyle.MarginBottom(1).Render("? toggle help / q to quit"))
	if m.help.ShowAll {
		s = append(s, m.help.View(m.keys))
	}

	// Send the sting back to BubbleTea for rendering
	return lipgloss.JoinVertical(lipgloss.Top, s...)
}

func (m RootModel) setCurrent() RootModel {
	switch m.cursor {
	case 0:
		m.current = idLive
	case 1:
		m.current = idSearch
	case 2:
		m.current = idUtils
	default:
		m.current = idRoot
	}

	return m
}

func (m RootModel) currentKey() string {
	switch m.current {
	case idRoot:
		return "root"
	case idLive:
		return "live"
	case idUtils:
		return "utils"
	}

	return ""
}

func (m RootModel) currentModel() tea.Model {
	return m.models[m.currentKey()]
}

func (m RootModel) setCurrentModel(cm tea.Model) RootModel {
	m.models[m.currentKey()] = cm

	return m
}
