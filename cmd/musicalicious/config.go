package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// Config holds the application configuration including Spotify credentials and OAuth tokens.
type Config struct {
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenExpiry  time.Time `json:"token_expiry"`
}

// needsRefresh returns true if the access token has expired or will expire within 5 minutes.
func (c *Config) needsRefresh() bool {
	if c.AccessToken == "" {
		return true
	}
	return time.Now().Add(5 * time.Minute).After(c.TokenExpiry)
}

// toOAuthToken converts the Config to an oauth2.Token for use with the Spotify client.
func (c *Config) toOAuthToken() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  c.AccessToken,
		RefreshToken: c.RefreshToken,
		Expiry:       c.TokenExpiry,
		TokenType:    "Bearer",
	}
}

// updateFromOAuthToken updates the Config from an oauth2.Token.
func (c *Config) updateFromOAuthToken(token *oauth2.Token) {
	c.AccessToken = token.AccessToken
	c.RefreshToken = token.RefreshToken
	c.TokenExpiry = token.Expiry
}

// getConfigPath returns the path to the configuration file.
// Uses XDG_CONFIG_HOME if set, otherwise defaults to ~/.config
func getConfigPath() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", errors.Wrap(err, "failed to get user home directory")
		}
		configDir = filepath.Join(homeDir, ".config")
	}

	appConfigDir := filepath.Join(configDir, "musicalicious")
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return "", errors.Wrap(err, "failed to create config directory")
	}

	return filepath.Join(appConfigDir, "config.json"), nil
}

// loadConfig loads the configuration from disk.
// If the config file doesn't exist, returns an empty Config without error.
func loadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	// If config doesn't exist, return empty config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read config file")
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, errors.Wrap(err, "failed to parse config file")
	}

	return &cfg, nil
}

// saveConfig saves the configuration to disk.
func saveConfig(cfg *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal config")
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return errors.Wrap(err, "failed to write config file")
	}

	return nil
}
