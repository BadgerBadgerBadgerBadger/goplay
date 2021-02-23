package main

import (
	"fmt"
	"strings"
)

func Rage(inp string) string {

	r := NewRager()

	upperCased := strings.ToUpper(inp)
	tripleExclaimed := strings.ReplaceAll(
		upperCased, "!",
		fmt.Sprintf("!!! %s", r.Rand()),
	)
	questionExclaimed := strings.ReplaceAll(
		tripleExclaimed, "?",
		fmt.Sprintf("?! %s", r.Rand()),
	)

	return questionExclaimed
}
