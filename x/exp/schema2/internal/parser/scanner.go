package parser

import (
	"bytes"
	"fmt"
	"io"
	"unicode/utf8"
)

const bufLen = 1024

const (
	specialRuneEOF = rune(-(iota + 1))
	specialRuneBOF
)

// scanner implements reading of Unicode characters and tokens from an io.Reader.
type scanner struct {
	src io.Reader

	// Source buffer
	srcBuf [bufLen + 1]byte
	srcPos int
	srcEnd int

	// Source position
	srcBufOffset int
	line         int
	column       int
	lastLineLen  int
	lastCharLen  int

	// Token text buffer
	tokBuf bytes.Buffer
	tokPos int
	tokEnd int

	// One character look-ahead
	ch rune

	// Last error encountered
	err error

	// Start position of most recently scanned token
	position Position
}

func newScanner(src io.Reader) (*scanner, error) {
	if src == nil {
		return nil, fmt.Errorf("nil reader")
	}
	s := &scanner{}
	s.init(src)
	return s, nil
}

func (s *scanner) init(src io.Reader) {
	s.src = src
	s.srcBuf[0] = utf8.RuneSelf
	s.srcPos = 0
	s.srcEnd = 0
	s.srcBufOffset = 0
	s.line = 1
	s.column = 0
	s.lastLineLen = 0
	s.lastCharLen = 0
	s.tokPos = -1
	s.ch = specialRuneBOF
	s.position.Line = 0
}

func (s *scanner) next() rune {
	ch, width := rune(s.srcBuf[s.srcPos]), 1

	if ch >= utf8.RuneSelf {
		for s.srcPos+utf8.UTFMax > s.srcEnd && !utf8.FullRune(s.srcBuf[s.srcPos:s.srcEnd]) {
			if s.tokPos >= 0 {
				s.tokBuf.Write(s.srcBuf[s.tokPos:s.srcPos])
				s.tokPos = 0
			}
			copy(s.srcBuf[0:], s.srcBuf[s.srcPos:s.srcEnd])
			s.srcBufOffset += s.srcPos
			i := s.srcEnd - s.srcPos
			n, err := s.src.Read(s.srcBuf[i:bufLen])
			s.srcPos = 0
			s.srcEnd = i + n
			s.srcBuf[s.srcEnd] = utf8.RuneSelf
			if err != nil {
				if err != io.EOF {
					s.error(err.Error())
				}
				if s.srcEnd == 0 {
					if s.lastCharLen > 0 {
						s.column++
					}
					s.lastCharLen = 0
					return specialRuneEOF
				}
				break
			}
		}
		ch = rune(s.srcBuf[s.srcPos])
		if ch >= utf8.RuneSelf {
			ch, width = utf8.DecodeRune(s.srcBuf[s.srcPos:s.srcEnd])
			if ch == utf8.RuneError && width == 1 {
				s.srcPos += width
				s.lastCharLen = width
				s.column++
				s.error("invalid UTF-8 encoding")
				return ch
			}
		}
	}

	s.srcPos += width
	s.lastCharLen = width
	s.column++

	switch ch {
	case 0:
		s.error("invalid character NUL")
	case '\n':
		s.line++
		s.lastLineLen = s.column
		s.column = 0
	}

	return ch
}

func (s *scanner) error(msg string) {
	s.tokEnd = s.srcPos - s.lastCharLen
	s.err = fmt.Errorf("%d:%d: %v", s.position.Line, s.position.Column, msg)
}

func isASCIILetter(ch rune) bool {
	return (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z')
}

func isASCIINumber(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isIdentRune(ch rune, first bool) bool {
	return ch == '_' || isASCIILetter(ch) || isASCIINumber(ch) && !first
}

func (s *scanner) scanIdentifier() rune {
	ch := s.next()
	for isIdentRune(ch, false) {
		ch = s.next()
	}
	return ch
}

func (s *scanner) scanString() rune {
	ch := s.next() // read character after opening quote
	for ch != '"' {
		if ch == '\n' || ch < 0 {
			s.error("string not terminated")
			return ch
		}
		if ch == '\\' {
			ch = s.next() // skip escape char
			if ch < 0 {
				s.error("string not terminated")
				return ch
			}
		}
		ch = s.next()
	}
	return s.next() // consume closing quote
}

func (s *scanner) scanComment(ch rune) rune {
	if ch == '/' {
		// line comment
		ch = s.next()
		for ch != '\n' && ch >= 0 {
			ch = s.next()
		}
		return ch
	}

	// block comment
	ch = s.next()
	for {
		if ch < 0 {
			s.error("comment not terminated")
			break
		}
		ch0 := ch
		ch = s.next()
		if ch0 == '*' && ch == '/' {
			ch = s.next()
			break
		}
	}
	return ch
}

func (s *scanner) scanOperator(ch0, ch rune) (TokenType, rune) {
	switch ch0 {
	case '@', '.', ',', ';', '(', ')', '{', '}', '[', ']', '<', '>', '=', '?':
		// single character operators
	case ':':
		if ch == ':' {
			ch = s.next()
		}
	default:
		return TokenUnknown, ch
	}
	return TokenOperator, ch
}

func isWhitespace(c rune) bool {
	switch c {
	case '\t', '\n', '\r', ' ':
		return true
	default:
		return false
	}
}

func (s *scanner) nextToken() Token {
	if s.ch == specialRuneBOF {
		s.ch = s.next()
	}

	ch := s.ch

	s.tokPos = -1
	s.position.Line = 0

redo:
	for isWhitespace(ch) {
		ch = s.next()
	}

	s.tokBuf.Reset()
	s.tokPos = s.srcPos - s.lastCharLen

	s.position.Offset = s.srcBufOffset + s.tokPos
	if s.column > 0 {
		s.position.Line = s.line
		s.position.Column = s.column
	} else {
		s.position.Line = s.line - 1
		s.position.Column = s.lastLineLen
	}

	var tt TokenType
	switch {
	case ch == specialRuneEOF:
		tt = TokenEOF
	case isIdentRune(ch, true):
		ch = s.scanIdentifier()
		tt = TokenIdent
	case ch == '"':
		ch = s.scanString()
		tt = TokenString
	case ch == '/':
		ch0 := ch
		ch = s.next()
		if ch == '/' || ch == '*' {
			s.tokPos = -1
			ch = s.scanComment(ch)
			goto redo
		}
		tt, ch = s.scanOperator(ch0, ch)
	default:
		tt, ch = s.scanOperator(ch, s.next())
	}

	s.tokEnd = s.srcPos - s.lastCharLen
	s.ch = ch

	return Token{
		Type: tt,
		Pos:  s.position,
		Text: s.tokenText(),
	}
}

func (s *scanner) tokenText() string {
	if s.tokPos < 0 {
		return ""
	}

	if s.tokBuf.Len() == 0 {
		return string(s.srcBuf[s.tokPos:s.tokEnd])
	}

	s.tokBuf.Write(s.srcBuf[s.tokPos:s.tokEnd])
	s.tokPos = s.tokEnd
	return s.tokBuf.String()
}

// Tokenize tokenizes the given source bytes.
func Tokenize(src []byte) ([]Token, error) {
	return TokenizeReader(bytes.NewBuffer(src))
}

// TokenizeReader tokenizes from an io.Reader.
func TokenizeReader(r io.Reader) ([]Token, error) {
	s, err := newScanner(r)
	if err != nil {
		return nil, err
	}
	var res []Token
	for tok := s.nextToken(); s.err == nil && tok.Type != TokenEOF; tok = s.nextToken() {
		res = append(res, tok)
	}
	if s.err != nil {
		return nil, s.err
	}
	res = append(res, Token{Type: TokenEOF, Pos: s.position})
	return res, nil
}
