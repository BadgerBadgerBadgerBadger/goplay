package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/BadgerBadgerBadgerBadger/goplay/pkg/util"
	"google.golang.org/api/youtube/v3"
)

func main() {

	watchDetailsIn := flag.String("watch-details-in", "", "Path to watch details JSON file")
	csvOut := flag.String("csv-out", "", "Path to CSV file to write")
	flag.Parse()

	if *watchDetailsIn == "" {
		flag.Usage()
		return
	}

	watchdetailsf, err := os.Open(*watchDetailsIn)
	if err != nil {
		util.ExitOnError(err, "failed to open watch details file")
	}
	defer watchdetailsf.Close()

	scanner := bufio.NewScanner(watchdetailsf)
	tags := map[string]int{}

	for scanner.Scan() {
		deet := youtube.Video{}
		err := json.Unmarshal([]byte(scanner.Text()), &deet)
		if err != nil {
			util.ExitOnError(err, "failed to unmarshal watch details")
		}

		for _, tag := range deet.Snippet.Tags {
			if _, ok := tags[tag]; !ok {
				tags[tag] = 0
			}

			tags[tag]++
		}
	}

	reverseCountMap := map[int][]string{}

	for tag, count := range tags {
		if _, ok := reverseCountMap[count]; !ok {
			reverseCountMap[count] = []string{}
		}

		reverseCountMap[count] = append(reverseCountMap[count], tag)
	}

	countsArr := []int{}
	for count, _ := range reverseCountMap {
		countsArr = append(countsArr, count)
	}

	sort.Slice(countsArr, func(i, j int) bool {
		return countsArr[i] > countsArr[j]
	})

	total := 0
	pairs := []Pair[string, int]{}
	for _, count := range countsArr {
		wordsForCount := reverseCountMap[count]

		for _, word := range wordsForCount {
			pairs = append(pairs, Pair[string, int]{
				A: word,
				B: count,
			})
			total++
		}

		if total >= 1500 {
			break
		}
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].A > pairs[j].A
	})

	if *csvOut != "" {
		csvf, err := os.Create(*csvOut)
		if err != nil {
			util.ExitOnError(err, "failed to create CSV file")
		}
		defer csvf.Close()

		fmt.Fprintf(csvf, "\"weight\";\"word\";\"color\";\"url\"\n")

		for _, pair := range pairs {
			fmt.Fprintf(csvf, "%d;\"%s\";\"\";\"\"\n", pair.B, pair.A)
		}
	}
}

type Pair[T, K any] struct {
	A T
	B K
}
