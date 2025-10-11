package bindef

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/big"
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

// SpecVersion is the version of the BinDef spec implemented by this runtime.
var SpecVersion = Version{Major: 0, Minor: 1}

// A Version describes a specification version with a major and minor version component.
type Version struct {
	Major int
	Minor int
}

// String represents the version in the form 'X.Y'.
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

// Meta describes the metadata of a BinDef document.
type Meta struct {
	Version Version  // The minimum BDF version this document supports.
	Name    string   // The name of the format being described.
	Mime    []string // The media or MIME type(s) for this format.
	Exts    []string // The file extensions used by this format.
	Doc     string   // Additional documentation for the format.
}

type SeekPos struct {
	Offset int64
	Whence int
}

type FormatType struct {
	Type   TypeName   // Format type (e.g. uint8, int8). May be more complex such as byte[n].
	Id     string     // Field identifier.
	Name   string     // Human-readable field name.
	Doc    string     // Documentation.
	At     SeekPos    // Seek position.
	Endian string     // For integer types only, the byte endianness (either "big" or "little").
	Match  []MagicTag // For magic types only, the pattern(s) that must match.
	Size   int64      // For byte types only, the size of the byte string.
	Strip  bool       // For byte types only, whether to strip whitespace or null bytes from the ends of the string.
}

type MagicTag struct {
	Contents string
	Offset   int64
}

func NumberResultAsInt(res Result) (int64, error) {
	switch res.Kind() {
	case ResultInt:
		resInt, err := ResultIs[IntegerResult](res)
		if err != nil {
			return -1, err
		}
		return resInt.Int64(), nil
	case ResultFloat:
		resFloat, err := ResultIs[FloatResult](res)
		if err != nil {
			return -1, err
		}
		return int64(resFloat), nil
	default:
		return -1, fmt.Errorf("result %s is not a numeric type", res.Kind())
	}
}

func parseMeta(meta Result) (Meta, error) {
	metadata, err := ResultIs[MapResult](meta)
	if err != nil {
		return Meta{}, fmt.Errorf("meta: %w", err)
	}

	// bdf version
	bdf, err := GetKeyByIdent[StringResult](metadata, "bdf", true)
	if err != nil {
		return Meta{}, fmt.Errorf("meta: %w", err)
	}

	version, err := NewVersionFromString(string(bdf))
	if err != nil {
		return Meta{}, err
	}

	// name
	name, err := GetKeyByIdent[StringResult](metadata, "name", true)
	if err != nil {
		return Meta{}, fmt.Errorf("meta: %w", err)
	}

	// doc
	doc, err := GetKeyByIdent[StringResult](metadata, "doc", false)
	if err != nil {
		return Meta{}, fmt.Errorf("meta: %w", err)
	}

	// media type(s)
	mime, err := GetKeyByIdent[ListResult](metadata, "mime", true)
	if err != nil {
		return Meta{}, fmt.Errorf("meta: %w", err)
	}

	mimeStrings := make([]string, len(mime))

	for idx, item := range mime {
		itemStr, err := ResultIs[StringResult](item)
		if err != nil {
			return Meta{}, fmt.Errorf("meta: mime[%d]: %w", idx, err)
		}

		mimeStrings[idx] = string(itemStr)
	}

	// exts
	exts, err := GetKeyByIdent[ListResult](metadata, "exts", true)
	if err != nil {
		return Meta{}, fmt.Errorf("meta: %w", err)
	}

	extStrings := make([]string, len(exts))

	for idx, item := range exts {
		itemStr, err := ResultIs[StringResult](item)
		if err != nil {
			return Meta{}, fmt.Errorf("meta: exts[%d]: %w", idx, err)
		}

		extStrings[idx] = string(itemStr)
	}

	return Meta{
		Version: version,
		Name:    string(name),
		Mime:    mimeStrings,
		Exts:    extStrings,
		Doc:     string(doc),
	}, nil
}

func parseFormatType(format Result, ns Namespace) (FormatType, error) {
	bin, err := ResultIs[MapResult](format)
	if err != nil {
		return FormatType{}, err
	}

	genTypeRes, err := GetKeyByIdent[Result](bin, "type", true)
	if err != nil {
		return FormatType{}, err
	}

	var typeRes TypeResult
	switch genTypeRes.Kind() {
	case ResultLazy:
		tempRes, err := genTypeRes.(LazyResult)(ns)
		if err != nil {
			return FormatType{}, err
		}

		typeRes, err = ResultIs[TypeResult](tempRes)
		if err != nil {
			return FormatType{}, err
		}
	case ResultType:
		typeRes = genTypeRes.(TypeResult)
	}

	lazyIdRes, err := GetKeyByIdent[LazyResult](bin, "id", false)
	if err != nil {
		return FormatType{}, err
	}

	var idRes IdentResult
	if lazyIdRes != nil {
		genIdRes, err := lazyIdRes(nil)
		if err != nil {
			return FormatType{}, err
		}

		if idRes, err = ResultIs[IdentResult](genIdRes); err != nil {
			return FormatType{}, err
		}

	}

	nameRes, err := GetKeyByIdent[StringResult](bin, "name", false)
	if err != nil {
		return FormatType{}, err
	}

	docRes, err := GetKeyByIdent[StringResult](bin, "doc", false)
	if err != nil {
		return FormatType{}, err
	}

	atRes, err := GetKeyByIdent[Result](bin, "at", false)
	if err != nil {
		return FormatType{}, err
	}

	var atVal SeekPos

	if atRes != nil {
		switch atRes.Kind() {
		case ResultInt, ResultFloat:
			offsetInt, err := NumberResultAsInt(atRes)
			if err != nil {
				return FormatType{}, err
			}

			atVal = SeekPos{Offset: offsetInt, Whence: io.SeekStart}
		case ResultList:
			atValList, err := ResultIs[ListResult](atRes)
			if err != nil {
				return FormatType{}, err
			}

			if len(atValList) < 2 {
				return FormatType{}, fmt.Errorf("value 'at' must contain 2 items")
			}

			offsetInt, err := NumberResultAsInt(atValList[0])
			if err != nil {
				return FormatType{}, fmt.Errorf("at[0]: %w", err)
			}

			whenceStr, err := ResultIs[StringResult](atValList[1])
			if err != nil {
				return FormatType{}, fmt.Errorf("at[1]: %w", err)
			}

			switch whenceStr {
			case StringResult("start"):
				atVal = SeekPos{Offset: offsetInt, Whence: io.SeekStart}
			case StringResult("end"):
				atVal = SeekPos{Offset: offsetInt, Whence: io.SeekEnd}
			case StringResult("current"):
				atVal = SeekPos{Offset: offsetInt, Whence: io.SeekCurrent}
			default:
				return FormatType{}, fmt.Errorf("at[1]: whence is not a valid seek identifier")
			}
		default:
			return FormatType{}, fmt.Errorf("value 'at' is not a list or number")
		}
	}

	baseFormat := FormatType{
		Type: typeRes.Name,
		Id:   string(idRes),
		Name: string(nameRes),
		Doc:  string(docRes),
		At:   atVal,
	}

	switch typeRes.Name {
	case TypeMagic:
		matchRes, err := GetKeyByIdent[Result](bin, "match", true)
		if err != nil {
			return FormatType{}, err
		}

		switch matchRes.Kind() {
		case ResultString:
			baseFormat.Match = []MagicTag{
				{
					Contents: string(matchRes.(StringResult)),
					Offset:   0,
				},
			}
			return baseFormat, nil
		case ResultList:
			for idx, res := range matchRes.(ListResult) {
				matchStr, err := ResultIs[StringResult](res)
				if err != nil {
					return FormatType{}, fmt.Errorf("%d: %w", idx, err)
				}

				baseFormat.Match = append(baseFormat.Match, MagicTag{
					Contents: string(matchStr),
					Offset:   0,
				})
			}
			return baseFormat, nil
		}
	case TypeUint16, TypeUint32, TypeUint64, TypeInt16, TypeInt32, TypeInt64:
		endianRes, err := GetKeyByIdent[StringResult](bin, "endian", true)
		if err != nil {
			return FormatType{}, err
		}

		baseFormat.Endian = strings.ToLower(string(endianRes))
		if baseFormat.Endian != "little" && baseFormat.Endian != "big" {
			return FormatType{}, fmt.Errorf("endian is not 'little' or 'big'")
		}

		return baseFormat, nil
	case TypeByte:
		strip, err := GetKeyByIdent[BooleanResult](bin, "strip", false)
		if err != nil {
			return FormatType{}, err
		}

		if strip {
			baseFormat.Strip = bool(strip)
		}

		if len(typeRes.Params) <= 0 {
			baseFormat.Size = 1
			return baseFormat, nil
		}

		var sizeRes Result

		if maybeLazySize := typeRes.Params[0]; maybeLazySize.Kind() == ResultLazy {
			var err error
			sizeRes, err = maybeLazySize.(LazyResult)(ns)
			if err != nil {
				return FormatType{}, err
			}
		} else {
			sizeRes = typeRes.Params[0]
		}

		switch sizeRes.Kind() {
		case ResultInt:
			baseFormat.Size = sizeRes.(IntegerResult).Int64()
		case ResultFloat:
			baseFormat.Size = int64(math.Trunc(float64(sizeRes.(FloatResult))))
		default:
			return FormatType{}, fmt.Errorf("byte size must be numeric")
		}

		return baseFormat, nil
	}

	return baseFormat, nil
}

type ErrMagic struct {
	Offset int64
}

func (e ErrMagic) Error() string {
	return fmt.Sprintf("did not find magic at offset %d", e.Offset)
}

func checkMagic(handle *os.File, format FormatType) error {
	baseOffset, err := handle.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	for _, tag := range format.Match {
		if _, err := handle.Seek(baseOffset+tag.Offset, io.SeekStart); err != nil {
			return err
		}

		contents := []byte(tag.Contents)

		matchBytes := make([]byte, len(contents))
		if _, err := handle.Read(matchBytes); err != nil {
			return err
		}

		if slices.Equal(matchBytes, contents) {
			return nil
		}
	}

	return ErrMagic{Offset: baseOffset}
}

func readInt(handle *os.File, format FormatType) (Result, error) {
	var numBytes int
	switch format.Type {
	case TypeUint8, TypeInt8:
		numBytes = 1
	case TypeUint16, TypeInt16:
		numBytes = 2
	case TypeUint32, TypeInt32:
		numBytes = 4
	case TypeUint64, TypeInt64:
		numBytes = 8
	default:
		return nil, fmt.Errorf("%s is not an integer type", format.Type)
	}

	bytesToRead := make([]byte, numBytes)
	if _, err := handle.Read(bytesToRead); err != nil {
		return nil, err
	}

	if format.Type == TypeUint8 || format.Type == TypeInt8 {
		// byte endianness here is not relevant
		return IntegerResult{big.NewInt(int64(bytesToRead[0]))}, nil
	}

	switch format.Endian {
	case "little":
		switch format.Type {
		case TypeUint16, TypeInt16:
			return IntegerResult{big.NewInt(int64(binary.LittleEndian.Uint16(bytesToRead)))}, nil
		case TypeUint32, TypeInt32:
			return IntegerResult{big.NewInt(int64(binary.LittleEndian.Uint32(bytesToRead)))}, nil
		case TypeUint64, TypeInt64:
			return IntegerResult{big.NewInt(int64(binary.LittleEndian.Uint64(bytesToRead)))}, nil
		}
	case "big":
		switch format.Type {
		case TypeUint16, TypeInt16:
			return IntegerResult{big.NewInt(int64(binary.BigEndian.Uint16(bytesToRead)))}, nil
		case TypeUint32, TypeInt32:
			return IntegerResult{big.NewInt(int64(binary.BigEndian.Uint32(bytesToRead)))}, nil
		case TypeUint64, TypeInt64:
			return IntegerResult{big.NewInt(int64(binary.BigEndian.Uint64(bytesToRead)))}, nil
		}
	}

	return nil, fmt.Errorf("not an integer")
}

type MetaPair struct {
	Key   string
	Value any
}

func ApplyBDF(document Result, targetFile string) ([]MetaPair, error) {
	contents := []MetaPair{}

	root, err := ResultIs[MapResult](document)
	if err != nil {
		return nil, fmt.Errorf("root: %w", err)
	}

	handle, err := os.Open(targetFile)
	if err != nil {
		return nil, err
	}
	defer handle.Close()

	binarySeq, err := GetKeyByIdent[ListResult](root, "binary", true)
	if err != nil {
		return nil, fmt.Errorf("binary: %w", err)
	}

	ns := map[Result]Result{}

	for idx, res := range binarySeq {
		formatType, err := parseFormatType(res, ns)
		if err != nil {
			return nil, fmt.Errorf("binary[%d]: %w", idx, err)
		}

		var emptyPos SeekPos
		if formatType.At != emptyPos {
			if _, err := handle.Seek(formatType.At.Offset, formatType.At.Whence); err != nil {
				return nil, fmt.Errorf("binary[%d].at: %w", idx, err)
			}
		}

		var value Result
		isMagic := false
		switch formatType.Type {
		case TypeMagic:
			if err := checkMagic(handle, formatType); err != nil {
				return nil, err
			}

			isMagic = true
		case TypeUint8, TypeUint16, TypeUint32, TypeUint64, TypeInt8, TypeInt16, TypeInt32, TypeInt64:
			num, err := readInt(handle, formatType)

			if err != nil {
				return nil, fmt.Errorf("binary[%d]: %w", idx, err)
			}

			value = num
		case TypeByte:
			byteSlice := make([]byte, formatType.Size)
			if _, err := handle.Read(byteSlice); err != nil {
				return nil, fmt.Errorf("binary[%d]: %w", idx, err)
			}

			if formatType.Strip {
				trimmed := bytes.TrimFunc(byteSlice, func(r rune) bool {
					return unicode.IsSpace(r) || r == '\x00'
				})

				value = StringResult(trimmed)
			} else {
				value = StringResult(byteSlice)
			}
		default:
			return nil, fmt.Errorf("binary[%d].type: %s is not currently supported", idx, formatType.Type)
		}

		if isMagic {
			continue
		}

		var zero string
		if formatType.Id != zero {
			ns[IdentResult(formatType.Id)] = value
		}

		if formatType.Name != zero {
			contents = append(contents, MetaPair{formatType.Name, value})
		} else if formatType.Id != zero && !strings.HasPrefix(formatType.Id, "_") {
			contents = append(contents, MetaPair{formatType.Id, value})
		}
	}

	return contents, nil
}

// GetMetadata returns the metadata described in the 'meta' key of document.
func GetMetadata(document Result) (Meta, error) {
	rootMap, err := ResultIs[MapResult](document)
	if err != nil {
		return Meta{}, err
	}

	metaMap, err := GetKeyByIdent[MapResult](rootMap, "meta", true)
	if err != nil {
		return Meta{}, err
	}

	return parseMeta(metaMap)
}
