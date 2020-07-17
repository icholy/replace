package replace

import (
	"bytes"

	"golang.org/x/text/transform"
)

// Transformer is a transformer that replaces text
// See: http://golang.org/x/text/transform
type Transformer struct {
	old, new []byte
	oldlen   int

	transform.NopResetter
}

var _ transform.Transformer = (*Transformer)(nil)

// Bytes returns a transformer that replaces all instances of old with new.
// Unlike bytes.Replace, empty old values don't match anything.
func Bytes(old, new []byte) Transformer {
	return Transformer{old: old, new: new, oldlen: len(old)}
}

// String returns a transformer that replaces all instances of old with new.
// Unlike strings.Replace, empty old values don't match anything.
func String(old, new string) Transformer {
	return Bytes([]byte(old), []byte(new))
}

// Transform implements golang.org/x/text/transform#Transformer
func (t Transformer) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	// don't do anything for empty old string. We're forced to do this because an optimization in
	// transform.String prevents us from generating any output when the src is empty.
	// see: https://github.com/golang/text/blob/master/transform/transform.go#L570-L576
	if t.oldlen == 0 {
		n, err := fullcopy(dst, src)
		return n, n, err
	}
	// make sure there's enough to even find a match
	if len(src) < t.oldlen {
		if atEOF {
			n, err := fullcopy(dst, src)
			return n, n, err
		}
		return 0, 0, transform.ErrShortSrc
	}
	// replace all instances of old with new
	for {
		i := bytes.Index(src[nSrc:], t.old)
		if i == -1 {
			break
		}
		// copy everything up to the match
		n1, err := fullcopy(dst[nDst:], src[nSrc:nSrc+i])
		if err != nil {
			return nDst, nSrc, err
		}
		// copy the new value
		n2, err := fullcopy(dst[nDst+i:], t.new)
		if err != nil {
			return nDst, nSrc, err
		}
		nDst += n1 + n2
		nSrc += i + len(t.old)
	}
	// skip everything except the trailing len(r.old) - 1
	if skip := len(src[nSrc:]) - t.oldlen + 1; skip > 0 {
		n, err := fullcopy(dst[nDst:], src[nSrc:nSrc+skip])
		nSrc += n
		nDst += n
		if err != nil {
			return nDst, nSrc, err
		}
	}
	// if we're at the end, tack on any remaining bytes
	if atEOF {
		n, err := fullcopy(dst[nDst:], src[nSrc:])
		nDst += n
		nSrc += n
		return nDst, nSrc, err
	}
	return nDst, nSrc, transform.ErrShortSrc
}

func fullcopy(dst, src []byte) (int, error) {
	n := copy(dst, src)
	if n < len(src) {
		return n, transform.ErrShortDst
	}
	return n, nil
}
