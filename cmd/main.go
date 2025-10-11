package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"github.com/aescarias/binid/bindef"
)

var VERSION = "0.1.0"

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

func GetDefs() (map[string]bindef.Result, error) {
	defs := map[string]bindef.Result{}

	err := filepath.Walk("formats", func(path string, info fs.FileInfo, err error) error {
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

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("BinID version %s\n", VERSION)
		fmt.Printf("usage: binid [filename]\n")
		os.Exit(1)
	}

	handle, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer handle.Close()

	defs, err := GetDefs()
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Println("'formats' definition folder missing")
			os.Exit(1)
		}

		fmt.Printf("failed to load definitions:\n%s\n", err)
		os.Exit(1)
	}

	fmt.Printf("found %d definition(s)\n", len(defs))
	if len(defs) <= 0 {
		os.Exit(1)
	}

	found := false
	for defPath, defResult := range defs {
		match, err := bindef.ApplyBDF(defResult, os.Args[1])
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
			switch refVal := reflect.ValueOf(pair.Value); refVal.Kind() {
			case reflect.Slice, reflect.Array:
				fmt.Printf("%s (%v):\n", pair.Key, refVal.Len())
				for idx := range refVal.Len() {
					fmt.Printf("  %s\n", refVal.Index(idx).String())
				}
			case reflect.Map:
				fmt.Printf("%s (%v):\n", pair.Key, refVal.Len())
				for _, inKey := range refVal.MapKeys() {
					fmt.Printf("  %s: %s\n", inKey.String(), refVal.MapIndex(inKey).String())
				}
			case reflect.String:
				fmt.Printf("%s: %q\n", pair.Key, pair.Value)
			default:
				fmt.Printf("%s: %v\n", pair.Key, pair.Value)
			}
		}
	}

	if !found {
		fmt.Println("no definitions matched")
	}
}
