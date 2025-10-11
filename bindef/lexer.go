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
	TokenDot                           // .
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
	TokenRemainder                     // %
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
	case TokenDot:
		return "Dot"
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
	case TokenRemainder:
		return "Remainder"
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

// IsDecimalDigit reports whether a character ch is a decimal digit, i.e. a character
// in the range 0-9.
func IsDecimalDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// IsHexDigit reports whether a character ch is a hexadecimal digit, i.e. a character
// in the range 0-9 then A-F or a-f.
func IsHexDigit(ch byte) bool {
	return ('0' <= ch && ch <= '9') || ('a' <= ch && ch <= 'f') || ('A' <= ch && ch <= 'F')
}

// IsOctalDigit reports whether a character ch is an octal digit, i.e. a character
// in the range 0-7.
func IsOctalDigit(ch byte) bool {
	return '0' <= ch && ch <= '7'
}

// IsBinaryDigit reports whether a character ch is a binary digit, i.e. 0 or 1.
func IsBinaryDigit(ch byte) bool {
	return ch == '0' || ch == '1'
}

// IsASCIIWhitespace reports whether a character ch is ASCII whitespace, i.e. any
// of the following characters: horizontal tab ('\t'), line feed ('\n'),
// vertical tab ('\v'), form feed ('\f'), carriage return ('\r'), space (' '),
// next line (U+0085; NEL), and non-breaking space (U+00A0; NBSP).
func IsASCIIWhitespace(ch byte) bool {
	return slices.Contains([]byte{'\t', '\n', '\v', '\f', '\r', ' ', '\x85', '\xa0'}, ch)
}

// IsIdentifier reports whether a character ch can be part of a valid identifier.
//
// An identifier is a string of alphanumeric characters including the minus sign (-)
// and the underscore (_). An identifier must not start with a number or the minus
// sign and must not contain whitespace within.
func IsIdentifier(ch byte) bool {
	return IsASCIILetter(ch) || IsDecimalDigit(ch) || ch == '_' || ch == '-'
}

// IsStartOfIdentifier reports whether a character ch can be the start of a valid identifier.
func IsStartOfIdentifier(ch byte) bool {
	return (IsASCIILetter(ch) || ch == '_') && !IsDecimalDigit(ch) && ch != '-'
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
	Scanner[byte]
	Tokens []Token
}

func NewLexer(Contents []byte) Lexer {
	return Lexer{Scanner: NewScanner(Contents), Tokens: []Token{}}
}

func (lx *Lexer) LexNumeric() Token {
	start := lx.Position
	digits := []byte{lx.Cursor()}
	lx.Advance(1)

	type match func(ch byte) bool
	var isDigit match

	if digits[0] == '0' {
		switch lx.Cursor() {
		case 'b': // binary
			isDigit = IsBinaryDigit
			digits = append(digits, lx.Cursor())
			lx.Advance(1)
		case 'o': // octal
			isDigit = IsOctalDigit
			digits = append(digits, lx.Cursor())
			lx.Advance(1)
		case 'x': // hex
			isDigit = IsHexDigit
			digits = append(digits, lx.Cursor())
			lx.Advance(1)
		}
	}

	if isDigit == nil {
		isDigit = func(ch byte) bool { return IsDecimalDigit(ch) || lx.Cursor() == '.' }
	}

	for !lx.IsDone() && (isDigit(lx.Cursor()) || lx.Cursor() == '_') {
		if lx.Cursor() == '_' {
			lx.Advance(1)
			continue
		}

		digits = append(digits, lx.Cursor())
		lx.Advance(1)
	}

	pos := Position{start, lx.Position}
	if bytes.Contains(digits, []byte{'.'}) {
		return Token{Kind: TokenFloat, Value: string(digits), Position: pos}
	} else {
		return Token{Kind: TokenInteger, Value: string(digits), Position: pos}
	}
}

func (lx *Lexer) LexIdentifier() Token {
	start := lx.Position
	ident := []byte{lx.Cursor()}
	lx.Advance(1)

	for !lx.IsDone() && IsIdentifier(lx.Cursor()) {
		ident = append(ident, lx.Cursor())
		lx.Advance(1)
	}

	pos := Position{start, lx.Position}
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
	start := lx.Position

	strSeq := []byte{}
	lx.Advance(1) // for the single-byte start quote

	closed := false
	for !lx.IsDone() {
		cur := lx.Cursor()

		if cur == '\\' {
			escapeStart := lx.Position
			lx.Advance(1)

			nc := lx.Cursor()
			if escape, ok := escapeMap[nc]; ok {
				strSeq = append(strSeq, escape)
				lx.Advance(1)
				continue
			}

			if nc == 'x' {
				const byteSize int = 2
				hexSeq := string(lx.Peek(byteSize))
				hexVal, err := strconv.ParseInt(hexSeq, 16, 16)
				if err != nil {
					return Token{}, LangError{
						ErrorSyntax,
						Position{escapeStart, lx.Position + byteSize + 1},
						fmt.Sprintf("invalid hex sequence %s", hexSeq),
					}
				}

				lx.Advance(byteSize + 1)
				strSeq = append(strSeq, byte(hexVal))
				continue
			} else if IsDecimalDigit(nc) {
				const octSize int = 3

				octSeq := string(lx.Cursor()) + string(lx.Peek(octSize-1))
				octVal, err := strconv.ParseInt(octSeq, 8, 8)
				if err != nil {
					return Token{}, LangError{
						ErrorSyntax,
						Position{escapeStart, lx.Position + octSize},
						fmt.Sprintf("invalid octal sequence %s", octSeq),
					}
				}

				lx.Advance(octSize)
				strSeq = append(strSeq, byte(octVal))
				continue
			}

			return Token{}, LangError{
				ErrorSyntax,
				Position{escapeStart, lx.Position},
				fmt.Sprintf("unknown escape sequence %q", lx.Cursor()),
			}
		}

		if cur == delimiter {
			lx.Advance(1)
			closed = true
			break
		}

		strSeq = append(strSeq, cur)
		lx.Advance(1)
	}

	if !closed {
		return Token{}, LangError{
			ErrorSyntax,
			Position{start, lx.Position},
			"string was never closed",
		}
	}

	return Token{
		Kind:     TokenString,
		Value:    string(strSeq),
		Position: Position{start, lx.Position},
	}, nil
}

var singleTokenMap = map[byte]TokenKind{
	'(': TokenLParen,
	')': TokenRParen,
	'{': TokenLBrace,
	'}': TokenRBrace,
	'[': TokenLBracket,
	']': TokenRBracket,
	',': TokenComma,
	':': TokenColon,
	'.': TokenDot,
	'+': TokenPlus,
	'-': TokenMinus,
	'!': TokenNot,
	'%': TokenRemainder,
	'^': TokenBitwiseXor,
	'~': TokenBitwiseNot,
}

func (lx *Lexer) addToken(kind TokenKind, value string) {
	pos := Position{Start: lx.Position, End: lx.Position + len(value)}
	lx.Tokens = append(lx.Tokens, Token{Kind: kind, Value: value, Position: pos})
}

func (lx *Lexer) Process() error {
	for !lx.IsDone() {
		ch := lx.Cursor()

		// single-byte tokens
		if kind, ok := singleTokenMap[ch]; ok {
			lx.addToken(kind, string(ch))
			lx.Advance(1)
			continue
		}

		// (possibly) two-byte tokens
		switch ch {
		case '=':
			switch nc := string(lx.Peek(1)); nc {
			case "=":
				lx.addToken(TokenEquals, string(ch)+nc)
				lx.Advance(1)
				continue
			}
		case '>':
			switch nc := string(lx.Peek(1)); nc {
			case "=":
				lx.addToken(TokenGtEq, string(ch)+nc)
				lx.Advance(1)
			case ">":
				lx.addToken(TokenBitwiseRight, string(ch)+nc)
				lx.Advance(1)
			default:
				lx.addToken(TokenGt, string(ch))
			}
			lx.Advance(1)
			continue

		case '<':
			switch nc := string(lx.Peek(1)); nc {
			case "=":
				lx.addToken(TokenLtEq, string(ch)+nc)
				lx.Advance(1)
			case "<":
				lx.addToken(TokenBitwiseLeft, string(ch)+nc)
				lx.Advance(1)
			default:
				lx.addToken(TokenLt, string(ch))
			}
			lx.Advance(1)
			continue

		case '/':
			switch nc := string(lx.Peek(1)); nc {
			case "/":
				for !lx.IsDone() && lx.Cursor() != '\n' {
					lx.Advance(1)
				}
				continue
			default:
				lx.addToken(TokenDiv, string(ch))
				lx.Advance(1)
				continue
			}
		case '*':
			switch nc := string(lx.Peek(1)); nc {
			case "*":
				lx.addToken(TokenPow, string(ch)+nc)
				lx.Advance(1)
			default:
				lx.addToken(TokenMul, string(ch))
			}
			lx.Advance(1)
			continue
		case '!':
			switch nc := string(lx.Peek(1)); nc {
			case "=":
				lx.addToken(TokenNotEq, string(ch)+nc)
				lx.Advance(1)
			default:
				lx.addToken(TokenNot, string(ch))
			}
			lx.Advance(1)
			continue
		case '|':
			switch nc := string(lx.Peek(1)); nc {
			case "|":
				lx.addToken(TokenLogicalOr, string(ch)+nc)
				lx.Advance(1)
			default:
				lx.addToken(TokenBitwiseOr, string(ch))
			}
			lx.Advance(1)
			continue
		case '&':
			switch nc := string(lx.Peek(1)); nc {
			case "&":
				lx.addToken(TokenLogicalAnd, string(ch)+nc)
				lx.Advance(1)
			default:
				lx.addToken(TokenBitwiseAnd, string(ch))
			}
			lx.Advance(1)
			continue
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

		if IsDecimalDigit(ch) {
			lx.Tokens = append(lx.Tokens, lx.LexNumeric())
			continue
		}

		if IsASCIIWhitespace(ch) {
			lx.Advance(1)
			continue
		}

		return LangError{
			Position: Position{Start: lx.Position, End: lx.Position + 1},
			Message:  fmt.Sprintf("unknown character %q", lx.Cursor()),
		}
	}

	return nil
}
