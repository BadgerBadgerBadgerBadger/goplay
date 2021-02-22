package util

import (
	"encoding/json"
	"io/ioutil"
)

// ConfigFromJsonFile reads a json file from the given path and
// loads into the target interface. An error is returned if any
// is encountered during the file reading or json unmarshalling.
func ConfigFromJsonFile(path string, target interface{}) error {

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, target)
	return err
}
