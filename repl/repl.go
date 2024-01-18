package repl

import (
	"bufio"
	"fmt"
	"io"
	"monkey/evaluator"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"strings"
)

const PROMPT = "==> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironemnt()

	for {
		fmt.Fprint(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		switch strings.Trim(line, " ") {
		case "quit":
			return
		default:
			l := lexer.New(line)
			p := parser.New(l)

			if res, err := p.ParseProgram(); err == nil {
				t := &evaluator.TreeWalker{}
				if eval, err := t.Eval(res, env); err == nil {
					io.WriteString(out, eval.Inspect())
				} else {
					io.WriteString(out, err.Error())
				}
			} else {
				io.WriteString(out, err.Error())
			}
			io.WriteString(out, "\n")
		}
	}
}
