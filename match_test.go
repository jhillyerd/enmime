package enmime

import (
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
	if p == nil {
		t.Fatal("BreadthMatchFirst should have returned a result for text/plain")
	}
	if p.(*memMIMEPart) != a2 {
		t.Error("BreadthMatchFirst should have returned a2, got:", p.FileName())
	}

	p = BreadthMatchFirst(root, func(pt MIMEPart) bool {
		return pt.ContentType() == "text/html"
	})
	if p == nil {
		t.Fatal("BreadthMatchFirst should have returned a result for text/html")
	}
	if p.(*memMIMEPart) != a3 {
		t.Error("BreadthMatchFirst should have returned a3, got:", p.FileName())
	}
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
	if len(ps) != 2 {
		t.Fatal("BreadthMatchAll should have returned two matches, got:", len(ps))
	}
	if ps[0].(*memMIMEPart) != a2 {
		t.Error("BreadthMatchAll should have returned a2, got:", ps[0].FileName())
	}
	if ps[1].(*memMIMEPart) != b1 {
		t.Error("BreadthMatchAll should have returned b1, got:", ps[1].FileName())
	}

	ps = BreadthMatchAll(root, func(pt MIMEPart) bool {
		return pt.ContentType() == "text/html"
	})
	if len(ps) != 2 {
		t.Fatal("BreadthMatchAll should have returned two matches, got:", len(ps))
	}
	if ps[0].(*memMIMEPart) != a3 {
		t.Error("BreadthMatchAll should have returned a3, got:", ps[0].FileName())
	}
	if ps[1].(*memMIMEPart) != b2 {
		t.Error("BreadthMatchAll should have returned b2, got:", ps[1].FileName())
	}
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
	if p == nil {
		t.Fatal("DepthMatchFirst should have returned a result for text/plain")
	}
	if p.(*memMIMEPart) != b1 {
		t.Error("DepthMatchFirst should have returned b1, got:", p.FileName())
	}

	p = DepthMatchFirst(root, func(pt MIMEPart) bool {
		return pt.ContentType() == "text/html"
	})
	if p.(*memMIMEPart) != b2 {
		t.Error("DepthMatchFirst should have returned b2, got:", p.FileName())
	}
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
	if len(ps) != 2 {
		t.Fatal("DepthMatchAll should have returned two matches, got:", len(ps))
	}
	if ps[0].(*memMIMEPart) != b1 {
		t.Error("DepthMatchAll should have returned b1, got:", ps[0].FileName())
	}
	if ps[1].(*memMIMEPart) != a2 {
		t.Error("DepthMatchAll should have returned a2, got:", ps[1].FileName())
	}

	ps = DepthMatchAll(root, func(pt MIMEPart) bool {
		return pt.ContentType() == "text/html"
	})
	if len(ps) != 2 {
		t.Fatal("DepthMatchAll should have returned two matches, got:", len(ps))
	}
	if ps[0].(*memMIMEPart) != b2 {
		t.Error("DepthMatchAll should have returned b2, got:", ps[0].FileName())
	}
	if ps[1].(*memMIMEPart) != a3 {
		t.Error("DepthMatchAll should have returned a3, got:", ps[1].FileName())
	}
}
