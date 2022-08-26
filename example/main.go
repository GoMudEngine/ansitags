package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/Volte6/ansigo"
)

// accepts a string via shell pipe and processes with ansigo.Parse()
func main() {

	info, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if info.Mode()&os.ModeCharDevice != 0 || info.Size() <= 0 {

		fmt.Printf("\n%s %s\n\n",
			ansigo.Parse("<ansi fg=red bold=true>Usage:</ansi>"),
			"echo \"<ansi fg=red>Bingo</ansi>\" | "+os.Args[0])

		return
	}

	input := bufio.NewReader(os.Stdin)
	output := bufio.NewWriter(os.Stdout)

	if err := ansigo.ParseStreaming(input, output); err != nil {
		panic(err)
	}

}
