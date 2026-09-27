package mediatype

import (
	"strings"
	"testing"
)

func TestFixMangledMediaType(t *testing.T) {
	testCases := []struct {
		input   string
		sep     rune
		want    string
		options ParseOptions
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
			// Removes empty parameters in the middle
			input: `text/html; =""; charset=""`,
			sep:   ';',
			want:  `text/html; charset=""`,
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
		{
			// Keeps a distinct param whose name is a suffix of an earlier
			// one (previously dropped by the substring match) (#162).
			input: `application/octet-stream; filename="a.txt"; name="b.txt"`,
			sep:   ';',
			want:  `application/octet-stream; filename="a.txt"; name="b.txt"`,
		},
		{
			// Keeps a distinct param whose "key=" appears inside an earlier
			// param's value (#162).
			input: `text/plain; note="size=5"; size="10"`,
			sep:   ';',
			want:  `text/plain; note="size=5"; size="10"`,
		},
		{
			// Duplicate param names are case-insensitive (RFC 2045 §5.1).
			input: `text/plain; name="x"; NAME="y"`,
			sep:   ';',
			want:  `text/plain; name="x"`,
		},

		// remove extra content type parts
		{
			input: `application/pdf/.pdf; name=1337.pdf`,
			sep:   ';',
			want:  `application/pdf; name=1337.pdf`,
		},
		{
			input: `application/pdf/pdf/pdf; name=1337.pdf`,
			sep:   ';',
			want:  `application/pdf; name=1337.pdf`,
		},
		// invalid media type characters
		{
			input: `text/html>`,
			sep:   ';',
			want:  `text/html>`,
		},
		// invalid media type characters with stripping invalid characters sanitation enabled
		{
			input:   `text/html>`,
			sep:     ';',
			want:    `text/html`,
			options: ParseOptions{StripMediaTypeInvalidCharacters: true},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got := fixMangledMediaType(tc.input, tc.sep, tc.options)
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
			input: `application/octet-stream; param1="value1"`,
			want:  `application/octet-stream; param1="value1"`,
		},
		{
			input: `application/octet-stream; param1="value1"\`,
			want:  `application/octet-stream; param1="value1"\`,
		},
		{
			input: "application/octet-stream; param1=value1",
			want:  "application/octet-stream; param1=value1",
		},
		{
			input: `application/octet-stream; param1=value1\`,
			want:  "application/octet-stream; param1=value1",
		},
		{
			input: `application/octet-stream; param1=value1\"`,
			want:  `application/octet-stream; param1="value1\""`,
		},
		{
			input: `application/octet-stream; param1=value"1"`,
			want:  `application/octet-stream; param1="value\"1\""`,
		},
		{
			input: `application/octet-stream; param1="value\"1\""`,
			want:  `application/octet-stream; param1="value\"1\""`,
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
			input: `application/octet-stream; param1="value1;"`,
			want:  `application/octet-stream; param1="value1;"`,
		},
		{
			input: `application/octet-stream; param1="value1;2.txt"`,
			want:  `application/octet-stream; param1="value1;2.txt"`,
		},
		{
			input: `application/octet-stream; param1="value 1"`,
			want:  `application/octet-stream; param1="value 1"`,
		},
		{
			// Preserve quoted whitespace.
			input: `application/octet-stream; param1=" value 1"`,
			want:  `application/octet-stream; param1=" value 1"`,
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
			input: `application/octet-stream; param1="value(1).pdf"`,
			want:  `application/octet-stream; param1="value(1).pdf"`,
		},
		{
			input: `application/octet-stream; param1=value(1).pdf`,
			want:  `application/octet-stream; param1="value(1).pdf"`,
		},
		{
			input: `application/octet-stream; param1=value(1).pdf; param2=value(2).pdf`,
			want:  `application/octet-stream; param1="value(1).pdf"; param2="value(2).pdf"`,
		},
		{
			input: "application/octet-stream; param1=value(1).pdf;\tparam2=value2.pdf;",
			want:  "application/octet-stream; param1=\"value(1).pdf\";\tparam2=value2.pdf;",
		},
		{
			input: `application/octet-stream; param1=value(1).pdf;param2=value2.pdf;`,
			want:  `application/octet-stream; param1="value(1).pdf";param2=value2.pdf;`,
		},
		{
			input: `application/octet-stream; param1=value/1`,
			want:  `application/octet-stream; param1="value/1"`,
		},
		{
			input: `multipart/alternative; boundary=?UOAwFjScLp1is-162467503201177404728935166502-`,
			want:  `multipart/alternative; boundary="?UOAwFjScLp1is-162467503201177404728935166502-"`,
		},
		{
			input: `text/HTML; charset="UTF-8Return-Path: bounce-810_HTML-1070564-43@example.com`,
			want:  `text/HTML; charset="UTF-8Return-Path: bounce-810_HTML-1070564-43@example.com"`,
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
		{
			// Value with spaces is surrounded with quotes.
			input: `text/plain; name=Untitled document.txt`,
			want:  `text/plain; name="Untitled document.txt"`,
		},
		{
			// Value with spaces is surrounded with quotes.
			input: `text/plain; name=Untitled document.txt; disposition=inline`,
			want:  `text/plain; name="Untitled document.txt"; disposition=inline`,
		},
		{
			// Escaped character at the beginning of param.
			input: `application/pdf; name=\"pdf\"pdf\"`,
			want:  `application/pdf; name="\"pdf\"pdf\""`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got := fixUnquotedSpecials(tc.input)
			if got != tc.want {
				t.Errorf("\ngot  : %s\nwant : %s\ninput: %s", got, tc.want, tc.input)
			}
		})
	}
}

func TestFixUnEscapedQuotes(t *testing.T) {
	testCases := []struct {
		input, want string
	}{
		{
			input: `application/rtf; charset=iso-8859-1; name=""V047411.rtf".rtf"`,
			want:  `application/rtf; charset=iso-8859-1; name="\"V047411.rtf\".rtf"`,
		},
		{
			input: `application/octet-stream; param1="`,
			want:  `application/octet-stream; param1=""`,
		},
		{
			input: `application/octet-stream; param1="\""`,
			want:  `application/octet-stream; param1="\""`,
		},
		{
			input: `application/rtf; charset=iso-8859-1; name=b"V047411.rtf".rtf`,
			want:  `application/rtf; charset=iso-8859-1; name="b\"V047411.rtf\".rtf"`,
		},
		{
			input: `application/rtf; charset=iso-8859-1; name="V047411.rtf".rtf`,
			want:  `application/rtf; charset=iso-8859-1; name="\"V047411.rtf\".rtf"`,
		},
		{
			input: `application/rtf; charset=iso-8859-1; name="V047411.rtf;".rtf`,
			want:  `application/rtf; charset=iso-8859-1; name="\"V047411.rtf;\".rtf"`,
		},
		{
			input: `application/rtf; charset=utf-8; name="žába.jpg"`,
			want:  `application/rtf; charset=utf-8; name="žába.jpg"`,
		},
		{
			input: `application/rtf; charset=utf-8; name=""žába".jpg"`,
			want:  `application/rtf; charset=utf-8; name="\"žába\".jpg"`,
		},
		{
			input: `multipart/mixed; boundary="aaa;bbb;ccc"`, // `;` inside quoted text, should ignore it
			want:  `multipart/mixed; boundary="aaa;bbb;ccc"`,
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
			input:  `text/html; name="index.html"`,
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
			input:  `text/html; name="index.html";`,
			mtype:  "text/html",
			params: map[string]string{"name": "index.html"},
		},
		{
			label:  "unclosed quoted filename",
			input:  `text/html; name="index.html`,
			mtype:  "text/html",
			params: map[string]string{"name": "index.html"},
		},
		{
			label:  "quoted filename with separator",
			input:  `text/html; name="index;a.html"`,
			mtype:  "text/html",
			params: map[string]string{"name": "index;a.html"},
		},
		{
			label:  "quoted separator mid-string",
			input:  `text/html; name="index;a.html"; hash=8675309`,
			mtype:  "text/html",
			params: map[string]string{"name": "index;a.html", "hash": "8675309"},
		},
		{
			label:  "with prefix whitespace",
			input:  `text/plain; charset= "UTF-8"; format=flowed`,
			mtype:  "text/plain",
			params: map[string]string{"charset": "UTF-8", "format": "flowed"},
		},
		{
			label:  "with double prefix whitespace",
			input:  `text/plain; charset = "UTF-8"; format=flowed`,
			mtype:  "text/plain",
			params: map[string]string{"charset": "UTF-8", "format": "flowed"},
		},
		{
			label:  "with postfix whitespace",
			input:  `text/plain; charset="UTF-8" ; format=flowed`,
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
		{
			label:  "unquoted with spaces",
			input:  "application/pdf; x-unix-mode=0644; name=File name with spaces.pdf",
			mtype:  "application/pdf",
			params: map[string]string{"x-unix-mode": "0644", "name": "File name with spaces.pdf"},
		},
		{
			label:  "Outlook-Logo with newlines",
			input:  `application/octet-stream; name="=?utf-8?B?T3V0bG9vay1Mb2dvCgpEZXNj?="`,
			mtype:  "application/octet-stream",
			params: map[string]string{"name": "Outlook-Logo  Desc"},
		},
		{
			label:  "remove HTML tag",
			input:  `text/html; charset="iso-8859-1"<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">`,
			mtype:  "text/html",
			params: map[string]string{"charset": "iso-8859-1"},
		},
		{
			label:  "ignore quoted HTML tag",
			input:  `multipart/alternative; boundary="<myboundary>"`,
			mtype:  "multipart/alternative",
			params: map[string]string{"boundary": "<myboundary>"},
		},
		{
			label:  "encoded equal sign",
			input:  "application/pdf; name=\"=?iso-8859-1?Q?key=3Dvalue?=\"",
			mtype:  "application/pdf",
			params: map[string]string{"name": "key=value"},
		},
		{
			label:  "distinct suffix-name param kept",
			input:  `application/octet-stream; filename="a.txt"; name="b.txt"`,
			mtype:  "application/octet-stream",
			params: map[string]string{"filename": "a.txt", "name": "b.txt"},
		},
		{
			label:  "distinct param whose key appears in earlier value kept",
			input:  `text/plain; note="size=5"; size="10"`,
			mtype:  "text/plain",
			params: map[string]string{"note": "size=5", "size": "10"},
		},
		{
			label:  "exact duplicate keeps first",
			input:  `text/plain; name="first"; name="second"`,
			mtype:  "text/plain",
			params: map[string]string{"name": "first"},
		},
		{
			label:  "case-insensitive duplicate keeps first",
			input:  `text/plain; name="x"; NAME="y"`,
			mtype:  "text/plain",
			params: map[string]string{"name": "x"},
		},
		{
			label:  "RFC 2231 multi-segment filename with euc-kr charset",
			input:  `attachment; filename*0*=euc-kr''%B0%B3; filename*1*=%2E%74%78%74`,
			mtype:  "attachment",
			params: map[string]string{"filename": "개.txt"},
		},
		{
			label:  "RFC 2231 multi-segment filename with us-ascii charset",
			input:  `attachment; filename*0*=us-ascii''hello; filename*1*=-world.txt`,
			mtype:  "attachment",
			params: map[string]string{"filename": "hello-world.txt"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			mtype, params, _, err := Parse(tc.input)

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

func TestRemoveTrailingHTMLTags(t *testing.T) {
	tests := map[string]struct {
		has  string
		want string
	}{
		"singe tag": {
			has:  `text/html; charset="iso-8859-1"<tag blah-blah-blah>`,
			want: `text/html; charset="iso-8859-1"`,
		},
		"no tag": {
			has:  `text/html; charset="iso-8859-1"`,
			want: `text/html; charset="iso-8859-1"`,
		},
		"no tag to params": {
			has:  `text/html"`,
			want: `text/html"`,
		},
		"multiple tags": {
			has:  `text/html; charset="iso-8859-1"<tag1> <tag2>`,
			want: `text/html; charset="iso-8859-1"`,
		},
		"nested tags": {
			has:  `text/html; charset="iso-8859-1"<tag1 blah-blah-blah <tag2>>`,
			want: `text/html; charset="iso-8859-1"`,
		},
		"with spaces": {
			has:  `text/html; charset="iso-8859-1"  <tag1>`,
			want: `text/html; charset="iso-8859-1"  `,
		},
		"multiple, nested and with spaces": {
			has:  `text/html; charset="iso-8859-1"  <tag1<tag2>> <tag3>`,
			want: `text/html; charset="iso-8859-1"  `,
		},
		"no params single tag": {
			has:  `text/html<tag "some text">`,
			want: `text/html`,
		},
		"unclosed tag": {
			has:  `text/html <unclosed tag`,
			want: `text/html <unclosed tag`,
		},
		"unopened tag": {
			has:  `text/html unopened tag>`,
			want: `text/html unopened tag>`,
		},
		"broken tag": {
			has:  `text/html >text<`,
			want: `text/html >text<`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := removeTrailingHTMLTags(tt.has)
			if got != tt.want {
				t.Errorf("should remove trainling HTML tags, has: %q, want: %q, got: %q", tt.has, tt.want, got)
			}
		})
	}
}

// TestAssembleRFC2231Params exercises the RFC 2231 continuation-parameter assembly added for
// https://github.com/jhillyerd/enmime/issues/109. Each case runs end-to-end through Parse so
// that failure modes (dropped params, leftover segments, garbage concatenation) are caught.
func TestAssembleRFC2231Params(t *testing.T) {
	testCases := []struct {
		label  string
		input  string
		mtype  string
		params map[string]string
	}{
		{
			label:  "multi-segment utf-8 happy path",
			input:  `attachment; filename*0*=utf-8''%E2%82%AC%20; filename*1*=%EB%AC%B8%EC%84%9C.txt`,
			mtype:  "attachment",
			params: map[string]string{"filename": "€ 문서.txt"},
		},
		{
			label:  "multi-segment euc-kr happy path",
			input:  `attachment; filename*0*=euc-kr''%C0%CC; filename*1*=%2E%74%78%74`,
			mtype:  "attachment",
			params: map[string]string{"filename": "이.txt"},
		},
		{
			label:  "single-segment with charset decodes as percent-encoded",
			input:  `attachment; filename*=euc-kr''%C0%CC.txt`,
			mtype:  "attachment",
			params: map[string]string{"filename": "이.txt"},
		},
		{
			label:  "single-segment ascii charset handled by stdlib",
			input:  `attachment; filename*=us-ascii''report-1.txt`,
			mtype:  "attachment",
			params: map[string]string{"filename": "report-1.txt"},
		},
		{
			label:  "repeated charset prefix on every segment (Outlook)",
			input:  `attachment; filename*0*=euc-kr''%C6%C4; filename*1*=euc-kr''%C0%CF`,
			mtype:  "attachment",
			params: map[string]string{"filename": "파일"},
		},
		{
			label:  "segment gap assembles present segments only",
			input:  `attachment; filename*0*=euc-kr''%C0%CC; filename*2*=%2E%74%78%74`,
			mtype:  "attachment",
			params: map[string]string{"filename": "이.txt"},
		},
		{
			label:  "out-of-order segments assembled by index",
			input:  `attachment; filename*1*=%2E%74%78%74; filename*0*=euc-kr''%C0%CC`,
			mtype:  "attachment",
			params: map[string]string{"filename": "이.txt"},
		},
		{
			label:  "empty charset prefix segment",
			input:  `attachment; filename*0*=euc-kr''%C0%CC; filename*1*=''`,
			mtype:  "attachment",
			params: map[string]string{"filename": "이"},
		},
		{
			label:  "unknown charset preserves value instead of dropping it",
			input:  `attachment; filename*0*=bogus-charset''%C0%CC; filename*1*=%2E%74%78%74`,
			mtype:  "attachment",
			params: map[string]string{"filename": "%C0%CC%2E%74%78%74"},
		},
		{
			label:  "malformed percent-encoding preserves raw value",
			input:  `attachment; filename*0*=euc-kr''%ZZ%41; filename*1*=%42`,
			mtype:  "attachment",
			params: map[string]string{"filename": "%ZZ%41%42"},
		},
		{
			label:  "apostrophe pair in plain segments survives literal",
			input:  `attachment; filename*0=don''t; filename*1=panic.txt`,
			mtype:  "attachment",
			params: map[string]string{"filename": "don''tpanic.txt"},
		},
		{
			label:  "charset prefix stripped only from extended segment",
			input:  `attachment; filename*0*=utf-8''doc; filename*1=don''t.txt`,
			mtype:  "attachment",
			params: map[string]string{"filename": "docdon''t.txt"},
		},
		{
			label:  "empty params ignored",
			input:  `attachment;;`,
			mtype:  "attachment",
			params: map[string]string{},
		},
		{
			label:  "no params untouched",
			input:  `text/plain`,
			mtype:  "text/plain",
			params: map[string]string{},
		},
		{
			label:  "plain filename param untouched",
			input:  `attachment; filename="x.txt"`,
			mtype:  "attachment",
			params: map[string]string{"filename": "x.txt"},
		},
		{
			label:  "single utf-8 extended param handled by stdlib",
			input:  `attachment; filename*=utf-8''%E2%82%AC%20a.txt`,
			mtype:  "attachment",
			params: map[string]string{"filename": "€ a.txt"},
		},
		{
			label:  "charset with language tag in extended prefix",
			input:  `attachment; filename*0*=euc-kr'ko'%C0%CC; filename*1*=%2E%74%78%74`,
			mtype:  "attachment",
			params: map[string]string{"filename": "이.txt"},
		},
		{
			label:  "charset with language tag, single segment",
			input:  `attachment; filename*=us-ascii'en-us'Hello.txt`,
			mtype:  "attachment",
			params: map[string]string{"filename": "Hello.txt"},
		},
		{
			label:  "quoted plain continuation segments are joined",
			input:  `test/plain; URL*0="ftp://"; URL*1="cs.utk.edu/pub"`,
			mtype:  "test/plain",
			params: map[string]string{"url": "ftp://cs.utk.edu/pub"},
		},
		{
			label:  "plain continuation segments with whitespace trimmed",
			input:  `attachment; filename*0= "long "; filename*1= name.txt `,
			mtype:  "attachment",
			params: map[string]string{"filename": "long name.txt"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			mtype, params, _, err := Parse(tc.input)
			if err != nil {
				t.Fatalf("got err %v, want nil", err)
			}
			if mtype != tc.mtype {
				t.Errorf("mtype got %q, want %q", mtype, tc.mtype)
			}
			for k, v := range tc.params {
				if params[k] != v {
					t.Errorf("params[%q] got %q, want %q", k, params[k], v)
				}
				delete(params, k)
			}
			for pname := range params {
				t.Errorf("Found unexpected param: %q=%q", pname, params[pname])
			}
		})
	}
}

func TestAssembleRFC2231ParamsString(t *testing.T) {
	// String-level checks of assembleRFC2231Params before it reaches the standard library.
	if got := assembleRFC2231Params("text/plain"); got != "text/plain" {
		t.Errorf("no-param input got %q, want unchanged", got)
	}
	if got := assembleRFC2231Params(`attachment; filename="x.txt"`); got != `attachment; filename="x.txt"` {
		t.Errorf("plain param got %q, want unchanged", got)
	}
	if got := assembleRFC2231Params(""); got != "" {
		t.Errorf("empty input got %q, want empty", got)
	}

	in := `attachment; filename*0*=euc-kr''%C6%C4; filename*1*=euc-kr''%C0%CF`
	out := assembleRFC2231Params(in)
	if strings.Count(out, "filename*") != 1 {
		t.Errorf("assembled string should collapse to one extended param, got %q", out)
	}
	if !strings.Contains(out, "utf-8''") {
		t.Errorf("assembled string should re-encode as utf-8 extended form, got %q", out)
	}

	// Failure paths must degrade to a recoverable quoted param, not an empty result.
	for _, in := range []string{
		`attachment; filename*0*=bogus-charset''%C0%CC; filename*1*=%2E%74%78%74`,
		`attachment; filename*0*=euc-kr''%ZZ%41; filename*1*=%42`,
	} {
		out := assembleRFC2231Params(in)
		mtype, params, _, err := Parse(out)
		if err != nil || mtype != "attachment" {
			t.Errorf("failure-path input %q did not survive re-parse: %q (%v)", in, out, err)
			continue
		}
		if params["filename"] == "" {
			t.Errorf("failure-path input %q silently dropped filename: %q", in, out)
		}
	}
}
