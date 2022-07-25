package util

import (
	"io/ioutil"
	"net/http"
)

func ReadBodyAsString(r *http.Request) string {

	body, err := ioutil.ReadAll(r.Body)
	Must(err, "failed to read body")

	return string(body)
}
