package tokens

import "fmt"

type TokenType int

const (

	// One or two character tokens.
	Bang TokenType = iota
	BangBang
	BangBangBang
	Question
	QuestionBang
	Period
	Dot

	// Literals
	BangStr
	Plain
)

type Token struct {
	Type    TokenType
	Literal interface{}
}

func NewToken(Type TokenType, Literal interface{}) Token {
	return Token{
		Type:    Type,
		Literal: Literal,
	}
}

func (t Token) String() string {
	return fmt.Sprintf("%d %v", t.Type, t.Literal)
}
