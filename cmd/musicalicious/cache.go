package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"
)

const (
	defaultCacheTTL = 24 * time.Hour
	cacheFileName   = "playlists.json"
)

// getCachePath returns the path to the cache directory.
// Uses XDG_CONFIG_HOME if set, otherwise defaults to ~/.config
func getCachePath() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", errors.Wrap(err, "failed to get user home directory")
		}
		configDir = filepath.Join(homeDir, ".config")
	}

	cacheDir := filepath.Join(configDir, "musicalicious", "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", errors.Wrap(err, "failed to create cache directory")
	}

	return filepath.Join(cacheDir, cacheFileName), nil
}

// loadPlaylistCache loads the playlist cache from disk.
// Returns an empty cache if the file doesn't exist.
func loadPlaylistCache() (*PlaylistCache, error) {
	cachePath, err := getCachePath()
	if err != nil {
		return nil, err
	}

	// If cache doesn't exist, return empty cache
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return &PlaylistCache{
			Playlists:   []CachedPlaylist{},
			LastUpdated: time.Time{},
		}, nil
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read cache file")
	}

	var cache PlaylistCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, errors.Wrap(err, "failed to parse cache file")
	}

	return &cache, nil
}

// savePlaylistCache saves the playlist cache to disk.
func savePlaylistCache(cache *PlaylistCache) error {
	cachePath, err := getCachePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal cache")
	}

	if err := os.WriteFile(cachePath, data, 0600); err != nil {
		return errors.Wrap(err, "failed to write cache file")
	}

	return nil
}

// isCacheValid returns true if the cache is fresh enough.
func isCacheValid(cache *PlaylistCache, ttl time.Duration) bool {
	if cache == nil || len(cache.Playlists) == 0 {
		return false
	}
	return time.Since(cache.LastUpdated) < ttl
}

// ProgressCallback is called during cache building to report progress.
type ProgressCallback func(current, total int, message string)

// buildPlaylistCache fetches all playlists and their tracks from Spotify.
// Calls the progress callback to report progress during the operation.
func buildPlaylistCache(ctx context.Context, client *spotify.Client, progressCb ProgressCallback) (*PlaylistCache, error) {
	// Get current user info
	user, err := client.CurrentUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current user")
	}

	// Fetch all playlists with pagination
	allPlaylists := []CachedPlaylist{}
	limit := 50
	offset := 0

	if progressCb != nil {
		progressCb(0, 1, "Fetching playlists...")
	}

	for {
		playlists, err := client.CurrentUsersPlaylists(ctx, spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch playlists")
		}

		if len(playlists.Playlists) == 0 {
			break
		}

		// For each playlist, fetch all tracks
		for i, pl := range playlists.Playlists {
			if progressCb != nil {
				progressCb(offset+i+1, int(playlists.Total), fmt.Sprintf("Fetching tracks for %s...", pl.Name))
			}

			tracks, err := fetchAllPlaylistTracks(ctx, client, pl.ID)
			if err != nil {
				// Log error but continue with other playlists
				fmt.Printf("Warning: failed to fetch tracks for playlist %s: %v\n", pl.Name, err)
				continue
			}

			cachedPlaylist := CachedPlaylist{
				ID:         string(pl.ID),
				Name:       pl.Name,
				Owner:      pl.Owner.DisplayName,
				Public:     pl.IsPublic,
				TrackCount: int(pl.Tracks.Total),
				Tracks:     tracks,
			}

			allPlaylists = append(allPlaylists, cachedPlaylist)
		}

		// Check if we have more playlists to fetch
		if len(playlists.Playlists) < limit {
			break
		}

		offset += limit
	}

	cache := &PlaylistCache{
		Playlists:   allPlaylists,
		LastUpdated: time.Now(),
		UserID:      string(user.ID),
	}

	if progressCb != nil {
		progressCb(len(allPlaylists), len(allPlaylists), "Cache built successfully!")
	}

	return cache, nil
}

// fetchAllPlaylistTracks fetches all tracks for a given playlist with pagination.
func fetchAllPlaylistTracks(ctx context.Context, client *spotify.Client, playlistID spotify.ID) ([]CachedTrack, error) {
	allTracks := []CachedTrack{}
	limit := 100
	offset := 0

	for {
		items, err := client.GetPlaylistItems(ctx, playlistID, spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch playlist items")
		}

		if len(items.Items) == 0 {
			break
		}

		for _, item := range items.Items {
			if item.Track.Track == nil {
				// Skip items without tracks (e.g., local files)
				continue
			}

			track := item.Track.Track

			// Extract artist names
			artistNames := make([]string, len(track.Artists))
			for i, artist := range track.Artists {
				artistNames[i] = artist.Name
			}

			// Get album art URL (prefer smallest image for efficiency)
			albumArtURL := ""
			if len(track.Album.Images) > 0 {
				// Images are usually sorted largest to smallest, so take the last one
				albumArtURL = track.Album.Images[len(track.Album.Images)-1].URL
			}

			cachedTrack := CachedTrack{
				ID:          string(track.ID),
				Name:        track.Name,
				Artists:     artistNames,
				Album:       track.Album.Name,
				AlbumArtURL: albumArtURL,
				Duration:    int(track.Duration),
				AddedAt:     item.AddedAt,
			}

			allTracks = append(allTracks, cachedTrack)
		}

		// Check if we have more tracks to fetch
		if len(items.Items) < limit {
			break
		}

		offset += limit
	}

	return allTracks, nil
}

// GetOrBuildCache loads the cache from disk or builds it if invalid.
func GetOrBuildCache(ctx context.Context, client *spotify.Client, forceRefresh bool, progressCb ProgressCallback) (*PlaylistCache, error) {
	if !forceRefresh {
		cache, err := loadPlaylistCache()
		if err == nil && isCacheValid(cache, defaultCacheTTL) {
			if progressCb != nil {
				progressCb(1, 1, "Using cached data")
			}
			return cache, nil
		}
	}

	// Build new cache
	cache, err := buildPlaylistCache(ctx, client, progressCb)
	if err != nil {
		return nil, err
	}

	// Save to disk
	if err := savePlaylistCache(cache); err != nil {
		// Log error but return the cache anyway
		fmt.Printf("Warning: failed to save cache: %v\n", err)
	}

	return cache, nil
}
