package evaluator

import (
	"monkey/ast"
	"monkey/object"
)

type TreeWalker struct{}

func (t *TreeWalker) Eval(node ast.Node) (object.Object, error) {
	switch node := node.(type) {
	// Statmements
	case *ast.Program:
		return t.evalStatements(node.Statements)
	case *ast.ExpressionStatement:
		return t.Eval(node.Expression)
	// Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}, nil
	// Else
	default:
		return nil, createEvalError("Unimplemented.")
	}
}

func (t *TreeWalker) evalStatements(stmts []ast.Statement) (object.Object, error) {
	var result object.Object

	for _, statement := range stmts {
		if res, err := t.Eval(statement); err == nil {
			result = res
		} else {
			return nil, err
		}
	}

	return result, nil
}
