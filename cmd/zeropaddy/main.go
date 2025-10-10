package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/BadgerBadgerBadgerBadger/goplay/pkg/util"
)

type playerScore struct {
	Name   string
	Scores []int
}

func main() {

	inp := flag.String("inp", "", "input file path")

	*inp = "/Users/shak/projects/shaun/goplay/cmd/zeropaddy/inp.csv"

	flag.Parse()

	f, err := os.Open(*inp)
	if err != nil {
		util.ExitOnError(err, "failed to open file")
	}
	defer f.Close()

	longestScores := 0
	playerScores := make(map[string]playerScore)

	// Read file line by line.
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// Do something with the line.
		line := scanner.Text()
		split := strings.Split(line, ",")

		name := split[0]
		scores := make([]int, 0)

		scoresCount := 0

		for _, s := range split[1:] {

			scoresCount++

			score, err := strconv.Atoi(s)
			if err != nil {
				util.ExitOnError(err, "failed to convert score to int")
			}

			scores = append(scores, score)
		}

		if scoresCount > longestScores {
			longestScores = scoresCount
		}

		ps := playerScore{
			Name:   name,
			Scores: scores,
		}
		playerScores[name] = ps
	}

	// Perform zero-padding for each player
	for _, ps := range playerScores {

		diff := longestScores - len(ps.Scores)
		if diff == 0 {
			continue
		}

		newScores := make([]int, longestScores)
		jump := len(ps.Scores) / diff

		for i := 1; i < diff; i++ {

		}
	}

	// Print the zero-padded player scores
	for name, ps := range playerScores {
		fmt.Printf("%s: %v\n", name, ps.Scores)
	}
}
