package textproto

import (
	"bytes"
	"errors"
	"math"
)

func (r *Reader) ReadEmailMIMEHeader() (MIMEHeader, error) {
	return readEmailMIMEHeader(r, math.MaxInt64)
}

func readEmailMIMEHeader(r *Reader, lim int64) (MIMEHeader, error) {
	// Avoid lots of small slice allocations later by allocating one
	// large one ahead of time which we'll cut up into smaller
	// slices. If this isn't big enough later, we allocate small ones.
	var strs []string
	hint := r.upcomingHeaderNewlines()
	if hint > 0 {
		strs = make([]string, hint)
	}

	m := make(MIMEHeader, hint)

	// The first line cannot start with a leading space.
	if buf, err := r.R.Peek(1); err == nil && (buf[0] == ' ' || buf[0] == '\t') {
		line, err := r.readLineSlice()
		if err != nil {
			return m, err
		}
		return m, ProtocolError("malformed MIME header initial line: " + string(line))
	}

	for {
		kv, err := r.readContinuedLineSlice(mustHaveFieldNameColon)
		if len(kv) == 0 {
			return m, err
		}

		// Key ends at first colon.
		k, v, ok := bytes.Cut(kv, colon)
		if !ok {
			return m, ProtocolError("malformed MIME header line: " + string(kv))
		}
		key, ok := canonicalEmailMIMEHeaderKey(k)
		if !ok {
			return m, ProtocolError("malformed MIME header line: " + string(kv))
		}
		// for _, c := range v {
		// 	if !validHeaderValueByte(c) {
		// 		return m, ProtocolError("malformed MIME header line: " + string(kv))
		// 	}
		// }

		// As per RFC 7230 field-name is a token, tokens consist of one or more chars.
		// We could return a ProtocolError here, but better to be liberal in what we
		// accept, so if we get an empty key, skip it.
		if key == "" {
			continue
		}

		// Skip initial spaces in value.
		value := string(bytes.TrimLeft(v, " \t"))

		vv := m[key]
		if vv == nil {
			lim -= int64(len(key))
			lim -= 100 // map entry overhead
		}
		lim -= int64(len(value))
		if lim < 0 {
			// TODO: This should be a distinguishable error (ErrMessageTooLarge)
			// to allow mime/multipart to detect it.
			return m, errors.New("message too large")
		}
		if vv == nil && len(strs) > 0 {
			// More than likely this will be a single-element key.
			// Most headers aren't multi-valued.
			// Set the capacity on strs[0] to 1, so any future append
			// won't extend the slice into the other strings.
			vv, strs = strs[:1:1], strs[1:]
			vv[0] = value
			m[key] = vv
		} else {
			m[key] = append(vv, value)
		}

		if err != nil {
			return m, err
		}
	}
}

func CanonicalEmailMIMEHeaderKey(s string) string {
	// Quick check for canonical encoding.
	upper := true
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !validEmailHeaderFieldByte(c) {
			return s
		}
		if upper && 'a' <= c && c <= 'z' {
			s, _ = canonicalEmailMIMEHeaderKey([]byte(s))
			return s
		}
		if !upper && 'A' <= c && c <= 'Z' {
			s, _ = canonicalEmailMIMEHeaderKey([]byte(s))
			return s
		}
		upper = c == '-'
	}
	return s
}

func canonicalEmailMIMEHeaderKey(a []byte) (_ string, ok bool) {
	// See if a looks like a header key. If not, return it unchanged.
	noCanon := false
	for _, c := range a {
		if validEmailHeaderFieldByte(c) {
			continue
		}
		// Don't canonicalize.
		if c == ' ' {
			// We accept invalid headers with a space before the
			// colon, but must not canonicalize them.
			// See https://go.dev/issue/34540.
			noCanon = true
			continue
		}
		return string(a), false
	}
	if noCanon {
		return string(a), true
	}

	return string(a), true
}

// validEmailHeaderFieldByte Valid characters in email header field.
//
// According to [RFC 5322](https://www.rfc-editor.org/rfc/rfc5322#section-2.2),
//
//	> A field name MUST be composed of printable US-ASCII characters (i.e.,
//	> characters that have values between 33 and 126, inclusive), except
//	> colon.
func validEmailHeaderFieldByte(c byte) bool {
	const mask = 0 |
		(1<<(10)-1)<<'0' |
		(1<<(26)-1)<<'a' |
		(1<<(26)-1)<<'A' |
		1<<'!' |
		1<<'"' |
		1<<'#' |
		1<<'$' |
		1<<'%' |
		1<<'&' |
		1<<'\'' |
		1<<'(' |
		1<<')' |
		1<<'*' |
		1<<'+' |
		1<<',' |
		1<<'-' |
		1<<'.' |
		1<<'/' |
		1<<';' |
		1<<'<' |
		1<<'=' |
		1<<'>' |
		1<<'?' |
		1<<'@' |
		1<<'[' |
		1<<'\\' |
		1<<']' |
		1<<'^' |
		1<<'_' |
		1<<'`' |
		1<<'{' |
		1<<'|' |
		1<<'}' |
		1<<'~'
	return ((uint64(1)<<c)&(mask&(1<<64-1)) |
		(uint64(1)<<(c-64))&(mask>>64)) != 0
}
