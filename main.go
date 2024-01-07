package main

import (
	"fmt"
	"go-live/internal/root"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	f, err := tea.LogToFile("bubbletea.log", "debug")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	m := root.NewModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
