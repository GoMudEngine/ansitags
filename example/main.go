package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

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

	reader := bufio.NewReader(os.Stdin)
	var sBuilder strings.Builder

	for {
		input, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}
		sBuilder.WriteRune(input)
	}

	fmt.Print(ansigo.Parse(sBuilder.String()))

}
