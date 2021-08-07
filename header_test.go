package enmime

import (
	"bufio"
	"net/textproto"
	"strings"
	"testing"
)

// Ensure that a single plain text token passes unharmed
func TestPlainSingleToken(t *testing.T) {
	in := "Test"
	want := in
	got := decodeHeader(in)
	if got != want {
		t.Error("got:", got, "want:", want)
	}
}

// Ensure that a string of plain text tokens do not get mangled
func TestPlainTokens(t *testing.T) {
	in := "Testing One two 3 4"
	want := in
	got := decodeHeader(in)
	if got != want {
		t.Error("got:", got, "want:", want)
	}
}

// Test control character detection & abort
func TestCharsetControlDetect(t *testing.T) {
	in := "=?US\nASCII?Q?Keith_Moore?="
	want := in
	got := decodeHeader(in)
	if got != want {
		t.Error("got:", got, "want:", want)
	}
}

// Test control character detection & abort
func TestEncodingControlDetect(t *testing.T) {
	in := "=?US-ASCII?\r?Keith_Moore?="
	want := in
	got := decodeHeader(in)
	if got != want {
		t.Error("got:", got, "want:", want)
	}
}

// Test mangled termination
func TestInvalidTermination(t *testing.T) {
	in := "=?US-ASCII?Q?Keith_Moore?!"
	want := in
	got := decodeHeader(in)
	if got != want {
		t.Error("got:", got, "want:", want)
	}
}

// Try decoding a simple ASCII quoted-printable encoded word
func TestAsciiQ(t *testing.T) {
	in := "=?US-ASCII?Q?Keith_Moore?="
	want := "Keith Moore"
	got := decodeHeader(in)
	if got != want {
		t.Error("got:", got, "want:", want)
	}
}

// Try decoding a simple ASCII quoted-printable encoded word
func TestAsciiB64(t *testing.T) {
	in := "=?US-ASCII?B?SGVsbG8gV29ybGQ=?="
	want := "Hello World"
	got := decodeHeader(in)
	if got != want {
		t.Error("got:", got, "want:", want)
	}
}

// Try decoding an embedded ASCII quoted-printable encoded word
func TestEmbeddedAsciiQ(t *testing.T) {
	var testTable = []struct {
		in, want string
	}{
		// Abutting a MIME header comment is legal
		{"(=?US-ASCII?Q?Keith_Moore?=)", "(Keith Moore)"},
		// The entire header does not need to be encoded
		{"(Keith =?US-ASCII?Q?Moore?=)", "(Keith Moore)"},
	}

	for _, tt := range testTable {
		got := decodeHeader(tt.in)
		if got != tt.want {
			t.Errorf("DecodeHeader(%q) == %q, want: %q", tt.in, got, tt.want)
		}
	}
}

// Spacing rules from RFC 2047
func TestSpacing(t *testing.T) {
	var testTable = []struct {
		in, want string
	}{
		{"(=?ISO-8859-1?Q?a?=)", "(a)"},
		{"(=?ISO-8859-1?Q?a?= b)", "(a b)"},
		{"(=?ISO-8859-1?Q?a?= =?ISO-8859-1?Q?b?=)", "(ab)"},
		{"(=?ISO-8859-1?Q?a?=  =?ISO-8859-1?Q?b?=)", "(ab)"},
		{"(=?ISO-8859-1?Q?a?=\r\n  =?ISO-8859-1?Q?b?=)", "(ab)"},
		{"(=?ISO-8859-1?Q?a_b?=)", "(a b)"},
		{"(=?ISO-8859-1?Q?a?= =?ISO-8859-2?Q?_b?=)", "(a b)"},
	}

	for _, tt := range testTable {
		got := decodeHeader(tt.in)
		if got != tt.want {
			t.Errorf("DecodeHeader(%q) == %q, want: %q", tt.in, got, tt.want)
		}
	}
}

// Test some different character sets
func TestCharsets(t *testing.T) {
	var testTable = []struct {
		in, want string
	}{
		{"=?utf-8?q?abcABC_=24_=c2=a2_=e2=82=ac?=", "abcABC $ \u00a2 \u20ac"},
		{"=?iso-8859-1?q?#=a3_c=a9_r=ae_u=b5?=", "#\u00a3 c\u00a9 r\u00ae u\u00b5"},
		{"=?big5?q?=a1=5d_=a1=61_=a1=71?=", "\uff08 \uff5b \u3008"},
	}

	for _, tt := range testTable {
		got := decodeHeader(tt.in)
		if got != tt.want {
			t.Errorf("DecodeHeader(%q) == %q, want: %q", tt.in, got, tt.want)
		}
	}
}

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
	}

	for _, tt := range testTable {
		got := decodeToUTF8Base64Header(tt.in)
		if got != tt.want {
			t.Errorf("DecodeHeader(%q) == %q, want: %q", tt.in, got, tt.want)
		}
	}
}

func TestFixMangledMediaType(t *testing.T) {
	testCases := []struct {
		input string
		sep   rune
		want  string
	}{
		{
			input: "",
			sep:   ';',
			want:  "",
		},
		{
			input: `text/HTML; charset=UTF-8; format=flowed; content-transfer-encoding: 7bit=`,
			sep:   ';',
			want:  "text/HTML; charset=UTF-8; format=flowed",
		},
		{
			input: "text/html;charset=",
			sep:   ';',
			want:  "text/html;charset=",
		},
		{
			input: "text/;charset=",
			sep:   ';',
			want:  "text/plain;charset=",
		},
		{
			input: "multipart/;charset=",
			sep:   ';',
			want:  "multipart/mixed;charset=",
		},
		{
			input: "text/plain;",
			sep:   ';',
			want:  "text/plain",
		},
		{
			// Removes empty parameters.
			input: `image/png; name="abc.png"; =""`,
			sep:   ';',
			want:  `image/png; name="abc.png"`,
		},
		{
			input: "application/octet-stream;=?UTF-8?B?bmFtZT0iw7DCn8KUwoo=?=You've got a new voice miss call.msg",
			sep:   ';',
			want:  "application/octet-stream;name=\"ð\u009f\u0094\u008aYou've got a new voice miss call.msg\"",
		},
		{
			input: `application/; name="Voice message from =?UTF-8?B?4piOICsxIDI1MS0yNDUtODA0NC5tc2c=?=";`,
			sep:   ';',
			want:  `application/octet-stream; name="Voice message from ☎ +1 251-245-8044.msg"`,
		},
		{
			input: `application/pdf name="file.pdf"`,
			sep:   ' ',
			want:  `application/pdf;name="file.pdf"`,
		},
		{
			// Removes duplicate parameters.
			input: `one/two; name="file.two"; name="file.two"`,
			sep:   ';',
			want:  `one/two; name="file.two"`,
		},
		{
			// Removes duplicate parameters.
			input: `one/nosp;name="file.two"; name="file.two"`,
			sep:   ';',
			want:  `one/nosp;name="file.two"`,
		},
		{
			// Removes duplicate parameters.
			input: `one/; name="file.two"; name="file.two"`,
			sep:   ';',
			want:  `application/octet-stream; name="file.two"`,
		},
		{
			input: `application/octet-stream; =?UTF-8?B?bmFtZT3DsMKfwpTCii5tc2c=?=`,
			sep:   ' ',
			want:  "application/octet-stream;name=\"ð\u009f\u0094\u008a.msg\"",
		},
		{
			// Removes duplicate parameters.
			input: `one/two name="file.two" name="file.two"`,
			sep:   ' ',
			want:  `one/two;name="file.two"`,
		},
		{
			input: `; name="file.two"`,
			sep:   ';',
			want:  ctPlaceholder + `; name="file.two"`,
		},
		{
			input: `charset=binary; name="logoleft.jpg"`,
			sep:   ';',
			want:  `application/octet-stream; charset=binary; name="logoleft.jpg"`,
		},
		{
			input: `one/two;iso-8859-1`,
			sep:   ';',
			want:  `one/two;iso-8859-1=` + pvPlaceholder,
		},
		{
			input: `one/two; name="file.two"; iso-8859-1`,
			sep:   ';',
			want:  `one/two; name="file.two"; iso-8859-1=` + pvPlaceholder,
		},
		{
			input: `one/two; ; name="file.two"`,
			sep:   ';',
			want:  `one/two; name="file.two"`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got := fixMangledMediaType(tc.input, tc.sep)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFixUnquotedSpecials(t *testing.T) {
	testCases := []struct {
		input, want string
	}{
		{
			input: "",
			want:  "",
		},
		{
			input: "application/octet-stream",
			want:  "application/octet-stream",
		},
		{
			input: "application/octet-stream;",
			want:  "application/octet-stream;",
		},
		{
			input: "application/octet-stream; param1=\"value1\"",
			want:  "application/octet-stream; param1=\"value1\"",
		},
		{
			input: "application/octet-stream; param1=\"value1\"\\",
			want:  "application/octet-stream; param1=\"value1\"\\",
		},
		{
			input: "application/octet-stream; param1=value1",
			want:  "application/octet-stream; param1=value1",
		},
		{
			input: "application/octet-stream; param1=value1\\",
			want:  "application/octet-stream; param1=value1",
		},
		{
			input: "application/octet-stream; param1=value1\\\"",
			want:  "application/octet-stream; param1=\"value1\\\"\"",
		},
		{
			input: "application/octet-stream; param1=value\"1\"",
			want:  "application/octet-stream; param1=\"value\\\"1\\\"\"",
		},
		{
			input: "application/octet-stream; param1=\"value\\\"1\\\"\"",
			want:  "application/octet-stream; param1=\"value\\\"1\\\"\"",
		},
		{
			// Do not preserve unqoted whitespace.
			input: "application/octet-stream; param1= value1",
			want:  "application/octet-stream; param1=value1",
		},
		{
			// Do not preserve unqoted whitespace.
			input: "application/octet-stream; param1=\tvalue1",
			want:  "application/octet-stream; param1=value1",
		},
		{
			input: "application/octet-stream; param1=\"value1;\"",
			want:  "application/octet-stream; param1=\"value1;\"",
		},
		{
			input: "application/octet-stream; param1=\"value1;2.txt\"",
			want:  "application/octet-stream; param1=\"value1;2.txt\"",
		},
		{
			input: "application/octet-stream; param1=\"value 1\"",
			want:  "application/octet-stream; param1=\"value 1\"",
		},
		{
			// Preserve quoted whitespace.
			input: "application/octet-stream; param1=\" value 1\"",
			want:  "application/octet-stream; param1=\" value 1\"",
		},
		{
			// Preserve quoted whitespace.
			input: "application/octet-stream; param1=\"\tvalue 1\"",
			want:  "application/octet-stream; param1=\"\tvalue 1\"",
		},
		{
			// Preserve quoted whitespace.
			input: "application/octet-stream; param1=\"value\t1\"",
			want:  "application/octet-stream; param1=\"value\t1\"",
		},
		{
			input: "application/octet-stream; param1=\"value(1).pdf\"",
			want:  "application/octet-stream; param1=\"value(1).pdf\"",
		},
		{
			input: "application/octet-stream; param1=value(1).pdf",
			want:  "application/octet-stream; param1=\"value(1).pdf\"",
		},
		{
			input: "application/octet-stream; param1=value(1).pdf; param2=value(2).pdf",
			want:  "application/octet-stream; param1=\"value(1).pdf\"; param2=\"value(2).pdf\"",
		},
		{
			input: "application/octet-stream; param1=value(1).pdf;\tparam2=value2.pdf;",
			want:  "application/octet-stream; param1=\"value(1).pdf\";\tparam2=value2.pdf;",
		},
		{
			input: "application/octet-stream; param1=value(1).pdf;param2=value2.pdf;",
			want:  "application/octet-stream; param1=\"value(1).pdf\";param2=value2.pdf;",
		},
		{
			input: "application/octet-stream; param1=value/1",
			want:  "application/octet-stream; param1=\"value/1\"",
		},
		{
			input: "multipart/alternative; boundary=?UOAwFjScLp1is-162467503201177404728935166502-",
			want:  "multipart/alternative; boundary=\"?UOAwFjScLp1is-162467503201177404728935166502-\"",
		},
		{
			input: `text/HTML; charset="UTF-8Return-Path: bounce-810_HTML-769869545-477063-1070564-43@bounce.email.oflce57578375.com`,
			want:  `text/HTML; charset="UTF-8Return-Path: bounce-810_HTML-769869545-477063-1070564-43@bounce.email.oflce57578375.com"`,
		},
		{
			input: `text/html;charset=`,
			want:  `text/html;charset=""`,
		},
		{
			input: `text/html; charset=; format=flowed`,
			want:  `text/html; charset=""; format=flowed`,
		},
		{
			input: `text/html;charset="`,
			want:  `text/html;charset=""`,
		},
		{
			// Check unquoted 8bit is encoded
			input: `application/msword;name=管理.doc`,
			want:  `application/msword;name="=?utf-8?b?566h55CGLmRvYw==?="`,
		},
		{
			// Check mix of ascii and unquoted 8bit is encoded
			input: `application/msword;name=15管理.doc`,
			want:  `application/msword;name="=?utf-8?b?MTXnrqHnkIYuZG9j?="`,
		},
		{
			// Check quoted 8bit is encoded
			input: `application/msword;name="15管理.doc"`,
			want:  `application/msword;name="=?utf-8?b?MTXnrqHnkIYuZG9j?="`,
		},
		{
			// Check quoted 8bit with missing closing quote is encoded
			input: `application/msword;name="15管理.doc`,
			want:  `application/msword;name="=?utf-8?b?MTXnrqHnkIYuZG9j?="`,
		},
		{
			// Trailing quote without starting quote is considered as part of param text for simplicity
			input: `application/msword;name=15管理.doc"`,
			want:  `application/msword;name="=?utf-8?b?MTXnrqHnkIYuZG9jXCI=?="`,
		},
		{
			// Invalid UTF-8 sequence does not cause any fatal error
			input: "application/msword;name=\xe2\x28\xa1.doc",
			want:  `application/msword;name="=?utf-8?b?77+9KO+/vS5kb2M=?="`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got := fixUnquotedSpecials(tc.input)
			if got != tc.want {
				t.Errorf("\ngot:  %s\nwant: %s", got, tc.want)
			}
		})
	}
}

func TestFixUnEscapedQuotes(t *testing.T) {
	testCases := []struct {
		input, want string
	}{
		{
			input: "application/rtf; charset=iso-8859-1; name=\"\"V047411.rtf\".rtf\"",
			want:  "application/rtf; charset=iso-8859-1; name=\"\\\"V047411.rtf\\\".rtf\"",
		},
		{
			input: "application/octet-stream; param1=\"",
			want:  "application/octet-stream; param1=\"\"",
		},
		{
			input: "application/octet-stream; param1=\"\\\"\"",
			want:  "application/octet-stream; param1=\"\\\"\"",
		},
		{
			input: "application/rtf; charset=iso-8859-1; name=b\"V047411.rtf\".rtf",
			want:  "application/rtf; charset=iso-8859-1; name=\"b\\\"V047411.rtf\\\".rtf\"",
		},
		{
			input: "application/rtf; charset=iso-8859-1; name=\"V047411.rtf\".rtf",
			want:  "application/rtf; charset=iso-8859-1; name=\"\\\"V047411.rtf\\\".rtf\"",
		},
		{
			input: "application/rtf; charset=iso-8859-1; name=\"V047411.rtf;\".rtf",
			want:  "application/rtf; charset=iso-8859-1; name=\"\\\"V047411.rtf;\\\".rtf\"",
		},
		{
			input: "application/rtf; charset=utf-8; name=\"žába.jpg\"",
			want:  "application/rtf; charset=utf-8; name=\"žába.jpg\"",
		},
		{
			input: "application/rtf; charset=utf-8; name=\"\"žába\".jpg\"",
			want:  "application/rtf; charset=utf-8; name=\"\\\"žába\\\".jpg\"",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got := fixUnescapedQuotes(tc.input)
			if got != tc.want {
				t.Errorf("\ngot:  %s\nwant: %s", got, tc.want)
			}
		})
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

func TestParseMediaType(t *testing.T) {
	testCases := []struct {
		label  string            // Test case label.
		input  string            // Content type to parse.
		mtype  string            // Expected media type returned.
		params map[string]string // Expected params returned.
	}{
		{
			label:  "basic filename",
			input:  "text/html; name=index.html",
			mtype:  "text/html",
			params: map[string]string{"name": "index.html"},
		},
		{
			label:  "quoted filename",
			input:  "text/html; name=\"index.html\"",
			mtype:  "text/html",
			params: map[string]string{"name": "index.html"},
		},
		{
			label:  "basic filename trailing separator",
			input:  "text/html; name=index.html;",
			mtype:  "text/html",
			params: map[string]string{"name": "index.html"},
		},
		{
			label:  "quoted filename trailing separator",
			input:  "text/html; name=\"index.html\";",
			mtype:  "text/html",
			params: map[string]string{"name": "index.html"},
		},
		{
			label:  "unclosed quoted filename",
			input:  "text/html; name=\"index.html",
			mtype:  "text/html",
			params: map[string]string{"name": "index.html"},
		},
		{
			label:  "quoted filename with separator",
			input:  "text/html; name=\"index;a.html\"",
			mtype:  "text/html",
			params: map[string]string{"name": "index;a.html"},
		},
		{
			label:  "quoted separator mid-string",
			input:  "text/html; name=\"index;a.html\"; hash=8675309",
			mtype:  "text/html",
			params: map[string]string{"name": "index;a.html", "hash": "8675309"},
		},
		{
			label:  "with prefix whitespace",
			input:  "text/plain; charset= \"UTF-8\"; format=flowed",
			mtype:  "text/plain",
			params: map[string]string{"charset": "UTF-8", "format": "flowed"},
		},
		{
			label:  "with double prefix whitespace",
			input:  "text/plain; charset = \"UTF-8\"; format=flowed",
			mtype:  "text/plain",
			params: map[string]string{"charset": "UTF-8", "format": "flowed"},
		},
		{
			label:  "with postfix whitespace",
			input:  "text/plain; charset=\"UTF-8\" ; format=flowed",
			mtype:  "text/plain",
			params: map[string]string{"charset": "UTF-8", "format": "flowed"},
		},
		{
			label:  "with whitespace tab",
			input:  "text/plain; charset=\"UTF-8\"\t; format=flowed",
			mtype:  "text/plain",
			params: map[string]string{"charset": "UTF-8", "format": "flowed"},
		},
		{
			label:  "with newline and tab",
			input:  "text/plain; charset=\"UTF-8\"\n\t; format=flowed",
			mtype:  "text/plain",
			params: map[string]string{"charset": "UTF-8", "format": "flowed"},
		},
		{
			label:  "with newline and space",
			input:  "application/pdf; name=foo\n ; format=flowed",
			mtype:  "application/pdf",
			params: map[string]string{"name": "foo", "format": "flowed"},
		},
		{
			label:  "with more spaces",
			input:  "application/pdf; name=foo      ; format=flowed",
			mtype:  "application/pdf",
			params: map[string]string{"name": "foo", "format": "flowed"},
		},
		{
			label:  "with more tabs",
			input:  "application/pdf; name=foo \t\t; format=flowed",
			mtype:  "application/pdf",
			params: map[string]string{"name": "foo", "format": "flowed"},
		},
		{
			label:  "with more newlines",
			input:  "application/pdf; name=foo \n\n; format=flowed",
			mtype:  "application/pdf",
			params: map[string]string{"name": "foo", "format": "flowed"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			mtype, params, _, err := ParseMediaType(tc.input)

			if err != nil {
				t.Errorf("got err %v, want nil", err)
				return
			}

			if mtype != tc.mtype {
				t.Errorf("mtype got %q, want %q", mtype, tc.mtype)
			}

			for k, v := range tc.params {
				if params[k] != v {
					t.Errorf("params[%q] got %q, want %q", k, params[k], v)
				}
				// Delete param to allow check for unexpected below.
				delete(params, k)
			}
			for pname := range params {
				t.Errorf("Found unexpected param: %q=%q", pname, params[pname])
			}
		})
	}
}
