package repl

import (
	"bufio"
	"fmt"
	"io"
	"monkey/lexer"
	"monkey/parser"
	"strings"
)

const PROMPT = "==> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

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
				io.WriteString(out, res.String())
			} else {
				io.WriteString(out, err.Error())
			}
			io.WriteString(out, "\n")
		}
	}
}
