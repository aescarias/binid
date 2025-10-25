package bindef

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

// ShowSyntaxTree prints the abstract syntax tree (AST) defined by node starting at the
// specified indent level.
func ShowSyntaxTree(node Node, indent int) {
	tabbed := strings.Repeat(" ", indent)

	switch node.Type() {
	case NodeBinOp:
		binOp := node.(*BinOpNode)
		fmt.Printf("%s- %s (%s)\n", tabbed, binOp.Type(), binOp.Op.Value)

		ShowSyntaxTree(binOp.Left, indent+1)
		ShowSyntaxTree(binOp.Right, indent+1)
	case NodeUnaryOp:
		unaryOp := node.(*UnaryOpNode)
		fmt.Printf("%s- %s (%s)\n", tabbed, unaryOp.Type(), unaryOp.Op.Value)

		ShowSyntaxTree(unaryOp.Node, indent+1)
	case NodeLiteral:
		litNode := node.(*LiteralNode)

		fmt.Printf("%s- %s (%s)\n", tabbed, litNode.Type(), litNode.Token.Value)
	case NodeMap:
		mapNode := node.(*MapNode)

		fmt.Printf("%s- %s\n", tabbed, mapNode.Type())

		for key, value := range mapNode.Items {
			ShowSyntaxTree(key, indent+1)
			ShowSyntaxTree(value, indent+2)
		}
	case NodeList:
		listNode := node.(*ListNode)

		fmt.Printf("%s- %s\n", tabbed, listNode.Type())

		for _, key := range listNode.Items {
			ShowSyntaxTree(key, indent+1)
		}
	case NodeAttr:
		attrNode := node.(*AttrNode)
		fmt.Printf("%s- %s\n", tabbed, attrNode.Type())

		ShowSyntaxTree(attrNode.Expr, indent+1)
		ShowSyntaxTree(attrNode.Attr, indent+1)
	case NodeSubscript:
		subNode := node.(*SubscriptNode)
		fmt.Printf("%s- %s\n", tabbed, subNode.Type())

		ShowSyntaxTree(subNode.Expr, indent+1)
		ShowSyntaxTree(subNode.Item, indent+1)
	case NodeCall:
		callNode := node.(*CallNode)
		fmt.Printf("%s- %s\n", tabbed, callNode.Type())

		ShowSyntaxTree(callNode.Expr, indent+1)
		for _, arg := range callNode.Arguments {
			ShowSyntaxTree(arg, indent+1)
		}
	default:
		fmt.Printf("%s- %#v\n", tabbed, node)
	}
}

// ReportError prints an error report for a file at filepath with byte contents source
// containing details about err.
func ReportError(filepath string, source []byte, err error) {
	if lerr, ok := err.(LangError); ok {
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
				fmt.Println(lerr)

				trimmed := strings.TrimLeftFunc(string(lineStr), unicode.IsSpace)
				diff := len(string(lineStr)) - len(trimmed)

				arrowAlign := max(column-diff-1, 0)
				fmt.Println("   ", trimmed)
				fmt.Println("   ", strings.Repeat(" ", arrowAlign)+strings.Repeat("^", length))
				break
			}
		}
	} else {
		fmt.Println(err)
	}
}

// ShowMetadataField prints a pair containing a format type and a value with the
// specified indent level. Spaces are used for indentation. fullBytes determines
// whether the complete byte sequence is shown. If false, then only the first 256
// bytes are displayed.
func ShowMetadataField(pair MetaPair, indent int, fullBytes bool) {
	indentStr := strings.Repeat("  ", indent)

	var key string
	if pair.Field.Name != "" {
		key = pair.Field.Name
	} else {
		key = pair.Field.Id
		if strings.HasPrefix(key, "_") || key == "" {
			return
		}
	}

	switch f := pair.Field; f.Type {
	case TypeMagic:
		return
	case TypeByte:
		str := string(pair.Value.(StringResult))

		const cutoff int = 256
		if !fullBytes && len(str) > cutoff {
			str = fmt.Sprintf("%q", str[:cutoff]) + fmt.Sprintf(" (%d bytes remain)", len(str)-cutoff)
		} else {
			str = fmt.Sprintf("%q", str)
		}

		fmt.Printf("%s%s: %s\n", indentStr, key, str)
	case TypeStruct:
		mapping := pair.Value.(MapResult)

		fmt.Printf("%s%s:\n", indentStr, key)
		for _, field := range f.ProcFields {
			id := IdentResult(field.Id)

			ShowMetadataField(
				MetaPair{Field: field, Value: mapping[id]},
				indent+1, fullBytes,
			)
		}
	case TypeArray:
		list := pair.Value.(ListResult)
		fmt.Printf("%s%s (%d):\n", indentStr, key, len(list))

		for _, field := range list {
			ShowMetadataField(
				MetaPair{Field: *f.ProcArrItem, Value: field},
				indent+1, fullBytes,
			)
		}
	case TypeEnum:
		var friendlyName string
		for _, member := range f.EnumMembers {
			if doBinOpEquals(member.Value, pair.Value) {
				if member.Doc != "" {
					friendlyName = member.Doc
				} else {
					friendlyName = member.Id
				}
				break
			}
		}
		fmt.Printf("%s%s: %s (%#x)\n", indentStr, key, friendlyName, pair.Value)
	default:
		fmt.Printf("%s%s: %v\n", indentStr, key, pair.Value)
	}
}
