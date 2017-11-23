package decimal

import (
	"fmt"
	"os"
	"strings"
)

func ExampleBig_reversePolishNotationCalculator() {
	const input = "15 7 1 1 + - / 3 * 2 1 1 + + - 5 * 3 / ="
	var stack []*Big
Loop:
	for _, tok := range strings.Split(input, " ") {
		last := len(stack) - 1
		switch tok {
		case "+":
			x := stack[last-1]
			x.Add(x, stack[last])
			stack = stack[:last]
		case "-":
			x := stack[last-1]
			x.Sub(x, stack[last])
			stack = stack[:last]
		case "/":
			x := stack[last-1]
			x.Quo(x, stack[last])
			stack = stack[:last]
		case "*":
			x := stack[last-1]
			x.Mul(x, stack[last])
			stack = stack[:last]
		case "=":
			break Loop
		default:
			x := WithContext(Context128)
			if _, ok := x.SetString(tok); !ok {
				fmt.Fprintf(os.Stderr, "invalid decimal: %v\n", x.Context.Err)
				os.Exit(1)
			}
			stack = append(stack, x)
		}
	}
	fmt.Printf("%+6.4g\n", stack[0])
	// Output: +8.333
}
