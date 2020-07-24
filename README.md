# Streaming text replacement

[![GoDoc](https://godoc.org/github.com/icholy/replace?status.svg)](https://godoc.org/github.com/icholy/replace)

> This package provides a [x/text/transform.Transformer](https://godoc.org/golang.org/x/text/transform#Transformer)
> implementation that replaces text

## Example

``` go
package main

import (
	"io"
	"os"
	"regexp"

	"github.com/icholy/replace"
	"golang.org/x/text/transform"
)

func main() {
	f, _ := os.Open("file")
	defer f.Close()

	r := transform.NewReader(f, transform.Chain(
		// simple replace
		replace.String("foo", "bar"),
		replace.Bytes([]byte("thing"), []byte("test")),

		// remove all words that start with baz
		replace.Regexp(regexp.MustCompile(`baz\w*`), nil),

		// surround all words with parentheses
		replace.RegexpString(regexp.MustCompile(`\w+`), "($0)"),

		// increment all numbers
		replace.RegexpStringFunc(regexp.MustCompile(`\d+`), func(match string) string {
			x, _ := strconv.Atoi(match)
			return strconv.Itoa(x+1)
		}),
	))

	_, _ = io.Copy(os.Stdout, r)
}
```

## Notes Regexp* functions

* The `replace` functions should not save or modify any `[]byte` parameters they recieve.
* For better performance, reduce the `MaxSourceBuffer` size to the largest possible match (Default 64kb).
* If a match is longer than `MaxSourceBuffer` it may be skipped.