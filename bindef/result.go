package bindef

import (
	"fmt"
	"math"
	"math/big"
)

// A Namespace represents a mapping of results (usually identifiers) to other results.
// This is used in identifier resolution.
type Namespace map[Result]Result

// A ResultKind represents the type of result (e.g. integer, float, list, etc.)
type ResultKind string

const (
	ResultInt     ResultKind = "Int"
	ResultFloat   ResultKind = "Float"
	ResultBoolean ResultKind = "Boolean"
	ResultMap     ResultKind = "Map"
	ResultList    ResultKind = "List"
	ResultString  ResultKind = "String"
	ResultIdent   ResultKind = "Identifier"
	ResultLazy    ResultKind = "Lazy"
	ResultType    ResultKind = "Type"
)

type Result interface {
	Kind() ResultKind
}

// IntegerResult is a result that represents a signed multi-precision integer.
type IntegerResult struct {
	*big.Int
}

func (ir IntegerResult) Kind() ResultKind { return ResultInt }

// FloatResult is a result representing a 64-bit float
type FloatResult float64

func (fr FloatResult) Kind() ResultKind { return ResultFloat }

// BooleanResult is a result representing a boolean.
type BooleanResult bool

func (br BooleanResult) Kind() ResultKind { return ResultBoolean }

// MapResult is a result representing a key-value map.
type MapResult map[Result]Result

func (mr MapResult) Kind() ResultKind { return ResultMap }

// ListResult is a result representing a list of items.
type ListResult []Result

func (lr ListResult) Kind() ResultKind { return ResultList }

// StringResult is a result representing a sequence of Unicode codepoints.
type StringResult string

func (sr StringResult) Kind() ResultKind { return ResultString }

// IdentResult is a result representing an identifier.
type IdentResult string

func (ir IdentResult) Kind() ResultKind { return ResultIdent }

// LazyResult is a result representing a lazily-evaluated expression.
type LazyResult func(ns Namespace) (Result, error)

func (lr LazyResult) Kind() ResultKind { return ResultLazy }

// TypeResult is a result representing a type.
type TypeResult struct {
	Name   TypeName
	Params []Result
}

type TypeName string

const (
	TypeMagic   TypeName = "magic"
	TypeBool    TypeName = "bool"
	TypeByte    TypeName = "byte"
	TypeStruct  TypeName = "struct"
	TypeArray   TypeName = "array"
	TypeVar     TypeName = "var"
	TypeEnum    TypeName = "enum"
	TypeUint8   TypeName = "uint8"
	TypeUint16  TypeName = "uint16"
	TypeUint24  TypeName = "uint24"
	TypeUint32  TypeName = "uint32"
	TypeUint64  TypeName = "uint64"
	TypeInt8    TypeName = "int8"
	TypeInt16   TypeName = "int16"
	TypeInt24   TypeName = "int24"
	TypeInt32   TypeName = "int32"
	TypeInt64   TypeName = "int64"
	TypeFloat32 TypeName = "float32"
	TypeFloat64 TypeName = "float64"
)

func (tr TypeResult) Kind() ResultKind { return ResultType }

var AvailableTypeNames = []TypeName{
	TypeMagic, TypeBool, TypeByte, TypeStruct, TypeArray, TypeVar, TypeEnum,
	TypeUint8, TypeUint16, TypeUint24, TypeUint32, TypeUint64,
	TypeInt8, TypeInt16, TypeInt24, TypeInt32, TypeInt64,
	TypeFloat32, TypeFloat64,
}

var AvailableNumericTypes = []TypeName{
	TypeUint8, TypeUint16, TypeUint24, TypeUint32, TypeUint64,
	TypeInt8, TypeInt16, TypeInt24, TypeInt32, TypeInt64,
	TypeFloat32, TypeFloat64,
}

// GetEvalKeyByIdent pipes its arguments to [GetKeyByIdent] and returns its result but
// handling cases where mapping is expected to contain a key that is either a result
// of type T or a lazy result that evaluates to one of type T. This evaluation is
// done using the namespace ns.
func GetEvalKeyByIdent[T Result](mapping map[Result]Result, key string, required bool, ns Namespace) (T, error) {
	var zero T

	maybeEvalRes, err := GetKeyByIdent[Result](mapping, key, required)
	if err != nil {
		return zero, err
	}

	if maybeEvalRes == nil {
		return zero, err
	}

	var res Result
	if maybeEvalRes.Kind() == ResultLazy {
		if res, err = maybeEvalRes.(LazyResult)(ns); err != nil {
			return zero, err
		}
	} else {
		res = maybeEvalRes
	}

	asserted, ok := res.(T)
	if !ok {
		panic(fmt.Sprintf("type assertion failed for %s (input is %s)", zero.Kind(), res.Kind()))
	}

	return asserted, nil
}

// GetKeyByIdent converts a string key to an identifier and returns the value of the
// key if it exists in mapping. If required is specified, the key must exist or an
// error is returned; otherwise, a zero value is returned.
func GetKeyByIdent[T Result](mapping map[Result]Result, key string, required bool) (T, error) {
	res, ok := mapping[IdentResult(key)]

	var zero T
	if !ok && required {
		return zero, fmt.Errorf("missing key %q is required", key)
	} else if !ok {
		return zero, nil
	}

	var value any
	var err error

	switch res.Kind() {
	case ResultMap:
		value, err = ResultIs[MapResult](res)
	case ResultList:
		value, err = ResultIs[ListResult](res)
	case ResultIdent:
		value, err = ResultIs[IdentResult](res)
	case ResultFloat:
		value, err = ResultIs[FloatResult](res)
	case ResultInt:
		value, err = ResultIs[IntegerResult](res)
	case ResultString:
		value, err = ResultIs[StringResult](res)
	case ResultBoolean:
		value, err = ResultIs[BooleanResult](res)
	case ResultLazy:
		value, err = ResultIs[LazyResult](res)
	case ResultType:
		value, err = ResultIs[TypeResult](res)
	default:
		return zero, fmt.Errorf("cannot assert unsupported result %s", res.Kind())
	}

	if err != nil {
		return zero, err
	}

	asserted, ok := value.(T)
	if !ok {
		panic(fmt.Sprintf("type assertion failed for %s (input is %s)", zero.Kind(), res.Kind()))
	}

	return asserted, nil
}

// EvalResultIs has the same behavior as [ResultIs], that is, it checks whether
// a result res is of type T. However, if res is a lazy result, it is first evaluated
// before checking.
func EvalResultIs[T Result](res Result, ns Namespace) (T, error) {
	var zero T

	var evalRes Result

	if res.Kind() == ResultLazy {
		var err error
		if evalRes, err = res.(LazyResult)(ns); err != nil {
			return zero, err
		}
	} else {
		evalRes = res
	}

	if evalRes.Kind() != zero.Kind() {
		return zero, fmt.Errorf(
			"expected result of type %s, received %s", zero.Kind(), res.Kind(),
		)
	}

	val, ok := evalRes.(T)
	if !ok {
		panic(fmt.Sprintf("type assertion failed for %s (input is %s)", zero.Kind(), evalRes.Kind()))
	}

	return val, nil

}

// ResultIs validates whether a result res is of type T. If it is, an asserted
// result is returned. Otherwise, a zero value and an error are returned.
func ResultIs[T Result](res Result) (T, error) {
	var zero T

	if res.Kind() != zero.Kind() {
		return zero, fmt.Errorf(
			"expected result of type %s, received %s", zero.Kind(), res.Kind(),
		)
	}

	val, ok := res.(T)
	if !ok {
		panic(fmt.Sprintf("type assertion failed for %s (input is %s)", zero.Kind(), res.Kind()))
	}

	return val, nil
}

// NumberResultAsInt converts a numeric result res into an int64. If res is a float,
// the returned value is a truncated integer. If res is an integer, the returned value
// is simply the integer. Otherwise, a value of -1 and an error is returned.
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

// IntegerInBounds reports whether the integer value is in bounds with the type
// bound provided.
func IntegerInBounds(bound TypeName, value IntegerResult) bool {
	bigUint := func(value uint64) *big.Int {
		return new(big.Int).SetUint64(value)
	}
	bigInt := func(value int64) *big.Int {
		return new(big.Int).SetInt64(value)
	}

	intValue := value.Int

	var minBound, maxBound *big.Int

	switch bound {
	case TypeUint8:
		minBound, maxBound = bigUint(0), bigUint(0xff)
	case TypeUint16:
		minBound, maxBound = bigUint(0), bigUint(0xffff)
	case TypeUint24:
		minBound, maxBound = bigUint(0), bigUint(0xffffff)
	case TypeUint32:
		minBound, maxBound = bigUint(0), bigUint(math.MaxUint32)
	case TypeUint64:
		minBound, maxBound = bigUint(0), bigUint(math.MaxUint64)
	case TypeInt8:
		minBound, maxBound = bigInt(math.MinInt8), bigInt(math.MaxInt8)
	case TypeInt16:
		minBound, maxBound = bigInt(math.MinInt16), bigInt(math.MaxInt16)
	case TypeInt24:
		minBound, maxBound = bigInt(-8_388_608), bigInt(8_388_607)
	case TypeInt32:
		minBound, maxBound = bigInt(math.MinInt32), bigInt(math.MaxInt32)
	case TypeInt64:
		minBound, maxBound = bigInt(math.MinInt64), bigInt(math.MaxInt64)
	}

	return intValue.Cmp(minBound) >= 0 && intValue.Cmp(maxBound) <= 0
}
