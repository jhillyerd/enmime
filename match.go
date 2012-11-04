package enmime

import (
	"container/list"
)

type MIMEPartMatcher func(part MIMEPart) bool

// BreadthMatchFirst performs a breadth first search of the MIMEPart tree
// and returns the first part that causes the given matcher to return true
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

// BreadthMatchAll performs a breadth first search of the MIMEPart tree
// and returns all parts that cause the given matcher to return true
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
