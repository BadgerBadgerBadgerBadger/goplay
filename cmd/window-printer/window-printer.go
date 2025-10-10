package main

import (
	"fmt"
	"time"
)

func main() {

	toPrint := []string{
		"Read lines from a file or standard input.",
		"Maintain a slice of 10 lines (a scrolling window).",
		"Instead of clearing the whole screen, move the cursor up 10 times, erase each line, and then reprint.",
		"Delay or wait for input to proceed to the next line.",
	}

	index := 0
	printed := 0
	ticker := time.NewTicker(1 * time.Second)

	for {
		if index == len(toPrint) {
			break
		}

		<-ticker.C

		if printed >= 2 {
			fmt.Print("\033[1A\033[2K\r")
			printed = -1
			continue
		}

		fmt.Println(toPrint[index])
		printed++
		index++
	}
}
