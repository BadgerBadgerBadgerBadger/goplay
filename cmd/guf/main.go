package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/BadgerBadgerBadgerBadger/goplay/pkg/util"
	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"github.com/yoheimuta/go-protoparser"
	"github.com/yoheimuta/go-protoparser/parser"
)

const (
	guffileName = "guf.toml"

	fileSuffixProto = "proto"
)

func main() {

	gufDir, err := getGufDir(os.Args[1:])
	util.ExitOnError(err, "failed to get current directory")

	gf, err := getGufFile(gufDir)
	if errors.Cause(err) == os.ErrNotExist {
		fmt.Fprintf(os.Stderr, "could not find "+guffileName+" in working directory")
		os.Exit(1)
	}
	util.ExitOnError(err, "failed to get guffile")

	var remoteContents []*remoteFile

	for _, dep := range gf.Deps {
		remoteFile, err := fetchRemoteFile(dep)
		util.ExitOnError(err, "failed to fetch remote file")

		remoteContents = append(remoteContents, remoteFile)
	}
}

func fetchRemoteFile(d dep) (*remoteFile, error) {

}

type remoteFile struct {
	name       string
	conentents []byte
}

func getGufDir(inp []string) (string, error) {

	if len(inp) > 0 && inp[0] != "" {
		return inp[0], nil
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "failed to get current directory")
	}

	return currentDir, nil
}

func getGufFile(gufDir string) (gufFile, error) {

	gufModFile, err := os.Open(gufDir + "/" + guffileName)
	if err != nil {
		return gufFile{}, errors.Wrap(err, "failed to open "+guffileName)
	}

	defer gufModFile.Close()

	gf := gufFile{}
	_, err = toml.NewDecoder(gufModFile).Decode(&gf)
	if err != nil {
		return gufFile{}, errors.Wrap(err, "failed to decode "+guffileName)
	}

	return gf, nil
}

func getProtoFiles(gufDir string) ([]*parser.Proto, error) {

	dirEntries, err := os.ReadDir(gufDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read guf directory")
	}

	var protoFiles []*parser.Proto

	for _, dirEntry := range dirEntries {

		if dirEntry.IsDir() {
			continue
		}

		name := dirEntry.Name()
		if isProtoFile(name) {

			proto, err := parseProtoFile(gufDir + "/" + name)
			if err != nil {
				return nil, err
			}

			protoFiles = append(protoFiles, proto)
		}
	}

	return protoFiles, nil
}

func parseProtoFile(path string) (*parser.Proto, error) {

	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open protofile %s", path)
	}
	defer f.Close()

	proto, err := protoparser.Parse(f)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse protofile %s", path)
	}

	return proto, nil
}

func isProtoFile(name string) bool {

	split := strings.Split(name, ".")
	last := split[len(split)-1]

	return last == fileSuffixProto
}

func parseProtoFiles(protFile) {}
