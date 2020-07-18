package replace

import (
	"fmt"
	"regexp"
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
		{"test", "", "x", "test"},
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
			assert.DeepEqual(t, result, tt.out)
		})
	}
}

func TestRegex(t *testing.T) {
	tests := []struct {
		in  string
		re  string
		new string
		out string
	}{
		{"a", "a", "b", "b"},
		{"testing", ".*", "x", "x"},
		{strings.Repeat("--ax-- --bx--", 4<<10), `--\wx--`, "--xx--", strings.Repeat("--xx-- --xx--", 4<<10)},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			re := regexp.MustCompile(tt.re)
			tr := RegexpString(re, tt.new)
			result, _, err := transform.String(tr, tt.in)
			assert.NilError(t, err)
			assert.DeepEqual(t, result, tt.out)
		})
	}
}

func TestRegexStringSubmatchFunc(t *testing.T) {
	in := strings.Repeat("--1x-- --2x-- --3x--", 8<<10)
	re := regexp.MustCompile(`--(\d)x--`)
	tr := RegexpStringSubmatchFunc(re, func(match []string) string {
		x, _ := strconv.Atoi(match[1])
		return fmt.Sprintf("--%dx--", x-1)
	})
	result, _, err := transform.String(tr, in)
	assert.NilError(t, err)
	want := strings.Repeat("--0x-- --1x-- --2x--", 8<<10)
	assert.DeepEqual(t, result, want)
}
