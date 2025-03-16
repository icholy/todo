package todo

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"unicode"
)

// parseLine parses a single TODO line.
// Does not set the Location or Line fields.
func parseLine(line []byte) (Todo, bool) {
	var t Todo
	// ignore everything up to the first TODO
	_, line, ok := bytes.Cut(line, []byte("TODO"))
	if !ok {
		return t, false
	}
	br := bufio.NewReader(bytes.NewReader(line))
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

// parseValue checks if next is a quoted or unquoted value.
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
