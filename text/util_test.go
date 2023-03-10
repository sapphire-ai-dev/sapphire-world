package text

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertEmptyNotNil(t *testing.T, a any) {
	assert.Empty(t, a)
	assert.NotNil(t, a)
}
