package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPass(t *testing.T) {
	assert.Equal(t, 1, 1)
}

func TestFail(t *testing.T) {
	assert.Equal(t, 1, 2)
}

func TestSubTests(t *testing.T) {
	for i, v := range []string{"one", "two", "three"} {
		i := i
		v := v
		t.Run(fmt.Sprintf("%d - %s", i, v), func(t *testing.T) {
			assert.True(t, i < 2)
			assert.Equal(t, 3, len(v))
		})
	}
}

func TestWithSkip(t *testing.T) {
	t.Skip()
}

func TestWithPanic(t *testing.T) {
	type Foo struct {
		t *time.Time
	}
	Foo{}.t.Format("2006")
}
