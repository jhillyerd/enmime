/*
	Package enmime implements a MIME parsing library for Go.  It's built ontop
	of Go's included mime/multipart support, but is geared towards parsing MIME
	encoded emails.

	The basics:

	Calling ParseMIMEBody causes enmime to parse the body of the message object
	into a tree of MIMEPart objects, each of which is aware of its content
	type, filename and headers.  If the part was encoded in quoted-printable or
	base64, it is decoded before being stored in the MIMEPart object.

	ParseMIMEBody returns a MIMEBody struct.  The struct contains both the
	plain text and HTML portions of the email (if available).  The root of the
	tree, as well as slices of the email's inlines and attachments are
	available in the struct.

	If you need to locate a particular MIMEPart, you can pass a custom
	MIMEPartMatcher function into BreadthMatchFirst() or DepthMatchFirst() to
	search the MIMEPart tree.  BreadthMatchAll() and DepthMatchAll() will
	collect all matching parts.

	Please note that enmime parses messages into memory, so it is not likely to
	perform well with multi-gigabyte attachments.

	enmime is open source software released under the MIT License.  The latest
	version can be found at https://github.com/jhillyerd/go.enmime
*/
package enmime
