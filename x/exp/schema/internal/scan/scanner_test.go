package scan

import (
	"testing"
)

func TestTokenKindString(t *testing.T) {
	tests := []struct {
		kind TokenKind
		want string
	}{
		{TokenEOF, "EOF"},
		{TokenIdent, "identifier"},
		{TokenString, "string"},
		{TokenDoubleColon, "::"},
		{TokenLBrace, "{"},
		{TokenRBrace, "}"},
		{TokenLBracket, "["},
		{TokenRBracket, "]"},
		{TokenLAngle, "<"},
		{TokenRAngle, ">"},
		{TokenLParen, "("},
		{TokenRParen, ")"},
		{TokenComma, ","},
		{TokenSemicolon, ";"},
		{TokenColon, ":"},
		{TokenQuestion, "?"},
		{TokenEquals, "="},
		{TokenAt, "@"},
		{TokenKind(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.kind.String(); got != tt.want {
			t.Errorf("TokenKind(%d).String() = %q, want %q", tt.kind, got, tt.want)
		}
	}
}

func TestScannerTokens(t *testing.T) {
	src := `entity User in [Group] { name: String };`
	s := New([]byte(src))
	var kinds []TokenKind
	for {
		tok, err := s.Next()
		if err != nil {
			t.Fatal(err)
		}
		kinds = append(kinds, tok.Kind)
		if tok.Kind == TokenEOF {
			break
		}
	}
	want := []TokenKind{
		TokenIdent, TokenIdent, TokenIdent, TokenLBracket, TokenIdent, TokenRBracket,
		TokenLBrace, TokenIdent, TokenColon, TokenIdent, TokenRBrace, TokenSemicolon, TokenEOF,
	}
	if len(kinds) != len(want) {
		t.Fatalf("got %d tokens, want %d", len(kinds), len(want))
	}
	for i := range kinds {
		if kinds[i] != want[i] {
			t.Errorf("token %d: got %s, want %s", i, kinds[i], want[i])
		}
	}
}

func TestScannerSymbols(t *testing.T) {
	src := `{}[]<>(),;:?=@::`
	s := New([]byte(src))
	want := []TokenKind{
		TokenLBrace, TokenRBrace, TokenLBracket, TokenRBracket,
		TokenLAngle, TokenRAngle, TokenLParen, TokenRParen,
		TokenComma, TokenSemicolon, TokenColon, TokenQuestion,
		TokenEquals, TokenAt, TokenDoubleColon, TokenEOF,
	}
	for _, w := range want {
		tok, err := s.Next()
		if err != nil {
			t.Fatal(err)
		}
		if tok.Kind != w {
			t.Errorf("got %s, want %s", tok.Kind, w)
		}
	}
}

func TestScannerIdent(t *testing.T) {
	src := `_foo bar123 Baz`
	s := New([]byte(src))
	wantTexts := []string{"_foo", "bar123", "Baz"}
	for _, want := range wantTexts {
		tok, err := s.Next()
		if err != nil {
			t.Fatal(err)
		}
		if tok.Kind != TokenIdent {
			t.Fatalf("got %s, want identifier", tok.Kind)
		}
		if tok.Text != want || tok.Value != want {
			t.Errorf("got text=%q value=%q, want %q", tok.Text, tok.Value, want)
		}
	}
}

func TestScannerString(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{`"hello"`, "hello"},
		{`"with \"quotes\""`, `with "quotes"`},
		{`"line\nbreak"`, "line\nbreak"},
		{`"tab\there"`, "tab\there"},
		{`"null\0char"`, "null\x00char"},
	}
	for _, tt := range tests {
		s := New([]byte(tt.src))
		tok, err := s.Next()
		if err != nil {
			t.Errorf("scanning %q: %v", tt.src, err)
			continue
		}
		if tok.Kind != TokenString {
			t.Errorf("scanning %q: got %s, want string", tt.src, tok.Kind)
			continue
		}
		if tok.Value != tt.want {
			t.Errorf("scanning %q: value = %q, want %q", tt.src, tok.Value, tt.want)
		}
	}
}

func TestScannerUnterminatedString(t *testing.T) {
	s := New([]byte(`"hello`))
	_, err := s.Next()
	if err == nil {
		t.Fatal("expected error for unterminated string")
	}
}

func TestScannerInvalidChar(t *testing.T) {
	s := New([]byte(`~`))
	_, err := s.Next()
	if err == nil {
		t.Fatal("expected error for unexpected character")
	}
}

func TestScannerComments(t *testing.T) {
	src := "// line comment\nfoo /* block */ bar"
	s := New([]byte(src))
	tok1, err := s.Next()
	if err != nil {
		t.Fatal(err)
	}
	if tok1.Text != "foo" {
		t.Errorf("got %q, want foo", tok1.Text)
	}
	tok2, err := s.Next()
	if err != nil {
		t.Fatal(err)
	}
	if tok2.Text != "bar" {
		t.Errorf("got %q, want bar", tok2.Text)
	}
}

func TestScannerPosition(t *testing.T) {
	src := "ab\ncd"
	s := New([]byte(src))
	tok, _ := s.Next()
	if tok.Pos.Line != 1 || tok.Pos.Column != 1 {
		t.Errorf("first token: got %d:%d, want 1:1", tok.Pos.Line, tok.Pos.Column)
	}
	tok, _ = s.Next()
	if tok.Pos.Line != 2 || tok.Pos.Column != 1 {
		t.Errorf("second token: got %d:%d, want 2:1", tok.Pos.Line, tok.Pos.Column)
	}
}

func TestScannerColon(t *testing.T) {
	src := `: ::`
	s := New([]byte(src))
	tok1, _ := s.Next()
	if tok1.Kind != TokenColon {
		t.Errorf("got %s, want :", tok1.Kind)
	}
	tok2, _ := s.Next()
	if tok2.Kind != TokenDoubleColon {
		t.Errorf("got %s, want ::", tok2.Kind)
	}
}

func TestQuote(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", `"hello"`},
		{`quote"here`, `"quote\"here"`},
		{"back\\slash", `"back\\slash"`},
		{"new\nline", `"new\nline"`},
		{"car\rret", `"car\rret"`},
		{"tab\there", `"tab\there"`},
		{"null\x00char", `"null\0char"`},
		{"\x01ctrl", `"\u{1}ctrl"`},
	}
	for _, tt := range tests {
		got := Quote(tt.input)
		if got != tt.want {
			t.Errorf("Quote(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestScannerEOF(t *testing.T) {
	s := New([]byte(""))
	tok, err := s.Next()
	if err != nil {
		t.Fatal(err)
	}
	if tok.Kind != TokenEOF {
		t.Errorf("got %s, want EOF", tok.Kind)
	}
}

func TestScannerBlockCommentUnterminated(t *testing.T) {
	s := New([]byte("/* unterminated"))
	tok, err := s.Next()
	if err != nil {
		t.Fatal(err)
	}
	if tok.Kind != TokenEOF {
		t.Errorf("got %s, want EOF after unterminated block comment", tok.Kind)
	}
}

func TestScannerStringEscape(t *testing.T) {
	s := New([]byte(`"a\\b"`))
	tok, err := s.Next()
	if err != nil {
		t.Fatal(err)
	}
	if tok.Value != `a\b` {
		t.Errorf("got %q, want %q", tok.Value, `a\b`)
	}
}

func TestScannerStringEscapeAtEnd(t *testing.T) {
	// Backslash at end of string (before closing quote not found)
	s := New([]byte(`"\`))
	_, err := s.Next()
	if err == nil {
		t.Fatal("expected error for backslash at end")
	}
}

func TestScannerInvalidStringEscape(t *testing.T) {
	s := New([]byte(`"\q"`))
	_, err := s.Next()
	if err == nil {
		t.Fatal("expected error for invalid escape")
	}
}

func TestScannerStringRawText(t *testing.T) {
	s := New([]byte(`"hello"`))
	tok, err := s.Next()
	if err != nil {
		t.Fatal(err)
	}
	if tok.Text != `"hello"` {
		t.Errorf("Text = %q, want %q", tok.Text, `"hello"`)
	}
}
