package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Volte6/ansitags"
)

// accepts a string via shell pipe and processes with ansitags.Parse()
func main() {

	if os.Getenv(`COLOR_MODE`) == `256` {
		ansitags.SetColorMode(ansitags.Color256)
	}

	modePtr := flag.String("mode", "parse", "[parse|generate]")
	flag.Parse()

	// Useful for testing... example:
	// ./example/bin/ansitags -mode=generate | ./example/bin/ansitags
	if *modePtr == "generate" {

		str := "Some normal text <ansi fg=black bg=\"white\">A</ansi><ansi fg=\"red\" bg=\"cyan\">B</ansi><ansi fg=\"green\" bg=\"magenta\">C</ansi><ansi fg=\"yellow\" bg=\"blue\">D</ansi><ansi fg=\"blue\" bg=\"yellow\">E</ansi><ansi fg=\"magenta\" bg=\"green\">F</ansi><ansi fg=\"cyan\" bg=\"red\">G</ansi><ansi fg=\"white\" bg=\"black\">H</ansi> some more normal text... "
		strlen := len(str)

		for {
			for i := 0; i < strlen; i++ {
				fmt.Print(string(str[i]))
				time.Sleep(10 * time.Millisecond)
			}
		}

		return
	}

	info, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if info.Mode()&os.ModeCharDevice != 0 || info.Size() <= 0 {

		fmt.Printf("\n%s %s\n\n",
			ansitags.Parse("<ansi fg=red bold=true>Usage:</ansi>"),
			"echo \"<ansi fg=red>Bingo</ansi>\" | "+os.Args[0])

		return
	}

	input := bufio.NewReader(os.Stdin)
	output := bufio.NewWriterSize(os.Stdout, 1)

	ansitags.ParseStreaming(input, output)

}
