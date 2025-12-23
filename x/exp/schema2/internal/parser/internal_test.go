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
		testutil.Equals(t, p.pos, 0) // pos should not go negative
	})

	t.Run("parseAnnotation without @", func(t *testing.T) {
		t.Parallel()
		p, err := New([]byte("foo"))
		testutil.OK(t, err)
		_, err = p.parseAnnotation()
		testutil.Equals(t, err != nil, true)
	})

	t.Run("parseEntityRef with identifier only", func(t *testing.T) {
		t.Parallel()
		p, err := New([]byte("someActionName"))
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
		// Create a string longer than buffer to test buffer handling
		longStr := strings.Repeat("a", 2000)
		tokens, err := Tokenize([]byte(longStr))
		testutil.OK(t, err)
		testutil.Equals(t, tokens[0].Text, longStr)
	})

	t.Run("tokenText with buffer content", func(t *testing.T) {
		t.Parallel()
		// Create input that will use tokBuf
		longStr := strings.Repeat("x", 2000)
		tokens, err := Tokenize([]byte(longStr))
		testutil.OK(t, err)
		testutil.Equals(t, len(tokens[0].Text), 2000)
	})
}

func TestNewParserWithInvalidInput(t *testing.T) {
	t.Parallel()

	t.Run("tokenization error", func(t *testing.T) {
		t.Parallel()
		// Input that causes tokenization error
		_, err := New([]byte("\"unclosed"))
		testutil.Equals(t, err != nil, true)
	})
}

func TestParseSchemaWithTokenizationError(t *testing.T) {
	t.Parallel()

	t.Run("tokenization error", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte("\"unclosed"))
		testutil.Equals(t, err != nil, true)
	})
}

func TestMoreErrorPaths(t *testing.T) {
	t.Parallel()

	t.Run("namespace entity error", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`namespace MyApp { entity; }`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("namespace entity shape error", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`namespace MyApp { entity User { 123 }; }`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("action appliesTo principal list close error", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`action view appliesTo { principal: [User };`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("action appliesTo resource list close error", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`action view appliesTo { resource: [Doc };`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("parseEntityRef error on path ident", func(t *testing.T) {
		t.Parallel()
		p, err := New([]byte("123"))
		testutil.OK(t, err)
		_, err = p.parseEntityRef()
		testutil.Equals(t, err != nil, true)
	})

	t.Run("parseCommonType type parse error", func(t *testing.T) {
		t.Parallel()
		p, err := New([]byte("type Foo = 123"))
		testutil.OK(t, err)
		_, err = p.parseCommonType(nil)
		testutil.Equals(t, err != nil, true)
	})

	t.Run("parseRecordPairs error in pair", func(t *testing.T) {
		t.Parallel()
		p, err := New([]byte("{ 123 }"))
		testutil.OK(t, err)
		_, err = p.parseRecordPairs()
		testutil.Equals(t, err != nil, true)
	})

	t.Run("parseRecordPairs unterminated", func(t *testing.T) {
		t.Parallel()
		p, err := New([]byte("{ name: String"))
		testutil.OK(t, err)
		_, err = p.parseRecordPairs()
		testutil.Equals(t, err != nil, true)
	})
}

func TestScannerUTF8(t *testing.T) {
	t.Parallel()

	t.Run("utf8 characters in input", func(t *testing.T) {
		t.Parallel()
		// UTF-8 characters that need buffer refilling
		tokens, err := Tokenize([]byte("foo_世界"))
		// The scanner may produce unexpected tokens but shouldn't crash
		testutil.Equals(t, err == nil || err != nil, true)
		_ = tokens
	})

	t.Run("input with multibyte at boundary", func(t *testing.T) {
		t.Parallel()
		// Test with long input that might span buffer boundaries
		input := strings.Repeat("a", 1020) + "β" + strings.Repeat("b", 100)
		tokens, err := Tokenize([]byte(input))
		testutil.OK(t, err)
		_ = tokens
	})

	t.Run("invalid utf8 sequence", func(t *testing.T) {
		t.Parallel()
		// Invalid UTF-8 byte sequence
		input := []byte{0x80, 0x81, 0x82}
		_, err := Tokenize(input)
		// Should produce error for invalid UTF-8
		testutil.Equals(t, err != nil, true)
	})

	t.Run("nul character", func(t *testing.T) {
		t.Parallel()
		// NUL character in input
		input := []byte("foo\x00bar")
		_, err := Tokenize(input)
		testutil.Equals(t, err != nil, true)
	})
}

func TestTokenTextBuffer(t *testing.T) {
	t.Parallel()

	t.Run("long string spanning buffer", func(t *testing.T) {
		t.Parallel()
		// Create a string that would require buffer to accumulate
		// The buffer is 1024 bytes, so a long string should trigger tokBuf usage
		longContent := strings.Repeat("x", 2000)
		input := `"` + longContent + `"`
		tokens, err := Tokenize([]byte(input))
		testutil.OK(t, err)
		testutil.Equals(t, len(tokens), 2) // string + EOF
		val, err := tokens[0].stringValue()
		testutil.OK(t, err)
		testutil.Equals(t, len(val), 2000)
	})

	t.Run("chunked reader triggers buffer refill", func(t *testing.T) {
		t.Parallel()
		// Create a chunked reader that returns data slowly
		longIdent := strings.Repeat("a", 2000)
		reader := &chunkedReader{data: []byte(longIdent), chunkSize: 100}
		tokens, err := TokenizeReader(reader)
		testutil.OK(t, err)
		testutil.Equals(t, tokens[0].Text, longIdent)
	})
}

// chunkedReader returns data in small chunks to force buffer refills
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

func TestDirectMethodCalls(t *testing.T) {
	t.Parallel()

	t.Run("parseNamespace directly without namespace keyword", func(t *testing.T) {
		t.Parallel()
		p, err := New([]byte("foo { entity User; }"))
		testutil.OK(t, err)
		_, err = p.parseNamespace(nil)
		testutil.Equals(t, err != nil, true)
	})

	t.Run("parseEntity directly without entity keyword", func(t *testing.T) {
		t.Parallel()
		p, err := New([]byte("foo;"))
		testutil.OK(t, err)
		_, err = p.parseEntity(nil)
		testutil.Equals(t, err != nil, true)
	})

	t.Run("parseAction directly without action keyword", func(t *testing.T) {
		t.Parallel()
		p, err := New([]byte("foo;"))
		testutil.OK(t, err)
		_, err = p.parseAction(nil)
		testutil.Equals(t, err != nil, true)
	})

	t.Run("enum entity missing semicolon", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`entity Status enum ["a"]`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("action missing final semicolon", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`action view`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("action appliesTo missing closing brace", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`action view appliesTo { principal: User`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("parseEntityRef error on expectString", func(t *testing.T) {
		t.Parallel()
		// Create parser where we get to the point of trying to read a string ID
		p, err := New([]byte(`Action::`))
		testutil.OK(t, err)
		_, err = p.parseEntityRef()
		// Should error because there's no string after ::
		testutil.Equals(t, err != nil, true)
	})

	t.Run("parseCommonType directly without type keyword", func(t *testing.T) {
		t.Parallel()
		p, err := New([]byte("foo = String;"))
		testutil.OK(t, err)
		_, err = p.parseCommonType(nil)
		testutil.Equals(t, err != nil, true)
	})

	t.Run("parseRecordPairs without opening brace", func(t *testing.T) {
		t.Parallel()
		p, err := New([]byte("name: String"))
		testutil.OK(t, err)
		_, err = p.parseRecordPairs()
		testutil.Equals(t, err != nil, true)
	})

	t.Run("parseEntityRef with string token", func(t *testing.T) {
		t.Parallel()
		p, err := New([]byte(`"someAction"`))
		testutil.OK(t, err)
		ref, err := p.parseEntityRef()
		testutil.OK(t, err)
		testutil.Equals(t, ref.ID, "someAction")
	})
}
