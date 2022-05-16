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
