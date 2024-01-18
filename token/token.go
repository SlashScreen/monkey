package token

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT  = "IDENT"
	INT    = "INT"
	STRING = "STRING"

	ASSIGN    = "="
	PLUS      = "+"
	MINUS     = "-"
	SLASH     = "/"
	ASTERISK  = "*"
	BANG      = "!"
	CARET     = "^"
	PIPE      = "|"
	AMPERSAND = "&"
	LANG      = "<"
	RANG      = ">"
	PERCENT   = "%"

	EQ      = "=="
	NEQ     = "!="
	GEQ     = ">="
	LEQ     = "<="
	XOR_EQ  = "^="
	OR_EQ   = "|="
	AND_EQ  = "&="
	PLUS_EQ = "+="
	MIN_EQ  = "-="
	MUL_EQ  = "*="
	DIV_EQ  = "/="
	PERC_EQ = "%="
	AND     = "&&"
	OR      = "||"
	SHOVL   = "<<"
	SHOVR   = ">>"

	COMMA     = ","
	SEMICOLON = ";"

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"

	FUNCTION = "FUNCTION"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	RETURN   = "RETURN"
	IF       = "IF"
	ELSE     = "ELSE"
)

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

func New(t TokenType, v string) Token {
	return Token{Type: t, Literal: v}
}
