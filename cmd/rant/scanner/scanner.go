package scanner

import (
	"badgerbadgerbadgerbadger.dev/goplay/cmd/rant/tokens"
)

const (
	runeNull = 0x0
)

type Scanner struct {
	source []rune
	tokens []tokens.Token

	start   int
	current int
}

func NewScanner(source string) Scanner {
	return Scanner{
		source:  []rune(source),
		start:   0,
		current: 0,
	}
}

func (s *Scanner) ScanTokens() []tokens.Token {

	for !s.isAtEnd() {

		s.start = s.current
		s.scanToken()
	}

	return s.tokens
}

func (s Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *Scanner) scanToken() {

	c := s.advance()
	asStr := string(c)
	_ = asStr

	switch c {
	case '!', '?':

		// if second char isn't a banger or we're at the end,
		// we've got a single char token
		if !isBanger(s.peek()) || s.isAtEnd() {
			switch c {
			case '!':
				s.addToken(tokens.Bang)
			case '?':
				s.addToken(tokens.Question)
			}
			return
		}

		c = s.advance()

		// if third char isn't a banger, check for two-char token
		if !isBanger(s.peek()) || s.isAtEnd() {

			switch c {
			case '!':
				s.addToken(tokens.BangBang)
			case '?':
				s.addToken(tokens.QuestionBang)
			}
			return
		}

		c = s.advance()

		// if fourth char isn't a banger, check for three-char token
		if !isBanger(s.peek()) {

			switch c {
			case '!':
				s.addToken(tokens.BangBangBang)
			case '?':
				s.addToken(tokens.QuestionBang)
			}
			return
		}

		// okay, so now we're dealing with a bang string
		s.bangStr()
	case '.':
		// if the dot isn't followed by a space, we can assume
		// it's being used for something else like a username or
		// a floating point. not perfect but it's what we have
		// for now

		tokenToAdd := tokens.Dot

		if s.match(' ') || s.isAtEnd() {
			tokenToAdd = tokens.Period
		}

		s.addToken(tokenToAdd)
	default:

		s.plain()
	}
}

func (s *Scanner) bangStr() {
	for isBanger(s.peek()) {
		s.advance()
	}

	text := string(s.source[s.start:s.current])

	s.addTokenWithLiteral(tokens.BangStr, text)
}

// plain scans for a sequence of chars that isn't a period,
// or any of the bangers. Basically, everything else.
func (s *Scanner) plain() {

	// keep consuming till we hit a banger or a period
	for !s.isAtEnd() && !isBanger(s.peek()) && !isPeriod(s.peek(), s.peekNext()) {

		s.advance()
	}

	s.addTokenWithLiteral(tokens.Plain, string(s.source[s.start:s.current]))
}

// peek returns the next rune to be consumed without actually consuming it.
// If end of source has been reached, the NUL rune is returned.
func (s *Scanner) peek() rune {
	if s.isAtEnd() {
		return runeNull
	}

	return s.source[s.current]
}

// peekNext returns the next to next rune to be consumed without actually
// consuming it. If end of source has been reached, the NUL rune is returned.
func (s *Scanner) peekNext() rune {
	if s.current+1 >= len(s.source) {
		return 0x0
	}

	return s.source[s.current+1]
}

// peekNextNext returns the next to next to next rune to be consumed without
// actually consuming it. If end of source has been reached, the NUL rune is returned.
func (s *Scanner) peekNextNext() rune {
	if s.current+1 > len(s.source) {
		return 0x0
	}

	return s.source[s.current+1]
}

// advance consumes a single rune and returns it, having moved the
// scanning pointer one rune forward.
func (s *Scanner) advance() rune {
	s.current++
	return s.source[s.current-1]
}

// match returns true if the next rune to be consumed is the
// `expected` rune. The scanning pointer is moved
// ahead, in this case. If the rune does not match or end of
// source has been reached, it returns false.
func (s *Scanner) match(expected rune) bool {
	if s.isAtEnd() {
		return false
	}

	if s.source[s.current] != expected {
		return false
	}

	s.current++
	return true
}

func (s *Scanner) addToken(typeOfToken tokens.TokenType) {
	s.addTokenWithLiteral(typeOfToken, nil)
}

func (s *Scanner) addTokenWithLiteral(typeOfToken tokens.TokenType, literal interface{}) {
	s.tokens = append(
		s.tokens,
		tokens.NewToken(typeOfToken, literal),
	)
}

func isBanger(c rune) bool {
	return c == '!' || c == '?'
}

func isPeriod(first, second rune) bool {
	return first == '.' && (second == ' ' || second == runeNull)
}
