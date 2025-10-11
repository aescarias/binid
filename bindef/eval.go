package bindef

import (
	"errors"
	"fmt"
	"maps"
	"math"
	"math/big"
	"slices"
	"strconv"
)

func doBinOpEquals(left, right Result) BooleanResult {
	if left.Kind() == ResultInt && right.Kind() == ResultInt {
		cmp := left.(IntegerResult).Cmp(right.(IntegerResult).Int)
		return BooleanResult(cmp == 0)
	} else if left.Kind() == ResultFloat && right.Kind() == ResultFloat {
		return BooleanResult(left.(FloatResult) == right.(FloatResult))
	} else if left.Kind() == ResultInt && right.Kind() == ResultFloat {
		rightFloat := float64(right.(FloatResult))
		rightTrunc := math.Trunc(rightFloat)
		if rightTrunc != rightFloat {
			return BooleanResult(false)
		}

		cmp := left.(IntegerResult).Cmp(new(big.Int).SetInt64(int64(rightTrunc)))
		return BooleanResult(cmp == 0)
	} else if left.Kind() == ResultFloat && right.Kind() == ResultInt {
		leftFloat := float64(left.(FloatResult))
		leftTrunc := math.Trunc(leftFloat)
		if leftTrunc != leftFloat {
			return BooleanResult(false)
		}

		cmp := new(big.Int).SetInt64(int64(leftFloat)).Cmp(right.(IntegerResult).Int)
		return BooleanResult(cmp == 0)
	} else if left.Kind() == ResultString && right.Kind() == ResultString {
		return BooleanResult(left.(StringResult) == right.(StringResult))
	} else if left.Kind() == ResultBoolean && right.Kind() == ResultBoolean {
		return BooleanResult(left.(BooleanResult) == right.(BooleanResult))
	} else if left.Kind() == ResultList && right.Kind() == ResultList {
		equals := slices.Equal(left.(ListResult), right.(ListResult))
		return BooleanResult(equals)
	} else if left.Kind() == ResultMap && right.Kind() == ResultMap {
		equals := maps.Equal(left.(MapResult), right.(MapResult))
		return BooleanResult(equals)
	}

	return BooleanResult(false)
}

var ErrCannotCompare = fmt.Errorf("cannot compare these types")
var ErrCannotBoolean = fmt.Errorf("cannot convert result to boolean")

func doBinOpLt(left, right Result) (BooleanResult, error) {
	if left.Kind() == ResultInt && right.Kind() == ResultInt {
		cmp := left.(IntegerResult).Cmp(right.(IntegerResult).Int)
		return BooleanResult(cmp == -1), nil
	} else if left.Kind() == ResultFloat && right.Kind() == ResultFloat {
		return BooleanResult(left.(FloatResult) < right.(FloatResult)), nil
	} else if left.Kind() == ResultInt && right.Kind() == ResultFloat {
		leftFloat, _ := left.(IntegerResult).Float64()
		return BooleanResult(leftFloat < float64(right.(FloatResult))), nil
	} else if left.Kind() == ResultFloat && right.Kind() == ResultInt {
		rightFloat, _ := right.(IntegerResult).Float64()
		return BooleanResult(float64(left.(FloatResult)) < rightFloat), nil
	} else if left.Kind() == ResultString && right.Kind() == ResultString {
		return BooleanResult(left.(StringResult) < right.(StringResult)), nil
	}

	return BooleanResult(false), ErrCannotCompare
}

func doBinOpGt(left, right Result) (BooleanResult, error) {
	if left.Kind() == ResultInt && right.Kind() == ResultInt {
		cmp := left.(IntegerResult).Cmp(right.(IntegerResult).Int)
		return BooleanResult(cmp == 1), nil
	} else if left.Kind() == ResultFloat && right.Kind() == ResultFloat {
		return BooleanResult(left.(FloatResult) > right.(FloatResult)), nil
	} else if left.Kind() == ResultInt && right.Kind() == ResultFloat {
		leftFloat, _ := left.(IntegerResult).Float64()
		return BooleanResult(leftFloat > float64(right.(FloatResult))), nil
	} else if left.Kind() == ResultFloat && right.Kind() == ResultInt {
		rightFloat, _ := right.(IntegerResult).Float64()
		return BooleanResult(float64(left.(FloatResult)) > rightFloat), nil
	} else if left.Kind() == ResultString && right.Kind() == ResultString {
		return BooleanResult(left.(StringResult) > right.(StringResult)), nil
	}

	return BooleanResult(false), ErrCannotCompare
}

// EvaluateBinOp evaluates a binary operation node using namespace.
func EvaluateBinOp(node BinOpNode, namespace Namespace) (Result, error) {
	left, err := Evaluate(node.Left, namespace)
	if err != nil {
		return nil, err
	}

	right, err := Evaluate(node.Right, namespace)
	if err != nil {
		return nil, err
	}

	switch node.Op.Kind {
	case TokenPlus:
		if left.Kind() == ResultInt && right.Kind() == ResultInt {
			return IntegerResult{new(big.Int).Add(left.(IntegerResult).Int, right.(IntegerResult).Int)}, nil
		} else if left.Kind() == ResultFloat && right.Kind() == ResultFloat {
			return left.(FloatResult) + right.(FloatResult), nil
		} else if left.Kind() == ResultFloat && right.Kind() == ResultInt {
			rightFloat, _ := right.(IntegerResult).Float64()
			return left.(FloatResult) + FloatResult(rightFloat), nil
		} else if left.Kind() == ResultInt && right.Kind() == ResultFloat {
			leftFloat, _ := left.(IntegerResult).Float64()
			return FloatResult(leftFloat) + right.(FloatResult), nil
		} else if left.Kind() == ResultString && right.Kind() == ResultString {
			return left.(StringResult) + right.(StringResult), nil
		}

		return nil, LangError{
			ErrorType,
			node.Position(),
			fmt.Sprintf("binary operation %s is not defined on types %s and %s",
				node.Op.Value, left.Kind(), right.Kind()),
		}
	case TokenMinus:
		if left.Kind() == ResultInt && right.Kind() == ResultInt {
			return IntegerResult{new(big.Int).Sub(left.(IntegerResult).Int, right.(IntegerResult).Int)}, nil
		} else if left.Kind() == ResultFloat && right.Kind() == ResultFloat {
			return left.(FloatResult) - right.(FloatResult), nil
		} else if left.Kind() == ResultFloat && right.Kind() == ResultInt {
			rightFloat, _ := right.(IntegerResult).Float64()
			return left.(FloatResult) - FloatResult(rightFloat), nil
		} else if left.Kind() == ResultInt && right.Kind() == ResultFloat {
			leftFloat, _ := left.(IntegerResult).Float64()
			return FloatResult(leftFloat) - right.(FloatResult), nil
		}

		return nil, LangError{
			ErrorType,
			node.Position(),
			fmt.Sprintf("binary operation %s is not defined on types %s and %s",
				node.Op.Value, left.Kind(), right.Kind()),
		}
	case TokenMul:
		if left.Kind() == ResultInt && right.Kind() == ResultInt {
			return IntegerResult{new(big.Int).Mul(left.(IntegerResult).Int, right.(IntegerResult).Int)}, nil
		} else if left.Kind() == ResultFloat && right.Kind() == ResultFloat {
			return left.(FloatResult) * right.(FloatResult), nil
		} else if left.Kind() == ResultFloat && right.Kind() == ResultInt {
			rightFloat, _ := right.(IntegerResult).Float64()
			return left.(FloatResult) * FloatResult(rightFloat), nil
		} else if left.Kind() == ResultInt && right.Kind() == ResultFloat {
			leftFloat, _ := left.(IntegerResult).Float64()
			return FloatResult(leftFloat) * right.(FloatResult), nil
		}

		return nil, LangError{
			ErrorType,
			node.Position(),
			fmt.Sprintf("binary operation %s is not defined on types %s and %s",
				node.Op.Value, left.Kind(), right.Kind()),
		}
	case TokenPow:
		if left.Kind() == ResultInt && right.Kind() == ResultInt {
			leftFloat, _ := left.(IntegerResult).Float64()
			rightFloat, _ := right.(IntegerResult).Float64()

			pow := math.Pow(leftFloat, rightFloat)

			if pow != math.Trunc(pow) {
				return FloatResult(pow), nil
			}

			return IntegerResult{new(big.Int).SetInt64(int64(pow))}, nil
		} else if left.Kind() == ResultFloat && right.Kind() == ResultFloat {
			return FloatResult(math.Pow(float64(left.(FloatResult)), float64(right.(FloatResult)))), nil
		} else if left.Kind() == ResultFloat && right.Kind() == ResultInt {
			rightFloat, _ := right.(IntegerResult).Float64()
			return FloatResult(math.Pow(float64(left.(FloatResult)), rightFloat)), nil
		} else if left.Kind() == ResultInt && right.Kind() == ResultFloat {
			leftFloat, _ := left.(IntegerResult).Float64()
			return FloatResult(math.Pow(leftFloat, float64(right.(FloatResult)))), nil
		}

		return nil, LangError{
			ErrorType,
			node.Position(),
			fmt.Sprintf("binary operation %s is not defined on types %s and %s",
				node.Op.Value, left.Kind(), right.Kind()),
		}
	case TokenDiv:
		if left.Kind() == ResultInt && right.Kind() == ResultInt {
			leftFloat, _ := left.(IntegerResult).Float64()
			rightFloat, _ := right.(IntegerResult).Float64()

			if rightFloat == 0 {
				return nil, LangError{ErrorDomain, node.Position(), "division by zero"}
			}

			return FloatResult(leftFloat / rightFloat), nil
		} else if left.Kind() == ResultFloat && right.Kind() == ResultFloat {
			if right.(FloatResult) == 0 {
				return nil, LangError{ErrorDomain, node.Position(), "division by zero"}
			}

			return left.(FloatResult) / right.(FloatResult), nil
		} else if left.Kind() == ResultFloat && right.Kind() == ResultInt {
			rightFloat, _ := right.(IntegerResult).Float64()
			if rightFloat == 0.0 {
				return nil, LangError{ErrorDomain, node.Position(), "division by zero"}
			}

			return left.(FloatResult) / FloatResult(rightFloat), nil
		} else if left.Kind() == ResultInt && right.Kind() == ResultFloat {
			leftFloat, _ := left.(IntegerResult).Float64()
			if right.(FloatResult) == 0.0 {
				return nil, LangError{ErrorDomain, node.Position(), "division by zero"}
			}

			return FloatResult(leftFloat) / right.(FloatResult), nil
		}

		return nil, LangError{
			ErrorType,
			node.Position(),
			fmt.Sprintf("binary operation %s is not defined on types %s and %s",
				node.Op.Value, left.Kind(), right.Kind()),
		}
	case TokenRemainder:
		if left.Kind() == ResultInt && right.Kind() == ResultInt {
			if right.(IntegerResult).Int == big.NewInt(0) {
				return nil, LangError{ErrorDomain, node.Position(), "integer remainder by zero"}
			}

			return IntegerResult{new(big.Int).Rem(left.(IntegerResult).Int, right.(IntegerResult).Int)}, nil
		} else if left.Kind() == ResultFloat && right.Kind() == ResultFloat {
			if right.(FloatResult) == 0 {
				return nil, LangError{ErrorDomain, node.Position(), "float remainder by zero"}
			}

			rem := math.Remainder(float64(left.(FloatResult)), float64(right.(FloatResult)))
			return FloatResult(rem), nil
		} else if left.Kind() == ResultFloat && right.Kind() == ResultInt {
			rightFloat, _ := right.(IntegerResult).Float64()
			if rightFloat == 0.0 {
				return nil, LangError{ErrorDomain, node.Position(), "float remainder by zero"}
			}

			rem := math.Remainder(float64(left.(FloatResult)), rightFloat)
			return FloatResult(rem), nil
		} else if left.Kind() == ResultInt && right.Kind() == ResultFloat {
			leftFloat, _ := left.(IntegerResult).Float64()
			if right.(FloatResult) == 0.0 {
				return nil, LangError{ErrorDomain, node.Position(), "float remainder by zero"}
			}

			rem := math.Remainder(leftFloat, float64(right.(FloatResult)))
			return FloatResult(rem), nil
		}

		return nil, LangError{
			ErrorType,
			node.Position(),
			fmt.Sprintf("binary operation %s is not defined on types %s and %s",
				node.Op.Value, left.Kind(), right.Kind()),
		}
	case TokenBitwiseLeft:
		if left.Kind() == ResultInt && right.Kind() == ResultInt {
			return IntegerResult{new(big.Int).Lsh(left.(IntegerResult).Int, uint(right.(IntegerResult).Int.Uint64()))}, nil
		}

		return nil, LangError{
			ErrorType,
			node.Position(),
			fmt.Sprintf("binary operation %s is not defined on types %s and %s",
				node.Op.Value, left.Kind(), right.Kind()),
		}
	case TokenBitwiseRight:
		if left.Kind() == ResultInt && right.Kind() == ResultInt {
			return IntegerResult{new(big.Int).Rsh(left.(IntegerResult).Int, uint(right.(IntegerResult).Int.Uint64()))}, nil
		}

		return nil, LangError{
			ErrorType,
			node.Position(),
			fmt.Sprintf("binary operation %s is not defined on types %s and %s",
				node.Op.Value, left.Kind(), right.Kind()),
		}
	case TokenBitwiseAnd:
		if left.Kind() == ResultInt && right.Kind() == ResultInt {
			return IntegerResult{new(big.Int).And(left.(IntegerResult).Int, right.(IntegerResult).Int)}, nil
		}

		return nil, LangError{
			ErrorType,
			node.Position(),
			fmt.Sprintf("binary operation %s is not defined on types %s and %s",
				node.Op.Value, left.Kind(), right.Kind()),
		}
	case TokenBitwiseOr:
		if left.Kind() == ResultInt && right.Kind() == ResultInt {
			return IntegerResult{new(big.Int).Or(left.(IntegerResult).Int, right.(IntegerResult).Int)}, nil
		}

		return nil, LangError{
			ErrorType,
			node.Position(),
			fmt.Sprintf("binary operation %s is not defined on types %s and %s",
				node.Op.Value, left.Kind(), right.Kind()),
		}
	case TokenBitwiseXor:
		if left.Kind() == ResultInt && right.Kind() == ResultInt {
			return IntegerResult{new(big.Int).Xor(left.(IntegerResult).Int, right.(IntegerResult).Int)}, nil
		}

		return nil, LangError{
			ErrorType,
			node.Position(),
			fmt.Sprintf("binary operation %s is not defined on types %s and %s",
				node.Op.Value, left.Kind(), right.Kind()),
		}
	case TokenEquals:
		return doBinOpEquals(left, right), nil
	case TokenNotEq:
		return !doBinOpEquals(left, right), nil
	case TokenLt:
		result, err := doBinOpLt(left, right)
		if err != nil && errors.Is(err, ErrCannotCompare) {
			return nil, LangError{
				ErrorType,
				node.Position(),
				fmt.Sprintf("binary operation %s is not defined on types %s and %s",
					node.Op.Value, left.Kind(), right.Kind()),
			}
		}
		return result, nil
	case TokenLtEq:
		result, err := doBinOpLt(left, right)
		if err != nil && errors.Is(err, ErrCannotCompare) {
			return nil, LangError{
				ErrorType,
				node.Position(),
				fmt.Sprintf("binary operation %s is not defined on types %s and %s",
					node.Op.Value, left.Kind(), right.Kind()),
			}
		}

		return result || doBinOpEquals(left, right), nil
	case TokenGt:
		result, err := doBinOpGt(left, right)
		if err != nil && errors.Is(err, ErrCannotCompare) {
			return nil, LangError{
				ErrorType,
				node.Position(),
				fmt.Sprintf("binary operation %s is not defined on types %s and %s",
					node.Op.Value, left.Kind(), right.Kind()),
			}
		}
		return result, nil
	case TokenGtEq:
		result, err := doBinOpGt(left, right)
		if err != nil && errors.Is(err, ErrCannotCompare) {
			return nil, LangError{
				ErrorType,
				node.Position(),
				fmt.Sprintf("binary operation %s is not defined on types %s and %s",
					node.Op.Value, left.Kind(), right.Kind()),
			}
		}
		return result || doBinOpEquals(left, right), nil
	case TokenLogicalOr:
		leftBool, err := ResultAsBoolean(left)
		if err != nil && errors.Is(err, ErrCannotBoolean) {
			return nil, LangError{
				ErrorRuntime,
				node.Position(),
				fmt.Sprintf("left operand %s cannot be converted to be a boolean", left.Kind()),
			}
		}

		if leftBool {
			return BooleanResult(true), nil
		}

		rightBool, err := ResultAsBoolean(right)
		if err != nil && errors.Is(err, ErrCannotBoolean) {
			return nil, LangError{
				ErrorRuntime,
				node.Position(),
				fmt.Sprintf("right operand %s cannot be converted to be a boolean", right.Kind()),
			}
		}

		return leftBool || rightBool, nil
	case TokenLogicalAnd:
		leftBool, err := ResultAsBoolean(left)
		if err != nil && errors.Is(err, ErrCannotBoolean) {
			return nil, LangError{
				ErrorRuntime,
				node.Position(),
				fmt.Sprintf("left operand %s cannot be converted to be a boolean", left.Kind()),
			}
		}

		if !leftBool {
			return BooleanResult(false), nil
		}

		rightBool, err := ResultAsBoolean(right)
		if err != nil && errors.Is(err, ErrCannotBoolean) {
			return nil, LangError{
				ErrorRuntime,
				node.Position(),
				fmt.Sprintf("right operand %s cannot be converted to be a boolean", right.Kind()),
			}
		}

		return leftBool && rightBool, nil
	default:
		return nil, LangError{
			ErrorRuntime,
			node.Position(),
			fmt.Sprintf("behavior undefined for binary operation %s", node.Op.Value),
		}
	}
}

func EvaluateUnaryOp(node UnaryOpNode, namespace Namespace) (Result, error) {
	switch node.Op.Kind {
	case TokenPlus:
		result, err := Evaluate(node.Node, namespace)
		if err != nil {
			return nil, err
		}

		switch result.Kind() {
		case ResultInt:
			return result.(IntegerResult), nil
		case ResultFloat:
			return result.(FloatResult), nil
		default:
			return nil, LangError{
				ErrorType,
				node.Position(),
				fmt.Sprintf("%s does not support unary operation %s", result.Kind(), node.Op.Value),
			}
		}
	case TokenMinus:
		result, err := Evaluate(node.Node, namespace)
		if err != nil {
			return nil, err
		}

		switch result.Kind() {
		case ResultInt:
			return IntegerResult{new(big.Int).Neg(result.(IntegerResult).Int)}, nil
		case ResultFloat:
			return -result.(FloatResult), nil
		default:
			return nil, LangError{
				ErrorType,
				node.Position(),
				fmt.Sprintf("%s does not support unary operation %s", result.Kind(), node.Op.Value),
			}
		}
	case TokenBitwiseNot:
		result, err := Evaluate(node.Node, namespace)
		if err != nil {
			return nil, err
		}

		switch result.Kind() {
		case ResultInt:
			return IntegerResult{new(big.Int).Not(result.(IntegerResult).Int)}, nil
		default:
			return nil, LangError{
				ErrorType,
				node.Position(),
				fmt.Sprintf("%s does not support unary operation %s", result.Kind(), node.Op.Value),
			}
		}
	case TokenNot:
		result, err := Evaluate(node.Node, namespace)
		if err != nil {
			return nil, err
		}

		asBool, err := ResultAsBoolean(result)
		if err != nil && errors.Is(err, ErrCannotBoolean) {
			return nil, LangError{
				ErrorRuntime,
				node.Position(),
				fmt.Sprintf("%s cannot be converted to be a boolean", result.Kind()),
			}
		}

		return !asBool, nil
	default:
		return nil, LangError{
			ErrorRuntime,
			node.Position(),
			fmt.Sprintf("undefined binary operation %s", node.Op.Value),
		}
	}
}

func EvaluateLiteral(node LiteralNode, ns Namespace) (Result, error) {
	switch node.Token.Kind {
	case TokenInteger:
		number := new(big.Int)
		if _, ok := number.SetString(node.Token.Value, 0); !ok {
			return nil, LangError{ErrorSyntax, node.Position(), "invalid integer literal"}
		}

		return IntegerResult{number}, nil
	case TokenFloat:
		number, err := strconv.ParseFloat(node.Token.Value, 64)
		if err != nil {
			return nil, LangError{
				ErrorSyntax,
				node.Position(),
				fmt.Sprintf("invalid float literal: %s", err),
			}
		}

		return FloatResult(number), nil
	case TokenIdentifier:
		if tp := TypeName(node.Token.Value); slices.Contains(AvailableTypeNames, tp) {
			return TypeResult{Name: tp, Params: []Result{}}, nil
		}

		ident := IdentResult(node.Token.Value)

		if ns != nil {
			value, ok := ns[ident]
			if !ok {
				return nil, LangError{
					ErrorAccess,
					node.Position(),
					fmt.Sprintf("%q is not defined", node.Token.Value),
				}
			}
			return value, nil
		}

		return ident, nil
	case TokenKeyword:
		switch val := node.Token.Value; val {
		case string(KeywordTrue):
			return BooleanResult(true), nil
		case string(KeywordFalse):
			return BooleanResult(false), nil
		default:
			return nil, LangError{
				ErrorSyntax,
				node.Position(),
				fmt.Sprintf("unknown keyword %q", val),
			}
		}
	case TokenString:
		return StringResult(node.Token.Value), nil
	default:
		return nil, LangError{
			ErrorRuntime,
			node.Position(),
			fmt.Sprintf("evaluation undefined for literal type %s", node.Token.Kind),
		}
	}
}

func EvaluateMap(node MapNode, namespace Namespace) (Result, error) {
	items := map[Result]Result{}

	for key, val := range node.Items {
		keyRes, err := Evaluate(key, namespace)
		if err != nil {
			return nil, err
		}

		lazy, err := MustEvaluateLazily(val)
		if err != nil {
			return nil, err
		}

		var valueRes Result
		if lazy {
			valueRes = LazyResult(
				func(ns Namespace) (Result, error) { return Evaluate(val, ns) },
			)
		} else {
			valueRes, err = Evaluate(val, namespace)
			if err != nil {
				return nil, err
			}
		}

		items[keyRes] = valueRes
	}

	return MapResult(items), nil
}

func EvaluateList(node ListNode, namespace Namespace) (Result, error) {
	items := []Result{}

	for _, val := range node.Items {
		lazy, err := MustEvaluateLazily(val)
		if err != nil {
			return nil, err
		}

		var valRes Result
		if lazy {
			valRes = LazyResult(
				func(ns Namespace) (Result, error) { return Evaluate(val, ns) },
			)
		} else {
			valRes, err = Evaluate(val, namespace)
			if err != nil {
				return nil, err
			}
		}

		items = append(items, valRes)
	}

	return ListResult(items), nil
}

func EvaluateSubscript(node SubscriptNode, namespace Namespace) (Result, error) {
	expr, err := Evaluate(node.Expr, namespace)
	if err != nil {
		return nil, err
	}

	item, err := Evaluate(node.Item, namespace)
	if err != nil {
		return nil, err
	}

	var (
		value Result
		ok    bool
	)

	switch expr.Kind() {
	case ResultIdent:
		value, ok = namespace[item]
	case ResultMap:
		value, ok = expr.(MapResult)[item]
	case ResultType:
		typeRes := expr.(TypeResult)

		if typeRes.Name == TypeByte {
			return TypeResult{Name: TypeByte, Params: []Result{item}}, nil
		}

		return nil, LangError{
			ErrorType,
			node.Position(),
			fmt.Sprintf("type %s does not allow type parameters", typeRes.Name),
		}
	case ResultList:
		valueRes := expr.(ListResult)

		if item.Kind() != ResultInt {
			return nil, LangError{
				ErrorType,
				node.Position(),
				fmt.Sprintf("list indices must be %s, not %s", ResultInt, item.Kind()),
			}
		}

		intVal := item.(IntegerResult).Int64()
		if intVal >= int64(len(valueRes)) {
			return nil, LangError{ErrorAccess, node.Position(), "index out of bounds"}
		}

		value = valueRes[intVal]
		ok = true
	default:
		return nil, LangError{
			ErrorType,
			node.Position(),
			fmt.Sprintf("object of type %s does not support subscript access", expr.Kind()),
		}
	}

	if !ok {
		return nil, LangError{
			ErrorAccess,
			node.Position(),
			fmt.Sprintf("object of type %s does not have a member or key named %v", expr.Kind(), item),
		}
	}

	return value, nil
}

func EvaluateAttr(node AttrNode, namespace Namespace) (Result, error) {
	expr, err := Evaluate(node.Expr, namespace)
	if err != nil {
		return nil, err
	}

	attr, err := Evaluate(node.Attr, nil)
	if err != nil {
		return nil, err
	}

	var (
		value Result
		ok    bool
	)

	switch expr.Kind() {
	case ResultIdent:
		value, ok = namespace[attr]
	case ResultMap:
		value, ok = expr.(MapResult)[attr]
	default:
		return nil, LangError{
			ErrorType,
			node.Position(),
			fmt.Sprintf("object of type %s does not support attribute access", expr.Kind()),
		}
	}

	if !ok {
		return nil, LangError{
			ErrorAccess,
			node.Position(),
			fmt.Sprintf("object of type %s does not have a member named %v", expr.Kind(), attr),
		}
	}

	return value, nil
}

// Evaluate evaluates the result of node tree using namespace and returns the result
// and an error if any occurred.
func Evaluate(tree Node, namespace Namespace) (Result, error) {
	switch tree.Type() {
	case NodeBinOp:
		return EvaluateBinOp(*tree.(*BinOpNode), namespace)
	case NodeUnaryOp:
		return EvaluateUnaryOp(*tree.(*UnaryOpNode), namespace)
	case NodeLiteral:
		return EvaluateLiteral(*tree.(*LiteralNode), namespace)
	case NodeMap:
		return EvaluateMap(*tree.(*MapNode), namespace)
	case NodeAttr:
		return EvaluateAttr(*tree.(*AttrNode), namespace)
	case NodeList:
		return EvaluateList(*tree.(*ListNode), namespace)
	case NodeSubscript:
		return EvaluateSubscript(*tree.(*SubscriptNode), namespace)
	default:
		return nil, LangError{
			ErrorRuntime,
			tree.Position(),
			fmt.Sprintf("evaluation undefined for type %s", tree.Type()),
		}
	}
}

// ResultAsBoolean reports the result as a boolean.
func ResultAsBoolean(result Result) (BooleanResult, error) {
	switch result.Kind() {
	case ResultInt:
		isZero := result.(IntegerResult).Cmp(new(big.Int))
		return BooleanResult(isZero != 0), nil
	case ResultFloat:
		return BooleanResult(result.(FloatResult) != 0.0), nil
	case ResultBoolean:
		return result.(BooleanResult), nil
	case ResultMap:
		return BooleanResult(len(result.(MapResult)) > 0), nil
	case ResultList:
		return BooleanResult(len(result.(ListResult)) > 0), nil
	case ResultString:
		return BooleanResult(len(result.(StringResult)) > 0), nil
	default:
		return BooleanResult(false), ErrCannotBoolean
	}
}

// MustEvaluateLazily reports whether the provided node must be evaluated lazily,
// that is, whether the node must be evaluated on access rather than on parse.
func MustEvaluateLazily(node Node) (bool, error) {
	switch node.Type() {
	case NodeBinOp:
		binOp := node.(*BinOpNode)
		leftLazy, err := MustEvaluateLazily(binOp.Left)
		if err != nil {
			return false, err
		}

		rightLazy, err := MustEvaluateLazily(binOp.Right)
		if err != nil {
			return false, err
		}

		return leftLazy || rightLazy, nil
	case NodeUnaryOp:
		unary := node.(*UnaryOpNode)
		lazy, err := MustEvaluateLazily(unary.Node)
		if err != nil {
			return false, err
		}

		return lazy, nil
	case NodeAttr:
		// attr requires a namespace
		return true, nil
	case NodeSubscript:
		subscript := node.(*SubscriptNode)
		exprLazy, err := MustEvaluateLazily(subscript.Expr)
		if err != nil {
			return false, err
		}

		itemLazy, err := MustEvaluateLazily(subscript.Item)
		if err != nil {
			return false, err
		}

		return exprLazy || itemLazy, nil
	case NodeLiteral:
		lit := node.(*LiteralNode)
		switch lit.Token.Kind {
		case TokenIdentifier:
			if tp := TypeName(lit.Token.Value); slices.Contains(AvailableTypeNames, tp) {
				// type identifiers can be evaluated immediately.
				return false, nil
			}
			return true, nil
		default:
			return false, nil
		}
	case NodeMap, NodeList:
		// map and list may contain lazily evaluated nodes but, in general,
		// as they're literals, they can be evaluated immediately.
		return false, nil
	case NodeCall:
		call := node.(*CallNode)

		exprLazy, err := MustEvaluateLazily(call.Expr)
		if err != nil {
			return false, err
		}
		if exprLazy {
			return true, nil
		}

		for _, arg := range call.Arguments {
			argLazy, err := MustEvaluateLazily(arg)
			if err != nil {
				return false, err
			}

			if argLazy {
				return true, nil
			}
		}

		return false, nil
	default:
		return false, LangError{
			ErrorRuntime,
			node.Position(),
			fmt.Sprintf("evaluation undefined for type %s", node.Type()),
		}
	}
}
