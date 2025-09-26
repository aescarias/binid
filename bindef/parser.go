package bindef

import (
	"fmt"
)

type Node interface {
	Type() string
}

type UnaryOpNode struct {
	Op   Token
	Node Node
}

type BinOpNode struct {
	Left  Node
	Op    Token
	Right Node
}

type LiteralNode struct {
	Token Token
}

func (ln *LiteralNode) Type() string { return "Literal" }
func (bn *BinOpNode) Type() string   { return "BinOp" }
func (un *UnaryOpNode) Type() string { return "UnaryOp" }

type Parser struct {
	Scanner[Token]
}

func (ps *Parser) ParseLiteral() (Node, error) {
	switch ps.Cursor().Kind {
	case TokenInteger, TokenFloat, TokenIdentifier:
		lit := &LiteralNode{Token: ps.Cursor()}
		ps.Advance(1)
		return lit, nil
	case TokenPlus, TokenMinus, TokenBitwiseNot, TokenNot:
		tok := ps.Cursor()
		ps.Advance(1)
		expr, err := ps.ParseExpr()
		if err != nil {
			return nil, err
		}
		return &UnaryOpNode{Op: tok, Node: expr}, nil
	case TokenLParen:
		ps.Advance(1)
		expr, err := ps.ParseExpr()
		if err != nil {
			return nil, err
		}

		if ps.IsDone() || ps.Cursor().Kind != TokenRParen {
			return nil, fmt.Errorf("expected closing parenthesis")
		}
		ps.Advance(1)

		return expr, nil
	}

	return nil, fmt.Errorf("unknown literal type: %s", ps.Cursor().Kind)
}

func (ps *Parser) ParseFactor() (Node, error) {
	var (
		left Node
		err  error
	)

	if left, err = ps.ParseLiteral(); err != nil {
		return nil, err
	}

	for !ps.IsDone() && (ps.Cursor().Kind == TokenMul || ps.Cursor().Kind == TokenDiv) {
		tok := ps.Cursor()
		ps.Advance(1)

		right, err := ps.ParseLiteral()
		if err != nil {
			return nil, err
		}

		left = &BinOpNode{Left: left, Op: tok, Right: right}
	}

	return left, nil

}

func (ps *Parser) ParseExpr() (Node, error) {
	var (
		left Node
		err  error
	)

	if left, err = ps.ParseFactor(); err != nil {
		return nil, err
	}

	for !ps.IsDone() && (ps.Cursor().Kind == TokenPlus || ps.Cursor().Kind == TokenMinus) {
		tok := ps.Cursor()
		ps.Advance(1)

		right, err := ps.ParseFactor()
		if err != nil {
			return nil, err
		}

		left = &BinOpNode{Left: left, Op: tok, Right: right}
	}

	return left, nil
}
