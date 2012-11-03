package enmime

import (
	"container/list"
)

type MIMEPartMatcher func(part *MIMEPart) bool

func (n *MIMEPart) BreadthFirstSearch(matcher MIMEPartMatcher) *MIMEPart {
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
