package main

// CompletionHandler provides possible completions for given input
type CompletionHandler func(input string) []string

// DefaultCompletionHandler simply returns an empty slice.
var DefaultCompletionHandler = func(input string) []string {
	return make([]string, 0)
}

var complHandler = DefaultCompletionHandler

// SetCompletionHandler sets the CompletionHandler to be used for completion
func SetCompletionHandler(c CompletionHandler) {
	complHandler = c
}
