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
		in, out string
		tr      transform.Transformer
	}{
		{"test", "test", String("", "x")},
		{"a", "b", String("a", "b")},
		{"yes", "no", String("yes", "no")},
		{"what what what", "wut wut wut", String("what", "wut")},
		{"???????", "*******", String("?", "*")},
		{"no matches", "no matches", String("x", "y")},
		{"hello", "heLLo", String("l", "L")},
		{"hello", "hello", String("x", "X")},
		{"", "", String("x", "X")},
		{"radar", "<r>ada<r>", String("r", "<r>")},
		{"banana", "b<>n<>n<>", String("a", "<>")},
		{"banana", "b<><>a", String("an", "<>")},
		{"banana", "b<>na", String("ana", "<>")},
		{"banana", "banana", String("a", "a")},
		{"xxx", "", String("x", "")},
		{strings.Repeat("foo_", 8<<10), strings.Repeat("bar_", 8<<10), String("foo", "bar")},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result, _, err := transform.String(tt.tr, tt.in)
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
