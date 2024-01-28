package enmime

import (
	"fmt"
	"strings"
	"testing"
)

func TestSetCustomParseMediaType(t *testing.T) {
	alwaysReturnHTML := func(ctype string) (mtype string, params map[string]string, invalidParams []string, err error) {
		return "text/html", nil, nil, err
	}
	changeAndUtilizeDefault := func(ctype string) (mtype string, params map[string]string, invalidParams []string, err error) {
		modifiedStr := strings.ReplaceAll(ctype, "application/Pamir Viewer", "application/PamirViewer")
		return ParseMediaType(modifiedStr)
	}
	tcases := []struct {
		ctype                string
		want                 string
		customParseMediaType CustomParseMediaType
	}{
		{
			ctype:                "text/plain",
			want:                 "text/plain",
			customParseMediaType: nil,
		},
		{
			ctype:                "text/plain",
			want:                 "text/html",
			customParseMediaType: alwaysReturnHTML,
		},
		{
			ctype:                "text/plain; charset=utf-8",
			want:                 "text/html",
			customParseMediaType: alwaysReturnHTML,
		},
		{
			ctype:                "application/Pamir Viewer; name=\"2023-384.pmrv\"",
			want:                 "application/pamirviewer",
			customParseMediaType: changeAndUtilizeDefault,
		},
	}

	for _, tcase := range tcases {
		p := &Part{parser: NewParser(SetCustomParseMediaType(tcase.customParseMediaType))}

		got, _, _, _ := p.parseMediaType(tcase.ctype)
		if got != tcase.want {
			t.Errorf("Parser.parseMediaType(%v) == %v, want: %v",
				tcase.ctype, got, tcase.want)
		}
	}
}

func ExampleSetCustomParseMediaType() {
	// for the sake of simplicity replaces space in a very specific invalid content-type: "application/Pamir Viewer"
	replaceSpecificContentType := func(ctype string) (mtype string, params map[string]string, invalidParams []string, err error) {
		modifiedStr := strings.ReplaceAll(ctype, "application/Pamir Viewer", "application/PamirViewer")

		return ParseMediaType(modifiedStr)
	}

	invalidMessageContent := `From: <enmime@parser.git>
Content-Type: multipart/mixed;
	boundary="----=_NextPart_000_000F_01D9FAC6.09EB3B60"

------=_NextPart_000_000F_01D9FAC6.09EB3B60
Content-Type: application/Pamir Viewer;
	name="2023-10-13.pmrv"
Content-Transfer-Encoding: base64
Content-Disposition: attachment;
	filename="2023-10-13.pmrv"

f6En7vFpNql3tfMkoKABP1iBEf+M/qF6LCAIvyRbpH6uDCqcKKGmH3e6OiqN5eCfqUk=
`

	p := NewParser(SetCustomParseMediaType(replaceSpecificContentType))
	e, err := p.ReadEnvelope(strings.NewReader(invalidMessageContent))

	fmt.Println(err)
	fmt.Println(len(e.Attachments))
	fmt.Println(e.Attachments[0].ContentType)
	fmt.Println(e.Attachments[0].FileName)

	// Output:
	// <nil>
	// 1
	// application/pamirviewer
	// 2023-10-13.pmrv
}
