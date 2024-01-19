package lexer

import (
	"unicode/utf8"

	"monkey/token"
)

var singleCharMatch = map[rune]token.TokenType{
	'=': token.ASSIGN,
	';': token.SEMICOLON,
	'(': token.LPAREN,
	')': token.RPAREN,
	',': token.COMMA,
	'+': token.PLUS,
	'-': token.MINUS,
	'*': token.ASTERISK,
	'/': token.SLASH,
	'{': token.LBRACE,
	'}': token.RBRACE,
	'^': token.CARET,
	'|': token.PIPE,
	'&': token.AMPERSAND,
	'%': token.PERCENT,
	'<': token.LANG,
	'>': token.RANG,
	'[': token.LBRACKET,
	']': token.RBRACKET,
	':': token.COLON,
}

var doubleCharMatch = map[string]token.TokenType{
	token.EQ:      token.EQ,
	token.NEQ:     token.NEQ,
	token.GEQ:     token.GEQ,
	token.LEQ:     token.LEQ,
	token.XOR_EQ:  token.XOR_EQ,
	token.OR_EQ:   token.OR_EQ,
	token.AND_EQ:  token.AND_EQ,
	token.MIN_EQ:  token.MIN_EQ,
	token.MUL_EQ:  token.MUL_EQ,
	token.DIV_EQ:  token.DIV_EQ,
	token.PERC_EQ: token.PERC_EQ,
	token.AND:     token.AND,
	token.OR:      token.OR,
	token.SHOVL:   token.SHOVL,
	token.SHOVR:   token.SHOVR,
}

var keywordMatch = map[string]token.TokenType{
	"fn":     token.FUNCTION,
	"let":    token.LET,
	"if":     token.IF,
	"else":   token.ELSE,
	"return": token.RETURN,
	"true":   token.TRUE,
	"false":  token.FALSE,
	"!":      token.BANG, // putting bang here for convenience
}

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           rune
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	width := 1

	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch, width = utf8.DecodeRuneInString(l.input[l.readPosition:])
	}

	l.position = l.readPosition
	l.readPosition += width
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.eatWhitespace()

	if l.ch == 0 {
		tok = token.New(token.EOF, "")
	} else if val, ok := doubleCharMatch[string(l.ch)+string(l.peekChar())]; ok {
		tok = token.New(val, string(l.ch)+string(l.peekChar()))
		l.readChar()
		l.readChar()
	} else if isLetter(l.ch) {
		tok = token.New(l.handleIdentifier())
	} else if isDigit(l.ch) {
		tok = token.New(token.INT, l.readNumber())
	} else if val, ok := singleCharMatch[l.ch]; ok {
		tok = token.New(val, string(l.ch))
		l.readChar()
	} else {
		switch l.ch {
		case '"':
			tok = token.New(token.STRING, l.readString())
			l.readChar()
		default:
			tok = token.New(token.ILLEGAL, string(l.ch))
		}
	}

	return tok
}

func (l *Lexer) readNumber() string {
	pos := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[pos:l.position]
}

func (l *Lexer) readIdentifier() string {
	pos := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[pos:l.position]
}

func (l *Lexer) handleIdentifier() (token.TokenType, string) {
	val := l.readIdentifier()
	if match, ok := keywordMatch[val]; ok {
		return match, val
	} else {
		return token.IDENT, val
	}
}

func isLetter(r rune) bool {
	return 'a' <= r && r <= 'z' || r == '_' || r == '?' || r == '!' || 'A' <= r && r <= 'Z'
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) eatWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		ch, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
		return ch
	}
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}

	return l.input[position:l.position]
}
