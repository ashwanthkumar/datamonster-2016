package ngram

import (
	"errors"
	"regexp"
	"strings"
)

// Tokenize - Tokenize the input into n-gram tokens
func Tokenize(n int, input string) ([]string, error) {
	if n < 1 {
		return []string{}, errors.New("assertion failed: n > 0")
	}
	// TODO - Inject this from outside
	// Filter away non-alphanumeric words
	replaced := regexp.MustCompile("([^a-zA-Z0-9\\ ])").ReplaceAllString(input, "")
	words := strings.Split(replaced, " ")
	cleaned := words[:0]
	for _, word := range words {
		if len(word) > 0 {
			cleaned = append(cleaned, word)
		}
	}

	var tokens []string
	var next = 0
	for previous := 0; (next + n) <= len(cleaned); {
		ngram := strings.Join(cleaned[previous:previous+n], " ")
		tokens = append(tokens, ngram)
		previous++
		next++
	}

	return tokens, nil
}
