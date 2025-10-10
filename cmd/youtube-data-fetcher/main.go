package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/BadgerBadgerBadgerBadger/goplay/pkg/util"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func main() {

	watchHistoryIn := flag.String("watch-history-in", "", "Path to watch history JSON file")
	videoDetailsOut := flag.String("video-details-out", "", "Path to watch history JSON file")
	youtubeApiKey := flag.String("youtube-api-key", "", "YouTube API key")
	flag.Parse()

	if *watchHistoryIn == "" || *youtubeApiKey == "" || *videoDetailsOut == "" {
		flag.Usage()
		return
	}

	watchHistoryF, err := os.Open(*watchHistoryIn)
	if err != nil {
		util.Must(err, "failed to open watch history file")
	}
	defer watchHistoryF.Close()

	watchHistoryJsonDecoder := json.NewDecoder(watchHistoryF)
	watchHistory := make([]WatchHistoryInItem, 0)

	err = watchHistoryJsonDecoder.Decode(&watchHistory)
	if err != nil {
		util.Must(err, "failed to decode watch history file")
	}

	youtubeService, err := youtube.NewService(context.Background(), option.WithAPIKey(*youtubeApiKey))
	if err != nil {
		util.Must(err, "failed to create YouTube service")
	}

	vdof, err := os.Create(*videoDetailsOut)
	if err != nil {
		util.Must(err, "failed to open video details out file")
	}
	defer vdof.Close()

	videoIdsToCall := make([]string, 0)
	list := watchHistory

	log.Printf("Iterating over watch history, %d items\n", len(list))

	batch := 0
	for _, watchHistoryItem := range list {

		videoId := watchHistoryItem.VideoId()
		videoIdsToCall = append(videoIdsToCall, videoId)

		if len(videoIdsToCall) == 50 {

			batch++

			err := callYoutubeApi(batch, youtubeService, videoIdsToCall, vdof)
			if err != nil {
				util.Must(err, "failed to call YouTube API")
			}

			videoIdsToCall = make([]string, 0)
		}
	}

	if len(videoIdsToCall) > 0 {
		batch++

		err := callYoutubeApi(batch, youtubeService, videoIdsToCall, vdof)
		if err != nil {
			util.Must(err, "failed to call YouTube API")
		}
	}

	if err != nil {
		util.Must(err, "failed to get config from JSON")
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
}

func callYoutubeApi(batch int, youtubeService *youtube.Service, videoIds []string, vdof io.Writer) error {

	log.Printf("Calling YouTube API for batch %d\n", batch)

	call := youtubeService.Videos.List([]string{"snippet", "contentDetails", "statistics"}).
		Id(strings.Join(videoIds, ",")).
		MaxResults(50)

	response, err := call.Do()
	if err != nil {
		return errors.Wrap(err, "failed to call video list api")
	}

	err = writeToOut(response.Items, vdof)
	if err != nil {
		return errors.Wrap(err, "failed to write to out file")
	}

	return nil
}

func writeToOut(items []*youtube.Video, out io.Writer) error {

	enc := json.NewEncoder(out)

	for _, item := range items {
		err := enc.Encode(item)
		if err != nil {
			return errors.Wrap(err, "failed to encode video")
		}
	}

	return nil
}

type WatchHistoryInItem struct {
	Header    string `json:"header"`
	Title     string `json:"title"`
	TitleURL  string `json:"titleUrl"`
	Subtitles []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"subtitles"`
	Time             time.Time `json:"time"`
	Products         []string  `json:"products"`
	ActivityControls []string  `json:"activityControls"`
}

func (whit WatchHistoryInItem) VideoId() string {

	if whit.TitleURL == "" {
		return ""
	}
	split := strings.Split(whit.TitleURL, "https://www.youtube.com/watch?v=")

	return split[1]
}
