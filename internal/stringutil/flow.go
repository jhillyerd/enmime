package stringutil

import (
	"regexp"
	"strings"
)

const (
	rfc2646Space     = " "
	rfc2646Quote     = ">"
	rfc2646Signature = "-- "
	rfc2646CRLF      = "\r\n"
	rfc2646From      = "From "
	rfc2646Width     = 78
)

var lineTerm = regexp.MustCompile("\r\n|\n")

// Deflow decodes a text previously wrapped using "format=flowed".
//
// In order to decode, the input text must belong to a mail with headers similar to:
// Content-Type: text/plain; charset="CHARSET"; [delsp="yes|no"; ]format="flowed"
// (the quotes around CHARSET are not mandatory).
// Furthermore the header Content-Transfer-Encoding MUST NOT BE Quoted-Printable
// (see RFC3676 paragraph 4.2).(In fact this happens often for non 7bit messages).
func Deflow(text string, delSp bool) string {
	lines := regSplitAfter(text)
	var result *strings.Builder
	resultLine := &strings.Builder{}
	resultLineQuoteDepth := 0
	resultLineFlowed := false
	var line *string
	for i := 0; i <= len(lines); i++ {
		if i < len(lines) {
			line = &lines[i]
		} else {
			line = nil
		}
		actualQuoteDepth := 0
		if line != nil && len(*line) > 0 {
			tmpString := *line
			if tmpString == rfc2646Signature {
				// signature handling (the previous line is not flowed)
				resultLineFlowed = false
			}
			if strings.HasPrefix(*line, rfc2646Quote) {
				// Quote
				actualQuoteDepth = 1
				for actualQuoteDepth < len(tmpString) && string(tmpString[actualQuoteDepth]) == rfc2646Quote {
					actualQuoteDepth++
				}
				// if quote-depth changes wrt the previous line then this is not flowed
				if resultLineQuoteDepth != actualQuoteDepth {
					resultLineFlowed = false
				}
				tmpString = tmpString[actualQuoteDepth:]
				line = &tmpString
			} else {
				// id quote-depth changes wrt the first line then this is not flowed
				if resultLineQuoteDepth > 0 {
					resultLineFlowed = false
				}
			}

			if len(tmpString) > 0 && strings.HasPrefix(tmpString, rfc2646Space) {
				// Line space-stuffed
				tmpString = tmpString[1:]
				line = &tmpString
			}
		}

		// If the previous line was the last then it was not flowed.
		if line == nil {
			resultLineFlowed = false
		}

		// Add the PREVIOUS line.
		// This often will find the flow looking for a space as the last char of the line.
		// With quote changes or signatures it could be the following line to void the flow.
		if !resultLineFlowed && i > 0 {
			for j := 0; j < resultLineQuoteDepth; j++ {
				resultLine.WriteString(rfc2646Quote)
			}
			if resultLineQuoteDepth > 0 {
				resultLine.WriteString(rfc2646Space)
			}
			if result == nil {
				result = &strings.Builder{}
			} else {
				result.WriteString(rfc2646CRLF)
			}
			result.WriteString(resultLine.String())
			resultLine = &strings.Builder{}
			resultLineFlowed = false
		}
		resultLineQuoteDepth = actualQuoteDepth

		if line != nil {
			if !(*line == rfc2646Signature) && strings.HasSuffix(*line, rfc2646Space) && i < len(lines)-1 {
				// Line flowed (NOTE: for the split operation the line having i == len(lines) is the last that does not end with rfc2646CRLF)
				if delSp {
					tmpString := *line
					tmpString = tmpString[:len(tmpString)-1]
					line = &tmpString
				}
				resultLineFlowed = true
			} else {
				resultLineFlowed = false
			}

			resultLine.WriteString(*line)
		}
	}

	if result == nil {
		result = &strings.Builder{}
	}

	return result.String()
}

// Flow encodes a text (using standard width).
//
// When encoding the input text will be changed eliminating every space found before CRLF,
// otherwise it won't be possible to recognize hard breaks from soft breaks.
// In this scenario encoding and decoding a message will not return a message identical to
// the original (lines with hard breaks will be trimmed).
func Flow(text string, delSp bool) string {
	return FlowN(text, delSp, rfc2646Width)
}

// Flow encodes a text (using N with).
//
// When encoding the input text will be changed eliminating every space found before CRLF,
// otherwise it won't be possible to recognize hard breaks from soft breaks.
// In this scenario encoding and decoding a message will not return a message identical to
// the original (lines with hard breaks will be trimmed).
func FlowN(text string, delSp bool, n int) string {
	result := &strings.Builder{}
	lines := regSplitAfter(text)
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		notEmpty := len(line) > 0
		quoteDepth := 0
		for quoteDepth < len(line) && string(line[quoteDepth]) == rfc2646Quote {
			quoteDepth++
		}
		if quoteDepth > 0 {
			if quoteDepth+1 < len(line) && string(line[quoteDepth]) == rfc2646Space {
				line = line[quoteDepth+1:]
			} else {
				line = line[quoteDepth:]
			}
		}
		for notEmpty {
			extra := 0
			if quoteDepth == 0 {
				if strings.HasPrefix(line, rfc2646Space) || strings.HasPrefix(line, rfc2646Quote) || strings.HasPrefix(line, rfc2646From) {
					line = rfc2646Space + line
					extra = 1
				}
			} else {
				line = rfc2646Space + line
				for j := 0; j < quoteDepth; j++ {
					line = rfc2646Space + line
				}
				extra = quoteDepth + 1
			}

			j := n - 1
			if j > len(line) {
				j = len(line) - 1
			} else {
				for j >= extra && (delSp && isAlphaChar(text, j)) || (!delSp && string(line[j]) != rfc2646Space) {
					j--
				}
				if j < extra {
					// Not able to cut a word: skip to word end even if greater than the max width
					j = n - 1
					for j < len(line)-1 && (delSp && isAlphaChar(text, j)) || (!delSp && string(line[j]) != rfc2646Space) {
						j++
					}
				}
			}

			result.WriteString(line[:j+1])
			if j < len(line)-1 {
				if delSp {
					result.WriteString(rfc2646Space)
				}
				result.WriteString(rfc2646CRLF)
			}
			line = line[j+1:]
			notEmpty = len(line) > 0
		}

		if i < len(lines)-1 {
			// NOTE: Have to trim the spaces before, otherwise it won't recognize soft-break from hard break.
			// Deflow of flowed message will not be identical to the original.
			for result.Len() > 0 && string(result.String()[result.Len()-1]) == rfc2646Space {
				result.WriteString(rfc2646CRLF)
			}
		}
	}

	return result.String()
}

// isAlphaChar checks whether the char is part of a word.
// RFC asserts a word cannot be split (even if the length is greater than the maximum length).
func isAlphaChar(text string, index int) bool {
	// Note: a list of chars is available here:
	// http://www.zvon.org/tmRFC/RFC2646/Output/index.html
	c := text[index]
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

func regSplitAfter(s string) []string {
	var (
		r []string
		p int
	)
	is := lineTerm.FindAllStringIndex(s, -1)
	if is == nil {
		return append(r, s)
	}
	for _, i := range is {
		r = append(r, s[p:i[1]])
		p = i[1]
	}
	return append(r, s[p:])
}
