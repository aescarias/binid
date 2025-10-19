package bindef

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"maps"
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
	Type         TypeName     // Format type (e.g. uint8, int8). May be more complex such as byte[n].
	Id           string       // Field identifier.
	Name         string       // Human-readable field name.
	Doc          string       // Documentation.
	At           *SeekPos     // Seek position.
	Valid        LazyResult   // Validation function.
	If           LazyResult   // Only process value on condition.
	Endian       string       // For integer types only, the byte endianness (either "big" or "little").
	Match        []MagicTag   // For magic types only, the pattern(s) that must match.
	Size         int64        // For byte types only, the size of the byte string.
	Strip        bool         // For byte types only, whether to strip whitespace or null bytes from the ends of the string.
	RawFields    []MapResult  // For structures only, the fields contained in the struct.
	ProcFields   []FormatType // For structures only, the fields contained in the struct as format types.
	VarValue     LazyResult   // For variable definitions, the value contained.
	ArrSize      int64        // For array types, the amount of elements in the array.
	ArrSizeIsEos bool         // For array types, whether the array spans to the end of the sequence.
	RawArrItem   MapResult    // For array types, the format type as a result.
	ProcArrItem  *FormatType  // For array types, the format type for each element of the array.
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

func getFormatEndian(bin MapResult, base MapResult, ns Namespace) (string, error) {
	var endian string
	endianRes, err := GetEvalKeyByIdent[StringResult](bin, "endian", false, ns)
	if err != nil {
		return "", err
	}

	if endianRes == "" {
		baseEndianRes, err := GetEvalKeyByIdent[StringResult](base, "endian", true, ns)
		if err != nil {
			return "", err
		}

		endian = string(baseEndianRes)
	} else {
		endian = string(endianRes)
	}

	endian = strings.ToLower(endian)
	if endian != "little" && endian != "big" {
		return "", fmt.Errorf("endian is not 'little' or 'big'")
	}

	return endian, nil
}

var ErrSkipped = fmt.Errorf("format type was skipped because condition is false")

func ParseFormatType(format Result, ns Namespace, base MapResult) (FormatType, error) {
	bin, err := ResultIs[MapResult](format)
	if err != nil {
		return FormatType{}, err
	}

	ifRes, err := GetKeyByIdent[LazyResult](bin, "if", false)
	if err != nil {
		return FormatType{}, err
	}

	if ifRes != nil {
		willParseRes, err := ifRes(ns)
		if err != nil {
			return FormatType{}, err
		}

		willParse, err := ResultIs[BooleanResult](willParseRes)
		if err != nil {
			return FormatType{}, err
		}

		if !willParse {
			return FormatType{}, ErrSkipped
		}
	}

	genTypeRes, err := GetKeyByIdent[Result](bin, "type", true)
	if err != nil {
		return FormatType{}, err
	}

	var typeRes TypeResult
	switch genTypeRes.Kind() {
	case ResultLazy:
		typeNs := Namespace{}
		maps.Copy(typeNs, ns)
		typeNs[IdentResult("eos")] = IdentResult("eos")

		tempRes, err := genTypeRes.(LazyResult)(typeNs)
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

	var atVal *SeekPos = nil

	if atRes != nil {
		var resolvedAt Result
		if atRes.Kind() == ResultLazy {
			at, err := atRes.(LazyResult)(ns)
			if err != nil {
				return FormatType{}, err
			}
			resolvedAt = at
		} else {
			resolvedAt = atRes
		}

		switch resolvedAt.Kind() {
		case ResultInt, ResultFloat:
			offsetInt, err := NumberResultAsInt(resolvedAt)
			if err != nil {
				return FormatType{}, err
			}

			atVal = &SeekPos{Offset: offsetInt, Whence: io.SeekStart}
		case ResultList:
			atValList, err := ResultIs[ListResult](resolvedAt)
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
				atVal = &SeekPos{Offset: offsetInt, Whence: io.SeekStart}
			case StringResult("end"):
				atVal = &SeekPos{Offset: offsetInt, Whence: io.SeekEnd}
			case StringResult("current"):
				atVal = &SeekPos{Offset: offsetInt, Whence: io.SeekCurrent}
			default:
				return FormatType{}, fmt.Errorf("at[1]: whence is not a valid seek identifier")
			}
		default:
			return FormatType{}, fmt.Errorf("value 'at' is not a list or number")
		}
	}

	validRes, err := GetKeyByIdent[LazyResult](bin, "valid", false)
	if err != nil {
		return FormatType{}, err
	}

	baseFormat := FormatType{
		Type:  typeRes.Name,
		Id:    string(idRes),
		Name:  string(nameRes),
		Doc:   string(docRes),
		At:    atVal,
		If:    ifRes,
		Valid: validRes,
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
		endian, err := getFormatEndian(bin, base, ns)
		if err != nil {
			return FormatType{}, err
		}

		baseFormat.Endian = endian
		return baseFormat, nil
	case TypeVar:
		contents, err := GetKeyByIdent[LazyResult](bin, "value", true)
		if err != nil {
			return FormatType{}, err
		}

		baseFormat.VarValue = contents
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

		if baseFormat.Size < 0 {
			return FormatType{}, fmt.Errorf("byte size must be non-negative")
		}

		return baseFormat, nil
	case TypeArray:
		if len(typeRes.Params) <= 0 {
			return baseFormat, fmt.Errorf("array must specify length")
		}

		item, err := GetKeyByIdent[MapResult](bin, "item", true)
		if err != nil {
			return FormatType{}, err
		}

		var arrSizeRes Result
		if tp := typeRes.Params[0]; tp.Kind() == ResultLazy {
			var err error
			if arrSizeRes, err = tp.(LazyResult)(ns); err != nil {
				return FormatType{}, err
			}
		} else {
			arrSizeRes = typeRes.Params[0]
		}

		if arrSizeRes.Kind() == ResultIdent {
			if arrSizeRes.(IdentResult) == IdentResult("eos") {
				baseFormat.ArrSizeIsEos = true
			} else {
				return FormatType{}, fmt.Errorf("array size must be numeric")
			}
		} else {
			arrSize, err := ResultIs[IntegerResult](arrSizeRes)
			if err != nil {
				return FormatType{}, err
			}

			baseFormat.ArrSize = arrSize.Int64()
			if baseFormat.ArrSize < 0 {
				return FormatType{}, fmt.Errorf("array length must be non-negative")
			}
		}

		baseFormat.RawArrItem = item
		return baseFormat, nil
	case TypeStruct:
		fields, err := GetKeyByIdent[ListResult](bin, "fields", true)
		if err != nil {
			return FormatType{}, err
		}

		endian, err := getFormatEndian(bin, base, ns)
		if err != nil {
			return FormatType{}, err
		}

		for _, field := range fields {
			element, err := ResultIs[MapResult](field)
			if err != nil {
				return FormatType{}, err
			}

			baseFormat.RawFields = append(baseFormat.RawFields, element)
		}

		baseFormat.Endian = endian
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

func checkMagic(handle *os.File, format FormatType) (Result, error) {
	baseOffset, err := handle.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	for _, tag := range format.Match {
		if _, err := handle.Seek(baseOffset+tag.Offset, io.SeekStart); err != nil {
			return nil, err
		}

		contents := []byte(tag.Contents)

		matchBytes := make([]byte, len(contents))
		if _, err := handle.Read(matchBytes); err != nil {
			return nil, err
		}

		if slices.Equal(matchBytes, contents) {
			return StringResult(contents), nil
		}
	}

	return nil, ErrMagic{Offset: baseOffset}
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
	Field FormatType
	Value Result
}

func processType(handle *os.File, format *FormatType, ns Namespace) (res Result, err error) {
	if format.At != nil {
		if _, err := handle.Seek(format.At.Offset, format.At.Whence); err != nil {
			return nil, err
		}
	}

	var value Result
	switch format.Type {
	case TypeMagic:
		magic, err := checkMagic(handle, *format)
		if err != nil {
			return nil, err
		}

		value = magic
	case TypeUint8, TypeUint16, TypeUint32, TypeUint64, TypeInt8, TypeInt16, TypeInt32, TypeInt64:
		num, err := readInt(handle, *format)

		if err != nil {
			return nil, err
		}

		value = num
	case TypeVar:
		res, err := format.VarValue(ns)
		if err != nil {
			return nil, err
		}

		value = res
	case TypeByte:
		byteSlice := make([]byte, format.Size)
		if _, err := handle.Read(byteSlice); err != nil {
			return nil, err
		}

		if format.Strip {
			trimmed := bytes.TrimFunc(byteSlice, func(r rune) bool {
				return unicode.IsSpace(r) || r == '\x00'
			})

			value = StringResult(trimmed)
		} else {
			value = StringResult(byteSlice)
		}
	case TypeStruct:
		mapping := MapResult{}

		// special namespace for identifiers within a struct so that they can
		// be referred using their names directly while also avoiding polluting
		// the global namespace.
		currentNs := Namespace{}
		maps.Copy(currentNs, ns)

		inherited := MapResult{IdentResult("endian"): StringResult(format.Endian)}

		format.ProcFields = []FormatType{}
		for _, field := range format.RawFields {
			fieldFormat, err := ParseFormatType(field, currentNs, inherited)
			if err != nil {
				if errors.Is(err, ErrSkipped) {
					continue
				}
				return nil, err
			}

			res, err := processType(handle, &fieldFormat, currentNs)
			if err != nil {
				return nil, err
			}

			if fieldFormat.Id != "" {
				mapping[IdentResult(fieldFormat.Id)] = res
				currentNs[IdentResult(fieldFormat.Id)] = res
			}

			format.ProcFields = append(format.ProcFields, fieldFormat)
		}

		value = mapping
	case TypeArray:
		elements := ListResult{}

		idx := int64(0)

		for format.ArrSizeIsEos || idx < format.ArrSize {
			arrItem, err := ParseFormatType(format.RawArrItem, ns, nil)
			if err != nil {
				return nil, err
			}

			proc, err := processType(handle, &arrItem, ns)
			if err != nil {
				if format.ArrSizeIsEos && errors.Is(err, io.EOF) {
					break
				}
				return nil, err
			}

			format.ProcArrItem = &arrItem

			elements = append(elements, proc)
			idx += 1
		}

		value = elements
	default:
		return nil, fmt.Errorf("%s is not currently supported", format.Type)
	}

	if format.Id != "" {
		ns[IdentResult(format.Id)] = value
	}

	if format.Valid != nil {
		isValidRes, err := format.Valid(ns)
		if err != nil {
			return nil, err
		}

		isValid, err := ResultIs[BooleanResult](isValidRes)
		if err != nil {
			return nil, err
		}

		if !isValid {
			return nil, fmt.Errorf("value for %q is invalid (has value %v)", format.Id, value)
		}
	}

	return value, nil
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

	ns := Namespace{}

	for idx, res := range binarySeq {
		formatType, err := ParseFormatType(res, ns, nil)
		if err != nil {
			if errors.Is(err, ErrSkipped) {
				continue
			}
			return nil, fmt.Errorf("binary[%d]: %w", idx, err)
		}

		value, err := processType(handle, &formatType, ns)
		if err != nil {
			if err, ok := err.(ErrMagic); ok {
				return nil, err
			}

			return nil, fmt.Errorf("binary[%d]: %w", idx, err)
		}

		if formatType.Name != "" || formatType.Id != "" {
			contents = append(contents, MetaPair{formatType, value})
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
