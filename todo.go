package todo

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	treesitter "github.com/tree-sitter/go-tree-sitter"
)

var (
	languagesMu sync.Mutex
	languages   = map[string]*LanguageOptions{}
)

// LanguageOptions are treesitter options for a language.
type LanguageOptions struct {
	Name       string
	Extensions []string
	Language   *treesitter.Language
	Queries    []*treesitter.Query
}

// RegisterLanguage registers a language with the given extensions.
// If no queries are provided, the default queries will be used.
func RegisterLanguage(opt LanguageOptions) {
	languagesMu.Lock()
	defer languagesMu.Unlock()
	if len(opt.Queries) == 0 {
		names := []string{"comment", "line_comment", "block_comment"}
		for _, name := range names {
			query, err := treesitter.NewQuery(
				opt.Language,
				fmt.Sprintf(`(%s) @comment (#match? @comment "TODO")`, name),
			)
			if err == nil {
				opt.Queries = append(opt.Queries, query)
			}
		}
	}
	if len(opt.Queries) == 0 {
		panic(fmt.Sprintf("no queries for language: %s", opt.Name))
	}
	for _, ext := range opt.Extensions {
		languages[ext] = &opt
	}
}

// LanguageFor returns the language for the given file name.
func LanguageFor(file string) (*LanguageOptions, bool) {
	languagesMu.Lock()
	defer languagesMu.Unlock()
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
func ParseCode(ctx context.Context, file string, source []byte, opt *LanguageOptions) ([]Todo, error) {
	if opt == nil {
		var ok bool
		opt, ok = LanguageFor(file)
		if !ok {
			return nil, fmt.Errorf("no language for file: %s", file)
		}
	}
	var todos []Todo
	parser := treesitter.NewParser()
	defer parser.Close()
	parser.SetLanguage(opt.Language)
	tree := parser.Parse(source, nil)
	defer tree.Close()
	cursor := treesitter.NewQueryCursor()
	defer cursor.Close()
	for _, query := range opt.Queries {
		captures := cursor.Captures(query, tree.RootNode(), source)
		for {
			m, index := captures.Next()
			if m == nil {
				break
			}
			node := m.Captures[index].Node
			row := node.StartPosition().Row
			comment := source[node.StartByte():node.EndByte()]
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
