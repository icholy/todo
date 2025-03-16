//go:build !no_treesitter_grammars

package todo

import (
	treesitter "github.com/tree-sitter/go-tree-sitter"
	bash "github.com/tree-sitter/tree-sitter-bash/bindings/go"
	csharp "github.com/tree-sitter/tree-sitter-c-sharp/bindings/go"
	c "github.com/tree-sitter/tree-sitter-c/bindings/go"
	cpp "github.com/tree-sitter/tree-sitter-cpp/bindings/go"
	css "github.com/tree-sitter/tree-sitter-css/bindings/go"
	golang "github.com/tree-sitter/tree-sitter-go/bindings/go"
	html "github.com/tree-sitter/tree-sitter-html/bindings/go"
	java "github.com/tree-sitter/tree-sitter-java/bindings/go"
	javascript "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	ocaml "github.com/tree-sitter/tree-sitter-ocaml/bindings/go"
	php "github.com/tree-sitter/tree-sitter-php/bindings/go"
	python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	ruby "github.com/tree-sitter/tree-sitter-ruby/bindings/go"
	rust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
	scala "github.com/tree-sitter/tree-sitter-scala/bindings/go"
	typescript "github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

func init() {
	RegisterLanguage(LanguageOptions{
		Name:       "Golang",
		Language:   treesitter.NewLanguage(golang.Language()),
		Extensions: []string{".go"},
	})
	RegisterLanguage(LanguageOptions{
		Name:       "TypeScript",
		Language:   treesitter.NewLanguage(typescript.LanguageTypescript()),
		Extensions: []string{".ts"},
	})
	RegisterLanguage(LanguageOptions{
		Name:       "TypeScript TSX",
		Language:   treesitter.NewLanguage(typescript.LanguageTSX()),
		Extensions: []string{".tsx"},
	})
	RegisterLanguage(LanguageOptions{
		Name:       "JavaScript",
		Language:   treesitter.NewLanguage(javascript.Language()),
		Extensions: []string{".js"},
	})
	RegisterLanguage(LanguageOptions{
		Name:       "Ruby",
		Language:   treesitter.NewLanguage(ruby.Language()),
		Extensions: []string{".rb"},
	})
	RegisterLanguage(LanguageOptions{
		Name:       "Rust",
		Language:   treesitter.NewLanguage(rust.Language()),
		Extensions: []string{".rs"},
	})
	RegisterLanguage(LanguageOptions{
		Name:       "Python",
		Language:   treesitter.NewLanguage(python.Language()),
		Extensions: []string{".py"},
	})
	RegisterLanguage(LanguageOptions{
		Name:       "HTML",
		Language:   treesitter.NewLanguage(html.Language()),
		Extensions: []string{".html"},
	})
	RegisterLanguage(LanguageOptions{
		Name:       "CSS",
		Language:   treesitter.NewLanguage(css.Language()),
		Extensions: []string{".css"},
	})
	RegisterLanguage(LanguageOptions{
		Name:       "Bash",
		Language:   treesitter.NewLanguage(bash.Language()),
		Extensions: []string{".sh", ".bash"},
	})
	RegisterLanguage(LanguageOptions{
		Name:       "C",
		Language:   treesitter.NewLanguage(c.Language()),
		Extensions: []string{".c", ".h"},
	})
	RegisterLanguage(LanguageOptions{
		Name:       "C++",
		Language:   treesitter.NewLanguage(cpp.Language()),
		Extensions: []string{".cpp", ".cc", ".hpp"},
	})
	RegisterLanguage(LanguageOptions{
		Name:       "C#",
		Language:   treesitter.NewLanguage(csharp.Language()),
		Extensions: []string{".cs"},
	})
	RegisterLanguage(LanguageOptions{
		Name:       "Java",
		Language:   treesitter.NewLanguage(java.Language()),
		Extensions: []string{".java"},
	})
	RegisterLanguage(LanguageOptions{
		Name:       "OCaml",
		Language:   treesitter.NewLanguage(ocaml.LanguageOCaml()),
		Extensions: []string{".ml"},
	})
	RegisterLanguage(LanguageOptions{
		Name:       "OCaml Interface",
		Language:   treesitter.NewLanguage(ocaml.LanguageOCamlInterface()),
		Extensions: []string{".mli"},
	})
	RegisterLanguage(LanguageOptions{
		Name:       "PHP",
		Language:   treesitter.NewLanguage(php.LanguagePHP()),
		Extensions: []string{".php"},
	})
	RegisterLanguage(LanguageOptions{
		Name:       "Scala",
		Language:   treesitter.NewLanguage(scala.Language()),
		Extensions: []string{".scala", ".sc"},
	})
}
