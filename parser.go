package enmime

// Parser parses MIME.
// Default parser is a valid one.
type Parser struct {
	skipMalformedParts              bool
	multipartWOBoundaryAsSinglePart bool
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
