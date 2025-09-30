package bindef

import (
	"bytes"
	"fmt"
	"slices"
	"strconv"
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
	TokenDot                           // .
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
	TokenLogicalAnd                    // &&
	TokenLogicalOr                     // ||
	TokenIdentifier
	TokenKeyword // true, false
	TokenInteger
	TokenFloat
	TokenString
)

var TypeNames = []string{
	"bool", "byte",
	"uint8", "uint16", "uint32", "uint64",
	"int8", "int16", "int32", "int64",
}

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
	case TokenDot:
		return "Dot"
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
	case TokenLogicalAnd:
		return "LogicalAnd"
	case TokenLogicalOr:
		return "LogicalOr"
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

type KeywordKind string

const (
	KeywordTrue  KeywordKind = "true"
	KeywordFalse KeywordKind = "false"
)

// IsASCIILetter reports whether a character ch is an ASCII letter, i.e. a character
// in the range a-z or A-Z.
func IsASCIILetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

// IsASCIIDigit reports whether a character ch is an ASCII digit, i.e. a character
// in the range 0-9.
func IsASCIIDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// IsIdentifier reports whether a character ch can be part of a valid identifier.
//
// An identifier is a string of alphanumeric characters including the minus sign (-)
// and the underscore (_). An identifier must not start with a number or the minus
// sign and must not contain whitespace within.
func IsIdentifier(ch byte) bool {
	return IsASCIILetter(ch) || IsASCIIDigit(ch) || ch == '_' || ch == '-'
}

// IsStartOfIdentifier reports whether a character ch can be the start of a valid identifier.
func IsStartOfIdentifier(ch byte) bool {
	return IsASCIILetter(ch) && !IsASCIIDigit(ch) && ch != '-'
}

type Position struct {
	Start int
	End   int
}

type Token struct {
	Kind     TokenKind
	Value    string
	Position Position
}

type Lexer struct {
	Contents Scanner[byte]
	Tokens   []Token
}

func (lx *Lexer) LexNumeric() Token {
	start := lx.Contents.Current
	digits := []byte{lx.Contents.Cursor()}
	lx.Contents.Advance(1)

	for !lx.Contents.IsDone() && (IsASCIIDigit(lx.Contents.Cursor()) || lx.Contents.Cursor() == '.') {
		digits = append(digits, lx.Contents.Cursor())
		lx.Contents.Advance(1)
	}

	pos := Position{start, lx.Contents.Current}
	if bytes.Contains(digits, []byte{'.'}) {
		return Token{Kind: TokenFloat, Value: string(digits), Position: pos}
	} else {
		return Token{Kind: TokenInteger, Value: string(digits), Position: pos}
	}
}

func (lx *Lexer) LexIdentifier() Token {
	start := lx.Contents.Current
	ident := []byte{lx.Contents.Cursor()}
	lx.Contents.Advance(1)

	for !lx.Contents.IsDone() && IsIdentifier(lx.Contents.Cursor()) {
		ident = append(ident, lx.Contents.Cursor())
		lx.Contents.Advance(1)
	}

	pos := Position{start, lx.Contents.Current}
	if slices.Contains([]string{string(KeywordTrue), string(KeywordFalse)}, string(ident)) {
		return Token{Kind: TokenKeyword, Value: string(ident), Position: pos}
	} else {
		return Token{Kind: TokenIdentifier, Value: string(ident), Position: pos}
	}
}

var escapeMap = map[byte]byte{
	'\\': '\\',
	'\'': '\'',
	'"':  '"',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
}

func (lx *Lexer) LexString(delimiter byte) (Token, error) {
	start := lx.Contents.Current

	strSeq := []byte{}
	lx.Contents.Advance(1) // for the single-byte start quote

	closed := false
	for !lx.Contents.IsDone() {
		cur := lx.Contents.Cursor()

		if cur == '\\' {
			escapeStart := lx.Contents.Current
			lx.Contents.Advance(1)

			nc := lx.Contents.Cursor()
			if escape, ok := escapeMap[nc]; ok {
				strSeq = append(strSeq, escape)
				lx.Contents.Advance(1)
				continue
			}

			if nc == 'x' {
				const byteSize int = 2
				hexSeq := string(lx.Contents.Peek(byteSize))
				hexVal, err := strconv.ParseInt(hexSeq, 16, 8)
				if err != nil {
					return Token{}, LangError{
						Position{escapeStart, lx.Contents.Current + byteSize},
						fmt.Sprintf("invalid hex sequence %s", hexSeq),
					}
				}

				lx.Contents.Advance(byteSize + 1)
				strSeq = append(strSeq, byte(hexVal))
				continue
			} else if IsASCIIDigit(nc) {
				const octSize int = 3

				octSeq := string(lx.Contents.Cursor()) + string(lx.Contents.Peek(octSize-1))
				octVal, err := strconv.ParseInt(octSeq, 8, 8)
				if err != nil {
					return Token{}, LangError{
						Position{escapeStart, lx.Contents.Current + octSize},
						fmt.Sprintf("invalid octal sequence %s", octSeq),
					}
				}

				lx.Contents.Advance(octSize)
				strSeq = append(strSeq, byte(octVal))
				continue
			}

			return Token{}, LangError{
				Position{escapeStart, lx.Contents.Current},
				fmt.Sprintf("unknown escape sequence %q", lx.Contents.Cursor()),
			}
		}

		if cur == delimiter {
			lx.Contents.Advance(1)
			closed = true
			break
		}

		strSeq = append(strSeq, cur)
		lx.Contents.Advance(1)
	}

	if !closed {
		return Token{}, LangError{
			Position{start, lx.Contents.Current},
			"string was never closed",
		}
	}

	return Token{
		Kind:     TokenString,
		Value:    string(strSeq),
		Position: Position{start, lx.Contents.Current},
	}, nil
}

func (lx *Lexer) Process() error {
	for !lx.Contents.IsDone() {
		ch := lx.Contents.Cursor()

		singlePos := Position{lx.Contents.Current, lx.Contents.Current + 1}
		doublePos := Position{lx.Contents.Current, lx.Contents.Current + 2}

		switch ch {
		case '(':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenLParen, Value: string(ch), Position: singlePos})
		case ')':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenRParen, Value: string(ch), Position: singlePos})
		case '{':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenLBrace, Value: string(ch), Position: singlePos})
		case '}':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenRBrace, Value: string(ch), Position: singlePos})
		case '[':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenLBracket, Value: string(ch), Position: singlePos})
		case ']':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenRBracket, Value: string(ch), Position: singlePos})
		case ',':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenComma, Value: string(ch), Position: singlePos})
		case ':':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenColon, Value: string(ch), Position: singlePos})
		case '@':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenAt, Value: string(ch), Position: singlePos})
		case '.':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenDot, Value: string(ch), Position: singlePos})
		case '=':
			switch nc := string(lx.Contents.Peek(1)); nc {
			case "=":
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenEquals, Value: string(ch) + nc, Position: doublePos})
				lx.Contents.Advance(1)
			default:
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenAssign, Value: string(ch), Position: singlePos})
			}
		case '>':
			switch nc := string(lx.Contents.Peek(1)); nc {
			case "=":
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenGtEq, Value: string(ch) + nc, Position: doublePos})
				lx.Contents.Advance(1)
			case ">":
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenBitwiseRight, Value: string(ch) + nc, Position: doublePos})
				lx.Contents.Advance(1)
			default:
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenGt, Value: string(ch), Position: singlePos})
			}
		case '<':
			switch nc := string(lx.Contents.Peek(1)); nc {
			case "=":
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenLtEq, Value: string(ch) + nc, Position: doublePos})
				lx.Contents.Advance(1)
			case "<":
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenBitwiseLeft, Value: string(ch) + nc, Position: doublePos})
				lx.Contents.Advance(1)
			default:
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenLt, Value: string(ch), Position: singlePos})
			}
		case '+':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenPlus, Value: string(ch), Position: singlePos})
		case '-':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenMinus, Value: string(ch), Position: singlePos})
		case '/':
			switch nc := string(lx.Contents.Peek(1)); nc {
			case "/":
				for !lx.Contents.IsDone() && lx.Contents.Cursor() != '\n' {
					lx.Contents.Advance(1)
				}
			default:
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenDiv, Value: string(ch), Position: singlePos})
			}
		case '*':
			switch nc := string(lx.Contents.Peek(1)); nc {
			case "*":
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenPow, Value: string(ch) + nc, Position: doublePos})
				lx.Contents.Advance(1)
			default:
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenMul, Value: string(ch), Position: singlePos})
			}
		case '!':
			switch nc := string(lx.Contents.Peek(1)); nc {
			case "=":
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenNotEq, Value: string(ch) + nc, Position: doublePos})
				lx.Contents.Advance(1)
			default:
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenNot, Value: string(ch), Position: doublePos})
			}
		case '%':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenModulo, Value: string(ch), Position: singlePos})
		case '|':
			switch nc := string(lx.Contents.Peek(1)); nc {
			case "|":
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenLogicalOr, Value: string(ch) + nc, Position: doublePos})
				lx.Contents.Advance(1)
			default:
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenBitwiseOr, Value: string(ch), Position: doublePos})
			}
		case '&':
			switch nc := string(lx.Contents.Peek(1)); nc {
			case "&":
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenLogicalAnd, Value: string(ch) + nc, Position: doublePos})
				lx.Contents.Advance(1)
			default:
				lx.Tokens = append(lx.Tokens, Token{Kind: TokenBitwiseAnd, Value: string(ch), Position: doublePos})
			}
		case '^':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenBitwiseXor, Value: string(ch), Position: singlePos})
		case '~':
			lx.Tokens = append(lx.Tokens, Token{Kind: TokenBitwiseNot, Value: string(ch), Position: singlePos})
		case '\'', '"':
			strTok, err := lx.LexString(ch)
			if err != nil {
				return err
			}

			lx.Tokens = append(lx.Tokens, strTok)
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

	return nil
}
