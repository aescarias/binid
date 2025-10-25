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

var VERSION = "0.5.0"

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

func GetDefaultDefsPaths() (exec string, cwd string, err error) {
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

type ErrLookupFailed struct {
	Issues map[string]error
}

func (e ErrLookupFailed) Error() string {
	return fmt.Sprintf("failed to load definitions from %d location(s)", len(e.Issues))
}

func LoadDefs(paths []string) (map[string]bindef.Result, error) {
	lookupErrors := map[string]error{}
	for _, path := range paths {
		defs, err := GetDefs(path)
		if err != nil {
			lookupErrors[path] = err
			continue
		}
		return defs, nil
	}

	return nil, ErrLookupFailed{Issues: lookupErrors}
}

func main() {
	args := ParseCmdArgs(os.Args[1:])

	if args.ShowHelp {
		ShowHelp()
		os.Exit(0)
	}

	if args.ShowVersion {
		fmt.Println("BinID version", VERSION)
		os.Exit(0)
	}

	handle, err := os.Open(args.Filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer handle.Close()

	var lookupPaths []string
	if args.DefsPath != "" {
		lookupPaths = []string{args.DefsPath}
	} else {
		exePath, cwdPath, err := GetDefaultDefsPaths()
		if err != nil {
			fmt.Printf("failed definition lookup: %s\n", err)
			os.Exit(1)
		}
		lookupPaths = []string{exePath, cwdPath}
	}

	defs, err := LoadDefs(lookupPaths)
	if err != nil {
		if lerr, ok := err.(ErrLookupFailed); ok {
			fmt.Println(lerr)
			for path, err := range lerr.Issues {
				if errors.Is(err, fs.ErrNotExist) {
					fmt.Printf("%s:\n  the path does not exist\n", path)
				} else {
					fmt.Printf("%s:\n  %s\n", path, err)
				}
			}
		}
		os.Exit(1)
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
	failedMatches := map[string]error{}

	for defPath, defResult := range defs {
		match, err := bindef.ApplyBDF(defResult, args.Filename)
		if err != nil {
			if _, ok := err.(bindef.ErrMagic); !ok {
				failedMatches[defPath] = err
			}
			continue
		}

		meta, err := bindef.GetMetadata(defResult)
		if err != nil {
			failedMatches[defPath] = fmt.Errorf("metadata get failed: %w", err)
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

	if len(failedMatches) > 0 {
		fmt.Println("\n== errors")
		for defPath, err := range failedMatches {
			fmt.Printf("%s:\n  %s\n", defPath, err)
		}
	}

	if !found {
		fmt.Println("no definitions matched")
	}
}
