package enmime

import (
	"encoding/json"
	"net/textproto"
	"time"
)

// MarshalJSON implements json.Marshaler for Part.
//
// Only the part's own headers, metadata and content are emitted; the Parent,
// FirstChild and NextSibling links and the unexported parsing state are left
// out. Because the child parts are not encoded, marshaling a root Part on its
// own discards the rest of the tree, so callers usually want to marshal the
// enclosing Envelope instead to get a usable result.
func (p *Part) MarshalJSON() ([]byte, error) {
	return json.Marshal(partView{
		Header:                  p.Header,
		Boundary:                p.Boundary,
		ContentID:               p.ContentID,
		ContentType:             p.ContentType,
		ContentTypeParams:       p.ContentTypeParams,
		Disposition:             p.Disposition,
		FileName:                p.FileName,
		FileModDate:             p.FileModDate,
		Charset:                 p.Charset,
		OrigCharset:             p.OrigCharset,
		ContentTransferEncoding: p.ContentTransferEncoding,
		Errors:                  p.Errors,
		Content:                 p.Content,
		Epilogue:                p.Epilogue,
	})
}

type partView struct {
	Header                  textproto.MIMEHeader `json:"header,omitempty"`
	Boundary                string               `json:"boundary,omitempty"`
	ContentID               string               `json:"contentId,omitempty"`
	ContentType             string               `json:"contentType,omitempty"`
	ContentTypeParams       map[string]string    `json:"contentTypeParams,omitempty"`
	Disposition             string               `json:"disposition,omitempty"`
	FileName                string               `json:"fileName,omitempty"`
	FileModDate             time.Time            `json:"fileModDate,omitempty"`
	Charset                 string               `json:"charset,omitempty"`
	OrigCharset             string               `json:"origCharset,omitempty"`
	ContentTransferEncoding string               `json:"contentTransferEncoding,omitempty"`
	Errors                  []*Error             `json:"errors,omitempty"`
	Content                 []byte               `json:"content,omitempty"`
	Epilogue                []byte               `json:"epilogue,omitempty"`
}

// MarshalJSON implements json.Marshaler for Envelope.
//
// The default marshaler cannot encode an Envelope for two reasons: Root points
// into a Part tree whose child parts link back to their parent, and the same
// parts are reachable again through Attachments, Inlines and OtherParts. This
// emits the envelope content most callers want (the decoded text and HTML
// bodies, the parsed headers, and the flat attachment/inline/other part lists)
// and leaves out Root so the whole part tree isn't duplicated alongside it.
func (e *Envelope) MarshalJSON() ([]byte, error) {
	view := envelopeView{
		Text:        e.Text,
		HTML:        e.HTML,
		Attachments: e.Attachments,
		Inlines:     e.Inlines,
		OtherParts:  e.OtherParts,
		Errors:      e.Errors,
	}
	if e.header != nil {
		view.Header = textproto.MIMEHeader(*e.header)
	}
	return json.Marshal(view)
}

type envelopeView struct {
	Header      textproto.MIMEHeader `json:"header,omitempty"`
	Text        string               `json:"text,omitempty"`
	HTML        string               `json:"html,omitempty"`
	Attachments []*Part              `json:"attachments,omitempty"`
	Inlines     []*Part              `json:"inlines,omitempty"`
	OtherParts  []*Part              `json:"otherParts,omitempty"`
	Errors      []*Error             `json:"errors,omitempty"`
}
