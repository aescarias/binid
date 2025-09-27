package bindef

import (
	"fmt"
)

type Node interface {
	Type() string
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

type AttrAccessNode struct {
	Parent  Token
	Members []Token
}

func (ln *LiteralNode) Type() string { return "Literal" }
func (ln *LiteralNode) Position() Position {
	return ln.Token.Position
}

func (bn *BinOpNode) Type() string { return "BinOp" }
func (bn *BinOpNode) Position() Position {
	return Position{Start: bn.Left.Position().Start, End: bn.Right.Position().End}
}

func (un *UnaryOpNode) Type() string { return "UnaryOp" }
func (un *UnaryOpNode) Position() Position {
	return Position{Start: un.Op.Position.Start, End: un.Node.Position().End}
}

func (mn *MapNode) Type() string       { return "Map" }
func (mn *MapNode) Position() Position { return mn.pos }

func (ln *ListNode) Type() string       { return "List" }
func (ln *ListNode) Position() Position { return ln.pos }

func (an *AttrAccessNode) Type() string { return "AttrAccess" }
func (an *AttrAccessNode) Position() Position {
	return Position{
		Start: an.Parent.Position.Start,
		End:   an.Members[len(an.Members)-1].Position.End,
	}
}

type Parser struct {
	Scanner[Token]
}

func (ps *Parser) ParseLiteral() (Node, error) {
	switch ps.Cursor().Kind {
	case TokenInteger, TokenFloat:
		lit := &LiteralNode{Token: ps.Cursor()}
		ps.Advance(1)
		return lit, nil
	case TokenIdentifier, TokenString:
		parent := ps.Cursor()
		lookup := []Token{}

		ps.Advance(1)

		nextAttr := false

		for !ps.IsDone() {
			tok := ps.Cursor()
			if tok.Kind == TokenDot {
				ps.Advance(1)
				nextAttr = true
			} else if nextAttr && tok.Kind == TokenIdentifier {
				lookup = append(lookup, tok)
				ps.Advance(1)
				nextAttr = false
			} else {
				break
			}
		}

		if len(lookup) <= 0 {
			return &LiteralNode{Token: parent}, nil
		} else {
			return &AttrAccessNode{Parent: parent, Members: lookup}, nil
		}
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
			pos := Position{Start: expr.Position().End, End: expr.Position().End + 1}
			return nil, LangError{pos, "expected closing parenthesis"}
		}
		ps.Advance(1)

		return expr, nil
	case TokenLBrace:
		start := ps.Cursor().Position.Start
		ps.Advance(1)

		items := map[Node]Node{}
		for !ps.IsDone() && ps.Cursor().Kind != TokenRBrace {
			key, err := ps.ParseExpr()
			if err != nil {
				return nil, err
			}

			if ps.Cursor().Kind != TokenColon {
				pos := Position{Start: key.Position().End, End: key.Position().End + 1}
				return nil, LangError{pos, "expected colon after key in mapping"}
			}
			ps.Advance(1)

			value, err := ps.ParseExpr()
			if err != nil {
				return nil, err
			}

			if ps.Cursor().Kind == TokenComma {
				ps.Advance(1)
			} else if ps.IsDone() || ps.Cursor().Kind != TokenRBrace {
				pos := Position{Start: value.Position().End, End: value.Position().End + 1}
				return nil, LangError{pos, "expected closing brace for mapping"}
			}

			items[key] = value
		}

		end := ps.Cursor().Position.End
		ps.Advance(1)
		return &MapNode{Items: items, pos: Position{start, end}}, nil
	case TokenLBracket:
		start := ps.Cursor().Position.Start
		ps.Advance(1)

		items := []Node{}
		for !ps.IsDone() && ps.Cursor().Kind != TokenRBracket {
			item, err := ps.ParseExpr()
			if err != nil {
				return nil, err
			}

			if ps.Cursor().Kind == TokenComma {
				ps.Advance(1)
			} else if ps.IsDone() || ps.Cursor().Kind != TokenRBracket {
				pos := Position{Start: item.Position().End, End: item.Position().End + 1}
				return nil, LangError{pos, "expected closing bracket for list"}
			}

			items = append(items, item)
		}

		end := ps.Cursor().Position.End
		ps.Advance(1)
		return &ListNode{Items: items, pos: Position{start, end}}, nil
	}

	return nil, LangError{
		ps.Cursor().Position,
		fmt.Sprintf("unknown literal type %s", ps.Cursor().Kind),
	}
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
