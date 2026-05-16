package enmime

import (
	"bufio"
	"bytes"
	"net/mail"
	"net/textproto"

	"github.com/jhillyerd/enmime/v2/internal/coding"
	"github.com/jhillyerd/enmime/v2/internal/stringutil"
	inttp "github.com/jhillyerd/enmime/v2/internal/textproto"

	"github.com/pkg/errors"
)

const (
	// Standard MIME content dispositions
	cdAttachment = "attachment"
	cdInline     = "inline"

	// Standard MIME content types
	ctAppOctetStream   = "application/octet-stream"
	ctMultipartAltern  = "multipart/alternative"
	ctMultipartMixed   = "multipart/mixed"
	ctMultipartPrefix  = "multipart/"
	ctMultipartRelated = "multipart/related"
	ctTextPlain        = "text/plain"
	ctTextHTML         = "text/html"

	// Standard Transfer encodings
	cte7Bit            = "7bit"
	cte8Bit            = "8bit"
	cteBase64          = "base64"
	cteBinary          = "binary"
	cteQuotedPrintable = "quoted-printable"

	// Standard MIME header names
	hnContentDisposition = "Content-Disposition"
	hnContentEncoding    = "Content-Transfer-Encoding"
	hnContentID          = "Content-ID"
	hnContentType        = "Content-Type"
	hnMIMEVersion        = "MIME-Version"

	// Standard MIME header parameters
	hpBoundary = "boundary"
	hpCharset  = "charset"
	hpFile     = "file"
	hpFilename = "filename"
	hpName     = "name"
	hpModDate  = "modification-date"

	utf8 = "utf-8"
)

// AddressHeaders is the set of SMTP headers that contain email addresses, used by
// Envelope.AddressList().  Key characters must be all lowercase.
var AddressHeaders = map[string]bool{
	"bcc":             true,
	"cc":              true,
	"delivered-to":    true,
	"from":            true,
	"reply-to":        true,
	"to":              true,
	"sender":          true,
	"resent-bcc":      true,
	"resent-cc":       true,
	"resent-from":     true,
	"resent-reply-to": true,
	"resent-to":       true,
	"resent-sender":   true,
}

// ParseAddressList returns a mail.Address slice with RFC 2047 encoded names converted to UTF-8.
// It is more tolerant of malformed headers than the ParseAddressList func provided in Go's net/mail
// package.
func ParseAddressList(list string) ([]*mail.Address, error) {
	parser := mail.AddressParser{WordDecoder: coding.NewExtMimeDecoder()}

	ret, err := parser.ParseList(list)
	if err != nil {
		switch err.Error() {
		case "mail: expected comma":
			// Attempt to add commas and parse again.
			return parser.ParseList(stringutil.EnsureCommaDelimitedAddresses(list))
		case "mail: no address":
			return nil, mail.ErrHeaderNotPresent
		}
		return nil, err
	}

	for i := range ret {
		// try to additionally decode with less strict decoder
		ret[i].Name = coding.DecodeExtHeader(ret[i].Name)
		ret[i].Address = coding.DecodeExtHeader(ret[i].Address)
	}

	return ret, nil
}

// ReadHeader reads a block of SMTP or MIME headers and returns a
// textproto.MIMEHeader. Header parse warnings & errors will be added to
// ErrorCollector, io errors will be returned directly.
func ReadHeader(r *bufio.Reader, p ErrorCollector) (textproto.MIMEHeader, error) {
	// buf holds the massaged output for textproto.Reader.ReadMIMEHeader()
	buf := &bytes.Buffer{}
	tp := inttp.NewReader(r)
	firstHeader := true
line:
	for {
		// Pull out each line of the headers as a temporary slice s
		s, err := tp.ReadLineBytes()
		if err != nil {
			buf.Write([]byte{'\r', '\n'})
			break
		}

		firstColon := bytes.IndexByte(s, ':')
		firstSpace := bytes.IndexAny(s, " \t\n\r")
		if firstSpace == 0 {
			// Starts with space: continuation
			buf.WriteByte(' ')
			buf.Write(inttp.TrimBytes(s))
			continue
		}
		if firstColon == 0 {
			// Can't parse line starting with colon: skip
			p.AddError(ErrorMalformedHeader, "Header line %q started with a colon", s)
			continue
		}
		if firstColon > 0 {
			// Behavior change in net/textproto package in Golang 1.12.10 and 1.13.1:
			// A space preceding the first colon in a header line is no longer handled
			// automatically due to CVE-2019-16276 which takes advantage of this
			// particular violation of RFC-7230 to exploit HTTP/1.1
			if bytes.Contains(s[:firstColon+1], []byte{' ', ':'}) {
				s = bytes.Replace(s, []byte{' ', ':'}, []byte{':'}, 1)
				firstColon = bytes.IndexByte(s, ':')
			}

			// Behavior change in net/textproto package in Golang 1.20: invalid characters
			// in header keys are no longer allowed; https://github.com/golang/go/issues/53188
			for _, c := range s[:firstColon] {
				if c != ' ' && !inttp.ValidEmailHeaderFieldByte(c) {
					p.AddError(
						ErrorMalformedHeader, "Header name %q contains invalid character %q", s, c)
					continue line
				}
			}

			// Contains a colon, treat as a new header line
			if !firstHeader {
				// New Header line, end the previous
				buf.Write([]byte{'\r', '\n'})
			}

			s = inttp.TrimBytes(s)
			buf.Write(s)
			firstHeader = false
		} else {
			// No colon: potential non-indented continuation
			if len(s) > 0 {
				// Attempt to detect and repair a non-indented continuation of previous line
				buf.WriteByte(' ')
				buf.Write(s)
				p.AddWarning(ErrorMalformedHeader, "Continued line %q was not indented", s)
			} else {
				// Empty line, finish header parsing
				buf.Write([]byte{'\r', '\n'})
				break
			}
		}
	}

	buf.Write([]byte{'\r', '\n'})
	tr := inttp.NewReader(bufio.NewReader(buf))
	header, err := tr.ReadEmailMIMEHeader()
	return textproto.MIMEHeader(header), errors.WithStack(err)
}

// readHeader reads a block of SMTP or MIME headers and returns a textproto.MIMEHeader.
// Header parse warnings & errors will be added to p.Errors, io errors will be returned directly.
func readHeader(r *bufio.Reader, p *Part) (textproto.MIMEHeader, error) {
	return ReadHeader(r, &partErrorCollector{p})
}
