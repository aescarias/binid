package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aescarias/bindef/bindef"
)

type CmdArgs struct {
	DefinitionPath string
	TargetPath     string
	ShowAll        bool
	ShowHelp       bool
}

var usage string = "usage: lexbindef [options] [definition path] [target path]"

func ShowHelp() {
	fmt.Println(usage)
	fmt.Println()
	fmt.Println("arguments:")
	fmt.Println("  definition path      path of the definition file to check")
	fmt.Println("  target path          path of the target file to match")
	fmt.Println()
	fmt.Println("options:")
	fmt.Println("  -h, --help           show this help message")
	fmt.Println("  -a, --all            show all bytes of a byte sequence")
	fmt.Println("                       (this may produce large outputs)")
}

func ParseCmdArgs(args []string) CmdArgs {
	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}

	cmd := CmdArgs{}

	argPosition := 0
	done := false
	for argPosition < len(args) && !done {
		switch arg := args[argPosition]; arg {
		case "-h", "--help":
			cmd.ShowHelp = true
		case "-a", "--all":
			cmd.ShowAll = true
		default:
			if cmd.DefinitionPath == "" {
				cmd.DefinitionPath = arg
			} else if cmd.TargetPath == "" {
				cmd.TargetPath = arg
			} else {
				done = true
			}
		}
		argPosition += 1
	}

	if !cmd.ShowHelp {
		if cmd.DefinitionPath == "" {
			fmt.Println("error: missing required argument 'definition path'")
			fmt.Println("see lexbindef -h for help")
			os.Exit(1)
		} else if cmd.TargetPath == "" {
			fmt.Println("error: missing required argument 'target path'")
			fmt.Println("see lexbindef -h for help")
			os.Exit(1)
		}
	}

	return cmd
}

func main() {
	args := ParseCmdArgs(os.Args[1:])
	if args.ShowHelp {
		ShowHelp()
		os.Exit(0)
	}

	bdfData, err := os.ReadFile(args.DefinitionPath)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	lex := bindef.NewLexer(bdfData)

	if err := lex.Process(); err != nil {
		bindef.ReportError(args.DefinitionPath, bdfData, err)
		os.Exit(1)
	}

	fmt.Println("== tokens")
	for _, tok := range lex.Tokens {
		fmt.Println(tok)
	}

	ps := bindef.NewParser(lex.Tokens)

	tree, err := ps.Parse()
	if err != nil {
		bindef.ReportError(args.DefinitionPath, bdfData, err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("== syntax tree")
	bindef.ShowSyntaxTree(tree, 0)

	result, err := bindef.Evaluate(tree, nil)
	if err != nil {
		bindef.ReportError(args.DefinitionPath, bdfData, err)
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
	fmt.Printf("docs: %s\n", meta.Doc)

	fmt.Println()
	fmt.Println("== details")
	contents, err := bindef.ApplyBDF(result, args.TargetPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, pair := range contents {
		bindef.ShowMetadataField(pair, 0, args.ShowAll)
	}
}
