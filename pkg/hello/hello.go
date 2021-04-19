package hello

import "fmt"

// Greet creates a greeting.
func Greet(names ...interface{}) (greeting string) {
	if len(names) == 0 {
		return "Hello, World!"
	}

	for _, name := range names {
		greeting = fmt.Sprintf("%vHello, %v! ", greeting, name)
	}
	return greeting
}
