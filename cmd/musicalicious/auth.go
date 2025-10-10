package main

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const (
	redirectURI = "http://localhost:666/callback"
	state       = "musicalicious-auth-state"
)

// authenticate performs the OAuth2 authentication flow.
// It opens a browser for the user to authorize the app, starts a local server
// to receive the callback, and exchanges the authorization code for tokens.
func authenticate(ctx context.Context, cfg *Config) error {
	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPrivate,
			spotifyauth.ScopeUserReadEmail,
		),
		spotifyauth.WithClientID(cfg.ClientID),
		spotifyauth.WithClientSecret(cfg.ClientSecret),
	)

	// Create a channel to receive the token
	tokenChan := make(chan *oauth2.Token, 1)
	errChan := make(chan error, 1)

	// Start the callback server
	server := &http.Server{
		Addr:         ":666",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.Token(ctx, state, r)
		if err != nil {
			errChan <- errors.Wrap(err, "failed to get token")
			http.Error(w, "Authentication failed", http.StatusInternalServerError)
			return
		}

		tokenChan <- token

		// Show success page
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
			<html>
			<head>
				<style>
					body {
						font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
						display: flex;
						justify-content: center;
						align-items: center;
						height: 100vh;
						margin: 0;
						background: linear-gradient(135deg, #1DB954 0%, #191414 100%);
						color: white;
					}
					.container {
						text-align: center;
						padding: 40px;
						background: rgba(0, 0, 0, 0.5);
						border-radius: 10px;
					}
					h1 { font-size: 48px; margin: 0 0 20px 0; }
					p { font-size: 18px; margin: 10px 0; }
				</style>
			</head>
			<body>
				<div class="container">
					<h1>✓ Authentication Successful!</h1>
					<p>You can close this window and return to your terminal.</p>
				</div>
			</body>
			</html>
		`)
	})

	// Start server in background
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- errors.Wrap(err, "failed to start callback server")
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Open browser to authorization URL
	url := auth.AuthURL(state)
	fmt.Printf("\nOpening browser for authentication...\n")
	fmt.Printf("If the browser doesn't open automatically, visit:\n%s\n\n", url)

	if err := openBrowser(url); err != nil {
		fmt.Printf("Failed to open browser automatically: %v\n", err)
	}

	// Wait for token or error
	var token *oauth2.Token
	select {
	case token = <-tokenChan:
		// Success!
	case err := <-errChan:
		server.Shutdown(ctx)
		return err
	case <-time.After(5 * time.Minute):
		server.Shutdown(ctx)
		return errors.New("authentication timeout - please try again")
	}

	// Shutdown the server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(shutdownCtx)

	// Update config with tokens
	cfg.updateFromOAuthToken(token)

	return nil
}

// newSpotifyClient creates a new authenticated Spotify client.
func newSpotifyClient(cfg *Config) (*spotify.Client, error) {
	auth := spotifyauth.New(
		spotifyauth.WithClientID(cfg.ClientID),
		spotifyauth.WithClientSecret(cfg.ClientSecret),
	)

	token := cfg.toOAuthToken()
	httpClient := auth.Client(context.Background(), token)
	client := spotify.New(httpClient)

	return client, nil
}

// openBrowser opens the default browser to the specified URL.
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return errors.New("unsupported platform")
	}

	return cmd.Start()
}
