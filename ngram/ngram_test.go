package ngram

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenise(t *testing.T) {
	tokens, err := Tokenize(1, "input string")
	fmt.Printf("%q\n", tokens)
	assert.NoError(t, err)
	assert.Len(t, tokens, 2)
}
