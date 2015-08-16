package enmime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBreadthMatchFirst(t *testing.T) {
	// Setup test MIME tree:
	//    root
	//    ├── a1
	//    │   ├── b1
	//    │   └── b2
	//    ├── a2
	//    └── a3

	root := &memMIMEPart{contentType: "multipart/alternative", fileName: "root"}
	a1 := &memMIMEPart{contentType: "multipart/related", parent: root, fileName: "a1"}
	a2 := &memMIMEPart{contentType: "text/plain", parent: root, fileName: "a2"}
	a3 := &memMIMEPart{contentType: "text/html", parent: root, fileName: "a3"}
	b1 := &memMIMEPart{contentType: "text/plain", parent: a1, fileName: "b1"}
	b2 := &memMIMEPart{contentType: "text/html", parent: a1, fileName: "b2"}
	root.firstChild = a1
	a1.nextSibling = a2
	a2.nextSibling = a3
	a1.firstChild = b1
	b1.nextSibling = b2

	p := BreadthMatchFirst(root, func(pt MIMEPart) bool {
		return pt.ContentType() == "text/plain"
	})
	assert.NotNil(t, p, "BreadthMatchFirst should have returned a result for text/plain")
	assert.True(t, p.(*memMIMEPart) == a2,
		"BreadthMatchFirst should have returned a2, got %v", p.FileName())

	p = BreadthMatchFirst(root, func(pt MIMEPart) bool {
		return pt.ContentType() == "text/html"
	})
	assert.True(t, p.(*memMIMEPart) == a3,
		"BreadthMatchFirst should have returned a3, got %v", p.FileName())
}

func TestBreadthMatchAll(t *testing.T) {
	// Setup test MIME tree:
	//    root
	//    ├── a1
	//    │   ├── b1
	//    │   └── b2
	//    ├── a2
	//    └── a3

	root := &memMIMEPart{contentType: "multipart/alternative", fileName: "root"}
	a1 := &memMIMEPart{contentType: "multipart/related", parent: root, fileName: "a1"}
	a2 := &memMIMEPart{contentType: "text/plain", parent: root, fileName: "a2"}
	a3 := &memMIMEPart{contentType: "text/html", parent: root, fileName: "a3"}
	b1 := &memMIMEPart{contentType: "text/plain", parent: a1, fileName: "b1"}
	b2 := &memMIMEPart{contentType: "text/html", parent: a1, fileName: "b2"}
	root.firstChild = a1
	a1.nextSibling = a2
	a2.nextSibling = a3
	a1.firstChild = b1
	b1.nextSibling = b2

	ps := BreadthMatchAll(root, func(pt MIMEPart) bool {
		return pt.ContentType() == "text/plain"
	})
	assert.Equal(t, 2, len(ps), "BreadthMatchAll should have returned two matches")
	assert.True(t, ps[0].(*memMIMEPart) == a2,
		"BreadthMatchAll should have returned a2, got %v", ps[0].FileName())
	assert.True(t, ps[1].(*memMIMEPart) == b1,
		"BreadthMatchAll should have returned b1, got %v", ps[1].FileName())

	ps = BreadthMatchAll(root, func(pt MIMEPart) bool {
		return pt.ContentType() == "text/html"
	})
	assert.Equal(t, 2, len(ps), "BreadthMatchAll should have returned two matches")
	assert.True(t, ps[0].(*memMIMEPart) == a3,
		"BreadthMatchAll should have returned a3, got %v", ps[0].FileName())
	assert.True(t, ps[1].(*memMIMEPart) == b2,
		"BreadthMatchAll should have returned b2, got %v", ps[1].FileName())
}

func TestDepthMatchFirst(t *testing.T) {
	// Setup test MIME tree:
	//    root
	//    ├── a1
	//    │   ├── b1
	//    │   └── b2
	//    ├── a2
	//    └── a3

	root := &memMIMEPart{contentType: "multipart/alternative", fileName: "root"}
	a1 := &memMIMEPart{contentType: "multipart/related", parent: root, fileName: "a1"}
	a2 := &memMIMEPart{contentType: "text/plain", parent: root, fileName: "a2"}
	a3 := &memMIMEPart{contentType: "text/html", parent: root, fileName: "a3"}
	b1 := &memMIMEPart{contentType: "text/plain", parent: a1, fileName: "b1"}
	b2 := &memMIMEPart{contentType: "text/html", parent: a1, fileName: "b2"}
	root.firstChild = a1
	a1.nextSibling = a2
	a2.nextSibling = a3
	a1.firstChild = b1
	b1.nextSibling = b2

	p := DepthMatchFirst(root, func(pt MIMEPart) bool {
		return pt.ContentType() == "text/plain"
	})
	assert.NotNil(t, p, "DepthMatchFirst should have returned a result for text/plain")
	assert.True(t, p.(*memMIMEPart) == b1,
		"DepthMatchFirst should have returned b1, got %v", p.FileName())

	p = DepthMatchFirst(root, func(pt MIMEPart) bool {
		return pt.ContentType() == "text/html"
	})
	assert.True(t, p.(*memMIMEPart) == b2,
		"DepthMatchFirst should have returned b2, got %v", p.FileName())
}

func TestDepthMatchAll(t *testing.T) {
	// Setup test MIME tree:
	//    root
	//    ├── a1
	//    │   ├── b1
	//    │   └── b2
	//    ├── a2
	//    └── a3

	root := &memMIMEPart{contentType: "multipart/alternative", fileName: "root"}
	a1 := &memMIMEPart{contentType: "multipart/related", parent: root, fileName: "a1"}
	a2 := &memMIMEPart{contentType: "text/plain", parent: root, fileName: "a2"}
	a3 := &memMIMEPart{contentType: "text/html", parent: root, fileName: "a3"}
	b1 := &memMIMEPart{contentType: "text/plain", parent: a1, fileName: "b1"}
	b2 := &memMIMEPart{contentType: "text/html", parent: a1, fileName: "b2"}
	root.firstChild = a1
	a1.nextSibling = a2
	a2.nextSibling = a3
	a1.firstChild = b1
	b1.nextSibling = b2

	ps := DepthMatchAll(root, func(pt MIMEPart) bool {
		return pt.ContentType() == "text/plain"
	})
	assert.Equal(t, 2, len(ps), "DepthMatchAll should have returned two matches")
	assert.True(t, ps[0].(*memMIMEPart) == b1,
		"DepthMatchAll should have returned b1, got %v", ps[0].FileName())
	assert.True(t, ps[1].(*memMIMEPart) == a2,
		"DepthMatchAll should have returned a2, got %v", ps[1].FileName())

	ps = DepthMatchAll(root, func(pt MIMEPart) bool {
		return pt.ContentType() == "text/html"
	})
	assert.Equal(t, 2, len(ps), "DepthMatchAll should have returned two matches")
	assert.True(t, ps[0].(*memMIMEPart) == b2,
		"DepthMatchAll should have returned b2, got %v", ps[0].FileName())
	assert.True(t, ps[1].(*memMIMEPart) == a3,
		"DepthMatchAll should have returned a3, got %v", ps[1].FileName())
}
