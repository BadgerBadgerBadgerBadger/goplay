//go:build mage
// +build mage

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
	"github.com/pkg/errors"
)

const (
	indir  = "cmd"
	outdir = "bin"
)

// Build builds all programs under `./cmd` and saves the binaries
// to `./bin` with the same name as the directory under `./cmd`. Use `build all`
// to build all binaries at once or `build <name>` to build a specific program.
func Build(cmdToBuild string) error {
	return buildOrInstall("build", cmdToBuild)
}

// Install installs all programs under `./cmd` using the `go install`
// command. This will make the binaries available under $GOPATH/bin
func Install(cmdToBuild string) error {
	return buildOrInstall("install", cmdToBuild)
}

func buildOrInstall(action, cmdToBuild string) error {

	mg.Deps(InstallDeps)
	fmt.Println("Building...")

	currdir, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "failed to get current directory")
	}

	var filenames []string

	if cmdToBuild == "all" {
		files, err := os.ReadDir(fmt.Sprintf("%s/%s", currdir, indir))
		if err != nil {
			return errors.Wrap(err, "failed to list cmd dir")
		}

		for _, file := range files {

			if !file.IsDir() {
				continue
			}

			filenames = append(filenames, file.Name())
		}
	} else {
		filenames = []string{cmdToBuild}
	}

	for _, filename := range filenames {

		fmt.Printf("building binary for %s\n", filename)

		pathToIndir := fmt.Sprintf("%s/%s", currdir, indir)
		err = os.Chdir(pathToIndir)
		if err != nil {
			return err
		}

		inpath := fmt.Sprintf("%s/%s/%s/", currdir, indir, filename)
		outpath := fmt.Sprintf("%s/%s/%s", currdir, outdir, filename)

		cmd := func() *exec.Cmd {

			switch action {
			case "build":
				return exec.Command("go", action, "-o", outpath, inpath)

			case "install":
				return exec.Command("go", action, inpath)
			}

			return nil
		}()

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err = cmd.Run()
		if err != nil {
			fmt.Println(stderr.String())
			return err
		}
	}

	return nil
}

// InstallDeps downloads all dependencies for the project.
func InstallDeps() error {
	fmt.Println("Installing Deps...")

	cmd := exec.Command("go", "mod", "download")
	return cmd.Run()
}

// Clean up after yourself
func Clean() error {
	fmt.Println("Cleaning...")

	path, err := os.Getwd()
	if err != nil {
		return err
	}

	return os.RemoveAll(fmt.Sprintf("%s/%s", path, outdir))
}
