package bindef

import (
	"fmt"
	"math"
	"strconv"
)

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
)

type Result struct {
	Kind  ResultKind
	Value any
}

type Namespace map[Result]Result

func EvaluateBinOp(node BinOpNode, namespace Namespace) (Result, error) {
	left, err := Evaluate(node.Left, namespace)
	if err != nil {
		return Result{}, err
	}

	right, err := Evaluate(node.Right, namespace)
	if err != nil {
		return Result{}, err
	}

	switch node.Op.Kind {
	case TokenPlus:
		if left.Kind == ResultInt && right.Kind == ResultInt {
			return Result{Kind: ResultInt, Value: left.Value.(int) + right.Value.(int)}, nil
		} else if left.Kind == ResultFloat && right.Kind == ResultFloat {
			return Result{Kind: ResultFloat, Value: left.Value.(float64) + right.Value.(float64)}, nil
		} else if left.Kind == ResultFloat && right.Kind == ResultInt {
			rightFloat := float64(right.Value.(int))
			return Result{Kind: ResultFloat, Value: left.Value.(float64) + rightFloat}, nil
		} else if left.Kind == ResultInt && right.Kind == ResultFloat {
			leftFloat := float64(left.Value.(int))
			return Result{Kind: ResultFloat, Value: leftFloat + right.Value.(float64)}, nil
		}

		return Result{}, LangError{
			node.Position(),
			fmt.Sprintf("binary operation %s is not defined on types %s and %s",
				node.Op.Value, left.Kind, right.Kind),
		}
	case TokenMinus:
		if left.Kind == ResultInt && right.Kind == ResultInt {
			return Result{Kind: ResultInt, Value: left.Value.(int) - right.Value.(int)}, nil
		} else if left.Kind == ResultFloat && right.Kind == ResultFloat {
			return Result{Kind: ResultFloat, Value: left.Value.(float64) - right.Value.(float64)}, nil
		} else if left.Kind == ResultFloat && right.Kind == ResultInt {
			rightFloat := float64(right.Value.(int))
			return Result{Kind: ResultFloat, Value: left.Value.(float64) - rightFloat}, nil
		} else if left.Kind == ResultInt && right.Kind == ResultFloat {
			leftFloat := float64(left.Value.(int))
			return Result{Kind: ResultFloat, Value: leftFloat - right.Value.(float64)}, nil
		}

		return Result{}, LangError{
			node.Position(),
			fmt.Sprintf("binary operation %s is not defined on types %s and %s",
				node.Op.Value, left.Kind, right.Kind),
		}
	case TokenMul:
		if left.Kind == ResultInt && right.Kind == ResultInt {
			return Result{Kind: ResultInt, Value: left.Value.(int) * right.Value.(int)}, nil
		} else if left.Kind == ResultFloat && right.Kind == ResultFloat {
			return Result{Kind: ResultFloat, Value: left.Value.(float64) * right.Value.(float64)}, nil
		} else if left.Kind == ResultFloat && right.Kind == ResultInt {
			rightFloat := float64(right.Value.(int))
			return Result{Kind: ResultFloat, Value: left.Value.(float64) * rightFloat}, nil
		} else if left.Kind == ResultInt && right.Kind == ResultFloat {
			leftFloat := float64(left.Value.(int))
			return Result{Kind: ResultFloat, Value: leftFloat * right.Value.(float64)}, nil
		}

		return Result{}, LangError{
			node.Position(),
			fmt.Sprintf("binary operation %s is not defined on types %s and %s",
				node.Op.Value, left.Kind, right.Kind),
		}
	case TokenDiv:
		if left.Kind == ResultInt && right.Kind == ResultInt {
			if right.Value.(int) == 0 {
				return Result{}, LangError{node.Position(), "integer division by zero"}
			}

			return Result{Kind: ResultInt, Value: left.Value.(int) / right.Value.(int)}, nil
		} else if left.Kind == ResultFloat && right.Kind == ResultFloat {
			if right.Value.(float64) == 0 {
				return Result{}, LangError{node.Position(), "float division by zero"}
			}

			return Result{Kind: ResultFloat, Value: left.Value.(float64) / right.Value.(float64)}, nil
		} else if left.Kind == ResultFloat && right.Kind == ResultInt {
			rightFloat := float64(right.Value.(int))
			if rightFloat == 0.0 {
				return Result{}, LangError{node.Position(), "float division by zero"}
			}

			return Result{Kind: ResultFloat, Value: left.Value.(float64) / rightFloat}, nil
		} else if left.Kind == ResultInt && right.Kind == ResultFloat {
			leftFloat := float64(left.Value.(int))
			if right.Value.(float64) == 0.0 {
				return Result{}, LangError{node.Position(), "float division by zero"}
			}

			return Result{Kind: ResultFloat, Value: leftFloat / right.Value.(float64)}, nil
		}

		return Result{}, LangError{
			node.Position(),
			fmt.Sprintf("binary operation %s is not defined on types %s and %s",
				node.Op.Value, left.Kind, right.Kind),
		}
	case TokenModulo:
		if left.Kind == ResultInt && right.Kind == ResultInt {
			if right.Value.(int) == 0 {
				return Result{}, LangError{node.Position(), "integer modulo by zero"}
			}

			return Result{Kind: ResultInt, Value: left.Value.(int) % right.Value.(int)}, nil
		} else if left.Kind == ResultFloat && right.Kind == ResultFloat {
			if right.Value.(float64) == 0 {
				return Result{}, LangError{node.Position(), "float modulo by zero"}
			}

			mod := math.Mod(left.Value.(float64), right.Value.(float64))
			return Result{Kind: ResultFloat, Value: mod}, nil
		} else if left.Kind == ResultFloat && right.Kind == ResultInt {
			rightFloat := float64(right.Value.(int))
			if rightFloat == 0.0 {
				return Result{}, LangError{node.Position(), "float modulo by zero"}
			}

			mod := math.Mod(left.Value.(float64), rightFloat)
			return Result{Kind: ResultFloat, Value: mod}, nil
		} else if left.Kind == ResultInt && right.Kind == ResultFloat {
			leftFloat := float64(left.Value.(int))
			if right.Value.(float64) == 0.0 {
				return Result{}, LangError{node.Position(), "float modulo by zero"}
			}

			mod := math.Mod(leftFloat, right.Value.(float64))
			return Result{Kind: ResultFloat, Value: mod}, nil
		}

		return Result{}, LangError{
			node.Position(),
			fmt.Sprintf("binary operation %s is not defined on types %s and %s",
				node.Op.Value, left.Kind, right.Kind),
		}
	default:
		return Result{}, LangError{
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
			return Result{}, err
		}

		switch result.Kind {
		case ResultInt:
			return Result{Kind: ResultInt, Value: result.Value.(int)}, nil
		case ResultFloat:
			return Result{Kind: ResultFloat, Value: result.Value.(float64)}, nil
		default:
			return Result{}, LangError{
				node.Position(),
				fmt.Sprintf("%s does not support unary operation %s", result.Kind, node.Op.Value),
			}
		}
	case TokenMinus:
		result, err := Evaluate(node.Node, namespace)
		if err != nil {
			return Result{}, err
		}

		switch result.Kind {
		case ResultInt:
			return Result{Kind: ResultInt, Value: -result.Value.(int)}, nil
		case ResultFloat:
			return Result{Kind: ResultFloat, Value: -result.Value.(float64)}, nil
		default:
			return Result{}, LangError{
				node.Position(),
				fmt.Sprintf("%s does not support unary operation %s", result.Kind, node.Op.Value),
			}
		}
	case TokenBitwiseNot:
		result, err := Evaluate(node.Node, namespace)
		if err != nil {
			return Result{}, err
		}

		switch result.Kind {
		case ResultInt:
			return Result{Kind: ResultInt, Value: ^result.Value.(int)}, nil
		default:
			return Result{}, LangError{
				node.Position(),
				fmt.Sprintf("%s does not support unary operation %s", result.Kind, node.Op.Value),
			}
		}
	default:
		return Result{}, LangError{
			node.Position(),
			fmt.Sprintf("undefined binary operation %s", node.Op.Value),
		}
	}
}

func EvaluateLiteral(node LiteralNode, ns Namespace) (Result, error) {
	switch node.Token.Kind {
	case TokenInteger:
		number, err := strconv.Atoi(node.Token.Value)
		if err != nil {
			return Result{}, LangError{
				node.Position(),
				fmt.Sprintf("invalid integer literal: %s", err),
			}
		}
		return Result{Kind: ResultInt, Value: number}, nil
	case TokenFloat:
		number, err := strconv.ParseFloat(node.Token.Value, 64)
		if err != nil {
			return Result{}, LangError{
				node.Position(),
				fmt.Sprintf("invalid float literal: %s", err),
			}
		}
		return Result{Kind: ResultFloat, Value: number}, nil
	case TokenIdentifier:
		ident := Result{Kind: ResultIdent, Value: node.Token.Value}

		if ns != nil {
			value, ok := ns[ident]
			if !ok {
				return Result{}, LangError{
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
			return Result{Kind: ResultBoolean, Value: true}, nil
		case string(KeywordFalse):
			return Result{Kind: ResultBoolean, Value: false}, nil
		default:
			return Result{}, LangError{
				node.Position(),
				fmt.Sprintf("unknown keyword %q", val),
			}
		}
	case TokenString:
		return Result{Kind: ResultString, Value: node.Token.Value}, nil
	default:
		return Result{}, LangError{
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
			return Result{}, err
		}

		lazy, err := MustEvaluateLazily(val)
		if err != nil {
			return Result{}, err
		}

		var valueRes Result
		if lazy {
			valueRes = Result{
				Kind:  ResultLazy,
				Value: func(ns Namespace) (Result, error) { return Evaluate(val, ns) },
			}
		} else {
			valueRes, err = Evaluate(val, namespace)
			if err != nil {
				return Result{}, err
			}
		}

		items[keyRes] = valueRes
	}

	return Result{Kind: ResultMap, Value: items}, nil
}

func EvaluateList(node ListNode, namespace Namespace) (Result, error) {
	items := []Result{}

	for _, val := range node.Items {
		lazy, err := MustEvaluateLazily(val)
		if err != nil {
			return Result{}, err
		}

		var valRes Result
		if lazy {
			valRes = Result{
				Kind:  ResultLazy,
				Value: func(ns Namespace) (Result, error) { return Evaluate(val, ns) },
			}
		} else {
			valRes, err = Evaluate(val, namespace)
			if err != nil {
				return Result{}, err
			}
		}

		items = append(items, valRes)
	}

	return Result{Kind: ResultList, Value: items}, nil
}

func EvaluateAttr(node AttrNode, namespace Namespace) (Result, error) {
	expr, err := Evaluate(node.Expr, namespace)
	if err != nil {
		return Result{}, err
	}

	attr, err := Evaluate(node.Attr, nil)
	if err != nil {
		return Result{}, err
	}

	var (
		value Result
		ok    bool
	)

	switch expr.Kind {
	case ResultIdent:
		value, ok = namespace[attr]
	case ResultMap:
		value, ok = expr.Value.(map[Result]Result)[attr]
	default:
		return Result{}, LangError{
			node.Position(),
			fmt.Sprintf("object of type %s does not support attribute access", expr.Kind),
		}
	}

	if !ok {
		return Result{}, LangError{
			node.Position(),
			fmt.Sprintf("object of type %s does not have a member named %q", expr.Kind, attr.Value),
		}
	}

	return value, nil
}

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
	default:
		return Result{}, LangError{
			tree.Position(),
			fmt.Sprintf("evaluation undefined for type %s", tree.Type()),
		}
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
		return false, nil
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
			node.Position(),
			fmt.Sprintf("evaluation undefined for type %s", node.Type()),
		}
	}
}
