package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/BadgerBadgerBadgerBadger/goplay/pkg/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"
)

func main() {
	ctx := context.Background()
	err := run(ctx, os.Args[1:], os.Stdout)
	util.ExitOnError(err, "failed to run musicalicious")
}

func run(ctx context.Context, args []string, out io.Writer) error {

	// Load or create configuration
	cfg, err := loadConfig()
	if err != nil {
		return errors.Wrap(err, "failed to load configuration")
	}

	// Check if this is first run (no client ID/secret configured)
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		fmt.Fprintln(out, "Welcome to Musicalicious! Let's get you set up...")
		fmt.Fprintln(out, "\nFirst, you'll need to create a Spotify app at:")
		fmt.Fprintln(out, "https://developer.spotify.com/dashboard/applications")
		fmt.Fprintln(out, "\nMake sure to add this redirect URI to your app:")
		fmt.Fprintln(out, "  http://localhost:666/callback")
		fmt.Fprintln(out, "")

		// Launch config UI
		p := tea.NewProgram(newConfigModel(cfg), tea.WithAltScreen())
		m, err := p.StartReturningModel()
		if err != nil {
			return errors.Wrap(err, "failed to run config UI")
		}

		// Get updated config from model
		configModel, ok := m.(configModel)
		if !ok {
			return errors.New("unexpected model type from config UI")
		}

		if configModel.cancelled {
			return errors.New("configuration cancelled by user")
		}

		cfg = configModel.config
		if err := saveConfig(cfg); err != nil {
			return errors.Wrap(err, "failed to save configuration")
		}
	}

	// Check if we need to authenticate
	if cfg.AccessToken == "" || cfg.needsRefresh() {
		fmt.Fprintln(out, "Authentication required...")

		if err := authenticate(ctx, cfg); err != nil {
			return errors.Wrap(err, "authentication failed")
		}

		if err := saveConfig(cfg); err != nil {
			return errors.Wrap(err, "failed to save authentication tokens")
		}
	}

	// Create Spotify client
	client, err := newSpotifyClient(cfg)
	if err != nil {
		return errors.Wrap(err, "failed to create spotify client")
	}

	// Main menu loop
	for {
		// Show menu
		p := tea.NewProgram(newMenuModel(client), tea.WithAltScreen())
		m, err := p.StartReturningModel()
		if err != nil {
			return errors.Wrap(err, "failed to run menu")
		}

		menuModel, ok := m.(menuModel)
		if !ok {
			return errors.New("unexpected model type from menu")
		}

		// Handle menu selection
		switch menuModel.selectedOption {
		case MenuProfile:
			// Show profile
			p := tea.NewProgram(newProfileModel(client), tea.WithAltScreen())
			if _, err := p.StartReturningModel(); err != nil {
				return errors.Wrap(err, "failed to run profile UI")
			}

		case MenuSearch:
			// Show search interface
			p := tea.NewProgram(newSearchModel(client), tea.WithAltScreen())
			if _, err := p.StartReturningModel(); err != nil {
				return errors.Wrap(err, "failed to run search UI")
			}

		case MenuRefreshCache:
			// Refresh cache
			fmt.Fprintln(out, "\nRefreshing playlist cache...")
			cache, err := GetOrBuildCache(ctx, client, true, func(current, total int, message string) {
				fmt.Fprintf(out, "\r[%d/%d] %s", current, total, message)
			})
			if err != nil {
				fmt.Fprintf(out, "\nError refreshing cache: %v\n", err)
				fmt.Fprintln(out, "Press any key to continue...")
				// Wait for keypress
				var input string
				fmt.Scanln(&input)
			} else {
				fmt.Fprintf(out, "\n\nCache refreshed! %d playlists loaded.\n", len(cache.Playlists))
				fmt.Fprintln(out, "Press any key to continue...")
				// Wait for keypress
				var input string
				fmt.Scanln(&input)
			}

		case MenuQuit:
			return nil

		default:
			return nil
		}
	}
}
