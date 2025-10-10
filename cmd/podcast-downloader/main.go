package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

// Podcast represents the structure of each podcast item in the JSON file
type Podcast struct {
	Title         string `json:"title"`
	PublishedDate string `json:"published_date"`
	MP3Link       string `json:"mp3_link"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <path to json file> <number of concurrent downloads>")
		return
	}

	jsonFilePath := os.Args[1]
	concurrency := os.Args[2]

	// Read the JSON file
	jsonData, err := os.ReadFile(jsonFilePath)
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	// Parse the JSON data
	var podcasts []Podcast
	err = json.Unmarshal(jsonData, &podcasts)
	if err != nil {
		fmt.Println("Error parsing JSON data:", err)
		return
	}

	// Convert concurrency to an integer
	concurrencyLimit, err := strconv.Atoi(concurrency)
	if err != nil {
		fmt.Println("Error converting concurrency to integer:", err)
		return
	}

	// Create a wait group to wait for all downloads to complete
	var wg sync.WaitGroup
	// Create a buffered channel to limit the number of concurrent downloads
	sem := make(chan struct{}, concurrencyLimit)

	// Create a progress bar container
	p := mpb.New(mpb.WithWidth(60))

	// Create a total progress bar
	totalBar := p.AddBar(int64(len(podcasts)),
		mpb.PrependDecorators(
			decor.Name("Total Progress:", decor.WC{W: len("Total Progress:"), C: decor.DindentRight}),
			decor.CountersNoUnit("%d / %d", decor.WC{W: 10}),
		),
		mpb.AppendDecorators(
			decor.Percentage(),
		),
	)

	for _, podcast := range podcasts {
		wg.Add(1)
		go func(podcast Podcast) {
			defer wg.Done()
			sem <- struct{}{} // Acquire a slot
			downloadMP3(podcast, p)
			totalBar.Increment()
			<-sem // Release the slot
		}(podcast)
	}

	wg.Wait()
	p.Wait()
	fmt.Println("All downloads completed.")
}

func downloadMP3(podcast Podcast, p *mpb.Progress) {
	// Create a sanitized filename from the podcast title
	fileName := sanitizeFileName(podcast.Title) + ".mp3"
	outFile, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error creating file for", podcast.Title, ":", err)
		return
	}
	defer outFile.Close()

	// Download the MP3 file
	resp, err := http.Get(podcast.MP3Link)
	if err != nil {
		fmt.Println("Error downloading", podcast.MP3Link, ":", err)
		return
	}
	defer resp.Body.Close()

	// Get the content length for the progress bar
	contentLength := resp.ContentLength

	// Create a progress bar for the individual download
	bar := p.AddBar(contentLength,
		mpb.PrependDecorators(
			decor.Name(podcast.Title, decor.WC{W: len(podcast.Title), C: decor.DindentRight}),
		),
		mpb.AppendDecorators(
			decor.Percentage(),
		),
	)

	// Wrap the response body with the progress bar proxy reader
	barReader := bar.ProxyReader(resp.Body)
	defer barReader.Close()

	// Save the MP3 file to disk
	_, err = io.Copy(outFile, barReader)
	if err != nil {
		fmt.Println("Error saving", podcast.MP3Link, ":", err)
	}

	fmt.Println("Downloaded:", podcast.Title)
}

func sanitizeFileName(fileName string) string {
	return filepath.Clean(fileName)
}
