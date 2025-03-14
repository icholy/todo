package todo

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"unicode"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/bash"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/css"
	"github.com/smacker/go-tree-sitter/dockerfile"
	"github.com/smacker/go-tree-sitter/elixir"
	"github.com/smacker/go-tree-sitter/elm"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/hcl"
	"github.com/smacker/go-tree-sitter/html"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/kotlin"
	"github.com/smacker/go-tree-sitter/lua"
	"github.com/smacker/go-tree-sitter/ocaml"
	"github.com/smacker/go-tree-sitter/php"
	"github.com/smacker/go-tree-sitter/protobuf"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/scala"
	"github.com/smacker/go-tree-sitter/sql"
	"github.com/smacker/go-tree-sitter/svelte"
	"github.com/smacker/go-tree-sitter/swift"
	"github.com/smacker/go-tree-sitter/toml"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	"github.com/smacker/go-tree-sitter/yaml"
)

var languages = map[string]*sitter.Language{
	".go":         golang.GetLanguage(),
	".ts":         typescript.GetLanguage(),
	".tsx":        tsx.GetLanguage(),
	".js":         javascript.GetLanguage(),
	".rb":         ruby.GetLanguage(),
	".py":         python.GetLanguage(),
	".rs":         rust.GetLanguage(),
	".html":       html.GetLanguage(),
	".css":        css.GetLanguage(),
	".sh":         bash.GetLanguage(),
	".bash":       bash.GetLanguage(),
	".c":          c.GetLanguage(),
	".h":          c.GetLanguage(),
	".cpp":        cpp.GetLanguage(),
	".cc":         cpp.GetLanguage(),
	".hpp":        cpp.GetLanguage(),
	".cs":         csharp.GetLanguage(),
	".dockerfile": dockerfile.GetLanguage(),
	".ex":         elixir.GetLanguage(),
	".exs":        elixir.GetLanguage(),
	".elm":        elm.GetLanguage(),
	".tf":         hcl.GetLanguage(),
	".hcl":        hcl.GetLanguage(),
	".java":       java.GetLanguage(),
	".kt":         kotlin.GetLanguage(),
	".kts":        kotlin.GetLanguage(),
	".ml":         ocaml.GetLanguage(),
	".mli":        ocaml.GetLanguage(),
	".php":        php.GetLanguage(),
	".proto":      protobuf.GetLanguage(),
	".scala":      scala.GetLanguage(),
	".sc":         scala.GetLanguage(),
	".sql":        sql.GetLanguage(),
	".svelte":     svelte.GetLanguage(),
	".swift":      swift.GetLanguage(),
	".toml":       toml.GetLanguage(),
	".yaml":       yaml.GetLanguage(),
	".yml":        yaml.GetLanguage(),
	".lua":        lua.GetLanguage(),
}

// LanguageFor returns the language for the given file name
func LanguageFor(file string) (*sitter.Language, bool) {
	l, ok := languages[filepath.Ext(file)]
	return l, ok
}

// Attribute represents a key=value pair
type Attribute struct {
	Key   string
	Value string
	Quote bool
}

// Location represents a file location
type Location struct {
	File string
	Line int
}

// Todo represents a TODO line
type Todo struct {
	Line        string
	Location    Location
	Description string
	Attributes  []Attribute
}

// Parse parses the source and returns all TODO comments.
func Parse(ctx context.Context, file string, source []byte) ([]Todo, error) {
	if lang, ok := LanguageFor(file); ok {
		return ParseCode(ctx, file, source, lang)
	}
	return ParseText(file, string(source)), nil
}

// ParseCode parses the source code and returns all TODO comments
func ParseCode(ctx context.Context, file string, source []byte, lang *sitter.Language) ([]Todo, error) {
	if lang == nil {
		var ok bool
		lang, ok = LanguageFor(file)
		if !ok {
			return nil, fmt.Errorf("no language for file: %s", file)
		}
	}
	var todos []Todo
	parser := sitter.NewParser()
	parser.SetLanguage(lang)
	tree, err := parser.ParseCtx(ctx, nil, source)
	if err != nil {
		return nil, err
	}
	query, err := sitter.NewQuery([]byte("(comment) @comment"), lang)
	if err != nil {
		return nil, err
	}
	cursor := sitter.NewQueryCursor()
	cursor.Exec(query, tree.RootNode())
	for {
		m, ok := cursor.NextMatch()
		if !ok {
			break
		}
		m = cursor.FilterPredicates(m, source)
		for _, c := range m.Captures {
			row := c.Node.StartPoint().Row
			comment := c.Node.Content(source)
			for _, todo := range ParseText(file, comment) {
				todo.Location.Line += int(row)
				todos = append(todos, todo)
			}
		}
	}
	return todos, nil
}

// ParseText parses a text string and returns all TODO comments
func ParseText(file, text string) []Todo {
	var todos []Todo
	var row int
	for line := range strings.Lines(text) {
		line = strings.TrimSuffix(line, "\n")
		if todo, ok := ParseLine(line); ok {
			todo.Line = line
			todo.Location = Location{
				File: file,
				Line: row + 1,
			}
			todos = append(todos, todo)
		}
		row++
	}
	return todos
}

// ParseLine parses a single TODO line.
// Does not set the Location or Line fields.
func ParseLine(line string) (Todo, bool) {
	var t Todo
	// ignore everything up to the first TODO
	_, line, ok := strings.Cut(line, "TODO")
	if !ok {
		return t, false
	}
	br := bufio.NewReader(strings.NewReader(line))
	// After "TODO", optional attributes in parentheses
	if err := skipWhite(br); err != nil && !errors.Is(err, io.EOF) {
		return t, false
	}
	if peekByte(br) == '(' {
		if err := parseAttributes(br, &t); err != nil {
			return t, false
		}
	}
	// Skip whitespace
	if err := skipWhite(br); err != nil && !errors.Is(err, io.EOF) {
		return t, false
	}
	// Check for a colon
	if peekByte(br) != ':' {
		return t, false
	}
	// Consume the colon
	br.ReadByte()
	// Skip whitespace after colon
	if err := skipWhite(br); err != nil && !errors.Is(err, io.EOF) {
		return t, false
	}
	// Remainder is the description
	description, _ := io.ReadAll(br)
	t.Description = string(bytes.TrimSpace(description))
	return t, true
}

// parseAttributes consumes '(' ... ')' which may contain comma-separated attributes.
func parseAttributes(br *bufio.Reader, t *Todo) error {
	// consume '('
	if b, err := br.ReadByte(); err != nil || b != '(' {
		return errors.New("expected '('")
	}
	for {
		if err := skipWhite(br); err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		// check for ')'
		p, err := br.Peek(1)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		if len(p) == 1 && p[0] == ')' {
			br.ReadByte() // consume ')'
			return nil
		}
		// parse one attribute
		attr, err := parseOneAttribute(br)
		if err != nil {
			return err
		}
		t.Attributes = append(t.Attributes, attr)
		// after attribute, maybe ',' or ')'
		if err := skipWhite(br); err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		p, err = br.Peek(1)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		if len(p) == 1 && p[0] == ',' {
			br.ReadByte() // consume ','
			continue
		}
	}
}

// parseOneAttribute handles either:
//   - foo      (meaning Key="foo", no value)
//   - foo=bar  (unquoted)
//   - foo="bar" (quoted)
//   - etc.
func parseOneAttribute(br *bufio.Reader) (Attribute, error) {
	attr := Attribute{}
	// read the "key" portion, up to ',', ')', '=' or whitespace
	token, err := readUntilAny(br, []rune{',', ')', '='})
	if err != nil {
		return attr, err
	}
	attr.Key = strings.TrimSpace(token)
	// Now check what we hit: could be '=', ',', ')', or EOF
	r, err := br.Peek(1)
	if err != nil && !errors.Is(err, io.EOF) {
		return attr, err
	}
	if len(r) == 0 {
		// EOF => attribute has only a key
		return attr, nil
	}
	switch r[0] {
	case '=', ' ', '\t', '\n', '\r':
		// If it's '=', consume it & parse value. Otherwise, it might be whitespace (skip).
		if r[0] == '=' {
			br.ReadByte() // consume '='
			if err := skipWhite(br); err != nil && !errors.Is(err, io.EOF) {
				return attr, err
			}
			val, quote, err := parseValue(br)
			if err != nil {
				return attr, err
			}
			attr.Value = val
			attr.Quote = quote
		}
		// If it's whitespace, we skip it—but check the next character if it is '=' or not
		if unicode.IsSpace(rune(r[0])) {
			br.ReadByte() // consume the whitespace
			if err := skipWhite(br); err != nil && !errors.Is(err, io.EOF) {
				return attr, err
			}
			if peekByte(br) == '=' {
				br.ReadByte() // consume '='
				if err := skipWhite(br); err != nil && !errors.Is(err, io.EOF) {
					return attr, err
				}
				val, quote, e := parseValue(br)
				if e != nil {
					return attr, e
				}
				attr.Value = val
				attr.Quote = quote
			}
		}
	case ',', ')':
		// Means no value is present, so Key alone is the attribute
		// We'll leave Value = "" and Quote = false
	default:
		// Some unexpected character
		return attr, errors.New("unexpected character in attribute list")
	}
	return attr, nil
}

// parseValue checks if next is a quoted or unquoted value
func parseValue(br *bufio.Reader) (string, bool, error) {
	if peekByte(br) == '"' {
		// parse quoted
		v, err := parseQuotedValue(br)
		return v, true, err
	}
	// parse unquoted
	v, err := parseValueUnquoted(br)
	return v, false, err
}

// parseQuotedValue consumes an initial quote, reads until matching unescaped quote.
func parseQuotedValue(br *bufio.Reader) (string, error) {
	var sb strings.Builder
	// opening quote
	b, err := br.ReadByte()
	if err != nil {
		return "", err
	}
	if b != '"' {
		return "", errors.New("expected opening quote")
	}
	for {
		r, _, err := br.ReadRune()
		if err != nil {
			return "", err
		}
		if r == '\\' {
			// handle escape
			nxt, _, err := br.ReadRune()
			if err != nil {
				return "", err
			}
			switch nxt {
			case '\\':
				sb.WriteRune('\\')
			case '"':
				sb.WriteRune('"')
			default:
				// unrecognized escape, keep both
				sb.WriteRune(r)
				sb.WriteRune(nxt)
			}
			continue
		}
		if r == '"' {
			// end of quoted
			break
		}
		sb.WriteRune(r)
	}
	return sb.String(), nil
}

// parseValueUnquoted reads until ',', ')' or whitespace. It doesn't consume the stopping rune.
func parseValueUnquoted(br *bufio.Reader) (string, error) {
	var sb strings.Builder
	for {
		r, _, err := br.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", err
		}
		if r == ',' || r == ')' || unicode.IsSpace(r) {
			_ = br.UnreadRune()
			break
		}
		sb.WriteRune(r)
	}
	return strings.TrimSpace(sb.String()), nil
}

// readUntilAny reads until we hit one of the given runes, then unreads that rune.
// Used to grab an attribute key up to '=', ',', or ')'.
func readUntilAny(br *bufio.Reader, stop []rune) (string, error) {
	var sb strings.Builder
loop:
	for {
		r, size, err := br.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return sb.String(), nil
			}
			return "", err
		}
		for _, s := range stop {
			if r == s {
				// put it back
				_ = br.UnreadRune()
				break loop
			}
		}
		sb.WriteRune(r)
		// Also break if we read a whitespace (we can skip it later).
		if size > 0 && unicode.IsSpace(r) {
			break
		}
	}
	return sb.String(), nil
}

// skipWhite discards consecutive Unicode spaces, returning on the first non-space.
func skipWhite(br *bufio.Reader) error {
	for {
		r, _, err := br.ReadRune()
		if err != nil {
			return err
		}
		if !unicode.IsSpace(r) {
			_ = br.UnreadRune()
			return nil
		}
	}
}

func peekByte(br *bufio.Reader) byte {
	b, err := br.Peek(1)
	if err != nil {
		return 0
	}
	return b[0]
}
