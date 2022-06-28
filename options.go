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

// MultipartWOBoundaryAsSinglepart if set to true will treat a multi-part messages without boundary parameter as single-part.
// Otherwise, will return error that boundary is not found.
func MultipartWOBoundaryAsSinglepart(a bool) Option {
	return multipartWOBoundaryAsSinglepartOption(a)
}

type multipartWOBoundaryAsSinglepartOption bool

func (o multipartWOBoundaryAsSinglepartOption) apply(p *Parser) {
	p.multipartWOBoundaryAsSinglepart = bool(o)
}
