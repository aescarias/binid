package main

import (
	"fmt"
	"os"
)

type CmdArgs struct {
	Filename    string
	DefsPath    string
	ShowAll     bool
	ShowHelp    bool
	ShowVersion bool
}

func ShowHelp() {
	fmt.Println("BinID version", VERSION)

	fmt.Println("The Binary Identifier for determining file types")
	fmt.Println()
	fmt.Println("usage: binid [options] [filename]")
	fmt.Println()
	fmt.Println("arguments:")
	fmt.Println("  filename          path of the file to identify")
	fmt.Println()
	fmt.Println("options:")
	fmt.Println("  -h, --help        show this help message")
	fmt.Println("  -a, --all         show all bytes of a byte sequence")
	fmt.Println("                    (this may produce large outputs)")
	fmt.Println("  -d, --defs        path to the definitions folder")
	fmt.Println("                    (default is 'formats' in current directory)")
	fmt.Println("  -v, --version     print binid's version")
}

func ParseCmdArgs(args []string) CmdArgs {
	if len(os.Args) < 2 {
		fmt.Println("BinID version", VERSION)
		fmt.Println("usage: binid [options] [filename]. see binid -h for help.")
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
		case "-v", "--version":
			cmd.ShowVersion = true
		case "-d", "--defs":
			if argPosition+1 >= len(args) {
				fmt.Println("error: missing value for option 'defs'")
				os.Exit(1)
			}

			argPosition++
			cmd.DefsPath = args[argPosition]
		default:
			if cmd.Filename == "" {
				cmd.Filename = arg
			} else {
				done = true
			}
		}
		argPosition += 1
	}

	if !cmd.ShowHelp && cmd.Filename == "" {
		fmt.Println("error: missing required argument 'filename'")
		fmt.Println("see binid -h for help")
		os.Exit(1)
	}

	return cmd
}
