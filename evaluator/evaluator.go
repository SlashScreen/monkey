package evaluator

import (
	"fmt"
	"monkey/ast"
	"monkey/object"
)

type Evaluator interface {
	Eval(ast.Node) (object.Object, error)
}

type EvalError struct {
	msg string
}

func (e *EvalError) Error() string {
	return e.msg
}

func createEvalError(message string, args ...any) *EvalError {
	return &EvalError{msg: fmt.Sprintf(message, args...)}
}
