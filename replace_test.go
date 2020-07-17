package replace

import (
	"strconv"
	"strings"
	"testing"

	"golang.org/x/text/transform"
	"gotest.tools/v3/assert"
)

func TestTransformer(t *testing.T) {
	tests := []struct {
		in       string
		old, new string
		out      string
	}{
		{"a", "a", "b", "b"},
		{"yes", "yes", "no", "no"},
		{"what what what", "what", "wut", "wut wut wut"},
		{"???????", "?", "*", "*******"},
		{"no matches", "x", "y", "no matches"},
		{"hello", "l", "L", "heLLo"},
		{"hello", "x", "X", "hello"},
		{"", "x", "X", ""},
		{"radar", "r", "<r>", "<r>ada<r>"},
		{"banana", "a", "<>", "b<>n<>n<>"},
		{"banana", "an", "<>", "b<><>a"},
		{"banana", "ana", "<>", "b<>na"},
		{"banana", "a", "a", "banana"},
		{"xxx", "x", "", ""},
		{strings.Repeat("foo_", 8<<10), "foo", "bar", strings.Repeat("bar_", 8<<10)},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			tr := String(tt.old, tt.new)
			result, _, err := transform.String(tr, tt.in)
			assert.NilError(t, err)
			assert.Equal(t, result, tt.out)
		})
	}
}
