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
	RegisterLanguage(".go", sitter.NewLanguage(golang.Language()))
	RegisterLanguage(".ts", sitter.NewLanguage(typescript.LanguageTypescript()))
	RegisterLanguage(".tsx", sitter.NewLanguage(typescript.LanguageTSX()))
	RegisterLanguage(".js", sitter.NewLanguage(javascript.Language()))
	RegisterLanguage(".rb", sitter.NewLanguage(ruby.Language()))
	RegisterLanguage(".py", sitter.NewLanguage(python.Language()))
	RegisterLanguage(".rs", sitter.NewLanguage(rust.Language()))
	RegisterLanguage(".html", sitter.NewLanguage(html.Language()))
	RegisterLanguage(".css", sitter.NewLanguage(css.Language()))
	RegisterLanguage(".sh", sitter.NewLanguage(bash.Language()))
	RegisterLanguage(".bash", sitter.NewLanguage(bash.Language()))
	RegisterLanguage(".c", sitter.NewLanguage(c.Language()))
	RegisterLanguage(".h", sitter.NewLanguage(c.Language()))
	RegisterLanguage(".cpp", sitter.NewLanguage(cpp.Language()))
	RegisterLanguage(".cc", sitter.NewLanguage(cpp.Language()))
	RegisterLanguage(".hpp", sitter.NewLanguage(cpp.Language()))
	RegisterLanguage(".cs", sitter.NewLanguage(csharp.Language()))
	RegisterLanguage(".java", sitter.NewLanguage(java.Language()))
	RegisterLanguage(".ml", sitter.NewLanguage(ocaml.LanguageOCaml()))
	RegisterLanguage(".mli", sitter.NewLanguage(ocaml.LanguageOCamlInterface()))
	RegisterLanguage(".php", sitter.NewLanguage(php.LanguagePHP()))
	RegisterLanguage(".scala", sitter.NewLanguage(scala.Language()))
	RegisterLanguage(".sc", sitter.NewLanguage(scala.Language()))
}
