package replace

import (
	"bytes"
	"regexp"

	"golang.org/x/text/transform"
)

// Transformer replaces text in a stream
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
	// replace all instances of old with new
	for {
		i := bytes.Index(src[nSrc:], t.old)
		if i == -1 {
			break
		}
		// copy everything up to the match
		n, err := fullcopy(dst[nDst:], src[nSrc:nSrc+i])
		nSrc += n
		nDst += n
		if err != nil {
			return nDst, nSrc, err
		}
		// copy the new value
		n, err = fullcopy(dst[nDst:], t.new)
		if err != nil {
			return nDst, nSrc, err
		}
		nDst += n
		nSrc += t.oldlen
	}
	// if we're at the end, tack on any remaining bytes
	if atEOF {
		n, err := fullcopy(dst[nDst:], src[nSrc:])
		nDst += n
		nSrc += n
		return nDst, nSrc, err
	}
	// skip everything except the trailing len(r.old) - 1
	// we do this becasue there could be a match straddling
	// the boundary
	if skip := len(src[nSrc:]) - t.oldlen + 1; skip > 0 {
		n, err := fullcopy(dst[nDst:], src[nSrc:nSrc+skip])
		nSrc += n
		nDst += n
		if err != nil {
			return nDst, nSrc, err
		}
	}
	return nDst, nSrc, transform.ErrShortSrc
}

// RegexTransformer replaces regexp matches in a stream
// See: http://golang.org/x/text/transform
type RegexTransformer struct {
	transform.NopResetter
	// MaxSourceBuffer is the maximum size of the window used to search for the
	// regex match. (Default is 4mb).
	MaxSourceBuffer int

	re      *regexp.Regexp
	replace func(src []byte, index []int) []byte
}

var _ transform.Transformer = (*RegexTransformer)(nil)

// RegexBytes returns a transformer that replaces all matches of re with new
func RegexBytes(re *regexp.Regexp, new []byte) RegexTransformer {
	return RegexTransformer{
		re:              re,
		MaxSourceBuffer: 4 << 10,
		replace:         func(src []byte, index []int) []byte { return new },
	}
}

// RegexString returns a transformer that replaces all matches of re with new
func RegexString(re *regexp.Regexp, new []byte) RegexTransformer {
	return RegexBytes(re, []byte(new))
}

// RegexFunc returns a transformer that replaces all matches of re with the
// result of calling replace with the match. Replace may be called with the
// same match multiple times.
func RegexFunc(re *regexp.Regexp, replace func([]byte) []byte) RegexTransformer {
	return RegexTransformer{
		re:              re,
		MaxSourceBuffer: 4 << 10,
		replace: func(src []byte, index []int) []byte {
			match := make([]byte, index[1]-index[0])
			copy(match, src[index[0]:index[1]])
			return replace(match)
		},
	}
}

// RegexSubmatchFunc returns a transformer that replaces all matches of re with the
// result of calling replace with the submatch. Replace may be called with the
// same match multiple times.
func RegexSubmatch(re *regexp.Regexp, replace func([][]byte) []byte) RegexTransformer {
	return RegexTransformer{
		re:              re,
		MaxSourceBuffer: 4 << 10,
		replace: func(src []byte, index []int) []byte {
			match := make([][]byte, 1+re.NumSubexp())
			for i := range match {
				start, end := index[i*2], index[i*2+1]
				match[i] = make([]byte, end-start)
				copy(match[i], src[start:end])
			}
			return replace(match)
		},
	}
}

// Transform implements golang.org/x/text/transform#Transformer
func (t RegexTransformer) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	for _, index := range t.re.FindAllSubmatchIndex(src, -1) {
		// skip the match if it ends at the end the src buffer.
		// it could potentionally match more
		if index[1] == len(src)-1 && !atEOF {
			continue
		}
		// copy evertying up to the match
		n, err := fullcopy(dst[nDst:], src[nSrc:index[0]])
		nSrc += n
		nDst += n
		if err != nil {
			return nDst, nSrc, err
		}
		// copy the replacement
		n, err = fullcopy(dst[nDst:], t.replace(src, index))
		if err != nil {
			return nDst, nSrc, err
		}
		nDst += n
		nSrc = index[1]
	}
	// if we're at the end, tack on any remaining bytes
	if atEOF {
		n, err := fullcopy(dst[nDst:], src[nSrc:])
		nDst += n
		nSrc += n
		return nDst, nSrc, err
	}
	// skip any bytes which exceede the max source limit
	if skip := len(src[nSrc:]) - t.MaxSourceBuffer; skip > 0 {
		n, err := fullcopy(dst[nDst:], src[nSrc:nSrc+skip])
		nSrc += n
		nDst += n
		if err != nil {
			return nDst, nSrc, err
		}
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
