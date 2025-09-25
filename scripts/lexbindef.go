package main

import (
	"fmt"
	"os"

	"github.com/aescarias/binid/bindef"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("usage: lexbindef [filename]")
		os.Exit(1)
	}

	data, err := os.ReadFile(os.Args[1])

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	sc := bindef.Scanner{Data: string(data), Current: 0}
	ps := bindef.Lexer{Contents: sc, Tokens: []bindef.Token{}}

	ps.Process()

	for _, tok := range ps.Tokens {
		fmt.Println(tok)
	}
}
