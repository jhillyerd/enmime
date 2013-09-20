package enmime

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBreadthMatchFirst(t *testing.T) {
	// Setup test MIME tree:
	//    root
	//    ├── a1
	//    │   ├── b1
	//    │   └── b2
	//    ├── a2
	//    └── a3

	root := &memMIMEPart{contentType: "multipart/alternative"}
	a1 := &memMIMEPart{contentType: "multipart/related", parent: root}
	a2 := &memMIMEPart{contentType: "text/plain", parent: root}
	a3 := &memMIMEPart{contentType: "text/html", parent: root}
	b1 := &memMIMEPart{contentType: "text/plain", parent: a1}
	b2 := &memMIMEPart{contentType: "text/html", parent: a1}
	root.firstChild = a1
	a1.nextSibling = a2
	a2.nextSibling = a3
	a1.firstChild = b1
	b1.nextSibling = b2

	p := BreadthMatchFirst(root, func(pt MIMEPart) bool {
		return pt.ContentType() == "text/plain"
	})
	assert.NotNil(t, p, "BreathMatchFirst should have returned a result for text/plain")
	assert.True(t, p.(*memMIMEPart) == a2,
		"BreadthMatchFirst should have returned the first text/plain object")

	p = BreadthMatchFirst(root, func(pt MIMEPart) bool {
		return pt.ContentType() == "text/html"
	})
	assert.True(t, p.(*memMIMEPart) == a3,
		"BreadthMatchFirst should have returned the first text/html object")
}

func TestBreadthMatchAll(t *testing.T) {
	// Setup test MIME tree:
	//    root
	//    ├── a1
	//    │   ├── b1
	//    │   └── b2
	//    ├── a2
	//    └── a3

	root := &memMIMEPart{contentType: "multipart/alternative"}
	a1 := &memMIMEPart{contentType: "multipart/related", parent: root}
	a2 := &memMIMEPart{contentType: "text/plain", parent: root}
	a3 := &memMIMEPart{contentType: "text/html", parent: root}
	b1 := &memMIMEPart{contentType: "text/plain", parent: a1}
	b2 := &memMIMEPart{contentType: "text/html", parent: a1}
	root.firstChild = a1
	a1.nextSibling = a2
	a2.nextSibling = a3
	a1.firstChild = b1
	b1.nextSibling = b2

	ps := BreadthMatchAll(root, func(pt MIMEPart) bool {
		return pt.ContentType() == "text/plain"
	})
	assert.Equal(t, len(ps), 2, "BreadthMatchAll should have returned two matches")
	assert.True(t, ps[0].(*memMIMEPart) == a2,
		"BreadthMatchFirst should have returned the first text/plain object")
	assert.True(t, ps[1].(*memMIMEPart) == b1,
		"BreadthMatchFirst should have returned the second text/plain object")

	ps = BreadthMatchAll(root, func(pt MIMEPart) bool {
		return pt.ContentType() == "text/html"
	})
	assert.Equal(t, len(ps), 2, "BreadthMatchAll should have returned two matches")
	assert.True(t, ps[0].(*memMIMEPart) == a3,
		"BreadthMatchFirst should have returned the first text/html object")
	assert.True(t, ps[1].(*memMIMEPart) == b2,
		"BreadthMatchFirst should have returned the second text/html object")
}
