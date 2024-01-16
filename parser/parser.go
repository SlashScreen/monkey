package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
)

// Error

type ParseError struct {
	msg string
}

func (e *ParseError) Error() string {
	return e.msg
}

func createParseError(message string, args ...any) *ParseError {
	return &ParseError{msg: fmt.Sprintf(message, args...)}
}

// Parser

type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}

	// Set both tokens
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() (*ast.Program, error) {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		} else {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program, nil
}

// Statements

func (p *Parser) parseStatement() (ast.Statement, error) {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return nil, createParseError("Expected statement, got token %q", p.curToken.Type)
	}
}

func (p *Parser) parseLetStatement() (*ast.LetStatement, error) {
	stmt := &ast.LetStatement{Token: p.curToken}

	if res, err := p.expect(token.IDENT); !res {
		return nil, err
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if res, err := p.expect(token.ASSIGN); !res {
		return nil, err
	}

	exp, experr := p.parseExpression()

	if experr != nil {
		return nil, experr
	}

	stmt.Value = exp

	return stmt, nil
}

func (p *Parser) parseReturnStatement() (*ast.ReturnStatement, error) {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	exp, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	stmt.ReturnValue = exp
	return stmt, nil
}

// Expressions

func (p *Parser) parseExpression() (ast.Expression, error) {
	return nil, createParseError("Unimplemented.")
}

// Utilities

func (p *Parser) expect(t token.TokenType) (bool, error) {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true, nil
	} else {
		return false, createParseError("Expected token type %q, got %q instead", t, p.peekToken.Type)
	}
}

func (p *Parser) peekTokenIs(t token.TokenType) bool { return p.peekToken.Type == t }

func (p *Parser) curTokenIs(t token.TokenType) bool { return p.curToken.Type == t }
