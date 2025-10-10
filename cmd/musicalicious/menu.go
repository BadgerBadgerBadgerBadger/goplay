package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zmb3/spotify/v2"
)

var (
	menuTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#1DB954")).
			MarginBottom(1)

	menuItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#B3B3B3")).
			PaddingLeft(2)

	selectedMenuItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1DB954")).
				Bold(true).
				PaddingLeft(2)

	menuHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#535353")).
			Italic(true).
			MarginTop(2)
)

// menuModel is the Bubble Tea model for the main menu.
type menuModel struct {
	client         *spotify.Client
	cursor         int
	selectedOption MenuOption
	done           bool
}

// newMenuModel creates a new menu model.
func newMenuModel(client *spotify.Client) menuModel {
	return menuModel{
		client:         client,
		cursor:         0,
		selectedOption: -1,
		done:           false,
	}
}

func (m menuModel) Init() tea.Cmd {
	return nil
}

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.done = true
			m.selectedOption = MenuQuit
			return m, tea.Quit

		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = 3 // Wrap to bottom
			}

		case "down", "j":
			m.cursor++
			if m.cursor > 3 {
				m.cursor = 0 // Wrap to top
			}

		case "enter", " ":
			m.selectedOption = MenuOption(m.cursor)
			m.done = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m menuModel) View() string {
	if m.done {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(menuTitleStyle.Render("♫ Musicalicious"))
	b.WriteString("\n\n")

	// Menu options
	menuOptions := []string{
		"View Profile",
		"Search Playlists",
		"Refresh Cache",
		"Quit",
	}

	for i, option := range menuOptions {
		cursor := "  "
		if i == m.cursor {
			cursor = "▶ "
		}

		style := menuItemStyle
		if i == m.cursor {
			style = selectedMenuItemStyle
		}

		b.WriteString(style.Render(fmt.Sprintf("%s%s", cursor, option)))
		b.WriteString("\n")
	}

	// Help text
	b.WriteString(menuHelpStyle.Render("\n↑/↓ or j/k: navigate • enter/space: select • q: quit"))

	return b.String()
}
