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
	INDEX       // array[index]
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
	token.LPAREN:    CALL,
	token.LBRACKET:  INDEX,
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
	p.registerPrefix(token.LPAREN, p.parseGrouped)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	for k := range precedences {
		p.registerInfix(k, p.parseInfixExpression)
	}
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)

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
		if stmt, err := p.parseStatement(); err == nil {
			program.Statements = append(program.Statements, stmt)
		} else {
			return nil, err
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

	p.nextToken()

	if exp, err := p.parseExpression(LOWEST); err == nil {
		stmt.Value = exp
	} else {
		return nil, err
	}

	if p.peekTokenIs(token.SEMICOLON) { // FIXME: This
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseReturnStatement() (*ast.ReturnStatement, error) {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	if exp, err := p.parseExpression(LOWEST); err == nil {
		stmt.ReturnValue = exp
	} else {
		return nil, err
	}

	p.nextToken() // Could be an issue here

	return stmt, nil
}

func (p *Parser) parseExpressionStatement() (*ast.ExpressionStatement, error) {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	if exp, err := p.parseExpression(LOWEST); err == nil {
		stmt.Expression = exp
	} else {
		return nil, err
	}

	p.nextToken()

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseBlockStatement() (*ast.BlockStatement, error) {
	block := &ast.BlockStatement{Token: p.curToken, Statements: []ast.Statement{}}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) { // may cause weird behavior
		if stmt, err := p.parseStatement(); err == nil {
			block.Statements = append(block.Statements, stmt)
		} else {
			return nil, err
		}

		p.nextToken()
	}

	return block, nil
}

// Expressions

func (p *Parser) parseExpression(precedence int) (ast.Expression, error) {
	prefix := p.prefixParseFns[p.curToken.Type]

	if prefix == nil {
		return nil, createParseError("No prefix expression found for %q (%q).", p.curToken.Type, p.curToken.Literal)
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

	if value, err := strconv.ParseInt(p.curToken.Literal, 0, 64); err == nil {
		lit.Value = value
	} else {
		return nil, createParseError("Expected integer literal, got unparseable %q instead", p.curToken.Literal)
	}

	return lit, nil
}

func (p *Parser) parsePrefixExpression() (ast.Expression, error) {
	expr := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	if rhs, err := p.parseExpression(PREFIX); err == nil {
		expr.Right = rhs
	} else {
		return nil, err
	}

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

	if rhs, err := p.parseExpression(precedence); err == nil {
		expression.Right = rhs
	} else {
		return nil, err
	}

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

func (p *Parser) parseIfExpression() (ast.Expression, error) {
	expression := &ast.IfExpression{Token: p.curToken}

	if ok, err := p.expect(token.LPAREN); !ok {
		return nil, err
	}

	p.nextToken()
	if cond, err := p.parseExpression(LOWEST); err == nil {
		expression.Condition = cond
	} else {
		return nil, err
	}

	if ok, err := p.expect(token.RPAREN); !ok {
		return nil, err
	}
	if ok, err := p.expect(token.LBRACE); !ok {
		return nil, err
	}

	if cons, err := p.parseBlockStatement(); err == nil {
		expression.Consequence = cons
	} else {
		return nil, err
	}

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if ok, err := p.expect(token.LBRACE); !ok {
			return nil, err
		}

		if alt, err := p.parseBlockStatement(); err == nil {
			expression.Alternative = alt
		} else {
			return nil, err
		}
	}

	return expression, nil
}

func (p *Parser) parseFunctionLiteral() (ast.Expression, error) {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if ok, err := p.expect(token.LPAREN); !ok {
		return nil, err
	}

	if params, err := p.parseFunctionParameters(); err == nil {
		lit.Parameters = params
	} else {
		return nil, err
	}

	if ok, err := p.expect(token.LBRACE); !ok {
		return nil, err
	}

	if body, err := p.parseBlockStatement(); err == nil {
		lit.Body = body
	} else {
		return nil, err
	}

	return lit, nil
}

func (p *Parser) parseFunctionParameters() ([]*ast.Identifier, error) {
	ids := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) { // Empty list
		p.nextToken()
		return ids, nil
	}

	p.nextToken()

	// First ident
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	ids = append(ids, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()

		id := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		ids = append(ids, id)
	}

	if ok, err := p.expect(token.RPAREN); !ok {
		return nil, err
	}

	return ids, nil
}

func (p *Parser) parseCallExpression(function ast.Expression) (ast.Expression, error) {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	if args, err := p.parseCallArguments(); err == nil {
		exp.Arguments = args
	} else {
		return nil, err
	}
	return exp, nil
}

func (p *Parser) parseCallArguments() ([]ast.Expression, error) {
	args := []ast.Expression{}

	if p.peekTokenIs(token.RPAREN) { // Empty
		p.nextToken()
		return args, nil
	}

	p.nextToken()
	if arg, err := p.parseExpression(LOWEST); err == nil {
		args = append(args, arg)
	} else {
		return nil, err
	}

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()

		if arg, err := p.parseExpression(LOWEST); err == nil {
			args = append(args, arg)
		} else {
			return nil, err
		}
	}

	if ok, err := p.expect(token.RPAREN); !ok {
		return nil, err
	}

	return args, nil
}

func (p *Parser) parseStringLiteral() (ast.Expression, error) {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}, nil
}

func (p *Parser) parseArrayLiteral() (ast.Expression, error) {
	array := &ast.ArrayLiteral{Token: p.curToken}

	if els, err := p.parseExpressionList(token.RBRACKET); err == nil {
		array.Elements = els
	} else {
		return nil, err
	}

	return array, nil
}

func (p *Parser) parseExpressionList(end token.TokenType) ([]ast.Expression, error) {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list, nil
	}

	p.nextToken()

	if expr, err := p.parseExpression(LOWEST); err == nil {
		list = append(list, expr)
	} else {
		return []ast.Expression{}, err
	}

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		if expr, err := p.parseExpression(LOWEST); err == nil {
			list = append(list, expr)
		} else {
			return []ast.Expression{}, err
		}
	}

	if ok, err := p.expect(end); !ok {
		return []ast.Expression{}, err
	}
	return list, nil
}

func (p *Parser) parseIndexExpression(left ast.Expression) (ast.Expression, error) {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	if i, err := p.parseExpression(LOWEST); err == nil {
		exp.Index = i
	} else {
		return nil, err
	}

	if ok, err := p.expect(token.RBRACKET); !ok {
		return nil, err
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
