package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aescarias/binid/bindef"
)

func ShowTree(node bindef.Node, indent int) {
	tabbed := strings.Repeat(" ", indent)

	switch node.Type() {
	case "BinOp":
		binOp := node.(*bindef.BinOpNode)
		fmt.Printf("%s- %s (%s)\n", tabbed, binOp.Type(), binOp.Op.Value)

		ShowTree(binOp.Left, indent+1)
		ShowTree(binOp.Right, indent+1)
	case "UnaryOp":
		unaryOp := node.(*bindef.UnaryOpNode)
		fmt.Printf("%s- %s (%s)\n", tabbed, unaryOp.Type(), unaryOp.Op.Value)

		ShowTree(unaryOp.Node, indent+1)
	case "Literal":
		litNode := node.(*bindef.LiteralNode)

		fmt.Printf("%s- %s (%s)\n", tabbed, litNode.Type(), litNode.Token.Value)
	default:
		fmt.Printf("%s- %#v\n", tabbed, node)
	}
}

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

	sc := bindef.Scanner[byte]{Data: data, Current: 0}
	lx := bindef.Lexer{Contents: sc, Tokens: []bindef.Token{}}
	lx.Process()

	fmt.Println("== tokens")
	for _, tok := range lx.Tokens {
		fmt.Println(tok)
	}

	ps := bindef.Parser{
		Scanner: bindef.Scanner[bindef.Token]{Data: lx.Tokens, Current: 0},
	}

	tree, err := ps.ParseExpr()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("== syntax tree")
	ShowTree(tree, 0)

	result, err := bindef.Evaluate(tree)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("== result")
	fmt.Println(result)
}
