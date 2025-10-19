package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/aescarias/binid/bindef"
)

var VERSION = "0.4.0"

type CmdArgs struct {
	Filename string
	ShowAll  bool
	ShowHelp bool
}

func ShowHelp() {
	fmt.Println("BinID version", VERSION)

	fmt.Println("The Binary Identifier for determining file types")
	fmt.Println()
	fmt.Println("usage: binid [-h] [-a] [filename]")
	fmt.Println()
	fmt.Println("arguments:")
	fmt.Println("  filename      path of the file to identify")
	fmt.Println()
	fmt.Println("options:")
	fmt.Println("  -h, --help    show this help message")
	fmt.Println("  -a, --all     show all bytes of a byte sequence")
	fmt.Println("                (this may produce large outputs)")
}

func ParseCmdArgs(args []string) CmdArgs {
	if len(os.Args) < 2 {
		fmt.Println("BinID version", VERSION)
		fmt.Println("usage: binid [-h] [-a] [filename]. see binid -h for help.")
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
			if cmd.Filename != "" {
				done = true
			} else {
				cmd.Filename = arg
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

func ParseDef(filepath string) bindef.Result {
	bdfData, err := os.ReadFile(filepath)
	if slices.Equal(bdfData, []byte{}) {
		fmt.Printf("found empty bdf at %s\n", filepath)
		os.Exit(1)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	lex := bindef.NewLexer(bdfData)

	if err := lex.Process(); err != nil {
		bindef.ReportError(filepath, bdfData, err)
		os.Exit(1)
	}

	ps := bindef.NewParser(lex.Tokens)

	tree, err := ps.Parse()
	if err != nil {
		bindef.ReportError(filepath, bdfData, err)
		os.Exit(1)
	}

	result, err := bindef.Evaluate(tree, nil)
	if err != nil {
		bindef.ReportError(filepath, bdfData, err)
		os.Exit(1)
	}

	return result
}

func GetDefs(path string) (map[string]bindef.Result, error) {
	defs := map[string]bindef.Result{}

	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".bdf") {
			defs[info.Name()] = ParseDef(path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return defs, nil
}

func GetDefsPaths() (exec string, cwd string, err error) {
	exe, err := os.Executable()
	if err != nil {
		return "", "", err
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}

	return filepath.Join(filepath.Dir(exe), "formats"),
		filepath.Join(wd, "formats"),
		nil
}

func main() {
	args := ParseCmdArgs(os.Args[1:])

	if args.ShowHelp {
		ShowHelp()
		os.Exit(0)
	}

	handle, err := os.Open(args.Filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer handle.Close()

	exePath, cwdPath, err := GetDefsPaths()
	if err != nil {
		fmt.Printf("failed definition lookup: %s\n", err)
		os.Exit(1)
	}

	var defs map[string]bindef.Result
	if defs, err = GetDefs(exePath); err != nil {
		if defs, err = GetDefs(cwdPath); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				fmt.Printf("'formats' definition folder missing. looked in:\n  %s\n  %s\n", exePath, cwdPath)
			} else {
				fmt.Printf("failed to load definitions:\n%s\n", err)
			}

			os.Exit(1)
		}
	}

	fmt.Printf("found %d definition(s)\n", len(defs))
	if len(defs) <= 0 {
		os.Exit(1)
	}

	inputStat, err := os.Stat(args.Filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if inputStat.IsDir() {
		fmt.Printf("%s is a directory\n", args.Filename)
		os.Exit(0)
	}

	if inputStat.Size() <= 0 {
		fmt.Printf("%s is empty\n", args.Filename)
		os.Exit(0)
	}

	fmt.Printf("matching %s\n", args.Filename)

	found := false
	for defPath, defResult := range defs {
		match, err := bindef.ApplyBDF(defResult, args.Filename)
		if err != nil {
			if _, ok := err.(bindef.ErrMagic); !ok {
				fmt.Printf("%s:\n  %s\n", defPath, err)
			}
			continue
		}

		meta, err := bindef.GetMetadata(defResult)
		if err != nil {
			fmt.Printf("format %q matched but metadata get failed with %q\n", defPath, err)
			continue
		}

		found = true
		fmt.Println()
		fmt.Println("== match")
		fmt.Println("name:", meta.Name)

		if len(meta.Mime) > 0 {
			fmt.Println("mime(s):", strings.Join(meta.Mime, ", "))
		}

		if len(meta.Exts) > 0 {
			fmt.Println("extension(s):", strings.Join(meta.Exts, ", "))
		}

		if meta.Doc != "" {
			fmt.Println("details:", meta.Doc)
		}

		fmt.Println()
		fmt.Println("== metadata")

		if len(match) <= 0 {
			fmt.Println("no metadata extracted")
			continue
		}

		for _, pair := range match {
			bindef.ShowMetadataField(pair, 0, args.ShowAll)
		}
	}

	if !found {
		fmt.Println("no definitions matched")
	}
}
