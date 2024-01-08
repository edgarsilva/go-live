package utils

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const logo = `
   __  ____  _ __
  / / / / /_(_) /____
 / / / / __/ / / ___/
/ /_/ / /_/ / (__  )
\____/\__/_/_/____/
`

type choiceID int

const (
	idTable choiceID = iota
	idTimer
	idPing
	idProgress
)

var (
	logoStyle = lipgloss.
			NewStyle().
			PaddingTop(2).
			Foreground(lipgloss.Color("#01FAC6"))
	textStyle   = lipgloss.NewStyle().Align(lipgloss.Left).Foreground(lipgloss.Color("#FFE4E4"))
	activeStyle = lipgloss.
			NewStyle().
			Align(lipgloss.Left).
			Foreground(lipgloss.Color("#FF6E81")).
			Bold(true)
	titleStyle = lipgloss.NewStyle().
			MarginTop(1).
			MarginBottom(2).
			Bold(true)
	checkedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00C57A"))
)

type choice struct {
	label string
	key   choiceID
}

type tickMsg time.Time

type UtilsModel struct {
	choices       []choice
	state         map[choice]bool
	cursor        int
	timerSpent    bool
	timer         timer.Model
	progressSpent bool
	progress      progress.Model
	table         table.Model
}

func NewModel() UtilsModel {
	return UtilsModel{
		choices: []choice{
			idTable:    {"Table", idTable},
			idTimer:    {"Timer", idTimer},
			idPing:     {"Ping Google", idPing},
			idProgress: {"Progress", idProgress},
		},
		state:      map[choice]bool{},
		timerSpent: false,
		timer:      timer.NewWithInterval(5*time.Second, time.Millisecond),

		progressSpent: false,
		progress:      progress.New(progress.WithScaledGradient("#6A6094", "#FF6E81")),
		// progress: progress.New(progress.WithDefaultGradient()),
		table: newTable(),
	}
}

func (m UtilsModel) Init() tea.Cmd {
	return nil
}

func (m UtilsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Prob refactor this to focus update/actions based on the
	// current choice, if it's active and focused/focusable

	// log.Println("utils.Update msg:", msg)
	log.Println("utils.m.table.focus msg:", msg, m.table.Focused())
	if m.table.Focused() {
		log.Println("Table is focused")
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "backspace":
				m.setState(idTable, false)
				m.table.Blur()
			case "enter":
				log.Printf("Let's go to %s!", m.table.SelectedRow()[1])
			}
		}

		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)

		return m, cmd
	}

	// log.Println("utils.Update msg:", msg)
	switch msg := msg.(type) {
	// Is it a key press?
	case tea.KeyMsg:
		// Cool, what was the actual key pressed?
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter", " ":
			m.toggleCurrentState()
			return m.handleCurrentChoice(msg)
		}

	// Is it the timer?
	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		// m.keymap.stop.SetEnabled(m.timer.Running())
		// m.keymap.start.SetEnabled(!m.timer.Running())
		return m, cmd

	case timer.TimeoutMsg:
		m.setState(idTimer, false)
		m.timerSpent = true
		return m, nil

	// Is it the progress Ticker?
	case tickMsg:
		if m.progress.Percent() == 1.0 {
			return m, nil
		}

		// Note that you can also use progress.Model.SetPercent to set the
		// percentage value explicitly, too.
		cmd := m.progress.IncrPercent(0.25)
		return m, tea.Batch(startProgress(), cmd)

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}

	return m, nil
}

func (m UtilsModel) handleCurrentChoice(msg tea.Msg) (UtilsModel, tea.Cmd) {
	switch m.currentChoice().key {
	case idTable:
		if m.currentChoiceActive() {
			m.table.Focus()
		}

		var cmd tea.Cmd
		m.table, cmd = m.table.Update(tableFocusMsg())

		return m, cmd
	case idTimer:
		if m.currentChoiceActive() {
			return m, m.startTimer()
		}

		if m.timer.Running() {
			return m, m.stopTimer()
		}

		return m, nil
	case idPing:
		return m, pingGoogle()
	case idProgress:
		return m, startProgress()
	default:
		return m, nil
	}
}

func (m UtilsModel) View() string {
	s := []string{logoStyle.Render(logo)}

	for i, choice := range m.choices {
		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = "->" // cursor!
		}

		// Is this choice selected?
		checked := " " // not selected
		if m.choiceActive(choice) {
			checked = "✓" // selected!
		}

		// Render the row
		if i == m.cursor {
			s = append(s, activeStyle.Render(fmt.Sprintf(" %s [%s] %s", cursor, checked, choice.label)))
		} else {
			s = append(s,
				fmt.Sprintf("  %s%s%s",
					textStyle.Inline(true).Render(cursor+" ["),
					checkedStyle.Inline(true).Render(checked),
					textStyle.Inline(true).Render("]"+choice.label),
				),
			)
		}
	}

	// Table view
	s = append(s, "\n", tableStyle.Render(m.table.View(), "\n"))
	// return baseStyle.Render(m.table.View()) + "\n"

	s = append(s, "\nTimer running "+m.timer.View())

	// if !m.quitting {
	// 	s = "Exiting in " + s
	// 	s += m.helpView()
	// }

	// Progress bar view
	s = append(s, "\n", m.progress.View(), "\n")

	// The footer
	s = append(s, titleStyle.Render("🡠 Esc to go back"))

	// Send the sting back to BubbleTea for rendering
	return lipgloss.JoinVertical(lipgloss.Top, s...)
}

// Commands

// Pings Google and returns a tea.Cmd that will send a message back to the root model
type PingMsg string

func pingGoogle() tea.Cmd {
	url := "https://google.com"

	return func() tea.Msg {
		time.Sleep(5 * time.Second)
		c := &http.Client{
			Timeout: 5 * time.Second,
		}

		res, err := c.Get(url)
		if err != nil {
			var msg PingMsg = "ping:err"
			return msg
		}
		defer res.Body.Close()

		var msg PingMsg = "ping:ok"
		log.Println("ping:ok")
		return msg
	}
}

func startProgress() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *UtilsModel) startTimer() tea.Cmd {
	if m.timerSpent {
		m.timer = timer.NewWithInterval(5*time.Second, time.Millisecond)
		return m.timer.Init()
	}

	return m.timer.Start()
}

func (m UtilsModel) stopTimer() tea.Cmd {
	return m.timer.Stop()
}

func (m UtilsModel) choiceActive(c choice) bool {
	return m.state[c]
}

func (m UtilsModel) currentChoiceActive() bool {
	return m.state[m.currentChoice()]
}

func (m UtilsModel) toggleCurrentState() {
	m.state[m.currentChoice()] = !m.state[m.currentChoice()]
}

func (m UtilsModel) setState(id choiceID, state bool) {
	m.state[m.choices[id]] = state
}

func (m UtilsModel) currentChoice() choice {
	return m.choices[m.cursor]
}
