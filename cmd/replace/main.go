package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/icholy/replace"
	"golang.org/x/text/transform"
)

func main() {
	// parse flags
	var max int
	var old, new string
	flag.StringVar(&old, "old", "", "regular expression")
	flag.StringVar(&new, "new", "", "replacement expansion")
	flag.IntVar(&max, "max", 64<<10, "max match size")
	flag.Parse()
	// setup the transformer
	re, err := regexp.Compile(old)
	if err != nil {
		log.Fatalf("old: %v", err)
	}
	tr := replace.RegexpString(re, new)
	tr.MaxMatchSize = max
	// if there are no files, read from stdin
	if flag.NArg() == 0 {
		r := transform.NewReader(os.Stdin, tr)
		if _, err := io.Copy(os.Stdout, r); err != nil {
			log.Fatal(err)
		}
		return
	}
	// do replacements
	for _, name := range flag.Args() {
		if err := rewrite(name, tr); err != nil {
			log.Printf("%s: %v", name, err)
		}
	}
}

// rewrite the named file by passing it through the transformer.
// The rewritten content is first written to a temporary file in the same directory as the file.
// Once the transform is successful, the original file is replaced by the temporary file.
func rewrite(name string, t transform.Transformer) error {
	// open original file
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	// create temp file
	pattern := filepath.Base(name) + "-temp-*"
	tmp, err := ioutil.TempFile(filepath.Dir(name), pattern)
	if err != nil {
		return err
	}
	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()
	// replace while copying from f to tmp
	if _, err := io.Copy(tmp, transform.NewReader(f, t)); err != nil {
		return err
	}
	// make sure the tmp file was successfully written to
	if err := tmp.Close(); err != nil {
		return err
	}
	// close the file we're reading from
	if err := f.Close(); err != nil {
		return err
	}
	// overwrite the original file with the temp file
	return os.Rename(tmp.Name(), name)
}
