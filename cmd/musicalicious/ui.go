package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles for the config UI
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#1DB954")). // Spotify green
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#B3B3B3"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#535353")).
			Italic(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)
)

// configModel is the Bubble Tea model for the configuration UI.
type configModel struct {
	config    *Config
	inputs    []textinput.Model
	focused   int
	err       error
	cancelled bool
}

// newConfigModel creates a new configuration model.
func newConfigModel(cfg *Config) configModel {
	m := configModel{
		config: cfg,
		inputs: make([]textinput.Model, 2),
	}

	// Client ID input
	m.inputs[0] = textinput.New()
	m.inputs[0].Placeholder = "Enter your Spotify Client ID"
	m.inputs[0].Focus()
	m.inputs[0].CharLimit = 100
	m.inputs[0].Width = 50
	m.inputs[0].Prompt = "│ "
	m.inputs[0].SetValue(cfg.ClientID)

	// Client Secret input
	m.inputs[1] = textinput.New()
	m.inputs[1].Placeholder = "Enter your Spotify Client Secret"
	m.inputs[1].CharLimit = 100
	m.inputs[1].Width = 50
	m.inputs[1].Prompt = "│ "
	m.inputs[1].EchoMode = textinput.EchoPassword
	m.inputs[1].EchoCharacter = '•'
	m.inputs[1].SetValue(cfg.ClientSecret)

	return m
}

func (m configModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m configModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit

		case "enter":
			// Validate inputs
			if strings.TrimSpace(m.inputs[0].Value()) == "" {
				m.err = fmt.Errorf("Client ID cannot be empty")
				return m, nil
			}
			if strings.TrimSpace(m.inputs[1].Value()) == "" {
				m.err = fmt.Errorf("Client Secret cannot be empty")
				return m, nil
			}

			// Save values to config
			m.config.ClientID = strings.TrimSpace(m.inputs[0].Value())
			m.config.ClientSecret = strings.TrimSpace(m.inputs[1].Value())
			return m, tea.Quit

		case "tab", "shift+tab", "up", "down":
			// Cycle through inputs
			if msg.String() == "up" || msg.String() == "shift+tab" {
				m.focused--
			} else {
				m.focused++
			}

			if m.focused > len(m.inputs)-1 {
				m.focused = 0
			} else if m.focused < 0 {
				m.focused = len(m.inputs) - 1
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i < len(m.inputs); i++ {
				if i == m.focused {
					cmds[i] = m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input
	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *configModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m configModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("♫ Musicalicious Configuration"))
	b.WriteString("\n\n")

	// Client ID field
	b.WriteString(labelStyle.Render("Client ID:"))
	b.WriteString("\n")
	b.WriteString(m.inputs[0].View())
	b.WriteString("\n\n")

	// Client Secret field
	b.WriteString(labelStyle.Render("Client Secret:"))
	b.WriteString("\n")
	b.WriteString(m.inputs[1].View())
	b.WriteString("\n\n")

	// Error message
	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("✗ %s", m.err.Error())))
		b.WriteString("\n\n")
	}

	// Help text
	b.WriteString(helpStyle.Render("↑/↓ or tab: navigate • enter: save • esc: cancel"))
	b.WriteString("\n")

	return b.String()
}
