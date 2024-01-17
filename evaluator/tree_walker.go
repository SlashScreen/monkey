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
	case *ast.Boolean:
		return object.NativeToBooleanObject(node.Value), nil
	case *ast.PrefixExpression:
		if right, err := t.Eval(node.Right); err == nil {
			return t.evalPrefix(node.Operator, right)
		} else {
			return nil, err
		}
	case *ast.InfixExpression:
		left, err := t.Eval(node.Left)
		if err != nil {
			return object.NULL, err
		}
		right, err := t.Eval(node.Right)
		if err != nil {
			return object.NULL, err
		}
		return t.evalInfix(node.Operator, left, right)
	// Else
	default:
		return object.NULL, createEvalError("Unimplemented.")
	}
}

func (t *TreeWalker) evalStatements(stmts []ast.Statement) (object.Object, error) {
	var result object.Object

	for _, statement := range stmts {
		if res, err := t.Eval(statement); err == nil {
			result = res
		} else {
			return object.NULL, err
		}
	}

	return result, nil
}

func (t *TreeWalker) evalPrefix(op string, right object.Object) (object.Object, error) {
	switch op {
	case "!":
		return t.evalBangOperator(right)
	case "-":
		return t.evalNegOperator(right)
	default:
		return object.NULL, nil
	}
}

func (t *TreeWalker) evalBangOperator(right object.Object) (object.Object, error) {
	switch right {
	case object.TRUE:
		return object.FALSE, nil
	case object.FALSE:
		return object.TRUE, nil
	case object.NULL:
		return object.TRUE, nil
	default:
		return object.NULL, createEvalError("Cannot apply ! operator to %q", right.Type())
	}
}

func (t *TreeWalker) evalNegOperator(right object.Object) (object.Object, error) {
	if right.Type() != object.INTEGER_OBJ {
		return nil, createEvalError("Cannot negate a value of type %s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}, nil
}

func (t *TreeWalker) evalInfix(op string, left, right object.Object) (object.Object, error) {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return t.evalIntegerInfix(op, left, right)
	case op == "==":
		return object.NativeToBooleanObject(left == right), nil
	case op == "!=":
		return object.NativeToBooleanObject(left != right), nil

	default:
		return object.NULL, createEvalError("Operator %q cannot operate with a %q and %q", op, left.Type(), right.Type())
	}
}

func (t *TreeWalker) evalIntegerInfix(op string, left, right object.Object) (object.Object, error) {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch op {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}, nil
	case "-":
		return &object.Integer{Value: leftVal - rightVal}, nil
	case "*":
		return &object.Integer{Value: leftVal * rightVal}, nil
	case "/":
		return &object.Integer{Value: leftVal / rightVal}, nil
	case "%":
		return &object.Integer{Value: leftVal % rightVal}, nil
	case "|":
		return &object.Integer{Value: leftVal | rightVal}, nil
	case "&":
		return &object.Integer{Value: leftVal & rightVal}, nil
	case "^":
		return &object.Integer{Value: leftVal ^ rightVal}, nil
	case "<<":
		return &object.Integer{Value: leftVal << rightVal}, nil
	case ">>":
		return &object.Integer{Value: leftVal >> rightVal}, nil
	case "<":
		return object.NativeToBooleanObject(leftVal < rightVal), nil
	case ">":
		return object.NativeToBooleanObject(leftVal > rightVal), nil
	case "==":
		return object.NativeToBooleanObject(leftVal == rightVal), nil
	case "!=":
		return object.NativeToBooleanObject(leftVal != rightVal), nil
	default:
		return object.NULL, createEvalError("Operator %q cannot operate with a %q and %q", op, left.Type(), right.Type())
	}
}
