package mediatype

import (
	"testing"
)

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
