package main

import (
	"fmt"
	"strings"

	"badgerbadgerbadgerbadger.dev/goplay/cmd/rant/scanner"
	"badgerbadgerbadgerbadger.dev/goplay/cmd/rant/tokens"
)

func Rant(inp string) string {

	s := scanner.NewScanner(inp)
	scannedTokens := s.ScanTokens()

	// if the last token is a plain string, append an ending bang
	if scannedTokens[len(scannedTokens)-1].Type == tokens.Plain {
		scannedTokens = append(scannedTokens, tokens.NewToken(tokens.Bang, ""))
	}

	r := NewRager()

	var out strings.Builder

	for _, token := range scannedTokens {

		var result string

		switch token.Type {
		case tokens.Plain:

			result = strings.ToUpper(token.Literal.(string))

		case tokens.Dot:
			result = "."

		case tokens.Period:
			result = fmt.Sprintf("!!! %s ", r.Rand())

		case tokens.Question, tokens.QuestionBang:
			result = fmt.Sprintf("?! %s ", r.Rand())

		case tokens.Bang, tokens.BangBang, tokens.BangBangBang:
			result = fmt.Sprintf("!!! %s ", r.Rand())

		case tokens.BangStr:
			result = fmt.Sprintf("%s %s ", token.Literal, r.Rand())
		}

		_, err := fmt.Fprint(&out, result)
		if err != nil {
			panic(err)
		}
	}

	return out.String()
}
