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
	case bindef.NodeBinOp:
		binOp := node.(*bindef.BinOpNode)
		fmt.Printf("%s- %s (%s)\n", tabbed, binOp.Type(), binOp.Op.Value)

		ShowTree(binOp.Left, indent+1)
		ShowTree(binOp.Right, indent+1)
	case bindef.NodeUnaryOp:
		unaryOp := node.(*bindef.UnaryOpNode)
		fmt.Printf("%s- %s (%s)\n", tabbed, unaryOp.Type(), unaryOp.Op.Value)

		ShowTree(unaryOp.Node, indent+1)
	case bindef.NodeLiteral:
		litNode := node.(*bindef.LiteralNode)

		fmt.Printf("%s- %s (%s)\n", tabbed, litNode.Type(), litNode.Token.Value)
	case bindef.NodeMap:
		mapNode := node.(*bindef.MapNode)

		fmt.Printf("%s- %s\n", tabbed, mapNode.Type())

		for key, value := range mapNode.Items {
			ShowTree(key, indent+1)
			ShowTree(value, indent+2)
		}
	case bindef.NodeList:
		listNode := node.(*bindef.ListNode)

		fmt.Printf("%s- %s\n", tabbed, listNode.Type())

		for _, key := range listNode.Items {
			ShowTree(key, indent+1)
		}
	case bindef.NodeAttr:
		attrNode := node.(*bindef.AttrNode)
		fmt.Printf("%s- %s\n", tabbed, attrNode.Type())

		ShowTree(attrNode.Expr, indent+1)
		ShowTree(attrNode.Attr, indent+1)
	case bindef.NodeSubscript:
		subNode := node.(*bindef.SubscriptNode)
		fmt.Printf("%s- %s\n", tabbed, subNode.Type())

		ShowTree(subNode.Expr, indent+1)
		ShowTree(subNode.Item, indent+1)
	case bindef.NodeCall:
		callNode := node.(*bindef.CallNode)
		fmt.Printf("%s- %s\n", tabbed, callNode.Type())

		ShowTree(callNode.Expr, indent+1)
		for _, arg := range callNode.Arguments {
			ShowTree(arg, indent+1)
		}
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
				fmt.Printf("in %s:%d:%d-%d\n", filepath, line+1, column+1, column+1+length)
				fmt.Printf("error: %s\n", lerr.Message)

				trimmed := strings.TrimLeftFunc(string(lineStr), unicode.IsSpace)
				diff := len(string(lineStr)) - len(trimmed)

				fmt.Println("   ", trimmed)
				fmt.Println("   ", strings.Repeat(" ", column-diff-1)+strings.Repeat("^", length))
				break
			}
		}
	} else {
		fmt.Println(err)
	}
}

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
		ReportError(os.Args[1], bdfData, err)
		os.Exit(1)
	}

	fmt.Println("== tokens")
	for _, tok := range lex.Tokens {
		fmt.Println(tok)
	}

	ps := bindef.NewParser(lex.Tokens)

	tree, err := ps.Parse()
	if err != nil {
		ReportError(os.Args[1], bdfData, err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("== syntax tree")
	ShowTree(tree, 0)

	result, err := bindef.Evaluate(tree, nil)
	if err != nil {
		ReportError(os.Args[1], bdfData, err)
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

	fmt.Printf("bdf: %s\n", meta.BdfVersion)
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

	for key, val := range contents {
		fmt.Printf("%s: %v\n", key, val)
	}
}
