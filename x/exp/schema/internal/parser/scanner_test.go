package parser

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestTokenize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantCount int
		validate  func(t *testing.T, tokens []Token)
	}{
		{
			name:      "empty",
			input:     "",
			wantCount: 1,
			validate: func(t *testing.T, tokens []Token) {
				testutil.Equals(t, tokens[0].Type, TokenEOF)
			},
		},
		{
			name:      "identifiers",
			input:     "namespace entity action type",
			wantCount: 5,
			validate: func(t *testing.T, tokens []Token) {
				testutil.Equals(t, tokens[0].Type, TokenIdent)
				testutil.Equals(t, tokens[0].Text, "namespace")
				testutil.Equals(t, tokens[1].Text, "entity")
				testutil.Equals(t, tokens[2].Text, "action")
				testutil.Equals(t, tokens[3].Text, "type")
			},
		},
		{
			name:      "strings",
			input:     `"hello" "world"`,
			wantCount: 3,
			validate: func(t *testing.T, tokens []Token) {
				testutil.Equals(t, tokens[0].Type, TokenString)
				testutil.Equals(t, tokens[0].Text, `"hello"`)
				testutil.Equals(t, tokens[1].Text, `"world"`)
			},
		},
		{
			name:      "operators",
			input:     "@{};,:<>[]()=?::",
			wantCount: 16,
			validate: func(t *testing.T, tokens []Token) {
				expected := []string{"@", "{", "}", ";", ",", ":", "<", ">", "[", "]", "(", ")", "=", "?", "::"}
				for i, exp := range expected {
					testutil.Equals(t, tokens[i].Text, exp)
				}
			},
		},
		{
			name:      "double colon",
			input:     "Foo::Bar",
			wantCount: 4,
			validate: func(t *testing.T, tokens []Token) {
				testutil.Equals(t, tokens[0].Text, "Foo")
				testutil.Equals(t, tokens[1].Text, "::")
				testutil.Equals(t, tokens[2].Text, "Bar")
			},
		},
		{
			name:      "line comment",
			input:     "foo // comment\nbar",
			wantCount: 3,
			validate: func(t *testing.T, tokens []Token) {
				testutil.Equals(t, tokens[0].Text, "foo")
				testutil.Equals(t, tokens[1].Text, "bar")
			},
		},
		{
			name:      "block comment",
			input:     "foo /* comment */ bar",
			wantCount: 3,
			validate: func(t *testing.T, tokens []Token) {
				testutil.Equals(t, tokens[0].Text, "foo")
				testutil.Equals(t, tokens[1].Text, "bar")
			},
		},
		{
			name:      "whitespace",
			input:     "  \t\n\rfoo  \t\n\r  bar  ",
			wantCount: 3,
			validate: func(t *testing.T, tokens []Token) {
				testutil.Equals(t, tokens[0].Text, "foo")
				testutil.Equals(t, tokens[1].Text, "bar")
			},
		},
		{
			name:      "string with escapes",
			input:     `"hello\nworld"`,
			wantCount: 2,
			validate: func(t *testing.T, tokens []Token) {
				testutil.Equals(t, tokens[0].Type, TokenString)
				testutil.Equals(t, tokens[0].Text, `"hello\nworld"`)
			},
		},
		{
			name:      "position tracking",
			input:     "foo\nbar",
			wantCount: 3,
			validate: func(t *testing.T, tokens []Token) {
				testutil.Equals(t, tokens[0].Pos.Line, 1)
				testutil.Equals(t, tokens[0].Pos.Column, 1)
				testutil.Equals(t, tokens[1].Pos.Line, 2)
				testutil.Equals(t, tokens[1].Pos.Column, 1)
			},
		},
		{
			name:      "single colon operator",
			input:     ":",
			wantCount: 2,
			validate: func(t *testing.T, tokens []Token) {
				testutil.Equals(t, tokens[0].Type, TokenOperator)
				testutil.Equals(t, tokens[0].Text, ":")
			},
		},
		{
			name:      "dot operator",
			input:     ".",
			wantCount: 2,
			validate: func(t *testing.T, tokens []Token) {
				testutil.Equals(t, tokens[0].Type, TokenOperator)
				testutil.Equals(t, tokens[0].Text, ".")
			},
		},
		{
			name:      "slash without comment",
			input:     "/x",
			wantCount: 3,
			validate: func(t *testing.T, tokens []Token) {
				testutil.Equals(t, tokens[0].Type, TokenUnknown)
			},
		},
		{
			name:      "long identifier",
			input:     "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_",
			wantCount: 2,
			validate: func(t *testing.T, tokens []Token) {
				testutil.Equals(t, tokens[0].Text, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_")
			},
		},
		{
			name:      "identifier starting with underscore",
			input:     "_foo",
			wantCount: 2,
			validate: func(t *testing.T, tokens []Token) {
				testutil.Equals(t, tokens[0].Type, TokenIdent)
				testutil.Equals(t, tokens[0].Text, "_foo")
			},
		},
		{
			name:      "number not valid identifier start",
			input:     "123",
			wantCount: 4,
			validate: func(t *testing.T, tokens []Token) {
				testutil.Equals(t, tokens[0].Type, TokenUnknown)
			},
		},
		{
			name:      "empty lines",
			input:     "\n\n\nfoo\n\n",
			wantCount: 2,
			validate: func(t *testing.T, tokens []Token) {
				testutil.Equals(t, tokens[0].Text, "foo")
			},
		},
		{
			name:      "line comment at end",
			input:     "foo // comment",
			wantCount: 2,
			validate: func(t *testing.T, tokens []Token) {
				testutil.Equals(t, tokens[0].Text, "foo")
			},
		},
		{
			name:      "multiple block comments",
			input:     "foo /* comment */ bar /* another */",
			wantCount: 3,
			validate: func(t *testing.T, tokens []Token) {
				testutil.Equals(t, tokens[0].Text, "foo")
				testutil.Equals(t, tokens[1].Text, "bar")
			},
		},
		{
			name:      "unknown operator",
			input:     "~",
			wantCount: 2,
			validate: func(t *testing.T, tokens []Token) {
				testutil.Equals(t, tokens[0].Type, TokenUnknown)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tokens, err := Tokenize("", []byte(tt.input))
			testutil.OK(t, err)
			testutil.Equals(t, len(tokens), tt.wantCount)
			if tt.validate != nil {
				tt.validate(t, tokens)
			}
		})
	}
}

func TestTokenizeErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"unterminated string", `"hello`},
		{"unterminated block comment", "/* unterminated"},
		{"string with escape at end", `"hello\`},
		{"string with newline", "\"hello\nworld\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := Tokenize("", []byte(tt.input))
			testutil.Equals(t, err != nil, true)
		})
	}
}

func TestTokenStringValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		token     Token
		wantValue string
		wantErr   bool
	}{
		{
			name:      "simple string",
			token:     Token{Type: TokenString, Text: `"hello"`},
			wantValue: "hello",
		},
		{
			name:      "string with escapes",
			token:     Token{Type: TokenString, Text: `"hello\nworld"`},
			wantValue: "hello\nworld",
		},
		{
			name:    "non-string token",
			token:   Token{Type: TokenIdent, Text: "foo"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			val, err := tt.token.stringValue()
			if tt.wantErr {
				testutil.Equals(t, err != nil, true)
			} else {
				testutil.OK(t, err)
				testutil.Equals(t, val, tt.wantValue)
			}
		})
	}
}

func TestTokenMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		token  Token
		method string
		want   bool
	}{
		{"isEOF true", Token{Type: TokenEOF}, "isEOF", true},
		{"isEOF false", Token{Type: TokenIdent}, "isEOF", false},
		{"isIdent true", Token{Type: TokenIdent}, "isIdent", true},
		{"isIdent false", Token{Type: TokenString}, "isIdent", false},
		{"isString true", Token{Type: TokenString}, "isString", true},
		{"isString false", Token{Type: TokenIdent}, "isString", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var result bool
			switch tt.method {
			case "isEOF":
				result = tt.token.isEOF()
			case "isIdent":
				result = tt.token.isIdent()
			case "isString":
				result = tt.token.isString()
			}
			testutil.Equals(t, result, tt.want)
		})
	}
}
