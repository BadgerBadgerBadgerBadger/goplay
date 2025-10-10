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
		p := tea.NewProgram(newConfigModel(cfg))
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

	// Launch main UI
	client, err := newSpotifyClient(cfg)
	if err != nil {
		return errors.Wrap(err, "failed to create spotify client")
	}

	p := tea.NewProgram(newProfileModel(client))
	if _, err := p.StartReturningModel(); err != nil {
		return errors.Wrap(err, "failed to run profile UI")
	}

	return nil
}
