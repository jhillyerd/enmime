package enmime

// ReadPartErrorPolicy allows to recover the buffer (or not) on an error when reading a Part content.
//
// See AllowCorruptTextPartErrorPolicy for usage.
type ReadPartErrorPolicy func(*Part, error) bool

// AllowCorruptTextPartErrorPolicy recovers partial content from base64.CorruptInputError when content type is text/plain or text/html.
func AllowCorruptTextPartErrorPolicy(p *Part, err error) bool {
	if IsBase64CorruptInputError(err) && (p.ContentType == ctTextHTML || p.ContentType == ctTextPlain) {
		return true
	}
	return false
}

// CustomParseMediaType parses media type. See ParseMediaType for more details
type CustomParseMediaType func(ctype string) (mtype string, params map[string]string, invalidParams []string, err error)

// Parser parses MIME.
// Default parser is a valid one.
type Parser struct {
	maxStoredPartErrors             *int // TODO: Pointer until global var removed.
	multipartWOBoundaryAsSinglePart bool
	readPartErrorPolicy             ReadPartErrorPolicy
	skipMalformedParts              bool
	rawContent                      bool
	customParseMediaType            CustomParseMediaType
	stripMediaTypeInvalidCharacters bool
	disableTextConversion           bool
	disableCharacterDetection       bool
}

// defaultParser is a Parser with default configuration.
var defaultParser = Parser{}

// NewParser creates new parser with given options.
func NewParser(ops ...Option) *Parser {
	p := Parser{}

	for _, o := range ops {
		o.apply(&p)
	}

	return &p
}
