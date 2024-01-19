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
	case *ast.FunctionLiteral:
		return &object.Function{Parameters: node.Parameters, Body: node.Body, Env: env}, nil
	case *ast.CallExpression:
		function, err := t.Eval(node.Function, env)
		if err != nil {
			return function, err
		}

		args, err := t.evalExpressions(node.Arguments, env)
		if err != nil {
			return object.ErrorPair(err)
		}
		if len(args) == 1 && isError(args[0]) {
			return args[0], nil
		}

		return t.applyFunction(function, args)
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}, nil
	case *ast.ArrayLiteral:
		elements, err := t.evalExpressions(node.Elements, env)
		if len(elements) == 1 && err != nil {
			return elements[0], err
		}
		return &object.Array{Elements: elements}, nil
	case *ast.IndexExpression:
		left, err := t.Eval(node.Left, env)
		if err != nil {
			return left, err
		}
		index, err := t.Eval(node.Index, env)
		if err != nil {
			return index, err
		}
		return t.evalIndexExpression(left, index)
	case *ast.HashLiteral:
		return t.evalHashLiteral(node, env)
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
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return t.evalStringInfix(op, left, right)
	case left.Type() == object.ARRAY_OBJ:
		return t.evalArrayInfix(op, left, right)
	default:
		return object.ErrorPair(createEvalError("operator %s cannot operate with a %s and %s", op, left.Type(), right.Type()))
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
		return object.ErrorPair(createEvalError("operator %s cannot operate with a %s and %s", op, left.Type(), right.Type()))
	}
}

func (t *TreeWalker) evalStringInfix(op string, left, right object.Object) (object.Object, error) {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	switch op {
	case "+", "<<":
		return &object.String{Value: leftVal + rightVal}, nil
	default:
		return object.ErrorPair(createEvalError("operator %s cannot operate with a %s and %s", op, left.Type(), right.Type()))
	}
}

func (t *TreeWalker) evalArrayInfix(op string, left, right object.Object) (object.Object, error) {
	switch op {
	case "<<":
		val := builtins["push"].Fn(left, right)
		if isError(val) {
			return val, val.(*object.Error).Message
		} else {
			return val, nil
		}
	default:
		return object.ErrorPair(createEvalError("operator %s cannot operate with a %s and %s", op, left.Type(), right.Type()))
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

func (t *TreeWalker) evalExpressions(exps []ast.Expression, env *object.Environment) ([]object.Object, error) {
	var result []object.Object

	for _, exp := range exps {
		if evaluated, err := t.Eval(exp, env); err == nil {
			result = append(result, evaluated)
		} else {
			return []object.Object{evaluated}, err
		}
	}

	return result, nil
}

func (t *TreeWalker) applyFunction(fn object.Object, args []object.Object) (object.Object, error) {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := t.extendFunctionEnv(fn, args)
		evaluated, err := t.Eval(fn.Body, extendedEnv)
		if err != nil {
			return object.ErrorPair(err)
		}

		return t.unwrapReturnValue(evaluated), nil
	case *object.Builtin:
		return fn.Fn(args...), nil
	default:
		return object.ErrorPair(createEvalError("not a function: %s", fn.Type()))
	}
}

func (t *TreeWalker) extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIndex, param := range fn.Parameters {
		env.Set(param.Value, args[paramIndex])
	}

	return env
}

func (t *TreeWalker) unwrapReturnValue(obj object.Object) object.Object {
	if ret, ok := obj.(*object.ReturnValue); ok {
		return ret.Value
	}
	return obj
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
		if builtin, ok := builtins[node.Value]; ok {
			return builtin, nil
		}
		err := createEvalError("identifier not found: %s", node.Value)
		return &object.Error{Message: err}, err
	}
}

func (t *TreeWalker) evalIndexExpression(left, index object.Object) (object.Object, error) {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return t.evalArrayIndexExpression(left, index)
	default:
		return object.ErrorPair(createEvalError("Cannot index array with type %s", left.Type()))
	}
}

func (t *TreeWalker) evalArrayIndexExpression(array, index object.Object) (object.Object, error) {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > max {
		return object.ErrorPair(createEvalError("index out of bounds"))
	}
	return arrayObject.Elements[idx], nil
}

func (t *TreeWalker) evalHashLiteral(node *ast.HashLiteral, env *object.Environment) (object.Object, error) {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key, err := t.Eval(keyNode, env)
		if err != nil {
			return key, err
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return object.ErrorPair(createEvalError("unusable as hash key: %s", key.Type()))
		}

		value, err := t.Eval(valueNode, env)
		if err != nil {
			return value, err
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}, nil
}
