package bindef

import (
	"fmt"
	"math"
	"math/big"
	"slices"
)

func buildSliceFn(node CallNode, args []Result) LazyResult {
	return func(ns Namespace) (Result, error) {
		if len(args) != 3 {
			return nil, LangError{
				Kind:     ErrorType,
				Position: node.Position(),
				Message:  fmt.Sprintf("slice requires 3 arguments, received %d", len(args)),
			}
		}

		start, err := ResultIs[IntegerResult](args[1])
		if err != nil {
			return nil, LangError{
				Kind:     ErrorType,
				Position: node.Arguments[1].Position(),
				Message:  fmt.Sprintf("start argument of slice must be an integer, not %s", args[1].Kind()),
			}
		}

		end, err := ResultIs[IntegerResult](args[2])
		if err != nil {
			return nil, LangError{
				Kind:     ErrorType,
				Position: node.Arguments[2].Position(),
				Message:  fmt.Sprintf("end argument of slice must be an integer, not %s", args[2].Kind()),
			}
		}

		switch target := args[0]; target.Kind() {
		case ResultList:
			list := target.(ListResult)
			if int(start.Int64()) >= len(list) {
				return nil, LangError{
					Kind:     ErrorValue,
					Position: node.Arguments[1].Position(),
					Message:  "start argument out of bounds",
				}
			}

			endInt := min(int(end.Int64()), len(list))
			return list[start.Int64():endInt], nil
		case ResultString:
			str := target.(StringResult)
			if int(start.Int64()) >= len(str) {
				return nil, LangError{
					Kind:     ErrorValue,
					Position: node.Arguments[1].Position(),
					Message:  "start argument out of bounds",
				}
			}

			endInt := min(int(end.Int64()), len(str))
			return str[start.Int64():endInt], nil
		default:
			return nil, LangError{
				Kind:     ErrorType,
				Position: node.Arguments[0].Position(),
				Message:  fmt.Sprintf("first argument of slice must be a list or string, not %s", target.Kind()),
			}
		}
	}
}

func buildHasFn(node CallNode, args []Result) LazyResult {
	return func(ns Namespace) (Result, error) {
		if len(args) != 2 {
			return nil, LangError{
				Kind:     ErrorType,
				Position: node.Position(),
				Message:  fmt.Sprintf("has requires 2 arguments, received %d", len(args)),
			}
		}

		target, err := ResultIs[ListResult](args[0])
		if err != nil {
			return nil, LangError{
				Kind:     ErrorType,
				Position: node.Position(),
				Message:  fmt.Sprintf("first argument of has must be a list, not %s", args[0].Kind()),
			}
		}

		return BooleanResult(slices.Contains(target, args[1])), nil
	}
}

func buildParseIntFn(node CallNode, args []Result) LazyResult {
	return func(ns Namespace) (Result, error) {
		if len(args) != 1 {
			return nil, LangError{
				Kind:     ErrorType,
				Position: node.Position(),
				Message:  fmt.Sprintf("parseInt requires 1 argument, received %d", len(args)),
			}
		}

		switch target := args[0]; target.Kind() {
		case ResultString:
			strVal := target.(StringResult)
			intVal, ok := new(big.Int).SetString(string(strVal), 10)
			if !ok {
				return nil, LangError{
					Kind:     ErrorValue,
					Position: node.Position(),
					Message:  fmt.Sprintf("parseInt: invalid input %v", intVal),
				}
			}

			return IntegerResult{intVal}, nil
		default:
			return nil, LangError{
				Kind:     ErrorType,
				Position: node.Position(),
				Message:  fmt.Sprintf("parseInt argument must be a string, not %s", target.Kind()),
			}
		}
	}
}

func buildCeilFn(node CallNode, args []Result) LazyResult {
	return func(ns Namespace) (Result, error) {
		if len(args) != 1 {
			return nil, LangError{
				Kind:     ErrorType,
				Position: node.Position(),
				Message:  fmt.Sprintf("ceil requires 1 argument, received %d", len(args)),
			}
		}

		switch target := args[0]; target.Kind() {
		case ResultInt:
			return target.(IntegerResult), nil
		case ResultFloat:
			ceiling := math.Ceil(float64(target.(FloatResult)))
			return IntegerResult{new(big.Int).SetInt64(int64(ceiling))}, nil
		default:
			return nil, LangError{
				Kind:     ErrorType,
				Position: node.Position(),
				Message:  fmt.Sprintf("ceil argument must be numeric, not %s", target.Kind()),
			}
		}
	}
}

func buildFloorFn(node CallNode, args []Result) LazyResult {
	return func(ns Namespace) (Result, error) {
		if len(args) != 1 {
			return nil, LangError{
				Kind:     ErrorType,
				Position: node.Position(),
				Message:  fmt.Sprintf("floor requires 1 argument, received %d", len(args)),
			}
		}

		switch target := args[0]; target.Kind() {
		case ResultInt:
			return target.(IntegerResult), nil
		case ResultFloat:
			flooring := math.Floor(float64(target.(FloatResult)))
			return IntegerResult{new(big.Int).SetInt64(int64(flooring))}, nil
		default:
			return nil, LangError{
				Kind:     ErrorType,
				Position: node.Position(),
				Message:  fmt.Sprintf("floor argument must be numeric, not %s", target.Kind()),
			}
		}
	}
}

func buildAbsFn(node CallNode, args []Result) LazyResult {
	return func(ns Namespace) (Result, error) {
		if len(args) != 1 {
			return nil, LangError{
				Kind:     ErrorType,
				Position: node.Position(),
				Message:  fmt.Sprintf("abs requires 1 argument, received %d", len(args)),
			}
		}

		switch target := args[0]; target.Kind() {
		case ResultInt:
			return IntegerResult{new(big.Int).Abs(target.(IntegerResult).Int)}, nil
		case ResultFloat:
			abs := math.Abs(float64(target.(FloatResult)))
			return FloatResult(abs), nil
		default:
			return nil, LangError{
				Kind:     ErrorType,
				Position: node.Position(),
				Message:  fmt.Sprintf("abs argument must be numeric, not %s", target.Kind()),
			}
		}
	}
}

func buildLenFn(node CallNode, args []Result) LazyResult {
	return func(ns Namespace) (Result, error) {
		if len(args) != 1 {
			return nil, LangError{
				Kind:     ErrorType,
				Position: node.Position(),
				Message:  fmt.Sprintf("len requires 1 argument, received %d", len(args)),
			}
		}

		switch target := args[0]; target.Kind() {
		case ResultList:
			length := int64(len(target.(ListResult)))
			return IntegerResult{new(big.Int).SetInt64(length)}, nil
		case ResultMap:
			length := int64(len(target.(MapResult)))
			return IntegerResult{new(big.Int).SetInt64(length)}, nil
		case ResultString:
			length := int64(len(target.(StringResult)))
			return IntegerResult{new(big.Int).SetInt64(length)}, nil
		default:
			return nil, LangError{
				Kind:     ErrorType,
				Position: node.Position(),
				Message:  fmt.Sprintf("len: incompatible type %s", target.Kind()),
			}
		}
	}
}
