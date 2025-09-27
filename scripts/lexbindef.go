package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"unicode"

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
	case "Map":
		mapNode := node.(*bindef.MapNode)

		fmt.Printf("%s- %s\n", tabbed, mapNode.Type())

		for key, value := range mapNode.Items {
			ShowTree(key, indent+1)
			ShowTree(value, indent+2)
		}
	case "List":
		listNode := node.(*bindef.ListNode)

		fmt.Printf("%s- %s\n", tabbed, listNode.Type())

		for _, key := range listNode.Items {
			ShowTree(key, indent+1)
		}
	case "AttrAccess":
		attrNode := node.(*bindef.AttrAccessNode)

		tokStrs := []string{attrNode.Parent.Value}

		for _, tok := range attrNode.Members {
			tokStrs = append(tokStrs, tok.Value)
		}

		fmt.Printf("%s- %s (%s)\n", tabbed, attrNode.Type(), strings.Join(tokStrs, "."))
	default:
		fmt.Printf("%s- %#v\n", tabbed, node)
	}
}

func ReportError(filepath string, source []byte, err error) {
	if lerr, ok := err.(bindef.LangError); ok {
		line, column, offset := 0, 0, 0
		var ch byte

		for offset, ch = range source {
			column += 1

			if ch == '\n' {
				line += 1
				column = 0
			}

			if offset >= lerr.Position.Start {
				break
			}
		}

		for idx, lineStr := range bytes.Split(bytes.TrimSuffix(source, []byte("\n")), []byte("\n")) {
			if idx == line {
				length := lerr.Position.End - lerr.Position.Start
				fmt.Printf("in %s:%d:%d-%d\n", filepath, line+1, column, column+length)
				fmt.Printf("error: %s\n", lerr.Message)

				trimmed := strings.TrimLeftFunc(string(lineStr), unicode.IsSpace)
				diff := len(string(lineStr)) - len(trimmed)

				fmt.Println("   ", trimmed)
				fmt.Println("   ", strings.Repeat(" ", column-diff-1)+strings.Repeat("^", length))
			}
		}

	} else {
		fmt.Println(err)
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
		ReportError(os.Args[1], data, err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("== syntax tree")
	ShowTree(tree, 0)

	result, err := bindef.Evaluate(tree)
	if err != nil {
		ReportError(os.Args[1], data, err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("== result")
	fmt.Println(result)
}
