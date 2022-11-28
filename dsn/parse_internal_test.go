package dsn

import (
	"net/textproto"
	"testing"

	"github.com/jhillyerd/enmime"
	"github.com/stretchr/testify/assert"
)

func TestParseDeliveryStatusFields(t *testing.T) {
	tests := map[string]struct {
		status string
		want   []textproto.MIMEHeader
	}{
		"regular": {
			status: `Reporting-MTA: dns; cs.utk.edu

Action: failed
Status: 4.0.0
Diagnostic-Code: smtp; 426 connection timed out
`,
			want: []textproto.MIMEHeader{
				{
					"Reporting-Mta": []string{"dns; cs.utk.edu"},
				},
				{
					"Action":          []string{"failed"},
					"Status":          []string{"4.0.0"},
					"Diagnostic-Code": []string{"smtp; 426 connection timed out"},
				},
			},
		},
		"without per-message DSN": {
			status: `Action: failed
Status: 4.0.0
Diagnostic-Code: smtp; 426 connection timed out
`,
			want: []textproto.MIMEHeader{
				{
					"Action":          []string{"failed"},
					"Status":          []string{"4.0.0"},
					"Diagnostic-Code": []string{"smtp; 426 connection timed out"},
				},
			},
		},
		"without new line": {
			status: `Action: failed
Status: 4.0.0
Diagnostic-Code: smtp; 426 connection timed out`,
			want: []textproto.MIMEHeader{
				{
					"Action":          []string{"failed"},
					"Status":          []string{"4.0.0"},
					"Diagnostic-Code": []string{"smtp; 426 connection timed out"},
				},
			},
		},
		"empty": {
			status: "",
			want:   []textproto.MIMEHeader{{}},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := parseDeliveryStatusFields([]byte(tt.status))
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.want, got, "should correctly parse delivery/status message")
		})
	}
}

func TestSetExplanation(t *testing.T) {
	tests := map[string]struct {
		part *enmime.Part
		want Explanation
	}{
		"text/plain": {
			part: &enmime.Part{
				ContentType: "text/plain",
				Content:     []byte("text content"),
			},
			want: Explanation{Text: "text content"},
		},
		"text/html": {
			part: &enmime.Part{
				ContentType: "text/html",
				Content:     []byte("HTML content"),
			},
			want: Explanation{HTML: "HTML content"},
		},
		"multipart/alternative": {
			part: &enmime.Part{
				ContentType: "multipart/alternative",
				FirstChild: &enmime.Part{
					ContentType: "text/plain",
					Content:     []byte("text content"),
					NextSibling: &enmime.Part{
						ContentType: "text/html",
						Content:     []byte("HTML content"),
					},
				},
			},
			want: Explanation{Text: "text content", HTML: "HTML content"},
		},
		"no content-type": {
			part: &enmime.Part{
				ContentType: "",
				Content:     []byte("text content"),
			},
			want: Explanation{Text: "text content"},
		},
		"other": {
			part: &enmime.Part{
				ContentType: "other",
				Content:     []byte("some content"),
			},
			want: Explanation{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var e Explanation
			setExplanation(&e, tt.part)
			assert.Equal(t, tt.want, e)
		})
	}
}
