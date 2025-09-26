package bindef

import (
	"bytes"
	"fmt"
	"slices"
)

type TokenKind int

const (
	TokenLParen       TokenKind = iota // (
	TokenRParen                        // )
	TokenLBrace                        // {
	TokenRBrace                        // }
	TokenLBracket                      // [
	TokenRBracket                      // ]
	TokenMul                           // *
	TokenPow                           // **
	TokenDiv                           // /
	TokenPlus                          // +
	TokenMinus                         // -
	TokenComma                         // ,
	TokenColon                         // :
	TokenAt                            // @
	TokenAssign                        // =
	TokenEquals                        // ==
	TokenLt                            // <
	TokenGt                            // >
	TokenLtEq                          // <=
	TokenGtEq                          // >=
	TokenBitwiseOr                     // |
	TokenBitwiseAnd                    // &
	TokenBitwiseNot                    // ~
	TokenBitwiseXor                    // ^
	TokenBitwiseLeft                   // >>
	TokenBitwiseRight                  // <<
	TokenModulo                        // %
	TokenNot                           // !
	TokenNotEq                         // !=
	TokenIdentifier
	TokenKeyword // and, or, true, false
	TokenInteger
	TokenFloat
	TokenString
)

func (t TokenKind) String() string {
	switch t {
	case TokenLParen:
		return "LParen"
	case TokenRParen:
		return "RParen"
	case TokenLBrace:
		return "LBrace"
	case TokenRBrace:
		return "RBrace"
	case TokenLBracket:
		return "LBracket"
	case TokenRBracket:
		return "RBracket"
	case TokenMul:
		return "Mul"
	case TokenPow:
		return "Pow"
	case TokenDiv:
		return "Div"
	case TokenPlus:
		return "Plus"
	case TokenMinus:
		return "Minus"
	case TokenComma:
		return "Comma"
	case TokenColon:
		return "Colon"
	case TokenAt:
		return "At"
	case TokenAssign:
		return "Assign"
	case TokenEquals:
		return "Equals"
	case TokenLt:
		return "Lt"
	case TokenGt:
		return "Gt"
	case TokenLtEq:
		return "LtEq"
	case TokenGtEq:
		return "GtEq"
	case TokenBitwiseOr:
		return "BitwiseOr"
	case TokenBitwiseAnd:
		return "BitwiseAnd"
	case TokenBitwiseNot:
		return "BitwiseNot"
	case TokenBitwiseXor:
		return "BitwiseXor"
	case TokenBitwiseLeft:
		return "BitwiseLeft"
	case TokenBitwiseRight:
		return "BitwiseRight"
	case TokenModulo:
		return "Modulo"
	case TokenNot:
		return "Not"
	case TokenNotEq:
		return "NotEq"
	case TokenIdentifier:
		return "Identifier"
	case TokenKeyword:
		return "Keyword"
	case TokenInteger:
		return "Integer"
	case TokenFloat:
		return "Float"
	case TokenString:
		return "String"
	default:
		return fmt.Sprint(int(t))
	}
}

func IsASCIIChar(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

func IsASCIIDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// IsIdentifier reports whether a character ch can be part of a valid identifier.
//
// An identifier is a string of alphanumeric characters including the minus sign (-)
// and the underscore (_). An identifier must not start with a number or the minus
// sign and must not contain whitespace within.
func IsIdentifier(ch byte) bool {
	return IsASCIIChar(ch) || IsASCIIDigit(ch) || ch == '_' || ch == '-'
}

// IsStartOfIdentifier reports whether a character ch can be the start of a valid identifier.
func IsStartOfIdentifier(ch byte) bool {
	return IsASCIIChar(ch) && !IsASCIIDigit(ch) && ch != '-'
}

type Token struct {
	Kind  TokenKind
	Value string
}

type Lexer struct {
	Contents Scanner[byte]
	Tokens   []Token
}

func (lx *Lexer) LexNumeric() Token {
	digits := []byte{lx.Contents.Cursor()}
	lx.Contents.Advance(1)

	for !lx.Contents.IsDone() && (IsASCIIDigit(lx.Contents.Cursor()) || lx.Contents.Cursor() == '.') {
		digits = append(digits, lx.Contents.Cursor())
		lx.Contents.Advance(1)
	}

	if bytes.Contains(digits, []byte{'.'}) {
		return Token{Kind: TokenFloat, Value: string(digits)}
	} else {
		return Token{Kind: TokenInteger, Value: string(digits)}
	}
}

func (lx *Lexer) LexIdentifier() Token {
	ident := []byte{lx.Contents.Cursor()}
	lx.Contents.Advance(1)

	for !lx.Contents.IsDone() && IsIdentifier(lx.Contents.Cursor()) {
		ident = append(ident, lx.Contents.Cursor())
		lx.Contents.Advance(1)
	}

	if slices.Contains([]string{"and", "or", "true", "false"}, string(ident)) {
		return Token{Kind: TokenKeyword, Value: string(ident)}
	} else {
		return Token{Kind: TokenIdentifier, Value: string(ident)}
	}
}

func (lx *Lexer) LexString(delimiter byte) Token {
	strSeq := []byte{}
	lx.Contents.Advance(1) // for the single-byte start quote

	for !lx.Contents.IsDone() {
		cur := lx.Contents.Cursor()
		if cur == delimiter {
			lx.Contents.Advance(1)
			break
		}

		strSeq = append(strSeq, cur)
		lx.Contents.Advance(1)
	}

	return Token{Kind: TokenString, Value: string(strSeq)}
}

func (lx *Lexer) Process() {
	for !lx.Contents.IsDone() {
		ch := lx.Contents.Cursor()

		switch ch {
		case '(':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenLParen, Value: string(ch)})
		case ')':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenRParen, Value: string(ch)})
		case '{':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenLBrace, Value: string(ch)})
		case '}':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenRBrace, Value: string(ch)})
		case '[':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenLBracket, Value: string(ch)})
		case ']':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenRBracket, Value: string(ch)})
		case ',':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenComma, Value: string(ch)})
		case ':':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenColon, Value: string(ch)})
		case '@':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenAt, Value: string(ch)})
		case '=':
			switch nc := string(lx.Contents.Peek(1)); nc {
			case "=":
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenEquals, Value: string(ch) + nc})
				lx.Contents.Advance(1)
			default:
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenAssign, Value: string(ch)})
			}
		case '>':
			switch nc := string(lx.Contents.Peek(1)); nc {
			case "=":
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenGtEq, Value: string(ch) + nc})
				lx.Contents.Advance(1)
			case ">":
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenBitwiseRight, Value: string(ch) + nc})
				lx.Contents.Advance(1)
			default:
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenGt, Value: string(ch)})
			}
		case '<':
			switch nc := string(lx.Contents.Peek(1)); nc {
			case "=":
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenLtEq, Value: string(ch) + nc})
				lx.Contents.Advance(1)
			case "<":
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenBitwiseLeft, Value: string(ch) + nc})
				lx.Contents.Advance(1)
			default:
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenLt, Value: string(ch)})
			}
		case '+':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenPlus, Value: string(ch)})
		case '-':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenMinus, Value: string(ch)})
		case '/':
			switch nc := string(lx.Contents.Peek(1)); nc {
			case "/":
				for !lx.Contents.IsDone() && lx.Contents.Cursor() != '\n' {
					lx.Contents.Advance(1)
				}
			default:
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenDiv, Value: string(ch)})
			}
		case '*':
			switch nc := string(lx.Contents.Peek(1)); nc {
			case "*":
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenPow, Value: string(ch)})
				lx.Contents.Advance(1)
			default:
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenMul, Value: string(ch)})
			}
		case '!':
			switch nc := string(lx.Contents.Peek(1)); nc {
			case "=":
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenNotEq, Value: string(ch) + nc})
				lx.Contents.Advance(1)
			default:
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenNot, Value: string(ch)})
			}
		case '%':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenModulo, Value: string(ch)})
		case '|':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenBitwiseOr, Value: string(ch)})
		case '&':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenBitwiseAnd, Value: string(ch)})
		case '^':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenBitwiseXor, Value: string(ch)})
		case '~':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenBitwiseNot, Value: string(ch)})
		case '\'', '"':
			lx.Tokens = append(lx.Tokens, lx.LexString(ch))
			continue
		}

		if IsStartOfIdentifier(ch) {
			lx.Tokens = append(lx.Tokens, lx.LexIdentifier())
			continue
		}

		if IsASCIIDigit(ch) {
			lx.Tokens = append(lx.Tokens, lx.LexNumeric())
			continue
		}

		lx.Contents.Advance(1)
	}
}
