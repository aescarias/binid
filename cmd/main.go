package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/aescarias/binid/lib"
)

var VERSION = "0.1.0"

type Guess struct {
	Name    string
	Mime    string
	Details map[string]any
}

func GuessKind(handle *os.File) (*Guess, error) {
	for _, sig := range lib.MagicTags {
		matches, err := sig.Narrower(handle)
		if err != nil {
			return nil, err
		}

		if !matches {
			continue
		}

		guess := &Guess{Name: sig.Name, Mime: sig.Mime, Details: nil}
		if sig.Extractor == nil {
			return guess, nil
		}

		extracted, err := sig.Extractor(handle)
		if err != nil {
			return nil, err
		}

		guess.Details = extracted
		return guess, nil
	}

	return nil, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("BinID version %s\n", VERSION)
		fmt.Printf("usage: binid [filename]\n")
		os.Exit(1)
	}

	handle, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
	defer handle.Close()

	kind, err := GuessKind(handle)
	if err != nil {
		fmt.Printf("could not guess file type: %s\n", err)
		os.Exit(1)
	}

	if kind == nil {
		fmt.Printf("unknown file type\n")
		os.Exit(0)
	}

	fmt.Printf("name: %s\nmime: %s\n", kind.Name, kind.Mime)
	if kind.Details != nil {
		fmt.Printf("more: \n")
		for key, value := range kind.Details {
			switch refVal := reflect.ValueOf(value); refVal.Kind() {
			case reflect.Slice, reflect.Array:
				fmt.Printf("  %s (%v):\n", key, refVal.Len())
				for idx := range refVal.Len() {
					fmt.Printf("    %s\n", refVal.Index(idx).String())
				}
			case reflect.Map:
				fmt.Printf("  %s (%v):\n", key, refVal.Len())
				for _, inKey := range refVal.MapKeys() {
					fmt.Printf("    %s: %s\n", inKey.String(), refVal.MapIndex(inKey).String())
				}
			default:
				fmt.Printf("  %s: %v\n", key, value)
			}
		}
	}
}
