package config

import (
	"encoding/json"
	"io"
	"os"

	"badgerbadgerbadgerbadger.dev/goplay/pkg/util"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

var ErrTwoManyInputs = errors.New("too many input types, provide only one")

// FromJsonFile reads a json file from the given path and
// loads into the target interface. An error is returned if any
// is encountered during the file reading or json unmarshalling.
func FromJsonFile(path string, target interface{}) error {

	f, err := os.Open(path)
	if err != nil {
		return errors.Wrap(err, "failed to open file")
	}
	defer f.Close()

	return FromReader(f, target)
}

// FromReader reads bytes off the given reader and
// loads into the target interface. An error is
// returned if any is encountered during json unmarshalling.
func FromReader(r io.Reader, target interface{}) error {

	decoder := json.NewDecoder(r)

	err := decoder.Decode(target)
	if err != nil {
		return errors.Wrap(err, "failed to decide json")
	}

	return nil
}

type opts struct {
	filePath *string
	reader   io.Reader
	parseEnv bool
}

type optfunc = func(*opts)

func WithFilePath(path string) optfunc {
	return func(o *opts) {
		o.filePath = &path
	}
}

func WithReader(r io.Reader) optfunc {
	return func(o *opts) {
		o.reader = r
	}
}

func WithEnv() optfunc {
	return func(o *opts) {
		o.parseEnv = true
	}
}

// WithOptions provides many ways to load
// config into `target`.
func WithOptions(target any, funcs ...optfunc) error {

	o := opts{}
	util.ApplyOnType(&o, funcs)

	if o.filePath != nil && o.reader != nil {
		return ErrTwoManyInputs
	}

	if o.filePath != nil {
		err := FromJsonFile(*o.filePath, target)
		if err != nil {
			return err
		}
	}

	if o.reader != nil {
		err := FromReader(o.reader, target)
		if err != nil {
			return err
		}
	}

	if o.parseEnv {
		if err := envconfig.Process("", target); err != nil {
			return errors.Wrap(err, "failed to load config from env")
		}
	}

	return nil
}
