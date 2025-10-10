package main

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"time"

	"github.com/kevin-cantwell/dotmatrix"
	"github.com/nfnt/resize"
	"github.com/pkg/errors"
)

// fetchAlbumArtImage downloads an image from a URL and returns it as an image.Image.
func fetchAlbumArtImage(url string) (image.Image, error) {
	if url == "" {
		return nil, errors.New("empty album art URL")
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch album art")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("failed to fetch album art: status %d", resp.StatusCode)
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode image")
	}

	return img, nil
}

// resizeImage resizes an image to fit within the specified dimensions while maintaining aspect ratio.
func resizeImage(img image.Image, maxWidth, maxHeight uint) image.Image {
	bounds := img.Bounds()
	width := uint(bounds.Dx())
	height := uint(bounds.Dy())

	// If image is already smaller, return as-is
	if width <= maxWidth && height <= maxHeight {
		return img
	}

	// Calculate scaling factor to fit within bounds
	widthScale := float64(maxWidth) / float64(width)
	heightScale := float64(maxHeight) / float64(height)
	scale := widthScale
	if heightScale < widthScale {
		scale = heightScale
	}

	newWidth := uint(float64(width) * scale)
	newHeight := uint(float64(height) * scale)

	return resize.Resize(newWidth, newHeight, img, resize.Lanczos3)
}

// renderAlbumArt fetches and renders album art to a string using braille characters.
// Returns the rendered art as a string, or an empty string if rendering fails.
// The maxWidth and maxHeight are in characters (braille is 2x4 pixels per character).
func renderAlbumArt(url string, maxWidth, maxHeight uint) (string, error) {
	if url == "" {
		return "", nil
	}

	// Fetch image
	img, err := fetchAlbumArtImage(url)
	if err != nil {
		return "", err
	}

	// Resize to fit braille rendering
	// Each braille character represents 2x4 pixels, so:
	// - maxWidth characters = maxWidth * 2 pixels wide
	// - maxHeight characters = maxHeight * 4 pixels tall
	img = resizeImage(img, maxWidth*2, maxHeight*4)

	// Render to buffer using braille characters
	var buf bytes.Buffer
	err = dotmatrix.Print(&buf, img)
	if err != nil {
		return "", errors.Wrap(err, "failed to render image as braille")
	}

	return buf.String(), nil
}

// renderAlbumArtToString is a convenience function that returns album art or a placeholder.
// If rendering fails, returns an empty string instead of an error.
func renderAlbumArtToString(url string, maxWidth, maxHeight uint) string {
	// DISABLED: Album art removed per user preference
	return ""
}

// isTerminalGraphicsCapable checks if the terminal supports image display.
// Returns false since album art has been disabled.
func isTerminalGraphicsCapable() bool {
	return false
}

// getTerminalGraphicsInfo returns a string describing terminal graphics support.
func getTerminalGraphicsInfo() string {
	return "Album art: disabled"
}
