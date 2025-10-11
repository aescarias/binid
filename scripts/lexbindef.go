package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aescarias/binid/bindef"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: lexbindef [filepath to bdf] [filepath to target]")
		os.Exit(1)
	}

	bdfData, err := os.ReadFile(os.Args[1])

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	lex := bindef.NewLexer(bdfData)

	if err := lex.Process(); err != nil {
		bindef.ReportError(os.Args[1], bdfData, err)
		os.Exit(1)
	}

	fmt.Println("== tokens")
	for _, tok := range lex.Tokens {
		fmt.Println(tok)
	}

	ps := bindef.NewParser(lex.Tokens)

	tree, err := ps.Parse()
	if err != nil {
		bindef.ReportError(os.Args[1], bdfData, err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("== syntax tree")
	bindef.ShowSyntaxTree(tree, 0)

	result, err := bindef.Evaluate(tree, nil)
	if err != nil {
		bindef.ReportError(os.Args[1], bdfData, err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("== result")
	fmt.Println(result)

	fmt.Println()
	fmt.Println("== metadata")

	meta, err := bindef.GetMetadata(result)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("bdf: %s\n", meta.Version)
	fmt.Printf("name: %s\n", meta.Name)
	fmt.Printf("mime type(s): %s\n", strings.Join(meta.Mime, ", "))
	fmt.Printf("extension(s): %s\n", strings.Join(meta.Exts, ", "))

	fmt.Println()
	fmt.Println("== details")
	contents, err := bindef.ApplyBDF(result, os.Args[2])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, pair := range contents {
		fmt.Printf("%s: %v\n", pair.Key, pair.Value)
	}
}
