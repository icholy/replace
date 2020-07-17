# Streaming text replacement

> This package provides a x/text/transform.Transformer
> implementation that replaces text

## Example

``` go
package main

import (
	"io"
	"os"

	"github.com/icholy/replace"
	"golang.org/x/text/transform"
)

func main() {
	f, _ := os.Open("file")
	defer f.Close()

	r := transform.NewReader(f, transform.Chain(
		replace.String("foo", "bar"),
		replace.String("thing", "test"),
	))

	_, _ = io.Copy(os.Stdout, r)
}
```
