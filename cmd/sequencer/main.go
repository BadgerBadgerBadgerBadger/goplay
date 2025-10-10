package main

import (
	"fmt"
	"os"
	"time"
)

func main() {

	seq := newSequencer(20, os.Stdout, os.Stdin)

	seq.Add(13, func(r *reporter) error {
		<-time.After(time.Second * 3)
		r.Write([]byte("done\n"))
		return nil
	})

	seq.Add(8, func(r *reporter) error {
		<-time.After(time.Second * 5)
		r.Write([]byte("heyo\n"))
		return nil
	})

	seq.Add(12, func(r *reporter) error {
		<-time.After(time.Second * )
		return fmt.Errorf("some err")
	})

	code := seq.GetExitCode()
	os.Exit(code)
}
