package enmime

import (
	"net/textproto"
	"strings"
)

// detectMultipartMessage returns true if the message has a recognized multipart Content-Type header
func detectMultipartMessage(root *Part) bool {
	// Parse top-level multipart
	ctype := root.Header.Get(hnContentType)
	mediatype, _, err := parseMediaType(ctype)
	if err != nil {
		return false
	}
	// According to rfc2046#section-5.1.7 all other multipart should
	// be treated as multipart/mixed
	return strings.HasPrefix(mediatype, ctMultipartPrefix)
}

// detectAttachmentHeader returns true, if the given header defines an attachment.  First it checks
// if the Content-Disposition header defines an attachement or inline attachment. If this test is
// false, the Content-Type header is checked for attachment, but not inline.  Email clients use
// inline for their text bodies.
//
// Valid Attachment-Headers:
//
//  - Content-Disposition: attachment; filename="frog.jpg"
//  - Content-Disposition: inline; filename="frog.jpg"
//  - Content-Type: attachment; filename="frog.jpg"
func detectAttachmentHeader(header textproto.MIMEHeader) bool {
	mediatype, _, _ := parseMediaType(header.Get(hnContentDisposition))
	if strings.ToLower(mediatype) == cdAttachment ||
		strings.ToLower(mediatype) == cdInline {
		return true
	}

	mediatype, _, _ = parseMediaType(header.Get(hnContentType))
	return strings.ToLower(mediatype) == cdAttachment
}

// detectTextHeader returns true, if the the MIME headers define a valid 'text/plain' or 'text/html'
// part.  If the emptyContentTypeIsPlain argument is set to true, a missing Content-Type header will
// result in a positive plain part detection.
func detectTextHeader(header textproto.MIMEHeader, emptyContentTypeIsText bool) bool {
	ctype := header.Get(hnContentType)
	if ctype == "" && emptyContentTypeIsText {
		return true
	}

	mediatype, _, err := parseMediaType(ctype)
	if err != nil {
		return false
	}
	switch mediatype {
	case ctTextPlain, ctTextHTML:
		return true
	}

	return false
}

// detectBinaryBody returns true if the mail header defines a binary body.
func detectBinaryBody(root *Part) bool {
	if detectTextHeader(root.Header, true) {
		return false
	}

	isBin := detectAttachmentHeader(root.Header)
	if !isBin {
		// This must be an attachment, if the Content-Type is not
		// 'text/plain' or 'text/html'.
		// Example:
		// Content-Type: application/pdf; name="doc.pdf"
		mediatype, _, _ := parseMediaType(root.Header.Get(hnContentType))
		mediatype = strings.ToLower(mediatype)
		if mediatype != ctTextPlain && mediatype != ctTextHTML {
			return true
		}
	}

	return isBin
}
