package main

import (
	"strings"
)

// buildSearchIndex creates in-memory indexes for fast searching.
func buildSearchIndex(cache *PlaylistCache) *SearchIndex {
	index := &SearchIndex{
		SongIndex:   make(map[string][]PlaylistMatch),
		ArtistIndex: make(map[string][]PlaylistWithCount),
		Cache:       cache,
	}

	// Build song index: map song names to playlist positions
	for _, playlist := range cache.Playlists {
		for position, track := range playlist.Tracks {
			key := strings.ToLower(track.Name)

			index.SongIndex[key] = append(index.SongIndex[key], PlaylistMatch{
				Playlist: playlist,
				Position: position + 1, // 1-indexed
			})
		}
	}

	// Build artist index: map artist names to playlists with counts
	artistPlaylists := make(map[string]map[string][]CachedTrack)

	for _, playlist := range cache.Playlists {
		for _, track := range playlist.Tracks {
			for _, artist := range track.Artists {
				key := strings.ToLower(artist)

				if artistPlaylists[key] == nil {
					artistPlaylists[key] = make(map[string][]CachedTrack)
				}

				artistPlaylists[key][playlist.ID] = append(artistPlaylists[key][playlist.ID], track)
			}
		}
	}

	// Convert to final artist index format
	playlistMap := make(map[string]CachedPlaylist)
	for _, pl := range cache.Playlists {
		playlistMap[pl.ID] = pl
	}

	for artist, playlists := range artistPlaylists {
		var playlistsWithCount []PlaylistWithCount

		for playlistID, tracks := range playlists {
			if playlist, ok := playlistMap[playlistID]; ok {
				playlistsWithCount = append(playlistsWithCount, PlaylistWithCount{
					Playlist: playlist,
					Count:    len(tracks),
					Songs:    tracks,
				})
			}
		}

		index.ArtistIndex[artist] = playlistsWithCount
	}

	return index
}

// searchSongs searches for songs by name and returns matching results.
func searchSongs(query string, index *SearchIndex) []SongSearchResult {
	if query == "" {
		return nil
	}

	query = strings.ToLower(strings.TrimSpace(query))
	results := make(map[string]*SongSearchResult)

	// Exact and substring matches
	for songName, matches := range index.SongIndex {
		if strings.Contains(songName, query) {
			// Group by unique track
			for _, match := range matches {
				key := match.Playlist.Tracks[match.Position-1].ID
				if key == "" {
					key = match.Playlist.Tracks[match.Position-1].Name
				}

				if result, exists := results[key]; exists {
					result.Playlists = append(result.Playlists, match)
				} else {
					results[key] = &SongSearchResult{
						Track:     match.Playlist.Tracks[match.Position-1],
						Playlists: []PlaylistMatch{match},
					}
				}
			}
		}
	}

	// Convert map to slice
	var searchResults []SongSearchResult
	for _, result := range results {
		searchResults = append(searchResults, *result)
	}

	return searchResults
}

// searchArtists searches for artists by name and returns matching results.
func searchArtists(query string, index *SearchIndex) []ArtistSearchResult {
	if query == "" {
		return nil
	}

	query = strings.ToLower(strings.TrimSpace(query))
	var results []ArtistSearchResult

	// Exact and substring matches
	for artistName, playlists := range index.ArtistIndex {
		if strings.Contains(artistName, query) {
			results = append(results, ArtistSearchResult{
				Artist:    artistName,
				Playlists: playlists,
			})
		}
	}

	return results
}

// fuzzyMatch performs a simple fuzzy match between two strings.
// Returns a score from 0-100, where 100 is an exact match.
func fuzzyMatch(s1, s2 string) int {
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	if s1 == s2 {
		return 100
	}

	if strings.Contains(s1, s2) || strings.Contains(s2, s1) {
		// Substring match - score based on length similarity
		minLen := min(len(s1), len(s2))
		maxLen := max(len(s1), len(s2))
		return (minLen * 80) / maxLen
	}

	// Calculate similarity based on common characters
	commonChars := 0
	for _, c := range s1 {
		if strings.ContainsRune(s2, c) {
			commonChars++
		}
	}

	return (commonChars * 50) / len(s1)
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
