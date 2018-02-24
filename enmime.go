// Package enmime implements a MIME encoding and decoding library.  It's built on top of Go's
// included mime/multipart support where possible, but is geared towards parsing MIME encoded
// emails.
//
// Overview
//
// The enmime API has two conceptual layers.  The lower layer is a tree of Part structs,
// representing each component of a decoded MIME message.  The upper layer, called an Envelope
// provides an intuitive way to interact with a MIME message.
//
// Part Tree
//
// Calling ReadParts causes enmime to parse the body of a MIME message into a tree of Part objects,
// each of which is aware of its content type, filename and headers.  The content of a Part is
// available as a slice of bytes via the Content field.
//
// If the part was encoded in quoted-printable or base64, it is decoded prior to being placed in
// Content.  If the Part contains text in a character set other than utf-8, enmime will attempt to
// convert it to utf-8.
//
// To locate a particular Part, pass a custom PartMatcher function into the BreadthMatchFirst() or
// DepthMatchFirst() methods to search the Part tree.  BreadthMatchAll() and DepthMatchAll() will
// collect all Parts matching your criteria.
//
// The Envelope
//
// ReadEnvelope returns an Envelope struct.  Behind the scenes a Part tree is constructed, and then
// sorted into the correct fields of the Envelope.
//
// The Envelope contains both the plain text and HTML portions of the email.  If there was no plain
// text Part available, the HTML Part will be down-converted using the html2text library[1].  The
// root of the Part tree, as well as slices of the inline and attachment Parts are also available.
//
// Please note that enmime parses messages into memory, so it is not likely to perform well with
// multi-gigabyte attachments.
//
// enmime is open source software released under the MIT License.  The latest version can be found
// at https://github.com/jhillyerd/enmime
//
// [1]: https://github.com/jaytaylor/html2text
package enmime // import "github.com/jhillyerd/enmime"
