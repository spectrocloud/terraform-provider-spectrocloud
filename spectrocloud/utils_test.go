package spectrocloud

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestStringContains(t *testing.T) {
	ss := []string{"fizz_1", "bazz", "random", "nfizz_1", "fizz_2"}

	contains := stringContains(ss, "random")
	assert.Equal(t, true, contains, "Should be true.")

	assert.Equal(t, false, stringContains(ss, "doesnt"), "Should be false.")
}

func TestFilter(t *testing.T) {
	ss := []string{"fizz_1", "bazz", "random", "nfizz_1", "fizz_2"}

	mytest := func(s string) bool { return !strings.HasPrefix(s, "fizz_") && len(s) <= 7 }
	s3 := filter(ss, mytest)

	assert.Equal(t, 3, len(s3), "The two len should be the same.")
}

func TestIsMapSubset(t *testing.T) {
	a := map[string]string{"a": "b", "c": "d", "e": "f"}
	b := map[string]string{"a": "b", "e": "f"}
	c := map[string]string{"a": "b", "e": "g"}

	assert.Equal(t, true, IsMapSubset(a, b))
	assert.Equal(t, false, IsMapSubset(a, c))
	assert.Equal(t, false, IsMapSubset(b, a)) // a bigger than b
}
