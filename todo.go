package todo

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

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

// LanguageFor returns the language for the given file name.
func LanguageFor(file string) (*sitter.Language, bool) {
	l, ok := languages[filepath.Ext(file)]
	return l, ok
}

// Attribute represents a key=value pair.
type Attribute struct {
	Key   string
	Value string
	Quote bool
}

// String returns a string representation.
func (a Attribute) String() string {
	if a.Value == "" {
		return a.Key
	}
	if a.Quote {
		return fmt.Sprintf("%s=%q", a.Key, a.Value)
	}
	return fmt.Sprintf("%s=%s", a.Key, a.Value)
}

// Location represents a file location.
type Location struct {
	File string
	Line int
}

// String returns a string representation of the location.
func (l Location) String() string {
	return fmt.Sprintf("%s:%d", l.File, l.Line)
}

// Todo represents a TODO line.
type Todo struct {
	Line        string
	Location    Location
	Description string
	Attributes  []Attribute
}

// Attribute returns the value for the given key.
func (t Todo) Attribute(key string) (string, bool) {
	for _, a := range t.Attributes {
		if a.Key == key {
			return a.Value, true
		}
	}
	return "", false
}

// String returns a string representation.
func (t Todo) String() string {
	var b strings.Builder
	b.WriteString("TODO")
	if len(t.Attributes) > 0 {
		b.WriteByte('(')
		for i, a := range t.Attributes {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(a.String())
		}
		b.WriteByte(')')
	}
	b.WriteString(": ")
	b.WriteString(t.Description)
	return b.String()
}

// Parse parses the source and returns all TODO comments.
func Parse(ctx context.Context, file string, source []byte) ([]Todo, error) {
	if lang, ok := LanguageFor(file); ok {
		return ParseCode(ctx, file, source, lang)
	}
	return ParseText(file, source), nil
}

// ParseCode parses the source code and returns all TODO comments.
// If lang is nil, the language is inferred from the file extension.
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
	defer parser.Close()
	parser.SetLanguage(lang)
	tree, err := parser.ParseCtx(ctx, nil, source)
	if err != nil {
		return nil, err
	}
	defer tree.Close()
	query, err := sitter.NewQuery([]byte("(comment) @comment"), lang)
	if err != nil {
		return nil, err
	}
	defer query.Close()
	cursor := sitter.NewQueryCursor()
	defer cursor.Close()
	cursor.Exec(query, tree.RootNode())
	for {
		m, ok := cursor.NextMatch()
		if !ok {
			break
		}
		m = cursor.FilterPredicates(m, source)
		for _, c := range m.Captures {
			row := c.Node.StartPoint().Row
			comment := source[c.Node.StartByte():c.Node.EndByte()]
			for _, todo := range ParseText(file, comment) {
				todo.Location.Line += int(row)
				todos = append(todos, todo)
			}
		}
	}
	return todos, nil
}

// ParseText parses a text string and returns all TODO comments.
func ParseText(file string, text []byte) []Todo {
	var todos []Todo
	row := 0
	scanner := bufio.NewScanner(bytes.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		if todo, ok := parseLine(line); ok {
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
