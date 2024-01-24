package object

import "fmt"

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{
		"len", &Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return &Error{Message: newError("wrong number of arguments to 'len': %d", len(args))}
				}
				switch arg := args[0].(type) {
				case *String:
					return &Integer{Value: int64(len(arg.Value))}
				case *Array:
					return &Integer{Value: int64(len(arg.Elements))}
				default:
					return &Error{Message: newError("cannot take the length of %s", args[0].Type())}
				}
			},
		},
	},
	{
		"puts", &Builtin{
			Fn: func(args ...Object) Object {
				for _, arg := range args {
					fmt.Println(arg.Inspect())
				}
				return NULL
			},
		},
	},
	{
		"first",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return &Error{Message: newError("wrong number of arguments. got=%d, want=1",
					len(args))}
			}
			if args[0].Type() != ARRAY_OBJ {
				return &Error{Message: newError("argument to `first` must be ARRAY, got %s",
					args[0].Type())}
			}

			arr := args[0].(*Array)
			if len(arr.Elements) > 0 {
				return arr.Elements[0]
			}

			return nil
		},
		},
	},
	{
		"last",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return &Error{Message: newError("wrong number of arguments. got=%d, want=1",
					len(args))}
			}
			if args[0].Type() != ARRAY_OBJ {
				return &Error{Message: newError("argument to `last` must be ARRAY, got %s",
					args[0].Type())}
			}

			arr := args[0].(*Array)
			length := len(arr.Elements)
			if length > 0 {
				return arr.Elements[length-1]
			}

			return nil
		},
		},
	},
	{
		"rest",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return &Error{Message: newError("wrong number of arguments. got=%d, want=1",
					len(args))}
			}
			if args[0].Type() != ARRAY_OBJ {
				return &Error{Message: newError("argument to `rest` must be ARRAY, got %s",
					args[0].Type())}
			}

			arr := args[0].(*Array)
			length := len(arr.Elements)
			if length > 0 {
				newElements := make([]Object, length-1)
				copy(newElements, arr.Elements[1:length])
				return &Array{Elements: newElements}
			}

			return nil
		},
		},
	},
}

func newError(format string, a ...interface{}) error {
	return fmt.Errorf(format, a...)
}

func GetBuiltinByName(name string) *Builtin {
	for _, def := range Builtins {
		if def.Name == name {
			return def.Builtin
		}
	}
	return nil
}
