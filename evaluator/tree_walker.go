package evaluator

import (
	"monkey/ast"
	"monkey/object"
)

type TreeWalker struct{}

func (t *TreeWalker) Eval(node ast.Node, env *object.Environment) (object.Object, error) {
	switch node := node.(type) {
	// Statmements
	case *ast.Program:
		return t.evalProgram(node.Statements, env)
	case *ast.ExpressionStatement:
		return t.Eval(node.Expression, env)
	// Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}, nil
	case *ast.Boolean:
		return object.NativeToBooleanObject(node.Value), nil
	case *ast.PrefixExpression:
		if right, err := t.Eval(node.Right, env); err == nil {
			if isError(right) {
				return right, right.(*object.Error).Message
			}
			return t.evalPrefix(node.Operator, right)
		} else {
			return &object.Error{Message: err}, err
		}
	case *ast.InfixExpression:
		left, err := t.Eval(node.Left, env)
		if err != nil {
			return &object.Error{Message: err}, err
		}
		if isError(left) {
			return left, left.(*object.Error).Message
		}
		right, err := t.Eval(node.Right, env)
		if err != nil {
			return &object.Error{Message: err}, err
		}
		if isError(right) {
			return right, right.(*object.Error).Message
		}
		return t.evalInfix(node.Operator, left, right)
	case *ast.BlockStatement:
		return t.evalBlock(node, env)
	case *ast.IfExpression:
		return t.evalIfExpression(node, env)
	case *ast.ReturnStatement:
		if val, err := t.Eval(node.ReturnValue, env); err == nil {
			return &object.ReturnValue{Value: val}, nil
		} else {
			return &object.Error{Message: err}, err
		}
	case *ast.LetStatement:
		if val, err := t.Eval(node.Value, env); err == nil {
			env.Set(node.Name.Value, val)
			return val, nil
		} else {
			return object.ErrorPair(err)
		}
	case *ast.Identifier:
		return t.evalIdentifier(node, env)
	// Else
	default:
		return object.NULL, createEvalError("Unimplemented.")
	}
}

func (t *TreeWalker) evalProgram(stmts []ast.Statement, env *object.Environment) (object.Object, error) {
	var result object.Object

	for _, statement := range stmts {
		if res, err := t.Eval(statement, env); err == nil {
			result = res
		} else {
			return &object.Error{Message: err}, err
		}

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value, nil
		case *object.Error:
			return result, result.Message
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
		err := createEvalError("unknown operator: %s%s", op, right.Type())
		return &object.Error{Message: err}, err
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
		err := createEvalError("cannot apply ! operator to %s", right.Type())
		return &object.Error{Message: err}, err
	}
}

func (t *TreeWalker) evalNegOperator(right object.Object) (object.Object, error) {
	if right.Type() != object.INTEGER_OBJ {
		err := createEvalError("cannot apply - operator to %s", right.Type())
		return &object.Error{Message: err}, err
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
	case left.Type() != right.Type():
		err := createEvalError("type mismatch: %s %s %s", left.Type(), op, right.Type())
		return &object.Error{Message: err}, err
	default:
		err := createEvalError("operator %s cannot operate with a %s and %s", op, left.Type(), right.Type())
		return &object.Error{Message: err}, err
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
		err := createEvalError("operator %s cannot operate with a %s and %s", op, left.Type(), right.Type())
		return &object.Error{Message: err}, err
	}
}

func (t *TreeWalker) evalIfExpression(ie *ast.IfExpression, env *object.Environment) (object.Object, error) {
	condition, err := t.Eval(ie.Condition, env)
	if err != nil {
		return &object.Error{Message: err}, err
	}
	if isError(condition) {
		return condition, condition.(*object.Error).Message
	}

	if t.isTruthy(condition) {
		return t.Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return t.Eval(ie.Alternative, env)
	} else {
		return object.NULL, nil
	}
}

func (t *TreeWalker) isTruthy(obj object.Object) bool {
	switch obj {
	case object.NULL:
		return false
	case object.TRUE:
		return true
	case object.FALSE:
		return false
	default:
		return true
	}
}

func (t *TreeWalker) evalBlock(block *ast.BlockStatement, env *object.Environment) (object.Object, error) {
	var res object.Object

	for _, statement := range block.Statements {
		if result, err := t.Eval(statement, env); err == nil {
			res = result

			if result.Type() == object.RETURN_VALUE_OBJ || result.Type() == object.ERROR_OBJ {
				return result, nil
			}
		} else {
			return &object.Error{Message: err}, err
		}
	}

	return res, nil
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func (t *TreeWalker) evalIdentifier(node *ast.Identifier, env *object.Environment) (object.Object, error) {
	if val, ok := env.Get(node.Value); ok {
		return val, nil
	} else {
		err := createEvalError("identifier not found: %s", node.Value)
		return &object.Error{Message: err}, err
	}
}
