package testdata

// TODO: This is a simple todo item
func Simple() {
	// Just a function with a TODO
}

// TODO(priority=high): Fix this critical issue
// Another comment
func WithPriority() {
	// Implementation missing
	
	// TODO(author="john", issue=123): Add proper implementation
}

/*
 * Multi-line comment with TODOs
 * TODO: Support multi-line comments
 * TODO(deadline="2025-04-01"): Complete before April
 */
func MultiLineComment() {
	/* TODO: Inline multi-line comment */
	return
}

// Regular comment, not a TODO
func NoTodo() {
	/* 
	 * Just a regular multi-line comment
	 */
}