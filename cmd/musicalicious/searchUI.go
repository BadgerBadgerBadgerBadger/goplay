package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zmb3/spotify/v2"
)

var (
	searchTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#1DB954")).
				MarginBottom(1)

	searchInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#1DB954")).
				Padding(0, 1)

	modeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#B3B3B3")).
			MarginBottom(1)

	activeModeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1DB954")).
			Bold(true).
			Underline(true)

	resultHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				MarginTop(1).
				MarginBottom(1)

	playlistBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#1DB954")).
				Padding(0, 1).
				MarginBottom(1).
				Width(70)

	selectedPlaylistBoxStyle = lipgloss.NewStyle().
					Border(lipgloss.DoubleBorder()).
					BorderForeground(lipgloss.Color("#1DB954")).
					Padding(0, 1).
					MarginBottom(1).
					Width(70).
					Bold(true)

	playlistLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1DB954")).
				Bold(true)

	playlistValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#B3B3B3"))

	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1DB954")).
			Bold(true)

	searchHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#535353")).
			Italic(true).
			MarginTop(1)
)

type cacheLoadedMsg struct {
	cache *PlaylistCache
	index *SearchIndex
	err   error
}

type searchModel struct {
	client        *spotify.Client
	input         textinput.Model
	mode          SearchMode
	cache         *PlaylistCache
	index         *SearchIndex
	songResults   []SongSearchResult
	artistResults []ArtistSearchResult
	cursor        int
	viewport      int // Top of visible results window
	expanded      map[int]bool
	loading       bool
	err           error
	done          bool
	lastQuery     string // Track last searched query to prevent redundant searches
}

func newSearchModel(client *spotify.Client) searchModel {
	ti := textinput.New()
	ti.Placeholder = "Search for a song or artist..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 42 // Fits in 50-char column with border (2) + padding (2) + column padding (4)

	return searchModel{
		client:   client,
		input:    ti,
		mode:     SearchModeSong,
		expanded: make(map[int]bool),
		loading:  true,
	}
}

func (m searchModel) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.loadCache,
	)
}

func (m searchModel) loadCache() tea.Msg {
	ctx := context.Background()
	cache, err := GetOrBuildCache(ctx, m.client, false, nil)
	if err != nil {
		return cacheLoadedMsg{err: err}
	}

	index := buildSearchIndex(cache)
	return cacheLoadedMsg{cache: cache, index: index}
}

func (m searchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case cacheLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.cache = msg.cache
		m.index = msg.index
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.done = true
			return m, tea.Quit

		case "tab":
			// Toggle between song and artist search
			if m.mode == SearchModeSong {
				m.mode = SearchModeArtist
			} else {
				m.mode = SearchModeSong
			}
			m.cursor = 0
			m.expanded = make(map[int]bool)
			m.lastQuery = "" // Reset to force re-search in new mode
			m.performSearch()
			return m, nil

		case "up":
			if m.cursor > 0 {
				m.cursor--
				// Adjust viewport if cursor moved above visible area
				if m.cursor < m.viewport {
					m.viewport = m.cursor
				}
			}
			return m, nil

		case "down":
			maxCursor := 0
			if m.mode == SearchModeSong {
				maxCursor = len(m.songResults) - 1
			} else {
				maxCursor = len(m.artistResults) - 1
			}

			if m.cursor < maxCursor {
				m.cursor++
				// Adjust viewport if cursor moved below visible area (5 results visible)
				if m.cursor >= m.viewport+5 {
					m.viewport = m.cursor - 4
				}
			}
			return m, nil

		case "enter":
			// Toggle expansion
			m.expanded[m.cursor] = !m.expanded[m.cursor]
			return m, nil
		}
	}

	// Update input and perform search
	m.input, cmd = m.input.Update(msg)
	m.performSearch()

	return m, cmd
}

func (m *searchModel) performSearch() {
	if m.index == nil {
		return
	}

	query := m.input.Value()

	// Skip if query hasn't changed (prevents redundant searches and flickering)
	if query == m.lastQuery {
		return
	}

	m.lastQuery = query
	m.cursor = 0
	m.viewport = 0
	m.expanded = make(map[int]bool)

	// Only search if query has at least 3 characters
	if len(query) < 3 {
		m.songResults = nil
		m.artistResults = nil
		return
	}

	if m.mode == SearchModeSong {
		m.songResults = searchSongs(query, m.index)
	} else {
		m.artistResults = searchArtists(query, m.index)
	}
}

func (m searchModel) View() string {
	if m.loading {
		return loadingStyle.Render("\n  Loading playlist cache...\n  This may take a moment on first run.\n\n")
	}

	if m.err != nil {
		return fmt.Sprintf("\n  Error: %v\n\n  Press 'esc' to return to menu.\n", m.err)
	}

	// Build left column (controls)
	var leftCol strings.Builder

	// Title
	leftCol.WriteString(searchTitleStyle.Render("♫ Search Playlists"))
	leftCol.WriteString("\n\n")

	// Mode toggle
	songMode := modeStyle.Render("Song")
	artistMode := modeStyle.Render("Artist")
	if m.mode == SearchModeSong {
		songMode = activeModeStyle.Render("Song")
	} else {
		artistMode = activeModeStyle.Render("Artist")
	}

	leftCol.WriteString(fmt.Sprintf("Search by: %s | %s\n\n", songMode, artistMode))

	// Search input
	leftCol.WriteString(searchInputStyle.Render(m.input.View()))
	leftCol.WriteString("\n\n")

	// Cache age
	if m.cache != nil {
		age := time.Since(m.cache.LastUpdated)
		cacheInfo := fmt.Sprintf("Cache: %d playlists (updated %v ago)", len(m.cache.Playlists), age.Round(time.Minute))
		leftCol.WriteString(modeStyle.Render(cacheInfo))
		leftCol.WriteString("\n")
	}

	leftCol.WriteString("\n")

	leftCol.WriteString("\n")

	// Help
	leftCol.WriteString(searchHelpStyle.Render("tab: switch mode\n↑/↓: navigate\nenter: expand\nesc: back"))

	// Build right column (results)
	var rightCol strings.Builder

	inputLen := len(m.input.Value())
	if inputLen == 0 {
		rightCol.WriteString(playlistValueStyle.Render("Start typing to search..."))
	} else if inputLen < 3 {
		rightCol.WriteString(playlistValueStyle.Render(fmt.Sprintf("Type at least 3 characters to search... (%d/3)", inputLen)))
	} else if m.mode == SearchModeSong {
		m.renderSongResults(&rightCol)
	} else {
		m.renderArtistResults(&rightCol)
	}

	// Style columns
	leftColumn := lipgloss.NewStyle().
		Width(50).
		Padding(1, 2).
		Render(leftCol.String())

	rightColumn := lipgloss.NewStyle().
		Width(80).
		Padding(1, 2).
		Render(rightCol.String())

	// Join horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)
}

func (m searchModel) renderSongResults(b *strings.Builder) {
	if len(m.songResults) == 0 {
		b.WriteString("\n")
		b.WriteString(playlistValueStyle.Render("No songs found matching your search."))
		return
	}

	b.WriteString(resultHeaderStyle.Render(fmt.Sprintf("\nFound %d song(s):", len(m.songResults))))
	b.WriteString("\n")

	// Show 5 results at a time in viewport
	visibleCount := 5
	start := m.viewport
	end := min(start+visibleCount, len(m.songResults))

	for i := start; i < end; i++ {
		result := m.songResults[i]

		// Song info
		var songInfo strings.Builder
		songInfo.WriteString(playlistLabelStyle.Render(fmt.Sprintf("🎵 %s", result.Track.Name)))
		songInfo.WriteString("\n")
		songInfo.WriteString(playlistValueStyle.Render(fmt.Sprintf("   by %s", strings.Join(result.Track.Artists, ", "))))
		songInfo.WriteString("\n")
		songInfo.WriteString(playlistValueStyle.Render(fmt.Sprintf("   Album: %s", result.Track.Album)))
		songInfo.WriteString("\n")
		songInfo.WriteString(playlistValueStyle.Render(fmt.Sprintf("   Found in %d playlist(s)", len(result.Playlists))))

		// Expanded view with album art
		if m.expanded[i] {
			songInfo.WriteString("\n\n")

			// Show album art if available and terminal supports it
			if result.Track.AlbumArtURL != "" && isTerminalGraphicsCapable() {
				albumArt := renderAlbumArtToString(result.Track.AlbumArtURL, 20, 10)
				if albumArt != "" {
					songInfo.WriteString(albumArt)
					songInfo.WriteString("\n")
				}
			}

			// Show playlists
			for _, pm := range result.Playlists {
				songInfo.WriteString(playlistValueStyle.Render(fmt.Sprintf("   • %s (position #%d)", pm.Playlist.Name, pm.Position)))
				songInfo.WriteString("\n")
			}
		}

		// Apply style
		style := playlistBoxStyle
		if i == m.cursor {
			style = selectedPlaylistBoxStyle
		}

		b.WriteString(style.Render(songInfo.String()))
		b.WriteString("\n")
	}

	// Show scroll indicator
	if end < len(m.songResults) {
		b.WriteString(playlistValueStyle.Render(fmt.Sprintf("... %d more results (↓ to scroll)", len(m.songResults)-end)))
		b.WriteString("\n")
	}
}

func (m searchModel) renderArtistResults(b *strings.Builder) {
	if len(m.artistResults) == 0 {
		b.WriteString("\n")
		b.WriteString(playlistValueStyle.Render("No artists found matching your search."))
		return
	}

	b.WriteString(resultHeaderStyle.Render(fmt.Sprintf("\nFound %d artist(s):", len(m.artistResults))))
	b.WriteString("\n")

	// Show 5 results at a time in viewport
	visibleCount := 5
	start := m.viewport
	end := min(start+visibleCount, len(m.artistResults))

	for i := start; i < end; i++ {
		result := m.artistResults[i]

		// Artist info
		var artistInfo strings.Builder
		artistInfo.WriteString(playlistLabelStyle.Render(fmt.Sprintf("🎤 %s", result.Artist)))
		artistInfo.WriteString("\n")

		totalSongs := 0
		for _, pl := range result.Playlists {
			totalSongs += pl.Count
		}

		artistInfo.WriteString(playlistValueStyle.Render(fmt.Sprintf("   %d song(s) across %d playlist(s)", totalSongs, len(result.Playlists))))

		// Expanded view
		if m.expanded[i] {
			artistInfo.WriteString("\n\n")
			for _, pl := range result.Playlists {
				artistInfo.WriteString(playlistValueStyle.Render(fmt.Sprintf("   • %s (%d song(s))", pl.Playlist.Name, pl.Count)))
				artistInfo.WriteString("\n")

				// Show album art for first song if available and terminal supports it
				if len(pl.Songs) > 0 && pl.Songs[0].AlbumArtURL != "" && isTerminalGraphicsCapable() {
					albumArt := renderAlbumArtToString(pl.Songs[0].AlbumArtURL, 12, 6)
					if albumArt != "" {
						artistInfo.WriteString("      ")
						artistInfo.WriteString(albumArt)
						artistInfo.WriteString("\n")
					}
				}

				// Show song names
				for j, song := range pl.Songs {
					if j < 3 { // Show first 3 songs
						artistInfo.WriteString(playlistValueStyle.Render(fmt.Sprintf("      - %s (Album: %s)", song.Name, song.Album)))
						artistInfo.WriteString("\n")
					}
				}
				if len(pl.Songs) > 3 {
					artistInfo.WriteString(playlistValueStyle.Render(fmt.Sprintf("      ... and %d more", len(pl.Songs)-3)))
					artistInfo.WriteString("\n")
				}
			}
		}

		// Apply style
		style := playlistBoxStyle
		if i == m.cursor {
			style = selectedPlaylistBoxStyle
		}

		b.WriteString(style.Render(artistInfo.String()))
		b.WriteString("\n")
	}

	// Show scroll indicator
	if end < len(m.artistResults) {
		b.WriteString(playlistValueStyle.Render(fmt.Sprintf("... %d more results (↓ to scroll)", len(m.artistResults)-end)))
		b.WriteString("\n")
	}
}
