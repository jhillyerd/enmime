package enmime

import (
	"container/list"
)

// MIMEPartMatcher is a function type that you must implement to search for MIMEParts using
// the BreadthMatch* functions.  Implementators should inspect the provided MIMEPart and
// return true if it matches your criteria.
type MIMEPartMatcher func(part MIMEPart) bool

// BreadthMatchFirst performs a breadth first search of the MIMEPart tree and returns the
// first part that causes the given matcher to return true
func BreadthMatchFirst(p MIMEPart, matcher MIMEPartMatcher) MIMEPart {
	q := list.New()
	q.PushBack(p)

	// Push children onto queue and attempt to match in that order
	for q.Len() > 0 {
		e := q.Front()
		p := e.Value.(MIMEPart)
		if matcher(p) {
			return p
		}
		q.Remove(e)
		c := p.FirstChild()
		for c != nil {
			q.PushBack(c)
			c = c.NextSibling()
		}
	}

	return nil
}

// BreadthMatchAll performs a breadth first search of the MIMEPart tree and returns all parts
// that cause the given matcher to return true
func BreadthMatchAll(p MIMEPart, matcher MIMEPartMatcher) []MIMEPart {
	q := list.New()
	q.PushBack(p)

	matches := make([]MIMEPart, 0, 10)

	// Push children onto queue and attempt to match in that order
	for q.Len() > 0 {
		e := q.Front()
		p := e.Value.(MIMEPart)
		if matcher(p) {
			matches = append(matches, p)
		}
		q.Remove(e)
		c := p.FirstChild()
		for c != nil {
			q.PushBack(c)
			c = c.NextSibling()
		}
	}

	return matches
}

// DepthMatchFirst performs a depth first search of the MIMEPart tree and returns the
// first part that causes the given matcher to return true
func DepthMatchFirst(p MIMEPart, matcher MIMEPartMatcher) MIMEPart {
	root := p
	for {
		if matcher(p) {
			return p
		}
		c := p.FirstChild()
		if c != nil {
			p = c
		} else {
			for p.NextSibling() == nil {
				if p == root {
					return nil
				}
				p = p.Parent()
			}
			p = p.NextSibling()
		}
	}
}

// DepthMatchAll performs a depth first search of the MIMEPart tree and returns all parts
// that causes the given matcher to return true
func DepthMatchAll(p MIMEPart, matcher MIMEPartMatcher) []MIMEPart {
	root := p
	matches := make([]MIMEPart, 0, 10)
	for {
		if matcher(p) {
			matches = append(matches, p)
		}
		c := p.FirstChild()
		if c != nil {
			p = c
		} else {
			for p.NextSibling() == nil {
				if p == root {
					return matches
				}
				p = p.Parent()
			}
			p = p.NextSibling()
		}
	}
}
