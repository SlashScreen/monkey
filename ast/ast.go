package ast

import (
	"fmt"
	"monkey/token"
)

type Node interface {
	TokenLiteral() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode() // Perhaps a type resolution thing too, later
}

// PROGRAM

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

// LET STATEMENT

type LetStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }

// RETURN STATEMENT

type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

// IDENTIFIER

type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

// OPERATOR

type OperatorExpression struct {
	Left     Expression
	Right    Expression
	Operator token.TokenType
}

func (o *OperatorExpression) expressionNode() {}
func (o *OperatorExpression) TokenLiteral() string {
	return fmt.Sprintf("%q %q %q", o.Left.TokenLiteral(), o.Operator, o.Right.TokenLiteral())
}
