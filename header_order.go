package enmime

import (
	"sort"

	inttp "github.com/jhillyerd/enmime/v2/internal/textproto"
)

func (p *Part) recordHeaderKey(name string) {
	if p == nil {
		return
	}
	key := inttp.CanonicalEmailMIMEHeaderKey(name)
	for _, k := range p.headerKeys {
		if k == key {
			return
		}
	}
	p.headerKeys = append(p.headerKeys, key)
}

func (p *Part) removeHeaderKey(name string) {
	if p == nil {
		return
	}
	key := inttp.CanonicalEmailMIMEHeaderKey(name)
	for i, k := range p.headerKeys {
		if k == key {
			p.headerKeys = append(p.headerKeys[:i], p.headerKeys[i+1:]...)
			return
		}
	}
}

// headerKeysInWriteOrder returns header keys for encoding. When header order
// was captured during parsing, that order is preserved. Any headers added
// without updating headerKeys are appended in sorted order.
func (p *Part) headerKeysInWriteOrder() []string {
	if p == nil || len(p.Header) == 0 {
		return nil
	}
	if len(p.headerKeys) == 0 {
		keys := make([]string, 0, len(p.Header))
		for k := range p.Header {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		return keys
	}

	keys := make([]string, 0, len(p.Header))
	seen := make(map[string]struct{}, len(p.headerKeys))
	for _, k := range p.headerKeys {
		if _, ok := p.Header[k]; ok {
			keys = append(keys, k)
			seen[k] = struct{}{}
		}
	}
	var extra []string
	for k := range p.Header {
		if _, ok := seen[k]; !ok {
			extra = append(extra, k)
		}
	}
	sort.Strings(extra)
	return append(keys, extra...)
}
