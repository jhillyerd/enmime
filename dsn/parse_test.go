package dsn_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/jhillyerd/enmime"
	"github.com/jhillyerd/enmime/dsn"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseReport(t *testing.T) {
	tests := map[string]struct {
		filename string
		want     *dsn.Report
	}{
		"simple": {
			filename: "simple_dsn.raw",
			want: &dsn.Report{
				Explanation: dsn.Explanation{Text: "[human-readable explanation goes here]\n"},
				DeliveryStatus: dsn.DeliveryStatus{
					MessageDSNs: []textproto.MIMEHeader{
						{"Reporting-Mta": []string{"dns; cs.utk.edu"}},
					},
					RecipientDSNs: []textproto.MIMEHeader{
						{
							"Original-Recipient": []string{"rfc822;louisl@larry.slip.umd.edu"},
							"Final-Recipient":    []string{"rfc822;louisl@larry.slip.umd.edu"},
							"Action":             []string{"failed"},
							"Status":             []string{"4.0.0"},
							"Diagnostic-Code":    []string{"smtp; 426 connection timed out"},
							"Last-Attempt-Date":  []string{"Thu, 7 Jul 1994 17:15:49 -0400"},
						},
					},
				},
				OriginalMessage: []byte("[original message goes here]\n"),
			},
		},
		"simple with HTML": {
			filename: "simple_dsn_with_html.raw",
			want: &dsn.Report{
				Explanation: dsn.Explanation{HTML: "<div>[human-readable explanation goes here]</div>\n"},
				DeliveryStatus: dsn.DeliveryStatus{
					MessageDSNs: []textproto.MIMEHeader{
						{"Reporting-Mta": []string{"dns; cs.utk.edu"}},
					},
					RecipientDSNs: []textproto.MIMEHeader{
						{
							"Original-Recipient": []string{"rfc822;louisl@larry.slip.umd.edu"},
							"Final-Recipient":    []string{"rfc822;louisl@larry.slip.umd.edu"},
							"Action":             []string{"failed"},
							"Status":             []string{"4.0.0"},
							"Diagnostic-Code":    []string{"smtp; 426 connection timed out"},
							"Last-Attempt-Date":  []string{"Thu, 7 Jul 1994 17:15:49 -0400"},
						},
					},
				},
				OriginalMessage: []byte("[original message goes here]\n"),
			},
		},
		"multi-recipient": {
			filename: "multi_recipient_dsn.raw",
			want: &dsn.Report{
				Explanation: dsn.Explanation{Text: "[human-readable explanation goes here]\n"},
				DeliveryStatus: dsn.DeliveryStatus{
					MessageDSNs: []textproto.MIMEHeader{
						{"Reporting-Mta": []string{"dns; cs.utk.edu"}},
					},
					RecipientDSNs: []textproto.MIMEHeader{
						{
							"Original-Recipient": []string{"rfc822;arathib@vnet.ibm.com"},
							"Final-Recipient":    []string{"rfc822;arathib@vnet.ibm.com"},
							"Action":             []string{"failed"},
							"Status":             []string{"5.0.0 (permanent failure)"},
							"Diagnostic-Code":    []string{"smtp;  550 'arathib@vnet.IBM.COM' is not a registered gateway user"},
							"Remote-Mta":         []string{"dns; vnet.ibm.com"},
						},
						{
							"Original-Recipient": []string{"rfc822;johnh@hpnjld.njd.hp.com"},
							"Final-Recipient":    []string{"rfc822;johnh@hpnjld.njd.hp.com"},
							"Action":             []string{"delayed"},
							"Status":             []string{"4.0.0 (hpnjld.njd.jp.com: host name lookup failure)"},
						},
						{
							"Original-Recipient": []string{"rfc822;wsnell@sdcc13.ucsd.edu"},
							"Final-Recipient":    []string{"rfc822;wsnell@sdcc13.ucsd.edu"},
							"Action":             []string{"failed"},
							"Status":             []string{"5.0.0"},
							"Diagnostic-Code":    []string{"smtp; 550 user unknown"},
							"Remote-Mta":         []string{"dns; sdcc13.ucsd.edu"},
						},
					},
				},
				OriginalMessage: []byte("[original message goes here]\n"),
			},
		},
		"delayed": {
			filename: "delayed_dsn.raw",
			want: &dsn.Report{
				Explanation: dsn.Explanation{Text: "[human-readable explanation goes here]\n"},
				DeliveryStatus: dsn.DeliveryStatus{
					MessageDSNs: []textproto.MIMEHeader{
						{"Reporting-Mta": []string{"dns; sun2.nsfnet-relay.ac.uk"}},
					},
					RecipientDSNs: []textproto.MIMEHeader{
						{
							"Final-Recipient": []string{"rfc822;thomas@de-montfort.ac.uk"},
							"Status":          []string{"4.0.0 (unknown temporary failure)"},
							"Action":          []string{"delayed"},
						},
					},
				},
				OriginalMessage: nil,
			},
		},
		"with multipart/alternative": {
			filename: "dsn_with_multipart_alternative.raw",
			want: &dsn.Report{
				Explanation: dsn.Explanation{
					Text: "[human-readable explanation goes here]\n",
					HTML: "<div>[human-readable explanation goes here]</div>\n",
				},
				DeliveryStatus: dsn.DeliveryStatus{
					MessageDSNs: []textproto.MIMEHeader{
						{
							"Reporting-Mta":     []string{"dns;1234.prod.outlook.com"},
							"Received-From-Mta": []string{"dns;some.example.com"},
							"Arrival-Date":      []string{"Thu, 27 Jan 2022 08:03:02 +0000"},
						},
					},
					RecipientDSNs: []textproto.MIMEHeader{
						{
							"Final-Recipient": []string{"rfc822;non-existing@example.com"},
							"Action":          []string{"failed"},
							"Status":          []string{"5.1.10"},
							"Diagnostic-Code": []string{"smtp;550 5.1.10 RESOLVER.ADR.RecipientNotFound"},
						},
					},
				},
				OriginalMessage: []byte("[original message goes here]\n"),
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := readEnvelope(t, tt.filename)
			report, err := dsn.ParseReport(env.Root)
			require.NoError(t, err)
			assert.Equal(t, tt.want, report)
		})
	}
}

func readEnvelope(tb testing.TB, filename string) *enmime.Envelope {
	tb.Helper()

	data := readTestdata(tb, filename)
	env, err := enmime.ReadEnvelope(bytes.NewReader(data))
	if err != nil {
		tb.Fatalf("read envelope: %s", err)
	}

	return env
}

func readTestdata(tb testing.TB, filename string) []byte {
	tb.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		tb.Fatalf("read %s: %s", filename, err)
	}

	return data
}
