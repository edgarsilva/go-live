package utils

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	textStyle   = lipgloss.NewStyle().Align(lipgloss.Left).Foreground(lipgloss.Color("#FFE4E4"))
	activeStyle = lipgloss.
			NewStyle().
			Align(lipgloss.Left).
			Background(lipgloss.Color("#FFE4E4")).
			Foreground(lipgloss.Color("#C3515C")).
			Bold(true)
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFE4E4")).
			Background(lipgloss.Color("#C3525C")).
			MarginTop(1).
			MarginBottom(2)
)

type optID int

const (
	idSearch optID = iota
	idTimer
	idPing
	idProgress
)

var stateKeys = map[optID]string{
	idSearch:   "search",
	idTimer:    "timer",
	idPing:     "ping",
	idProgress: "progress",
}

type tickMsg time.Time

type UtilsModel struct {
	state    map[string]bool
	choices  []string
	cursor   int
	timer    timer.Model
	progress progress.Model
}

func NewModel() UtilsModel {
	return UtilsModel{
		choices: []string{
			"Search",
			"Timer",
			"Ping Google",
			"Progress",
		},
		state: map[string]bool{
			"search":   false,
			"timer":    false,
			"ping":     false,
			"progress": false,
		},
		timer:    timer.NewWithInterval(10*time.Second, time.Millisecond),
		progress: progress.New(progress.WithScaledGradient("#465979", "#FF878E")),
		// progress: progress.New(progress.WithDefaultGradient()),
	}
}

func (m UtilsModel) Init() tea.Cmd {
	// return m.timer.Init()
	return nil
}

func (m UtilsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			m.toggleState(m.cursor)

			if m.isHighlighted(idPing) && m.isSelected(idPing) {
				return m, pingGoogle()
			}

			if m.isHighlighted(idTimer) {
				if !m.isSelected(idTimer) && m.timer.Running() {
					return m, m.stopTimer()
				}

				return m, m.startTimer()
			}

			if m.isHighlighted(idProgress) && m.isSelected(idProgress) {
				log.Println("progress bar started...")
				return m, tickCmd()
			}
		}

	// Is it the timer?
	case timer.TickMsg:
		log.Println("timer.TickMsg")
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
		// m.quitting = true
		return m, nil

	// Is it the progress Ticker?
	case tickMsg:
		log.Println("progress tickMsg")
		if m.progress.Percent() == 1.0 {
			return m, nil
		}

		// Note that you can also use progress.Model.SetPercent to set the
		// percentage value explicitly, too.
		cmd := m.progress.IncrPercent(0.25)
		return m, tea.Batch(tickCmd(), cmd)

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		log.Println("progress frame msg")
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}

	return m, nil
}

func (m UtilsModel) View() string {
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
		if m.isIndexSelected(i) {
			checked = "✓" // selected!
		}

		// Render the row
		if i == m.cursor {
			s = append(s, activeStyle.Render(fmt.Sprintf("%s [%s] %s", cursor, checked, choice)))
		} else {
			s = append(s, textStyle.Render(fmt.Sprintf(" %s [%s] %s", cursor, checked, choice)))
		}
	}

	if m.timer.Running() || m.timer.Timedout() {
		s = append(s, "\nExiting in "+m.timer.View())
	}

	if m.timer.Timedout() {
		s = append(s, "All done!")
	}

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
		log.Println("pinged google init...")
		time.Sleep(5 * time.Second)
		c := &http.Client{
			Timeout: 5 * time.Second,
		}
		log.Println("pinged google exec...")
		res, err := c.Get(url)
		if err != nil {
			var msg PingMsg = "ping:err"
			return msg
		}
		defer res.Body.Close()

		log.Println("pinged google successfully!")

		var msg PingMsg = "ping:ok"
		return msg
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m UtilsModel) startTimer() tea.Cmd {
	log.Println("startTimer...")
	return m.timer.Start()
}

func (m UtilsModel) stopTimer() tea.Cmd {
	log.Println("stopTimer...")
	return m.timer.Stop()
}

func stateKey(id int) string {
	return stateKeys[optID(id)]
}

func (m UtilsModel) isIndexSelected(id int) bool {
	return m.state[stateKey(id)]
}

func (m UtilsModel) isSelected(id optID) bool {
	return m.state[stateKey(int(id))]
}

func (m UtilsModel) toggleState(id int) {
	m.state[stateKey(id)] = !m.state[stateKey(id)]
}

func (m UtilsModel) isHighlighted(id optID) bool {
	log.Println("isHighlighted", m.cursor, int(id))
	return m.cursor == int(id)
}
