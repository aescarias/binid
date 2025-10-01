package bindef

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

// SpecVersion is the version of the BinDef spec implemented by this runtime.
var SpecVersion = Version{Major: 0, Minor: 1}

// A Version describes a specification version with a major and minor version component.
type Version struct {
	Major int
	Minor int
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

// NewVersionFromString produces a Version struct from a string of the form
// "major.minor". If this parsing fails, an error is returned.
func NewVersionFromString(version string) (Version, error) {
	split := strings.SplitN(version, ".", 2)

	major, err := strconv.Atoi(split[0])
	if err != nil {
		return Version{}, err
	}

	minor, err := strconv.Atoi(split[1])
	if err != nil {
		return Version{}, err
	}

	return Version{Major: major, Minor: minor}, nil
}

func ident(value string) Result {
	return Result{Kind: ResultIdent, Value: value}
}

type Meta struct {
	BdfVersion Version
	Name       string
	Mime       []string
	Exts       []string
}

func parseMeta(meta Result) (Meta, error) {
	if meta.Kind != ResultMap {
		return Meta{}, fmt.Errorf("meta must be a mapping")
	}

	metadata := meta.Value.(map[Result]Result)
	bdf, ok := metadata[ident("bdf")]
	if !ok {
		return Meta{}, fmt.Errorf("'meta' map missing required key 'bdf'")
	}

	if bdf.Kind != ResultString {
		return Meta{}, fmt.Errorf("meta.bdf must be a constant string")
	}

	version, err := NewVersionFromString(bdf.Value.(string))
	if err != nil {
		return Meta{}, err
	}

	// name
	name, ok := metadata[ident("name")]
	if !ok {
		return Meta{}, fmt.Errorf("'meta' map missing required key 'name'")
	}

	if name.Kind != ResultString {
		return Meta{}, fmt.Errorf("meta.name must be a constant string")
	}

	// mime
	mime, ok := metadata[ident("mime")]
	if !ok {
		return Meta{}, fmt.Errorf("'meta' map missing required key 'mime'")
	}

	if mime.Kind != ResultList {
		return Meta{}, fmt.Errorf("meta.mime must be a constant list")
	}

	mimeRes := mime.Value.([]Result)
	mimeList := make([]string, len(mimeRes))

	for idx, item := range mimeRes {
		if item.Kind != ResultString {
			return Meta{}, fmt.Errorf("meta.mime: element at index %d is not a string", idx)
		}

		mimeList[idx] = item.Value.(string)
	}

	// exts
	exts, ok := metadata[ident("exts")]
	if !ok {
		return Meta{}, fmt.Errorf("'meta' map missing required key 'exts'")
	}

	if exts.Kind != ResultList {
		return Meta{}, fmt.Errorf("meta.exts must be a constant list")
	}

	extsRes := exts.Value.([]Result)
	extsList := make([]string, len(extsRes))

	for idx, item := range extsRes {
		if item.Kind != ResultString {
			return Meta{}, fmt.Errorf("meta.exts: element at index %d is not a string", idx)
		}

		extsList[idx] = item.Value.(string)
	}

	return Meta{
		BdfVersion: version,
		Name:       name.Value.(string),
		Mime:       mimeList,
		Exts:       extsList,
	}, nil
}

func resultTo[T any](res Result, kind ResultKind) (T, error) {
	if res.Kind != kind {
		var zero T
		return zero, fmt.Errorf("expected type %s, received %s", kind, res)
	}

	val, ok := res.Value.(T)
	if !ok {
		panic(fmt.Sprintf("type assertion failed for %s", kind))
	}

	return val, nil
}

func resultToMap[T map[Result]Result](res Result) (T, error) {
	return resultTo[T](res, ResultMap)
}

func resultToSlice[T []Result](res Result) (T, error) {
	return resultTo[T](res, ResultList)
}

func resultToString[T string](res Result) (T, error) {
	return resultTo[T](res, ResultString)
}

func resultToIdent[T string](res Result) (T, error) {
	return resultTo[T](res, ResultIdent)
}

func checkBinMagic(handle *os.File, item map[Result]Result) error {
	offset, err := handle.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	magicAt, err := resultToString(item[ident("match")])
	if err != nil {
		return err
	}

	matchBytes := make([]byte, len(magicAt))
	if _, err := handle.Read(matchBytes); err != nil {
		return err
	}

	if !slices.Equal(matchBytes, []byte(magicAt)) {
		return fmt.Errorf("did not find magic at position %d", offset)
	}

	return nil
}

type AnyUint interface {
	uint | uint8 | uint16 | uint32 | uint64
}

func getBinUint[T AnyUint](handle *os.File, item map[Result]Result) (T, error) {
	var zero T

	size := reflect.TypeOf(zero).Size()

	numBytes := make([]byte, size)
	if _, err := handle.Read(numBytes); err != nil {
		return zero, err
	}

	kind := reflect.TypeOf(zero).Kind()

	if kind == reflect.Uint8 {
		return T(numBytes[0]), nil
	}

	endian, err := resultToString(item[ident("endian")])
	if err != nil {
		return zero, err
	}

	switch endian {
	case "little":
		switch kind {
		case reflect.Uint16:
			return T(binary.LittleEndian.Uint16(numBytes)), nil
		case reflect.Uint, reflect.Uint32:
			return T(binary.LittleEndian.Uint32(numBytes)), nil
		case reflect.Uint64:
			return T(binary.LittleEndian.Uint64(numBytes)), nil
		}
	case "big":
		switch kind {
		case reflect.Uint16:
			return T(binary.BigEndian.Uint16(numBytes)), nil
		case reflect.Uint, reflect.Uint32:
			return T(binary.BigEndian.Uint32(numBytes)), nil
		case reflect.Uint64:
			return T(binary.BigEndian.Uint64(numBytes)), nil
		}
	}

	return zero, fmt.Errorf("not a uint")
}

func ApplyBDF(document Result, targetFile string) (map[string]any, error) {
	contents := map[string]any{}

	root, err := resultToMap(document)
	if err != nil {
		return nil, fmt.Errorf("root: %w", err)
	}

	handle, err := os.Open(targetFile)
	if err != nil {
		return nil, err
	}
	defer handle.Close()

	binarySeq, err := resultToSlice(root[ident("binary")])
	if err != nil {
		return nil, fmt.Errorf("binary: %w", err)
	}

	for idx, res := range binarySeq {
		binItem, err := resultToMap(res)
		if err != nil {
			return nil, fmt.Errorf("binary[%d]: %w", idx, err)
		}

		typeStr, err := resultToIdent(binItem[ident("type")])
		if err != nil {
			return nil, fmt.Errorf("binary[%d].type: %w", idx, err)
		}

		switch typeStr {
		case "magic":
			if err := checkBinMagic(handle, binItem); err != nil {
				return nil, fmt.Errorf("binary[%d]: %w", idx, err)
			}
		case "uint8", "uint16", "uint32", "uint64":
			var num any
			var err error

			switch typeStr {
			case "uint8":
				num, err = getBinUint[uint8](handle, binItem)
			case "uint16":
				num, err = getBinUint[uint16](handle, binItem)
			case "uint32":
				num, err = getBinUint[uint32](handle, binItem)
			case "uint64":
				num, err = getBinUint[uint64](handle, binItem)
			}

			if err != nil {
				return nil, fmt.Errorf("binary[%d].type: %w", idx, err)
			}

			ident, err := resultToIdent(binItem[ident("id")])
			if err != nil {
				return nil, fmt.Errorf("binary[%d].name: %w", idx, err)
			}

			contents[ident] = num
		}
	}

	return contents, nil
}

// GetMetadata returns the metadata described in the 'meta' key of document.
func GetMetadata(document Result) (Meta, error) {
	if document.Kind != ResultMap {
		return Meta{}, fmt.Errorf("document must be a mapping")
	}

	mapping := document.Value.(map[Result]Result)
	meta, err := parseMeta(mapping[ident("meta")])

	return meta, err
}
