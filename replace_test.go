package replace

import (
	"io/ioutil"
	"strconv"
	"testing"

	"golang.org/x/text/transform"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"
)

func TestReplacer(t *testing.T) {
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
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			r := New([]byte(tt.old), []byte(tt.new))
			result, _, err := transform.String(r, tt.in)
			assert.NilError(t, err)
			assert.Equal(t, result, tt.out)
		})
	}
}

func TestReader(t *testing.T) {
	f := golden.Open(t, "input.txt")
	defer f.Close()

	r := transform.NewReader(f, New([]byte("foo"), []byte("bar")))
	data, err := ioutil.ReadAll(r)
	assert.NilError(t, err)
	golden.Assert(t, string(data), "output.txt")
}
