package main

import "time"

// MenuOption represents the available menu options in the main menu.
type MenuOption int

const (
	MenuProfile MenuOption = iota
	MenuSearch
	MenuRefreshCache
	MenuQuit
)

// SearchMode represents the type of search being performed.
type SearchMode int

const (
	SearchModeSong SearchMode = iota
	SearchModeArtist
)

// CachedTrack represents a track stored in the local cache.
type CachedTrack struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Artists     []string `json:"artists"`
	Album       string   `json:"album"`
	AlbumArtURL string   `json:"album_art_url"` // URL to album artwork (small size ~64x64)
	Duration    int      `json:"duration_ms"`
	AddedAt     string   `json:"added_at"`
}

// CachedPlaylist represents a playlist with all its tracks in the local cache.
type CachedPlaylist struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Owner      string        `json:"owner"`
	Public     bool          `json:"public"`
	TrackCount int           `json:"track_count"`
	Tracks     []CachedTrack `json:"tracks"`
}

// PlaylistCache represents the entire cache structure with metadata.
type PlaylistCache struct {
	Playlists   []CachedPlaylist `json:"playlists"`
	LastUpdated time.Time        `json:"last_updated"`
	UserID      string           `json:"user_id"`
}

// PlaylistMatch represents a playlist that matches a search with additional context.
type PlaylistMatch struct {
	Playlist CachedPlaylist
	Position int // Position of the track in the playlist (1-indexed)
}

// SongSearchResult represents the result of a song search.
type SongSearchResult struct {
	Track     CachedTrack
	Playlists []PlaylistMatch
}

// PlaylistWithCount represents a playlist with a count of matching items.
type PlaylistWithCount struct {
	Playlist CachedPlaylist
	Count    int
	Songs    []CachedTrack
}

// ArtistSearchResult represents the result of an artist search.
type ArtistSearchResult struct {
	Artist    string
	Playlists []PlaylistWithCount
}

// SearchIndex holds the in-memory search indexes for fast lookups.
type SearchIndex struct {
	// Maps lowercase song name to all occurrences
	SongIndex map[string][]PlaylistMatch

	// Maps lowercase artist name to all playlists containing their songs
	ArtistIndex map[string][]PlaylistWithCount

	// Original cache for reference
	Cache *PlaylistCache
}
