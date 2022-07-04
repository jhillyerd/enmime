package enmime

// Option to configure parsing.
type Option interface {
	apply(p *Parser)
}

// SkipMalformedParts sets parsing to skip parts that's can't be parsed.
func SkipMalformedParts(s bool) Option {
	return skipMalformedPartsOption(s)
}

type skipMalformedPartsOption bool

func (o skipMalformedPartsOption) apply(p *Parser) {
	p.skipMalformedParts = bool(o)
}

// MultipartWOBoundaryAsSinglePart if set to true will treat a multi-part messages without boundary parameter as single-part.
// Otherwise, will return error that boundary is not found.
func MultipartWOBoundaryAsSinglePart(a bool) Option {
	return multipartWOBoundaryAsSinglePartOption(a)
}

type multipartWOBoundaryAsSinglePartOption bool

func (o multipartWOBoundaryAsSinglePartOption) apply(p *Parser) {
	p.multipartWOBoundaryAsSinglePart = bool(o)
}
