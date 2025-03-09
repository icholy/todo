# TODO

> A library for parsing structured TODO comments out of code.

## Implementation

* Tree-Sitter is used to parse comments out of source files.
* TODOs are extracted from the comments using a recursive descent parser.

## Syntax

Each valid TODO line **must** begin with `TODO`. It may optionally include a comma‐separated attribute list in parentheses immediately following `TODO`, but **must** always include a colon (`:`). Everything after the colon is treated as the description.

### Basic Structure

```
TODO(...) : description
```

Without attributes, the line still requires the colon:

```
TODO: description
```

### Attributes

- Attributes appear inside parentheses right after `TODO`.
- They are separated by commas.  
- Each attribute can be:

1. **Bare Key** - (ex: `2025-03-06`)
2. **Unquoted Key–Value**  - (ex: `key=value`)
3. **Quoted Key–Value**  - (ex: `author="John Doe"`)

Within quoted values, `\"` is interpreted as `"` and `\\` as a literal backslash `\`.

### Examples

```
// TODO: no attributes
// TODO(foo,bar): 2 key-only attributes
// TODO(created=2025-03-09, assigned=john): multiple key/value attributes
// TODO(deadline="June 2025"): quoted value 
```

## CLI Tool

This repository comes with a very minimal tool for testing this package.
It parses TODO comments from source files outputs them in JSON format.

### Install

```
go install github.com/icholy/todo/cmd/todo@latest
```

### Usage

```
todo **/*.go
```
