package common

import (
	"github.com/charmbracelet/bubbles/key"
)

// Keymap defines a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type Keymap struct {
	Up     key.Binding
	Down   key.Binding
	Help   key.Binding
	Quit   key.Binding
	Back   key.Binding
	Select key.Binding
	Blur   key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k Keymap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit, k.Back}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k Keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},   // first column
		{k.Help, k.Quit}, // second column
	}
}

var Keys = Keymap{
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
	Blur: key.NewBinding(
		key.WithKeys("backspace", "alt+left"),
		key.WithHelp("<-/backspace", "to focus on meny"),
	),
}
