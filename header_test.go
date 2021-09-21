package enmime

import (
	"bufio"
	"net/textproto"
	"strings"
	"testing"
)

// Test re-encoding to base64
func TestDecodeToUTF8Base64Header(t *testing.T) {
	var testTable = []struct {
		in, want string
	}{
		{"no encoding", "no encoding"},
		{"=?utf-8?q?abcABC_=24_=c2=a2_=e2=82=ac?=", "=?UTF-8?b?YWJjQUJDICQgwqIg4oKs?="},
		{"=?iso-8859-1?q?#=a3_c=a9_r=ae_u=b5?=", "=?UTF-8?b?I8KjIGPCqSBywq4gdcK1?="},
		{"=?big5?q?=a1=5d_=a1=61_=a1=71?=", "=?UTF-8?b?77yIIO+9myDjgIg=?="},
		// Must respect separate tokens
		{"=?UTF-8?Q?Miros=C5=82aw?= <u@h>", "=?UTF-8?b?TWlyb3PFgmF3?= <u@h>"},
		{"First Last <u@h> (=?iso-8859-1?q?#=a3_c=a9_r=ae_u=b5?=)",
			"First Last <u@h> (=?UTF-8?b?I8KjIGPCqSBywq4gdcK1?=)"},
		// Quoted display name without space before angle-addr spec, Issue #112
		{"\"=?UTF-8?b?TWlyb3PFgmF3?=\"<u@h>", "=?UTF-8?b?Ik1pcm9zxYJhdyI=?= <u@h>"},
		// Colon char, Issue #218
		{"=?utf-8?Q?re=3AStore=20?= =?utf-8?Q?Apple=20?= =?utf-8?Q?Premium=20?= =?utf-8?Q?Reseller?= <u@h>",
			"\"re:Store  Apple  Premium  Reseller\" <u@h>"},
	}

	for _, tt := range testTable {
		got := decodeToUTF8Base64Header(tt.in)
		if got != tt.want {
			t.Errorf("DecodeHeader(%q) == %q, want: %q", tt.in, got, tt.want)
		}
	}
}

func TestReadHeader(t *testing.T) {
	// These values will surround the test table input string.
	prefix := "From: hooman\n \n being\n"
	suffix := "Subject: hi\n\nPart body\n"

	data := make([]byte, 16*1024)
	for i := 0; i < len(data); i++ {
		data[i] = 'x'
	}
	sdata := string(data)
	var ttable = []struct {
		label, input, hname, want string
		correct                   bool
		extras                    []string
	}{
		{
			label:   "basic crlf",
			input:   "Foo: bar\r\n",
			hname:   "Foo",
			want:    "bar",
			correct: true,
		},
		{
			label:   "basic lf",
			input:   "To: anybody\n",
			hname:   "To",
			want:    "anybody",
			correct: true,
		},
		{
			label:   "hyphenated",
			input:   "Content-Language: en\r\n",
			hname:   "Content-Language",
			want:    "en",
			correct: true,
		},
		{
			label:   "numeric",
			input:   "Privilege: 127\n",
			hname:   "Privilege",
			want:    "127",
			correct: true,
		},
		{
			label:   "space before colon",
			input:   "SID : 0\r\n",
			hname:   "SID",
			want:    "0",
			correct: true,
		},
		{
			label:   "space in name",
			input:   "Audio Mode : None\r\n",
			hname:   "Audio Mode",
			want:    "None",
			correct: true,
		},
		{
			label:   "sdata",
			input:   "Cookie: " + sdata + "\r\n",
			hname:   "Cookie",
			want:    sdata,
			correct: true,
		},
		{
			label:   "missing name",
			input:   ": line1=foo\r\n",
			hname:   "",
			want:    "",
			correct: false,
		},
		{
			label: "blank line in continuation",
			input: "X-Continuation: line1=foo\r\n" +
				" \r\n" +
				" line2=bar\r\n",
			hname:   "X-Continuation",
			want:    "line1=foo  line2=bar",
			correct: true,
		},
		{
			label:   "lf-space continuation",
			input:   "Content-Type: text/plain;\n charset=us-ascii\n",
			hname:   "Content-Type",
			want:    "text/plain; charset=us-ascii",
			correct: true,
		},
		{
			label:   "lf-tab continuation",
			input:   "X-Tabbed-Continuation: line1=foo;\n\tline2=bar\n",
			hname:   "X-Tabbed-Continuation",
			want:    "line1=foo; line2=bar",
			correct: true,
		},
		{
			label:   "equals in name",
			input:   "name=value:text\n",
			hname:   "name=value",
			want:    "text",
			correct: true,
		},
		{
			label:   "no space before continuation",
			input:   "X-Bad-Continuation: line1=foo;\nline2=bar\n",
			hname:   "X-Bad-Continuation",
			want:    "line1=foo; line2=bar",
			correct: false,
		},
		{
			label:   "not really a continuation",
			input:   "X-Not-Continuation: line1=foo;\nline2: bar\n",
			hname:   "X-Not-Continuation",
			want:    "line1=foo;",
			correct: true,
			extras:  []string{"line2"},
		},
		{
			label:   "continuation with header style",
			input:   "X-Continuation: line1=foo;\n not-a-header 15 X-Not-Header: bar\n",
			hname:   "X-Continuation",
			want:    "line1=foo; not-a-header 15 X-Not-Header: bar",
			correct: true,
		},
		{
			label: "multiline continuation with header style, few spaces",
			input: "X-Continuation-DKIM-like: line1=foo;\n" +
				" h=Subject:From:Reply-To:To:Date:Message-ID: List-ID:List-Unsubscribe:\n" +
				" Content-Type:MIME-Version;\n",
			hname: "X-Continuation-DKIM-like",
			want: "line1=foo;" +
				" h=Subject:From:Reply-To:To:Date:Message-ID: List-ID:List-Unsubscribe:" +
				" Content-Type:MIME-Version;",
			correct: true,
		},
		{
			label: "multiline continuation, few colons",
			input: "Authentication-Results: mx.google.com;\n" +
				"       spf=pass (google.com: sender)\n" +
				"       dkim=pass header.i=@1;\n" +
				"       dkim=pass header.i=@2\n",
			hname: "Authentication-Results",
			want: "mx.google.com;" +
				" spf=pass (google.com: sender)" +
				" dkim=pass header.i=@1;" +
				" dkim=pass header.i=@2",
			correct: true,
		},
		{
			label: "continuation containing early name-colon",
			input: "DKIM-Signature: a=rsa-sha256; v=1; q=dns/txt;\r\n" +
				"  s=krs; t=1603674005; h=Content-Transfer-Encoding: Mime-Version:\r\n" +
				"  Content-Type: Subject: From: To: Message-Id: Sender: Date;\r\n",
			hname: "DKIM-Signature",
			want: "a=rsa-sha256; v=1; q=dns/txt;" +
				" s=krs; t=1603674005; h=Content-Transfer-Encoding: Mime-Version:" +
				" Content-Type: Subject: From: To: Message-Id: Sender: Date;",
			correct: true,
		},
	}

	for _, tt := range ttable {
		t.Run(tt.label, func(t *testing.T) {
			if lastc := tt.input[len(tt.input)-1]; lastc != '\r' && lastc != '\n' {
				t.Fatalf("Malformed test case, %q input does not end with a CR or LF", tt.label)
			}

			// Reader we will share with readHeader()
			r := bufio.NewReader(strings.NewReader(prefix + tt.input + suffix))

			p := &Part{}
			header, err := readHeader(r, p)
			if err != nil {
				t.Fatal(err)
			}

			// Check exepcted prefix header.
			got := header.Get("From")
			want := "hooman  being"
			if got != want {
				t.Errorf("Prefix (From) header mangled\ngot: %q, want: %q", got, want)
			}

			// Check exepcted suffix header.
			got = header.Get("Subject")
			want = "hi"
			if got != want {
				t.Errorf("Suffix (Subject) header mangled\ngot: %q, want: %q", got, want)
			}

			// Check exepcted header from ttable.
			got = header.Get(tt.hname)
			if got != tt.want {
				t.Errorf(
					"Stripped %q header value mismatch\ngot : %q,\nwant: %q", tt.hname, got, tt.want)
			}

			// Check error count.
			wantErrs := 0
			if !tt.correct {
				wantErrs = 1
			}
			gotErrs := len(p.Errors)
			if gotErrs != wantErrs {
				t.Errorf("Got %v p.Errors, want %v", gotErrs, wantErrs)
			}

			// Check for extra headers by removing expected ones.
			delete(header, "From")
			delete(header, "Subject")
			delete(header, textproto.CanonicalMIMEHeaderKey(tt.hname))
			for _, hname := range tt.extras {
				delete(header, textproto.CanonicalMIMEHeaderKey(hname))
			}
			for hname := range header {
				t.Errorf("Found unexpected header %q after parsing", hname)
			}

			// Output input if any check failed.
			if t.Failed() {
				t.Errorf("input: %q", tt.input)
			}

			// readHeader should have consumed the two header lines, and the blank line, but not the
			// body
			want = "Part body"
			line, isPrefix, err := r.ReadLine()
			got = string(line)
			if err != nil {
				t.Fatal(err)
			}
			if isPrefix {
				t.Fatal("isPrefix was true, wanted false")
			}
			if got != want {
				t.Errorf("Line got: %q, want: %q", got, want)
			}
		})
	}
}

func TestCommaDelimitedAddressLists(t *testing.T) {
	testData := []struct {
		have string
		want string
	}{
		{
			have: `"Joe @ Company" <joe@company.com> <other@company.com>`,
			want: `"Joe @ Company" <joe@company.com>, <other@company.com>`,
		},
		{
			have: `Joe Company <joe@company.com> <other@company.com>`,
			want: `Joe Company <joe@company.com>, <other@company.com>`,
		},
		{
			have: `Joe Company:Joey <joe@company.com> John <other@company.com>;`,
			want: `Joe Company:Joey <joe@company.com>, John <other@company.com>;`,
		},
		{
			have: `Joe Company:Joey <joe@company.com> John <other@company.com>; Jimmy John <jimmy.john@company.com>`,
			want: `Joe Company:Joey <joe@company.com>, John <other@company.com>;`,
		},
		{
			have: `Joe Company <joe@company.com> John Company <other@company.com>`,
			want: `Joe Company <joe@company.com>, John Company <other@company.com>`,
		},
		{
			have: `Joe Company <joe@company.com>,John Company <other@company.com>`,
			want: `Joe Company <joe@company.com>,John Company <other@company.com>`,
		},
		{
			have: `joe@company.com other@company.com`,
			want: `joe@company.com, other@company.com`,
		},
		{
			have: `Jimmy John <jimmy.john@company.com> joe@company.com other@company.com`,
			want: `Jimmy John <jimmy.john@company.com>, joe@company.com, other@company.com`,
		},
		{
			have: `Jimmy John <jimmy.john@company.com> joe@company.com John Company <other@company.com>`,
			want: `Jimmy John <jimmy.john@company.com>, joe@company.com, John Company <other@company.com>`,
		},
		{
			have: `<boss@nil.test> "Giant; \"Big\" Box" <sysservices@example.net>`,
			want: `<boss@nil.test>, "Giant; \"Big\" Box" <sysservices@example.net>`,
		},
		{
			have: `A Group:Ed Jones <c@a.test>,joe@where.test,John <jdoe@one.test>;`,
			want: `A Group:Ed Jones <c@a.test>,joe@where.test,John <jdoe@one.test>;`,
		},
		{
			have: `A Group:Ed Jones <c@a.test> joe@where.test John <jdoe@one.test>;`,
			want: `A Group:Ed Jones <c@a.test>, joe@where.test, John <jdoe@one.test>;`,
		},
	}
	for i := range testData {
		v := ensureCommaDelimitedAddresses(testData[i].have)
		if testData[i].want != v {
			t.Fatalf("Expected %s, but got %s", testData[i].want, v)
		}
	}
}
