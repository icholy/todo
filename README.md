# TODO

A library for parsing structured TODO comments from code.

[![Go Reference](https://pkg.go.dev/badge/github.com/icholy/todo.svg)](https://pkg.go.dev/github.com/icholy/todo)

## Overview

- **Tree-Sitter** is used to parse comments from source files.
- A simple recursive descent parser then extracts and interprets `TODO` lines.

## Syntax

`TODO` can appear anywhere in a comment line; anything before `TODO` is ignored. Once `TODO` is found:

1. An optional comma-separated list of attributes can follow, enclosed in parentheses.
2. A colon (`:`) must appear next.
3. Everything after the colon is the description.

### Examples

```
// TODO: no attributes
// TODO(foo,bar): two key-only attributes
// TODO(created=2025-03-09, assigned=john): multiple key/value
// TODO(deadline="June 2025"): quoted value 
```

### Grammar

```
todo-line  ::= (any text) "TODO" [ "(" attributes? ")" ] ":" description
attributes ::= attribute [ "," attribute ]*
attribute  ::= bare-key | key-value
key-value  ::= bare-key "=" (bare-key | quoted-value)
description ::= (any text to end of line)
bare-key   ::= (any non-whitespace sequence without parentheses, commas, or '=')
quoted-value ::= "\"" (any text) "\""
```

## API Usage

``` go
package main

import (
	"os"
	"fmt"
	"context"

	"github.com/icholy/todo"
)

func main() {
	// read file source
	file := "./main.go"
	source, _ := os.ReadFile(source)

	// parse todos
	ctx := context.Background()
	todos, _ := todo.Parse(ctx, file, source)

	// print todos with deadlines
	for _, t := range todos {
		if _, ok := t.Attribute("deadline"); ok {
			fmt.Println(t)
		}
	}
}
```

## CLI Tool

A minimal CLI tool is provided to parse and output these comments as JSON.

### Installation

```
go install github.com/icholy/todo/cmd/todo@latest
```

### Usage Example

```
todo ./**/*.go
./todo.go:88 TODO: investigate compilation error
```
