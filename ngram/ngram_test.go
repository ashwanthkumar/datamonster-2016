package ngram

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenise(t *testing.T) {
	tokens, err := Tokenize(1, "input string")
	assert.NoError(t, err)
	assert.Len(t, tokens, 2)
	assert.Equal(t, tokens[0], "input")
	assert.Equal(t, tokens[1], "string")
}
