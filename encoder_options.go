package enmime

type EncoderOption interface {
	apply(p *Encoder)
}

// Encoder implements MIME part encoding options
type Encoder struct {
	// Enforce Quoted encoding when selecting Content Transfer Encoding
	enforceQuotedCte bool
}

// EnforceQuotedPrintableCte enforces "quoted-printable" transfer encoding when selecting Content Transfer Encoding.
func EnforceQuotedPrintableCte(b bool) EncoderOption {
	return enforceQuotedCteOption(b)
}

type enforceQuotedCteOption bool

func (o enforceQuotedCteOption) apply(p *Encoder) {
	p.enforceQuotedCte = bool(o)
}

func NewEncoder(ops ...EncoderOption) *Encoder {
	e := Encoder{
		enforceQuotedCte: false,
	}

	for _, o := range ops {
		o.apply(&e)
	}

	return &e
}
