package main

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zmb3/spotify/v2"
)

// Profile display styles
var (
	// Spotify brand colors
	spotifyGreen = lipgloss.Color("#1DB954")
	spotifyBlack = lipgloss.Color("#191414")
	spotifyGray  = lipgloss.Color("#B3B3B3")
	spotifyWhite = lipgloss.Color("#FFFFFF")

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(spotifyGreen).
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(spotifyGreen).
			Padding(1, 2).
			MarginBottom(1)

	nameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(spotifyWhite).
			Align(lipgloss.Center).
			MarginBottom(1)

	fieldLabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(spotifyGreen).
			Width(15)

	fieldValueStyle = lipgloss.NewStyle().
			Foreground(spotifyGray)

	infoBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(spotifyGreen).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	footerStyle = lipgloss.NewStyle().
			Foreground(spotifyGray).
			Italic(true).
			MarginTop(2)
)

type profileLoadedMsg struct {
	user *spotify.PrivateUser
	err  error
}

// profileModel is the Bubble Tea model for displaying the user profile.
type profileModel struct {
	client  *spotify.Client
	user    *spotify.PrivateUser
	loading bool
	err     error
}

// newProfileModel creates a new profile model.
func newProfileModel(client *spotify.Client) profileModel {
	return profileModel{
		client:  client,
		loading: true,
	}
}

func (m profileModel) Init() tea.Cmd {
	return m.loadProfile
}

func (m profileModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}

	case profileLoadedMsg:
		m.loading = false
		m.user = msg.user
		m.err = msg.err
		return m, nil
	}

	return m, nil
}

func (m profileModel) View() string {
	if m.loading {
		return "\n  Loading your profile...\n\n"
	}

	if m.err != nil {
		return fmt.Sprintf("\n  Error loading profile: %v\n\n  Press 'q' to quit.\n", m.err)
	}

	var b strings.Builder

	// Header
	b.WriteString(headerStyle.Render("♫ Musicalicious"))
	b.WriteString("\n\n")

	// User name (large and centered)
	displayName := m.user.DisplayName
	if displayName == "" {
		displayName = string(m.user.ID)
	}
	b.WriteString(nameStyle.Render(fmt.Sprintf("🎵 %s 🎵", displayName)))
	b.WriteString("\n\n")

	// Info box with profile details
	var info strings.Builder

	// Email
	if m.user.Email != "" {
		info.WriteString(fieldLabelStyle.Render("Email:"))
		info.WriteString("  ")
		info.WriteString(fieldValueStyle.Render(m.user.Email))
		info.WriteString("\n")
	}

	// User ID
	info.WriteString(fieldLabelStyle.Render("Spotify ID:"))
	info.WriteString("  ")
	info.WriteString(fieldValueStyle.Render(string(m.user.ID)))
	info.WriteString("\n")

	// Product (account type)
	accountType := m.user.Product
	if accountType == "" {
		accountType = "free"
	}
	info.WriteString(fieldLabelStyle.Render("Account Type:"))
	info.WriteString("  ")

	accountStyle := fieldValueStyle
	if accountType == "premium" {
		accountStyle = accountStyle.Foreground(spotifyGreen).Bold(true)
	}
	info.WriteString(accountStyle.Render(strings.ToUpper(accountType)))
	info.WriteString("\n")

	// Country
	if m.user.Country != "" {
		info.WriteString(fieldLabelStyle.Render("Country:"))
		info.WriteString("  ")
		info.WriteString(fieldValueStyle.Render(m.user.Country))
		info.WriteString("\n")
	}

	// Followers
	info.WriteString(fieldLabelStyle.Render("Followers:"))
	info.WriteString("  ")
	info.WriteString(fieldValueStyle.Render(fmt.Sprintf("%d", m.user.Followers.Count)))
	info.WriteString("\n")

	// Spotify URI
	info.WriteString(fieldLabelStyle.Render("Profile URI:"))
	info.WriteString("  ")
	info.WriteString(fieldValueStyle.Render(string(m.user.URI)))

	b.WriteString(infoBoxStyle.Render(info.String()))
	b.WriteString("\n")

	// Footer with help
	b.WriteString(footerStyle.Render("Press 'q' to quit"))

	return b.String()
}

// loadProfile fetches the user profile from Spotify.
func (m profileModel) loadProfile() tea.Msg {
	ctx := context.Background()
	user, err := m.client.CurrentUser(ctx)
	if err != nil {
		return profileLoadedMsg{err: err}
	}

	return profileLoadedMsg{user: user}
}
