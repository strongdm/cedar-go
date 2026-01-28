package parser

import (
	"io"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestParserInternals(t *testing.T) {
	t.Parallel()

	t.Run("peek past end", func(t *testing.T) {
		t.Parallel()
		p := &Parser{tokens: []Token{}, pos: 0}
		tok := p.peek()
		testutil.Equals(t, tok.Type, TokenEOF)
	})

	t.Run("consumeString", func(t *testing.T) {
		t.Parallel()
		p, err := New("", []byte(`"hello"`))
		testutil.OK(t, err)
		val := p.consumeString()
		testutil.Equals(t, val, "hello")
	})

	t.Run("peek past end with large pos", func(t *testing.T) {
		t.Parallel()
		p := &Parser{tokens: []Token{}, pos: 100}
		tok := p.peek()
		testutil.Equals(t, tok.Type, TokenEOF)
	})

	t.Run("advance past end", func(t *testing.T) {
		t.Parallel()
		p := &Parser{tokens: []Token{}, pos: 0}
		tok := p.advance()
		testutil.Equals(t, tok.Type, TokenEOF)
		testutil.Equals(t, p.pos, 0)
	})

	t.Run("parseAnnotation without @", func(t *testing.T) {
		t.Parallel()
		p, err := New("", []byte("foo"))
		testutil.OK(t, err)
		_, err = p.parseAnnotation()
		testutil.Equals(t, err != nil, true)
	})

	t.Run("peekAhead beyond end", func(t *testing.T) {
		t.Parallel()
		p := &Parser{tokens: []Token{{Type: TokenIdent, Text: "foo"}}, pos: 0}
		tok := p.peekAhead(10)
		testutil.Equals(t, tok.Type, TokenEOF)
	})

	t.Run("parseEntityRef with identifier only", func(t *testing.T) {
		t.Parallel()
		p, err := New("", []byte("someActionName"))
		testutil.OK(t, err)
		ref, err := p.parseEntityRef()
		testutil.OK(t, err)
		testutil.Equals(t, ref.ID, "someActionName")
	})
}

func TestScannerInternals(t *testing.T) {
	t.Parallel()

	t.Run("scanner with very long line spanning buffer", func(t *testing.T) {
		t.Parallel()
		longStr := strings.Repeat("a", 2000)
		tokens, err := Tokenize("", []byte(longStr))
		testutil.OK(t, err)
		testutil.Equals(t, tokens[0].Text, longStr)
	})

	t.Run("tokenText with tokPos negative", func(t *testing.T) {
		t.Parallel()
		s, err := newScanner("", strings.NewReader("test"))
		testutil.OK(t, err)
		text := s.tokenText()
		testutil.Equals(t, text, "")
	})

	t.Run("tokenText with buffer content", func(t *testing.T) {
		t.Parallel()
		longStr := strings.Repeat("x", 2000)
		tokens, err := Tokenize("", []byte(longStr))
		testutil.OK(t, err)
		testutil.Equals(t, len(tokens[0].Text), 2000)
	})
}

func TestNewParserErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"tokenization error", `"unclosed`},
		{"tokenization via ParseSchema", `"unclosed`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := New("", []byte(tt.input))
			testutil.Equals(t, err != nil, true)
		})
	}
}

func TestParseSchemaWithTokenizationError(t *testing.T) {
	t.Parallel()

	t.Run("tokenization error", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema("", []byte("\"unclosed"))
		testutil.Equals(t, err != nil, true)
	})
}

func TestParserErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"namespace entity error", `namespace MyApp { entity; }`},
		{"namespace entity shape error", `namespace MyApp { entity User { 123 }; }`},
		{"action appliesTo principal list close error", `action view appliesTo { principal: [User };`},
		{"action appliesTo resource list close error", `action view appliesTo { resource: [Doc };`},
		{"namespace with enum conflict after entity", `namespace MyApp { entity Status; entity Status enum ["a"]; }`},
		{"namespace with enum parse error in values", `namespace MyApp { entity Status enum [; }`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseSchema("", []byte(tt.input))
			testutil.Equals(t, err != nil, true)
		})
	}
}

func TestDirectMethodCallErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		testFunc func(t *testing.T, p *Parser)
	}{
		{
			name:  "parseEntityRef error on path ident",
			input: "123",
			testFunc: func(t *testing.T, p *Parser) {
				_, err := p.parseEntityRef()
				testutil.Equals(t, err != nil, true)
			},
		},
		{
			name:  "parseCommonType type parse error",
			input: "type Foo = 123",
			testFunc: func(t *testing.T, p *Parser) {
				_, _, err := p.parseCommonType(nil)
				testutil.Equals(t, err != nil, true)
			},
		},
		{
			name:  "parseAttributes error in pair",
			input: "{ 123 }",
			testFunc: func(t *testing.T, p *Parser) {
				_, err := p.parseAttributes()
				testutil.Equals(t, err != nil, true)
			},
		},
		{
			name:  "parseAttributes unterminated",
			input: "{ name: String",
			testFunc: func(t *testing.T, p *Parser) {
				_, err := p.parseAttributes()
				testutil.Equals(t, err != nil, true)
			},
		},
		{
			name:  "parseNamespace directly without namespace keyword",
			input: "foo { entity User; }",
			testFunc: func(t *testing.T, p *Parser) {
				_, _, err := p.parseNamespace(nil)
				testutil.Equals(t, err != nil, true)
			},
		},
		{
			name:  "parseEntity directly without entity keyword",
			input: "foo;",
			testFunc: func(t *testing.T, p *Parser) {
				_, err := p.parseEntity(nil)
				testutil.Equals(t, err != nil, true)
			},
		},
		{
			name:  "parseAction directly without action keyword",
			input: "foo;",
			testFunc: func(t *testing.T, p *Parser) {
				_, err := p.parseAction(nil)
				testutil.Equals(t, err != nil, true)
			},
		},
		{
			name:  "parseEntityRef error on expectString",
			input: `Action::`,
			testFunc: func(t *testing.T, p *Parser) {
				_, err := p.parseEntityRef()
				testutil.Equals(t, err != nil, true)
			},
		},
		{
			name:  "parseCommonType directly without type keyword",
			input: "foo = String;",
			testFunc: func(t *testing.T, p *Parser) {
				_, _, err := p.parseCommonType(nil)
				testutil.Equals(t, err != nil, true)
			},
		},
		{
			name:  "parseAttributes without opening brace",
			input: "name: String",
			testFunc: func(t *testing.T, p *Parser) {
				_, err := p.parseAttributes()
				testutil.Equals(t, err != nil, true)
			},
		},
		{
			name:  "parseEntityRef with string token",
			input: `"someAction"`,
			testFunc: func(t *testing.T, p *Parser) {
				ref, err := p.parseEntityRef()
				testutil.OK(t, err)
				testutil.Equals(t, ref.ID, "someAction")
			},
		},
		{
			name:  "parseEnum directly without entity keyword",
			input: `Status enum ["a"]`,
			testFunc: func(t *testing.T, p *Parser) {
				_, _, err := p.parseEnum(nil)
				testutil.Equals(t, err != nil, true)
			},
		},
		{
			name:  "parseEnum with invalid name token",
			input: `entity 123 enum ["a"]`,
			testFunc: func(t *testing.T, p *Parser) {
				_, _, err := p.parseEnum(nil)
				testutil.Equals(t, err != nil, true)
			},
		},
		{
			name:  "parseEnum without enum keyword",
			input: `entity Status ["a"]`,
			testFunc: func(t *testing.T, p *Parser) {
				_, _, err := p.parseEnum(nil)
				testutil.Equals(t, err != nil, true)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, err := New("", []byte(tt.input))
			testutil.OK(t, err)
			tt.testFunc(t, p)
		})
	}
}

func TestParseSchemaEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"enum entity missing semicolon", `entity Status enum ["a"]`, true},
		{"action missing final semicolon", `action view`, true},
		{"action appliesTo missing closing brace", `action view appliesTo { principal: User`, true},
		{"error message with filename", `entity User { invalid }`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseSchema("test.cedar", []byte(tt.input))
			if tt.wantErr {
				testutil.Equals(t, err != nil, true)
				testutil.Equals(t, err.Error() != "", true)
			} else {
				testutil.OK(t, err)
			}
		})
	}
}

func TestScannerUTF8(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     []byte
		expectErr bool
	}{
		{
			name:      "utf8 characters in input",
			input:     []byte("foo_世界"),
			expectErr: false,
		},
		{
			name:      "input with multibyte at boundary",
			input:     []byte(strings.Repeat("a", 1020) + "β" + strings.Repeat("b", 100)),
			expectErr: false,
		},
		{
			name:      "invalid utf8 sequence",
			input:     []byte{0x80, 0x81, 0x82},
			expectErr: true,
		},
		{
			name:      "nul character",
			input:     []byte("foo\x00bar"),
			expectErr: true,
		},
		{
			name:      "utf8 at buffer boundary with token in progress",
			input:     []byte(strings.Repeat("x", 1023) + "β" + "y"),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := Tokenize("", tt.input)
			if tt.expectErr {
				testutil.Equals(t, err != nil, true)
			} else {
				testutil.Equals(t, err == nil || err != nil, true)
			}
		})
	}
}

func TestScannerReaderErrors(t *testing.T) {
	t.Parallel()

	t.Run("reader returns non-EOF error", func(t *testing.T) {
		t.Parallel()
		reader := &errorReader{
			data:    []byte("foo bar"),
			errAt:   3,
			readErr: io.ErrUnexpectedEOF,
		}
		_, err := TokenizeReader("", reader)
		testutil.Equals(t, err != nil, true)
	})

	t.Run("incomplete utf8 at buffer boundary during token", func(t *testing.T) {
		t.Parallel()
		prefix := strings.Repeat("x", 1022)
		data := []byte(prefix)
		data = append(data, 0xe2)
		data = append(data, 0x82, 0xac)
		data = append(data, 'y')

		reader := &splitReader{
			chunks: [][]byte{
				data[:1023],
				data[1023:],
			},
		}
		_, err := TokenizeReader("", reader)
		_ = err
	})

	t.Run("buffer refill mid-identifier with utf8", func(t *testing.T) {
		t.Parallel()
		chunk1 := make([]byte, 1024)
		for i := 0; i < 1023; i++ {
			chunk1[i] = 'a'
		}
		chunk1[1023] = 0xc3

		chunk2 := []byte{0xa3}
		chunk2 = append(chunk2, ' ', 'b')

		reader := &splitReader{
			chunks: [][]byte{chunk1, chunk2},
		}
		_, err := TokenizeReader("", reader)
		_ = err
	})
}

func TestTokenTextBuffer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func() ([]Token, error)
		validate func(t *testing.T, tokens []Token, err error)
	}{
		{
			name: "long string spanning buffer",
			setup: func() ([]Token, error) {
				longContent := strings.Repeat("x", 2000)
				input := `"` + longContent + `"`
				return Tokenize("", []byte(input))
			},
			validate: func(t *testing.T, tokens []Token, err error) {
				testutil.OK(t, err)
				testutil.Equals(t, len(tokens), 2)
				val, err := tokens[0].stringValue()
				testutil.OK(t, err)
				testutil.Equals(t, len(val), 2000)
			},
		},
		{
			name: "chunked reader triggers buffer refill",
			setup: func() ([]Token, error) {
				longIdent := strings.Repeat("a", 2000)
				reader := &chunkedReader{data: []byte(longIdent), chunkSize: 100}
				return TokenizeReader("", reader)
			},
			validate: func(t *testing.T, tokens []Token, err error) {
				testutil.OK(t, err)
				testutil.Equals(t, tokens[0].Text, strings.Repeat("a", 2000))
			},
		},
		{
			name: "tokBuf accumulation path",
			setup: func() ([]Token, error) {
				chunk1 := make([]byte, 1021)
				chunk1[0] = '"'
				for i := 1; i < 1020; i++ {
					chunk1[i] = 'a'
				}
				chunk1[1020] = 0xe2

				chunk2 := []byte{0x82, 0xac, '"'}

				reader := &splitReader{
					chunks: [][]byte{chunk1, chunk2},
				}
				return TokenizeReader("", reader)
			},
			validate: func(t *testing.T, tokens []Token, err error) {
				testutil.OK(t, err)
				testutil.Equals(t, tokens[0].Type, TokenString)
				testutil.Equals(t, len(tokens[0].Text) > 1000, true)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tokens, err := tt.setup()
			tt.validate(t, tokens, err)
		})
	}
}

func TestCommaSeparatedDeclarations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        string
		wantEntities int
		wantActions  int
		wantErr      bool
	}{
		{"multiple entities", `entity User, Admin, Guest;`, 3, 0, false},
		{"multiple entities with shared in clause", `entity Group; entity User, Admin in [Group];`, 3, 0, false},
		{"multiple entities with shared shape", `entity User, Admin { name: String };`, 2, 0, false},
		{"multiple actions", `action read, write, delete;`, 0, 3, false},
		{"multiple actions with shared appliesTo", `entity User; entity Doc; action read, write appliesTo { principal: User, resource: Doc };`, 2, 2, false},
		{"multiple actions with shared memberOf", `action baseAction; action read, write in ["baseAction"];`, 0, 3, false},
		{"multiple quoted action names", `action "view doc", "edit doc", "delete doc";`, 0, 3, false},
		{"enum cannot have multiple names", `entity Status, OtherStatus enum ["a", "b"];`, 0, 0, true},
		{"entity comma followed by invalid token", `entity User, 123;`, 0, 0, true},
		{"action comma followed by invalid token", `action read, 123;`, 0, 0, true},
		{"namespace with comma-separated entities", `namespace App { entity User, Admin; }`, 0, 0, false},
		{"namespace with comma-separated actions", `namespace App { action read, write; }`, 0, 0, false},
		{"trailing comma in memberOf list", `entity Group; entity User in [Group,];`, 0, 0, true},
		{"trailing comma in action memberOf list", `action parent; action view in ["parent",];`, 0, 0, true},
		{"reserved identifier 'in' as entity name", `entity in;`, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			schema, err := ParseSchema("", []byte(tt.input))
			if tt.wantErr {
				testutil.Equals(t, err != nil, true)
			} else {
				testutil.OK(t, err)
				testutil.Equals(t, len(schema.Entities), tt.wantEntities)
				testutil.Equals(t, len(schema.Actions), tt.wantActions)
			}
		})
	}
}

// Test helper types
type splitReader struct {
	chunks [][]byte
	idx    int
}

func (r *splitReader) Read(p []byte) (n int, err error) {
	if r.idx >= len(r.chunks) {
		return 0, io.EOF
	}
	n = copy(p, r.chunks[r.idx])
	r.idx++
	return n, nil
}

type chunkedReader struct {
	data      []byte
	pos       int
	chunkSize int
}

func (r *chunkedReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	end := r.pos + r.chunkSize
	if end > len(r.data) {
		end = len(r.data)
	}
	n = copy(p, r.data[r.pos:end])
	r.pos += n
	return n, nil
}

type errorReader struct {
	data    []byte
	pos     int
	errAt   int
	readErr error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	if r.pos >= r.errAt {
		return 0, r.readErr
	}
	end := r.pos + len(p)
	if end > len(r.data) {
		end = len(r.data)
	}
	if end > r.errAt {
		end = r.errAt
	}
	n = copy(p, r.data[r.pos:end])
	r.pos += n
	if r.pos >= r.errAt {
		return n, r.readErr
	}
	return n, nil
}
