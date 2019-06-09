package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOne(t *testing.T) {
	assert.Equal(t, 1, 2)
}

func TestTwo(t *testing.T) {
	assert.Equal(t, 1, 1)
}

func TestThree(t *testing.T) {
	for i, v := range []string{"one", "two", "three"} {
		i := i
		v := v
		t.Run(fmt.Sprintf("%d - %s", i, v), func(t *testing.T) {
			assert.True(t, i < 2)
			assert.Equal(t, 3, len(v))
		})
	}
}

func TestFour(t *testing.T) {
	t.Skip()
}
