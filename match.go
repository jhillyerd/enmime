package enmime

import (
	"container/list"
)

type MIMEPartMatcher func(part *MIMEPart) bool

// BreadthMatchFirst performs a breadth first search of the MIMEPart tree
// and returns the first part that causes the given matcher to return true
func (n *MIMEPart) BreadthMatchFirst(matcher MIMEPartMatcher) *MIMEPart {
	q := list.New()
	q.PushBack(n)

	// Push children onto queue and attempt to match in that order
	for q.Len() > 0 {
		e := q.Front()
		n := e.Value.(*MIMEPart)
		if matcher(n) {
			return n
		}
		q.Remove(e)
		c := n.FirstChild
		for c != nil {
			q.PushBack(c)
			c = c.NextSibling
		}
	}

	return nil
}

// BreadthMatchAll performs a breadth first search of the MIMEPart tree
// and returns all parts that cause the given matcher to return true
func (n *MIMEPart) BreadthMatchAll(matcher MIMEPartMatcher) []*MIMEPart {
	q := list.New()
	q.PushBack(n)

	matches := make([]*MIMEPart, 0, 10)

	// Push children onto queue and attempt to match in that order
	for q.Len() > 0 {
		e := q.Front()
		n := e.Value.(*MIMEPart)
		if matcher(n) {
			matches = append(matches, n)
		}
		q.Remove(e)
		c := n.FirstChild
		for c != nil {
			q.PushBack(c)
			c = c.NextSibling
		}
	}

	return matches
}
