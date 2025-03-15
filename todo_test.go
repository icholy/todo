package todo

import (
	"reflect"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name   string
		file   string
		source []byte
		lang   *sitter.Language
		want   []Todo
	}{
		{
			name:   "simple",
			file:   "test.go",
			source: []byte("// TODO: fix this\n"),
			want: []Todo{
				{
					Line: "// TODO: fix this",
					Location: Location{
						File: "test.go",
						Line: 1,
					},
					Description: "fix this",
				},
			},
		},
		{
			name:   "infer typescript",
			file:   "code.ts",
			source: []byte("/* \n TODO: does this work ?\n */"),
			want: []Todo{
				{
					Line: " TODO: does this work ?",
					Location: Location{
						File: "code.ts",
						Line: 2,
					},
					Description: "does this work ?",
				},
			},
		},
		{
			name:   "non treesitter",
			file:   "some.txt",
			source: []byte("// TODO(): fix this\nTODO: fix this again"),
			want: []Todo{
				{
					Line: "// TODO(): fix this",
					Location: Location{
						File: "some.txt",
						Line: 1,
					},
					Description: "fix this",
				},
				{
					Line: "TODO: fix this again",
					Location: Location{
						File: "some.txt",
						Line: 2,
					},
					Description: "fix this again",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(t.Context(), tt.file, tt.source)
			if err != nil {
				t.Fatalf("Parse(%q, %q, lang) error = %v", tt.file, tt.source, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse(%q, %q, lang) = %#v, want %#v", tt.file, tt.source, got, tt.want)
			}
		})
	}
}

func TestParseLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		ok   bool
		want Todo
	}{
		{
			name: "simple",
			line: "TODO: fix this",
			ok:   true,
			want: Todo{
				Description: "fix this",
				Attributes:  nil,
			},
		},
		{
			name: "not a todo",
			line: "This is just a comment",
			ok:   false,
		},
		{
			name: "do not match TODO without colon",
			line: "TODO fix this",
			ok:   false,
		},
		{
			name: "empty attribute list",
			line: "TODO(): fix this",
			ok:   true,
			want: Todo{
				Description: "fix this",
			},
		},
		{
			name: "with attributes",
			line: "TODO(created=2025-03-09,assigned=john): fix this",
			ok:   true,
			want: Todo{
				Description: "fix this",
				Attributes: []Attribute{
					{Key: "created", Value: "2025-03-09"},
					{Key: "assigned", Value: "john"},
				},
			},
		},
		{
			name: "quoted attribute value",
			line: `TODO(message="fix this, that, and the other"): implement feature`,
			ok:   true,
			want: Todo{
				Description: "implement feature",
				Attributes: []Attribute{
					{Key: "message", Value: "fix this, that, and the other", Quote: true},
				},
			},
		},
		{
			name: "mixed quoted and unquoted attributes",
			line: `TODO(created=2023-01-01,message="complex, value)"): do something`,
			ok:   true,
			want: Todo{
				Description: "do something",
				Attributes: []Attribute{
					{Key: "created", Value: "2023-01-01"},
					{Key: "message", Value: "complex, value)", Quote: true},
				},
			},
		},
		{
			name: "escaped quotes in attribute",
			line: `TODO(message="value with \"escaped\" quotes"): task`,
			ok:   true,
			want: Todo{
				Description: "task",
				Attributes: []Attribute{
					{Key: "message", Value: `value with "escaped" quotes`, Quote: true},
				},
			},
		},
		{
			name: "escaped backslashes in attribute",
			line: `TODO(path="C:\\Program Files\\App"): update path`,
			ok:   true,
			want: Todo{
				Description: "update path",
				Attributes: []Attribute{
					{Key: "path", Value: `C:\Program Files\App`, Quote: true},
				},
			},
		},
		{
			name: "attribute without value",
			line: `TODO(key, 2025-03-06, author=icholy): description`,
			ok:   true,
			want: Todo{
				Description: "description",
				Attributes: []Attribute{
					{Key: "key"},
					{Key: "2025-03-06"},
					{Key: "author", Value: "icholy"},
				},
			},
		},
		{
			name: "extra whitespace",
			line: `   TODO (key = value, key2 =  "value" ) : description`,
			ok:   true,
			want: Todo{
				Description: "description",
				Attributes: []Attribute{
					{Key: "key", Value: "value"},
					{Key: "key2", Value: "value", Quote: true},
				},
			},
		},
		{
			name: "ignore everything before 'TODO'",
			line: "# // * --- TODO: fix this",
			ok:   true,
			want: Todo{
				Description: "fix this",
				Attributes:  nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseLine(tt.line)
			if ok != tt.ok {
				t.Fatalf("ParseLine(%q) = got ok=%v, want ok=%v", tt.line, ok, tt.ok)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseLine(%q) = %#v, want %#v", tt.line, got, tt.want)
			}
		})
	}
}

func TestTodoString(t *testing.T) {
	tests := []struct {
		todo Todo
		want string
	}{
		{
			todo: Todo{
				Description: "fix this",
				Attributes: []Attribute{
					{Key: "created", Value: "2025-03-09"},
					{Key: "assigned", Value: "john"},
					{Key: "message", Value: "hello", Quote: true},
				},
			},
			want: `TODO(created=2025-03-09, assigned=john, message="hello"): fix this`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.todo.String(); got != tt.want {
				t.Errorf("Todo.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTodoAttribute(t *testing.T) {
	todo := Todo{
		Attributes: []Attribute{
			{Key: "created", Value: "2025-03-09"},
			{Key: "assigned", Value: "john"},
		},
	}
	tests := []struct {
		key   string
		value string
		ok    bool
	}{
		{
			key:   "created",
			value: "2025-03-09",
			ok:    true,
		},
		{
			key:   "assigned",
			value: "john",
			ok:    true,
		},
		{
			key:   "unknown",
			value: "",
			ok:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got, ok := todo.Attribute(tt.key)
			if ok != tt.ok {
				t.Fatalf("Todo.Attribute(%q) = got ok=%v, want ok=%v", tt.key, ok, tt.ok)
			}
			if got != tt.value {
				t.Errorf("Todo.Attribute(%q) = %q, want %q", tt.key, got, tt.value)
			}
		})
	}
}
