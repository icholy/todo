//go:build !no_treesitter_grammars

package todo

import (
	sitter "github.com/tree-sitter/go-tree-sitter"
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
	RegisterLanguage(sitter.NewLanguage(golang.Language()), ".go")
	RegisterLanguage(sitter.NewLanguage(typescript.LanguageTypescript()), ".ts")
	RegisterLanguage(sitter.NewLanguage(typescript.LanguageTSX()), ".tsx")
	RegisterLanguage(sitter.NewLanguage(javascript.Language()), ".js")
	RegisterLanguage(sitter.NewLanguage(ruby.Language()), ".rb")
	RegisterLanguage(sitter.NewLanguage(python.Language()), ".py")
	RegisterLanguage(sitter.NewLanguage(rust.Language()), ".rs")
	RegisterLanguage(sitter.NewLanguage(html.Language()), ".html")
	RegisterLanguage(sitter.NewLanguage(css.Language()), ".css")
	RegisterLanguage(sitter.NewLanguage(bash.Language()), ".sh")
	RegisterLanguage(sitter.NewLanguage(bash.Language()), ".bash")
	RegisterLanguage(sitter.NewLanguage(c.Language()), ".c")
	RegisterLanguage(sitter.NewLanguage(c.Language()), ".h")
	RegisterLanguage(sitter.NewLanguage(cpp.Language()), ".cpp")
	RegisterLanguage(sitter.NewLanguage(cpp.Language()), ".cc")
	RegisterLanguage(sitter.NewLanguage(cpp.Language()), ".hpp")
	RegisterLanguage(sitter.NewLanguage(csharp.Language()), ".cs")
	RegisterLanguage(sitter.NewLanguage(java.Language()), ".java")
	RegisterLanguage(sitter.NewLanguage(ocaml.LanguageOCaml()), ".ml")
	RegisterLanguage(sitter.NewLanguage(ocaml.LanguageOCamlInterface()), ".mli")
	RegisterLanguage(sitter.NewLanguage(php.LanguagePHP()), ".php")
	RegisterLanguage(sitter.NewLanguage(scala.Language()), ".scala")
	RegisterLanguage(sitter.NewLanguage(scala.Language()), ".sc")
}
