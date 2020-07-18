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
		replace.String("foo", "bar"),
		replace.Bytes([]byte("thing"), []byte("test")),
		replace.RegexpString(regexp.MustCompile(`\d+`), "a number")
	))

	_, _ = io.Copy(os.Stdout, r)
}
```
