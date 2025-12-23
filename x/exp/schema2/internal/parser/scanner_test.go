package parser

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestTokenize(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		tokens, err := Tokenize([]byte(""))
		testutil.OK(t, err)
		testutil.Equals(t, len(tokens), 1)
		testutil.Equals(t, tokens[0].Type, TokenEOF)
	})

	t.Run("identifiers", func(t *testing.T) {
		t.Parallel()
		tokens, err := Tokenize([]byte("namespace entity action type"))
		testutil.OK(t, err)
		testutil.Equals(t, len(tokens), 5) // 4 idents + EOF
		testutil.Equals(t, tokens[0].Type, TokenIdent)
		testutil.Equals(t, tokens[0].Text, "namespace")
		testutil.Equals(t, tokens[1].Text, "entity")
		testutil.Equals(t, tokens[2].Text, "action")
		testutil.Equals(t, tokens[3].Text, "type")
	})

	t.Run("strings", func(t *testing.T) {
		t.Parallel()
		tokens, err := Tokenize([]byte(`"hello" "world"`))
		testutil.OK(t, err)
		testutil.Equals(t, len(tokens), 3) // 2 strings + EOF
		testutil.Equals(t, tokens[0].Type, TokenString)
		testutil.Equals(t, tokens[0].Text, `"hello"`)
		testutil.Equals(t, tokens[1].Text, `"world"`)
	})

	t.Run("operators", func(t *testing.T) {
		t.Parallel()
		tokens, err := Tokenize([]byte("@{};,:<>[]()=?::"))
		testutil.OK(t, err)
		// @ { } ; , : < > [ ] ( ) = ? :: EOF
		testutil.Equals(t, len(tokens), 16) // 15 operators + EOF
		testutil.Equals(t, tokens[0].Text, "@")
		testutil.Equals(t, tokens[1].Text, "{")
		testutil.Equals(t, tokens[2].Text, "}")
		testutil.Equals(t, tokens[3].Text, ";")
		testutil.Equals(t, tokens[4].Text, ",")
		testutil.Equals(t, tokens[5].Text, ":")
		testutil.Equals(t, tokens[6].Text, "<")
		testutil.Equals(t, tokens[7].Text, ">")
		testutil.Equals(t, tokens[8].Text, "[")
		testutil.Equals(t, tokens[9].Text, "]")
		testutil.Equals(t, tokens[10].Text, "(")
		testutil.Equals(t, tokens[11].Text, ")")
		testutil.Equals(t, tokens[12].Text, "=")
		testutil.Equals(t, tokens[13].Text, "?")
		testutil.Equals(t, tokens[14].Text, "::")
	})

	t.Run("double colon", func(t *testing.T) {
		t.Parallel()
		tokens, err := Tokenize([]byte("Foo::Bar"))
		testutil.OK(t, err)
		testutil.Equals(t, len(tokens), 4) // Foo :: Bar EOF
		testutil.Equals(t, tokens[0].Text, "Foo")
		testutil.Equals(t, tokens[1].Text, "::")
		testutil.Equals(t, tokens[2].Text, "Bar")
	})

	t.Run("line comment", func(t *testing.T) {
		t.Parallel()
		tokens, err := Tokenize([]byte("foo // comment\nbar"))
		testutil.OK(t, err)
		testutil.Equals(t, len(tokens), 3) // foo bar EOF
		testutil.Equals(t, tokens[0].Text, "foo")
		testutil.Equals(t, tokens[1].Text, "bar")
	})

	t.Run("block comment", func(t *testing.T) {
		t.Parallel()
		tokens, err := Tokenize([]byte("foo /* comment */ bar"))
		testutil.OK(t, err)
		testutil.Equals(t, len(tokens), 3) // foo bar EOF
		testutil.Equals(t, tokens[0].Text, "foo")
		testutil.Equals(t, tokens[1].Text, "bar")
	})

	t.Run("whitespace", func(t *testing.T) {
		t.Parallel()
		tokens, err := Tokenize([]byte("  \t\n\rfoo  \t\n\r  bar  "))
		testutil.OK(t, err)
		testutil.Equals(t, len(tokens), 3)
		testutil.Equals(t, tokens[0].Text, "foo")
		testutil.Equals(t, tokens[1].Text, "bar")
	})

	t.Run("string with escapes", func(t *testing.T) {
		t.Parallel()
		tokens, err := Tokenize([]byte(`"hello\nworld"`))
		testutil.OK(t, err)
		testutil.Equals(t, len(tokens), 2)
		testutil.Equals(t, tokens[0].Type, TokenString)
		testutil.Equals(t, tokens[0].Text, `"hello\nworld"`)
	})

	t.Run("position tracking", func(t *testing.T) {
		t.Parallel()
		tokens, err := Tokenize([]byte("foo\nbar"))
		testutil.OK(t, err)
		testutil.Equals(t, tokens[0].Pos.Line, 1)
		testutil.Equals(t, tokens[0].Pos.Column, 1)
		testutil.Equals(t, tokens[1].Pos.Line, 2)
		testutil.Equals(t, tokens[1].Pos.Column, 1)
	})
}

func TestTokenStringValue(t *testing.T) {
	t.Parallel()

	t.Run("simple string", func(t *testing.T) {
		t.Parallel()
		tok := Token{Type: TokenString, Text: `"hello"`}
		val, err := tok.stringValue()
		testutil.OK(t, err)
		testutil.Equals(t, val, "hello")
	})

	t.Run("string with escapes", func(t *testing.T) {
		t.Parallel()
		tok := Token{Type: TokenString, Text: `"hello\nworld"`}
		val, err := tok.stringValue()
		testutil.OK(t, err)
		testutil.Equals(t, val, "hello\nworld")
	})

	t.Run("non-string token", func(t *testing.T) {
		t.Parallel()
		tok := Token{Type: TokenIdent, Text: "foo"}
		_, err := tok.stringValue()
		testutil.Equals(t, err != nil, true)
	})
}

func TestTokenMethods(t *testing.T) {
	t.Parallel()

	t.Run("isEOF", func(t *testing.T) {
		t.Parallel()
		tok := Token{Type: TokenEOF}
		testutil.Equals(t, tok.isEOF(), true)
		tok = Token{Type: TokenIdent}
		testutil.Equals(t, tok.isEOF(), false)
	})

	t.Run("isIdent", func(t *testing.T) {
		t.Parallel()
		tok := Token{Type: TokenIdent}
		testutil.Equals(t, tok.isIdent(), true)
		tok = Token{Type: TokenString}
		testutil.Equals(t, tok.isIdent(), false)
	})

	t.Run("isString", func(t *testing.T) {
		t.Parallel()
		tok := Token{Type: TokenString}
		testutil.Equals(t, tok.isString(), true)
		tok = Token{Type: TokenIdent}
		testutil.Equals(t, tok.isString(), false)
	})
}

func TestScannerEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("unknown operator", func(t *testing.T) {
		t.Parallel()
		tokens, err := Tokenize([]byte("~"))
		testutil.OK(t, err)
		testutil.Equals(t, tokens[0].Type, TokenUnknown)
	})

	t.Run("unterminated string", func(t *testing.T) {
		t.Parallel()
		_, err := Tokenize([]byte(`"hello`))
		// Unterminated string should produce an error
		testutil.Equals(t, err != nil, true)
	})

	t.Run("unterminated block comment", func(t *testing.T) {
		t.Parallel()
		_, err := Tokenize([]byte("/* unterminated"))
		testutil.Equals(t, err != nil, true)
	})
}
