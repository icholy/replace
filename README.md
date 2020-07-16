# Streaming text replacement

> This package provides a x/text/transform.Transformer
> implementation that replaces text

## Example

``` go

import (
  "os"
  "io"

  "github.com/icholy/replacer"
	"golang.org/x/text/transform"
)

func main() {
  f, _ := os.Open("file")
  defer f.Close()

  r := transform.NewReader(f, transform.Chain(
    replacer.New([]byte("foo"), []byte("bar")),
    replacer.New([]byte("thing") []byte("test")),
  ))

  _, _ = io.Copy(os.Stdout, r)
}

```
