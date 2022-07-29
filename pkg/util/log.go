package util

import log "github.com/sirupsen/logrus"

func LogOnError(err error, msg string) {
	if err == nil {
		return
	}

	log.Printf("Error: %s, %s\n", err, msg)
}
