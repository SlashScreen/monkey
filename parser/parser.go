package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > <
	SUM         // + -
	PRODUCT     // / * %
	PREFIX      // -, !
	SPECIAL     // bitwise
	CALL        // foo(x)
)

var precedences = map[token.TokenType]int{
	token.EQ:        EQUALS,
	token.NEQ:       EQUALS,
	token.LANG:      LESSGREATER,
	token.RANG:      LESSGREATER,
	token.PLUS:      SUM,
	token.MINUS:     SUM,
	token.SLASH:     PRODUCT,
	token.ASTERISK:  PRODUCT,
	token.PERCENT:   PRODUCT,
	token.PIPE:      SPECIAL,
	token.AMPERSAND: SPECIAL,
	token.CARET:     SPECIAL,
	token.SHOVL:     SPECIAL,
	token.SHOVR:     SPECIAL,
}

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

// Parser Functions

type (
	prefixParseFn func() (ast.Expression, error)
	infixParseFn  func(ast.Expression) (ast.Expression, error)
)

// Parser

type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, prefixParseFns: make(map[token.TokenType]prefixParseFn), infixParseFns: make(map[token.TokenType]infixParseFn)}
	p.registerPrefix(token.IDENT, p.parseIdent)
	p.registerPrefix(token.INT, p.parseInt)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.RPAREN, p.parseGrouped)
	for k := range precedences {
		p.registerInfix(k, p.parseInfixExpression)
	}

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
		return p.parseExpressionStatement()
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

	exp, experr := p.parseExpression(LOWEST)

	if experr != nil {
		return nil, experr
	}

	stmt.Value = exp

	return stmt, nil
}

func (p *Parser) parseReturnStatement() (*ast.ReturnStatement, error) {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	exp, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}

	stmt.ReturnValue = exp
	return stmt, nil
}

func (p *Parser) parseExpressionStatement() (*ast.ExpressionStatement, error) {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	exp, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	stmt.Expression = exp

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt, nil
}

// Expressions

func (p *Parser) parseExpression(precedence int) (ast.Expression, error) {
	prefix := p.prefixParseFns[p.curToken.Type]

	if prefix == nil {
		return nil, createParseError("No prefix expression found for %q.", p.curToken.Type)
	}

	lhs, err := prefix()
	if err != nil {
		return nil, err
	}

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return lhs, nil
		}

		p.nextToken()

		lhs, err = infix(lhs)
		if err != nil {
			return nil, err
		}
	}

	return lhs, nil
}

func (p *Parser) parseIdent() (ast.Expression, error) {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}, nil
}

func (p *Parser) parseInt() (ast.Expression, error) {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		return nil, createParseError("Expected integer literal, got unparseable %q instead", p.curToken.Literal)
	}

	lit.Value = value
	return lit, nil
}

func (p *Parser) parsePrefixExpression() (ast.Expression, error) {
	expr := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	rhs, err := p.parseExpression(PREFIX)
	if err != nil {
		return nil, err
	}
	expr.Right = rhs

	return expr, nil
}

func (p *Parser) parseInfixExpression(left ast.Expression) (ast.Expression, error) {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	rhs, err := p.parseExpression(precedence)
	if err != nil {
		return nil, err
	}
	expression.Right = rhs

	return expression, nil
}

func (p *Parser) parseBoolean() (ast.Expression, error) {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}, nil
}

func (p *Parser) parseGrouped() (ast.Expression, error) {
	p.nextToken()

	exp, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}

	if ok, _ := p.expect(token.RPAREN); !ok {
		return nil, createParseError("Expected closing parenthesis.")
	}

	return exp, nil
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

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}
