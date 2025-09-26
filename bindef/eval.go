package bindef

import (
	"fmt"
	"strconv"
)

type ResultKind string

const (
	ResultInt   ResultKind = "Int"
	ResultFloat ResultKind = "Float"
)

type Result struct {
	Kind  ResultKind
	Value any
}

func EvaluateBinOp(node BinOpNode) (Result, error) {
	left, err := Evaluate(node.Left)
	if err != nil {
		return Result{}, err
	}

	right, err := Evaluate(node.Right)
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

		return Result{}, fmt.Errorf("cannot perform binary operation %s on types %s and %s", node.Op.Value, left.Kind, right.Kind)
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

		return Result{}, fmt.Errorf("cannot perform binary operation %s on types %s and %s", node.Op.Value, left.Kind, right.Kind)
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

		return Result{}, fmt.Errorf("cannot perform binary operation %s on types %s and %s", node.Op.Value, left.Kind, right.Kind)
	case TokenDiv:
		if left.Kind == ResultInt && right.Kind == ResultInt {
			if right.Value.(int) == 0 {
				return Result{}, fmt.Errorf("integer division by zero")
			}

			return Result{Kind: ResultInt, Value: left.Value.(int) / right.Value.(int)}, nil
		} else if left.Kind == ResultFloat && right.Kind == ResultFloat {
			if right.Value.(float64) == 0 {
				return Result{}, fmt.Errorf("float division by zero")
			}

			return Result{Kind: ResultFloat, Value: left.Value.(float64) / right.Value.(float64)}, nil
		} else if left.Kind == ResultFloat && right.Kind == ResultInt {
			rightFloat := float64(right.Value.(int))
			if rightFloat == 0.0 {
				return Result{}, fmt.Errorf("float division by zero")
			}
			return Result{Kind: ResultFloat, Value: left.Value.(float64) / rightFloat}, nil
		} else if left.Kind == ResultInt && right.Kind == ResultFloat {
			leftFloat := float64(left.Value.(int))
			if right.Value.(float64) == 0.0 {
				return Result{}, fmt.Errorf("float division by zero")
			}
			return Result{Kind: ResultFloat, Value: leftFloat / right.Value.(float64)}, nil
		}

		return Result{}, fmt.Errorf("cannot perform binary operation %s on types %s and %s", node.Op.Value, left.Kind, right.Kind)
	default:
		return Result{}, fmt.Errorf("unknown binary operation %s", node.Op.Value)
	}
}

func EvaluateUnaryOp(node UnaryOpNode) (Result, error) {
	switch node.Op.Kind {
	case TokenPlus:
		result, err := Evaluate(node.Node)
		if err != nil {
			return Result{}, err
		}

		switch result.Kind {
		case ResultInt:
			return Result{Kind: ResultInt, Value: result.Value.(int)}, nil
		case ResultFloat:
			return Result{Kind: ResultFloat, Value: result.Value.(float64)}, nil
		default:
			return Result{}, fmt.Errorf("%s does not support unary op %s", result.Kind, node.Op.Value)
		}
	case TokenMinus:
		result, err := Evaluate(node.Node)
		if err != nil {
			return Result{}, err
		}

		switch result.Kind {
		case ResultInt:
			return Result{Kind: ResultInt, Value: -result.Value.(int)}, nil
		case ResultFloat:
			return Result{Kind: ResultFloat, Value: -result.Value.(float64)}, nil
		default:
			return Result{}, fmt.Errorf("%s does not support unary op %s", result.Kind, node.Op.Value)
		}
	case TokenBitwiseNot:
		result, err := Evaluate(node.Node)
		if err != nil {
			return Result{}, err
		}

		switch result.Kind {
		case ResultInt:
			return Result{Kind: ResultInt, Value: ^result.Value.(int)}, nil
		default:
			return Result{}, fmt.Errorf("%s does not support unary op %s", result.Kind, node.Op.Value)
		}
	default:
		return Result{}, fmt.Errorf("unknown unary op %s", node.Op.Value)
	}
}

func EvaluateLiteral(node LiteralNode) (Result, error) {
	switch node.Token.Kind {
	case TokenInteger:
		number, err := strconv.Atoi(node.Token.Value)
		if err != nil {
			return Result{}, fmt.Errorf("failed to convert integer: %w", err)
		}
		return Result{Kind: ResultInt, Value: number}, nil
	case TokenFloat:
		number, err := strconv.ParseFloat(node.Token.Value, 64)
		if err != nil {
			return Result{}, fmt.Errorf("failed to convert float: %w", err)
		}
		return Result{Kind: ResultFloat, Value: number}, nil
	default:
		return Result{}, fmt.Errorf("unknown literal type %s", node.Token.Kind)
	}
}

func Evaluate(tree Node) (Result, error) {
	switch tree.Type() {
	case "BinOp":
		return EvaluateBinOp(*tree.(*BinOpNode))
	case "UnaryOp":
		return EvaluateUnaryOp(*tree.(*UnaryOpNode))
	case "Literal":
		return EvaluateLiteral(*tree.(*LiteralNode))
	default:
		return Result{}, fmt.Errorf("cannot evaluate unknown type %s", tree.Type())
	}
}
