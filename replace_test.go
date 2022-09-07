package replace

import (
	"bytes"
	"fmt"
	"io"
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
		{
			in: "test", out: "test",
			tr: String("", "x"),
		},
		{
			in: "a", out: "b",
			tr: String("a", "b"),
		},
		{
			in: "yes", out: "no",
			tr: String("yes", "no"),
		},
		{
			in: "what what what", out: "wut wut wut",
			tr: String("what", "wut"),
		},
		{
			in: "???????", out: "*******",
			tr: String("?", "*"),
		},
		{
			in: "no matches", out: "no matches",
			tr: String("x", "y"),
		},
		{
			in: "hello", out: "heLLo",
			tr: String("l", "L"),
		},
		{
			in: "hello", out: "hello",
			tr: String("x", "X"),
		},
		{
			in: "", out: "",
			tr: String("x", "X"),
		},
		{
			in: "radar", out: "<r>ada<r>",
			tr: String("r", "<r>"),
		},
		{
			in: "banana", out: "b<>n<>n<>",
			tr: String("a", "<>"),
		},
		{
			in: "banana", out: "b<><>a",
			tr: String("an", "<>"),
		},
		{
			in: "banana", out: "b<>na",
			tr: String("ana", "<>"),
		},
		{
			in: "banana", out: "banana",
			tr: String("a", "a"),
		},
		{
			in: "xxx", out: "",
			tr: String("x", ""),
		},
		{
			in: strings.Repeat("foo_", 8<<10), out: strings.Repeat("bar_", 8<<10),
			tr: String("foo", "bar"),
		},
		{
			in: "a", out: "b",
			tr: RegexpString(regexp.MustCompile("a"), "b"),
		},
		{
			in: "testing", out: "x",
			tr: RegexpString(regexp.MustCompile(".*"), "x"),
		},
		{
			in: strings.Repeat("--ax-- --bx--", 4<<10), out: strings.Repeat("--xx-- --xx--", 4<<10),
			tr: RegexpString(regexp.MustCompile(`--\wx--`), "--xx--"),
		},
		{
			in: strings.Repeat("--1x-- --2x-- --3x--", 8<<10), out: strings.Repeat("--0x-- --1x-- --2x--", 8<<10),
			tr: RegexpStringSubmatchFunc(regexp.MustCompile(`--(\d)x--`), func(match []string) string {
				x, _ := strconv.Atoi(match[1])
				return fmt.Sprintf("--%dx--", x-1)
			}),
		},
		{
			in: "1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 ", out: strings.Repeat("num ", 16),
			tr: RegexpStringFunc(regexp.MustCompile(`\d+`), func(_ string) string {
				return "num"
			}),
		},
		{
			in: "bazzzz buzz foo what biz", out: "  foo what ",
			tr: Regexp(regexp.MustCompile(`b\w+z\w*`), nil),
		},
		{
			in: "a", out: "replaced",
			tr: RegexpSubmatchFunc(regexp.MustCompile("a(123)?"), func(_ [][]byte) []byte {
				return []byte("replaced")
			}),
		},
		{
			in: "this is a test", out: "(this) (is) (a) (test)",
			tr: RegexpString(regexp.MustCompile(`\w+`), "($0)"),
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result, _, err := transform.String(tt.tr, tt.in)
			assert.NilError(t, err)
			assert.DeepEqual(t, result, tt.out)
		})
	}
}

func TestOverflowDst(t *testing.T) {
	var calls int
	s := strings.Repeat("x", 8<<10)
	tr := RegexpStringFunc(regexp.MustCompile("x"), func(_ string) string {
		calls++
		return s
	})

	result, _, err := transform.String(tr, "x")
	assert.NilError(t, err)
	assert.Equal(t, result, s)
	assert.Equal(t, calls, 1)
}

func TestChain(t *testing.T) {
	var input bytes.Buffer
	input.WriteString(strings.Repeat("x", 1000))
	input.WriteString(strings.Repeat("y", 1000))
	input.WriteString(strings.Repeat("z", 1000))
	r := Chain(&input,
		RegexpString(regexp.MustCompile("x+"), "1"),
		RegexpString(regexp.MustCompile("y+"), "2"),
		RegexpString(regexp.MustCompile("z+"), "3"),
	)
	output, err := io.ReadAll(r)
	assert.NilError(t, err)
	assert.DeepEqual(t, string(output), "123")
}
