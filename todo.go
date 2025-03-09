package todo

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"path/filepath"
	"slices"
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
	// "github.com/smacker/go-tree-sitter/lua"
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

	// TODO: investigate compilation error
	// ".lua":        lua.GetLanguage(),
}

// FileExtensions returns all supported file extensions
func FileExtensions() []string {
	return slices.Sorted(maps.Keys(languages))
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

// Parse parses the source code and returns all TODO comments
func Parse(ctx context.Context, file string, source []byte, lang *sitter.Language) ([]Todo, error) {
	if lang == nil {
		var ok bool
		lang, ok = languages[filepath.Ext(file)]
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
			for line := range strings.Lines(comment) {
				line = strings.TrimSuffix(line, "\n")
				if todo, ok := ParseLine(line); ok {
					todo.Line = line
					todo.Location = Location{
						File: file,
						Line: int(row + 1),
					}
					todos = append(todos, todo)
				}
				row++
			}
		}
	}
	return todos, nil
}

// ParseLine parses a single TODO line.
// Does not set the Location or Line fields.
func ParseLine(line string) (Todo, bool) {
	// ignore everything up to the first TODO
	_, line, ok := strings.Cut(line, "TODO")
	if !ok {
		return Todo{}, false
	}
	var t Todo
	br := bufio.NewReader(strings.NewReader(line))
	// After "TODO", optional attributes in parentheses
	if err := skipWhite(br); err != nil && !errors.Is(err, io.EOF) {
		return t, false
	}
	if peek, _ := br.Peek(1); len(peek) == 1 && peek[0] == '(' {
		if err := parseAttributes(br, &t); err != nil {
			return t, false
		}
	}
	// Skip whitespace
	if err := skipWhite(br); err != nil && !errors.Is(err, io.EOF) {
		return t, false
	}
	// Check for a colon
	p, _ := br.Peek(1)
	if len(p) == 0 || p[0] != ':' {
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

func parseAttributes(br *bufio.Reader, t *Todo) error {
	// Consume '('
	if b, err := br.ReadByte(); err != nil || b != '(' {
		return errors.New("expected '('")
	}
	for {
		if err := skipWhite(br); err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		// If next is ')', we're done
		p, err := br.Peek(1)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		if len(p) == 1 && p[0] == ')' {
			br.ReadByte() // consume ')'
			return nil
		}
		// Parse one attribute (key=value)
		attr, err := parseOneAttribute(br)
		if err != nil {
			return err
		}
		t.Attributes = append(t.Attributes, attr)
		// After an attribute, we may see ',' or ')'
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

func parseOneAttribute(br *bufio.Reader) (Attribute, error) {
	var attr Attribute
	// Parse key (read until '=' or whitespace/comma/parenthesis)
	key, err := readUntil(br, '=')
	if err != nil {
		return attr, err
	}
	attr.Key = strings.TrimSpace(key)
	// We should now have '='
	b, err := br.ReadByte()
	if err != nil {
		return attr, err
	}
	if b != '=' {
		return attr, errors.New("expected '='")
	}
	// Check next rune; if it's a quote, parse quoted value
	if err := skipWhite(br); err != nil && !errors.Is(err, io.EOF) {
		return attr, err
	}
	r, err := br.Peek(1)
	if err != nil && !errors.Is(err, io.EOF) {
		return attr, err
	}
	if len(r) == 1 && r[0] == '"' {
		attr.Quote = true
		val, e := parseQuotedValue(br)
		if e != nil {
			return attr, e
		}
		attr.Value = val
	} else {
		// Unquoted: read until ',' or ')' or whitespace
		val, e := readValueUnquoted(br)
		if e != nil {
			return attr, e
		}
		attr.Value = val
	}
	return attr, nil
}

// parseQuotedValue reads a string that begins with a quote until the matching
// unescaped quote, handling backslash escapes (\" => ", \\ => \).
func parseQuotedValue(br *bufio.Reader) (string, error) {
	var sb strings.Builder
	// Consume the opening quote
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
		// If backslash, consume next to interpret escapes
		if r == '\\' {
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
				// If it's not a special escape, just write both
				sb.WriteRune(r)
				sb.WriteRune(nxt)
			}
			continue
		}
		// If we see a closing quote, we're done
		if r == '"' {
			break
		}
		sb.WriteRune(r)
	}
	return sb.String(), nil
}

// readValueUnquoted reads until ',' or ')' but does not consume them.
func readValueUnquoted(br *bufio.Reader) (string, error) {
	var sb strings.Builder
	for {
		r, _, err := br.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", err
		}
		// If we see ',' or ')', put it back and return
		if r == ',' || r == ')' {
			_ = br.UnreadRune()
			break
		}
		sb.WriteRune(r)
	}
	return strings.TrimSpace(sb.String()), nil
}

// readUntil reads runes until we see the given target or a comma/parenthesis/EOF.
func readUntil(br *bufio.Reader, target rune) (string, error) {
	var sb strings.Builder
	for {
		r, _, err := br.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return sb.String(), nil
			}
			return "", err
		}
		if r == target {
			_ = br.UnreadRune()
			return sb.String(), nil
		}
		if r == ',' || r == ')' {
			_ = br.UnreadRune()
			return sb.String(), nil
		}
		sb.WriteRune(r)
	}
}

// skipWhite consumes consecutive Unicode spaces.
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
