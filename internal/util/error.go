package util

type GenericMap map[string]interface{}

// Must is a quick way to panic in programs if an error exists
// with a helpful message being printed. It helps avoid having
// check for errors at every step of a top-level program.
func Must(err error, msgs ...string) {
	if err == nil {
		return
	}

	finalMessage := err.Error()

	for _, msg := range msgs {
		finalMessage = finalMessage + " " + msg
	}

	panic(finalMessage)
}
