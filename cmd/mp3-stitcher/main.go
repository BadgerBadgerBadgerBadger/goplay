package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/BadgerBadgerBadgerBadger/goplay/internal/media"
	"github.com/BadgerBadgerBadgerBadger/goplay/pkg/util"
	"github.com/bogem/id3v2"
)

func main() {

	inputFilesListPath := flag.String("inp", "", "path to json file containing array of input files")
	outputFile := flag.String("out", "", "path to mp3 output file")

	title := flag.String("title", "", "title of the mp3")
	artist := flag.String("artist", "", "artist of the mp3")
	album := flag.String("album", "", "album of the mp3")

	flag.Parse()

	*inputFilesListPath = "/Users/shak/projects/shaun/goplay/horses.json"
	*outputFile = "/Users/shak/projects/shaun/goplay/horses.mp3"
	*title = "The Valley of Horses"
	*artist = "Jean M. Auel"
	*album = "Earth's Children"

	tag := id3v2.NewEmptyTag()
	tag.SetTitle(*title)
	tag.SetArtist(*artist)
	tag.SetAlbum(*album)

	f, err := os.Open(*inputFilesListPath)
	util.Must(err, "failed to open chapter listing file")
	defer f.Close()

	inputFiles := []string{}

	decoder := json.NewDecoder(f)
	err = decoder.Decode(&inputFiles)
	util.Must(err, "failed to unmarshal input files list")

	stitcher := media.NewStitcher(media.WithId3v2(tag))

	for _, inputFile := range inputFiles {

		inpf, err := os.Open(inputFile)
		util.Must(err, "failed to open input file")
		defer inpf.Close()

		stitcher.Add(inpf)
	}

	outf, err := os.Create(*outputFile)
	util.Must(err, "failed to create output file")

	err = stitcher.Stitch(outf)
	util.Must(err, "failed to stitch mp3")

	fmt.Println("done")
}
