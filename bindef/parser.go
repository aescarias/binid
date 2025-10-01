package bindef

import (
	"fmt"
	"slices"
)

type Node interface {
	Type() NodeKind
	Position() Position
}

type LangError struct {
	Position Position
	Message  string
}

func (le LangError) Error() string {
	return le.Message
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

type MapNode struct {
	Items map[Node]Node
	pos   Position
}

type ListNode struct {
	Items []Node
	pos   Position
}

type AttrNode struct {
	Expr Node
	Attr Node
}

type SubscriptNode struct {
	Expr Node
	Item Node
}

type CallNode struct {
	Expr      Node
	Arguments []Node
	pos       Position
}

type NodeKind string

const (
	NodeLiteral   NodeKind = "Literal"
	NodeBinOp     NodeKind = "BinOp"
	NodeUnaryOp   NodeKind = "UnaryOp"
	NodeMap       NodeKind = "Map"
	NodeList      NodeKind = "List"
	NodeAttr      NodeKind = "Attr"
	NodeSubscript NodeKind = "Subscript"
	NodeCall      NodeKind = "Call"
)

func (ln *LiteralNode) Type() NodeKind { return NodeLiteral }
func (ln *LiteralNode) Position() Position {
	return ln.Token.Position
}

func (bn *BinOpNode) Type() NodeKind { return NodeBinOp }
func (bn *BinOpNode) Position() Position {
	return Position{Start: bn.Left.Position().Start, End: bn.Right.Position().End}
}

func (un *UnaryOpNode) Type() NodeKind { return NodeUnaryOp }
func (un *UnaryOpNode) Position() Position {
	return Position{Start: un.Op.Position.Start, End: un.Node.Position().End}
}

func (mn *MapNode) Type() NodeKind     { return NodeMap }
func (mn *MapNode) Position() Position { return mn.pos }

func (ln *ListNode) Type() NodeKind     { return NodeList }
func (ln *ListNode) Position() Position { return ln.pos }

func (an *AttrNode) Type() NodeKind { return NodeAttr }
func (an *AttrNode) Position() Position {
	return Position{
		Start: an.Expr.Position().Start,
		End:   an.Attr.Position().End,
	}
}

func (in *SubscriptNode) Type() NodeKind { return NodeSubscript }
func (in *SubscriptNode) Position() Position {
	return Position{
		Start: in.Expr.Position().Start,
		End:   in.Item.Position().End,
	}
}

func (cn *CallNode) Type() NodeKind     { return NodeCall }
func (cn *CallNode) Position() Position { return cn.pos }

type Parser struct {
	Scanner[Token]
}

func NewParser(tokens []Token) Parser {
	return Parser{Scanner: NewScanner(tokens)}
}

func (ps *Parser) tryPostfix(left Node) (Node, error) {
	for !ps.IsDone() {
		switch ps.Cursor().Kind {
		case TokenLBracket:
			ps.Advance(1)
			item, err := ps.ParseExpr()
			if err != nil {
				return nil, err
			}
			if ps.IsDone() || ps.Cursor().Kind != TokenRBracket {
				return nil, LangError{item.Position(), "expected closing bracket for subscript access"}
			}
			ps.Advance(1)
			left = &SubscriptNode{Expr: left, Item: item}
		case TokenLParen:
			start := left.Position().Start
			ps.Advance(1)
			arguments := []Node{}
			for !ps.IsDone() && ps.Cursor().Kind != TokenRParen {
				arg, err := ps.ParseExpr()
				if err != nil {
					return nil, err
				}
				arguments = append(arguments, arg)
				if ps.Cursor().Kind == TokenComma {
					ps.Advance(1)
				} else if ps.IsDone() || ps.Cursor().Kind != TokenRParen {
					return nil, LangError{arg.Position(), "expected closing paren in argument list"}
				}
			}
			end := ps.Cursor().Position.End
			ps.Advance(1)
			left = &CallNode{Expr: left, Arguments: arguments, pos: Position{start, end}}
		case TokenDot:
			ps.Advance(1)
			if ps.IsDone() || ps.Cursor().Kind != TokenIdentifier {
				return nil, LangError{left.Position(), "expected identifier after dot"}
			}
			attr := &LiteralNode{Token: ps.Cursor()}
			ps.Advance(1)
			left = &AttrNode{Expr: left, Attr: attr}
		default:
			return left, nil
		}
	}

	return left, nil
}

func (ps *Parser) ParseLiteral() (Node, error) {
	var left Node

	switch ps.Cursor().Kind {
	case TokenInteger, TokenFloat, TokenIdentifier, TokenString, TokenKeyword:
		left = &LiteralNode{Token: ps.Cursor()}
		ps.Advance(1)
	case TokenPlus, TokenMinus, TokenBitwiseNot, TokenNot:
		tok := ps.Cursor()
		ps.Advance(1)
		expr, err := ps.ParseExpr()
		if err != nil {
			return nil, err
		}
		left = &UnaryOpNode{Op: tok, Node: expr}
	case TokenLParen:
		ps.Advance(1)
		expr, err := ps.ParseExpr()
		if err != nil {
			return nil, err
		}

		if ps.IsDone() || ps.Cursor().Kind != TokenRParen {
			pos := Position{Start: expr.Position().End, End: expr.Position().End + 1}
			return nil, LangError{pos, "expected closing parenthesis"}
		}
		ps.Advance(1)

		left = expr
	case TokenLBrace:
		start := ps.Cursor().Position.Start
		ps.Advance(1)

		items := map[Node]Node{}
		for !ps.IsDone() && ps.Cursor().Kind != TokenRBrace {
			key, err := ps.ParseExpr()
			if err != nil {
				return nil, err
			}

			if ps.IsDone() || ps.Cursor().Kind != TokenColon {
				pos := Position{Start: key.Position().End, End: key.Position().End + 1}
				return nil, LangError{pos, "expected colon after key in mapping"}
			}
			ps.Advance(1)

			value, err := ps.ParseExpr()
			if err != nil {
				return nil, err
			}

			if !ps.IsDone() && ps.Cursor().Kind == TokenComma {
				ps.Advance(1)
			} else if ps.IsDone() || ps.Cursor().Kind != TokenRBrace {
				pos := Position{Start: value.Position().End, End: value.Position().End + 1}
				return nil, LangError{pos, "expected closing brace for mapping"}
			}

			items[key] = value
		}

		end := ps.Cursor().Position.End
		ps.Advance(1)

		left = &MapNode{Items: items, pos: Position{start, end}}
	case TokenLBracket:
		start := ps.Cursor().Position.Start
		ps.Advance(1)

		items := []Node{}
		for !ps.IsDone() && ps.Cursor().Kind != TokenRBracket {
			item, err := ps.ParseExpr()
			if err != nil {
				return nil, err
			}

			if !ps.IsDone() && ps.Cursor().Kind == TokenComma {
				ps.Advance(1)
			} else if ps.IsDone() || ps.Cursor().Kind != TokenRBracket {
				pos := Position{Start: item.Position().End, End: item.Position().End + 1}
				return nil, LangError{pos, "expected closing bracket or comma for list"}
			}

			items = append(items, item)
		}

		end := ps.Cursor().Position.End
		ps.Advance(1)

		left = &ListNode{Items: items, pos: Position{start, end}}
	}

	if left == nil {
		return nil, LangError{
			ps.Cursor().Position,
			fmt.Sprintf("unknown literal type %s", ps.Cursor().Kind),
		}
	}

	postfix, err := ps.tryPostfix(left)
	if err != nil {
		return nil, err
	}

	return postfix, nil
}

func (ps *Parser) ParseFactor() (Node, error) {
	var (
		left Node
		err  error
	)

	if left, err = ps.ParseLiteral(); err != nil {
		return nil, err
	}

	ops := []TokenKind{
		TokenMul, TokenDiv, TokenModulo,
		TokenBitwiseLeft, TokenBitwiseRight, TokenBitwiseAnd,
	}

	for !ps.IsDone() && slices.Contains(ops, ps.Cursor().Kind) {
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

	ops := []TokenKind{TokenPlus, TokenMinus, TokenBitwiseOr, TokenBitwiseXor}

	for !ps.IsDone() && slices.Contains(ops, ps.Cursor().Kind) {
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

func (ps *Parser) ParseComparison() (Node, error) {
	var (
		left Node
		err  error
	)

	if left, err = ps.ParseExpr(); err != nil {
		return nil, err
	}

	ops := []TokenKind{TokenEquals, TokenNotEq, TokenLt, TokenLtEq, TokenGt, TokenGtEq}

	for !ps.IsDone() && slices.Contains(ops, ps.Cursor().Kind) {
		tok := ps.Cursor()
		ps.Advance(1)

		right, err := ps.ParseExpr()
		if err != nil {
			return nil, err
		}

		left = &BinOpNode{Left: left, Op: tok, Right: right}
	}

	return left, nil
}

func (ps *Parser) ParseLogicalAnd() (Node, error) {
	var (
		left Node
		err  error
	)

	if left, err = ps.ParseComparison(); err != nil {
		return nil, err
	}

	ops := []TokenKind{TokenLogicalAnd}

	for !ps.IsDone() && slices.Contains(ops, ps.Cursor().Kind) {
		tok := ps.Cursor()
		ps.Advance(1)

		right, err := ps.ParseComparison()
		if err != nil {
			return nil, err
		}

		left = &BinOpNode{Left: left, Op: tok, Right: right}
	}

	return left, nil
}

func (ps *Parser) ParseLogicalOr() (Node, error) {
	var (
		left Node
		err  error
	)

	if left, err = ps.ParseLogicalAnd(); err != nil {
		return nil, err
	}

	for !ps.IsDone() && ps.Cursor().Kind == TokenLogicalOr {
		tok := ps.Cursor()
		ps.Advance(1)

		right, err := ps.ParseLogicalAnd()
		if err != nil {
			return nil, err
		}

		left = &BinOpNode{Left: left, Op: tok, Right: right}
	}

	return left, nil
}

func (ps *Parser) Parse() (Node, error) { return ps.ParseLogicalOr() }
