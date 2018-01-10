package test

import "testing"

func TestEqualLines(t *testing.T) {
	got := make([]byte, 0)
	want := make([]byte, 0)
	mockt := &testing.T{}
	DiffLines(mockt, got, want)
	if mockt.Failed() {
		t.Error("Two empty slices should succeed")
	}

	got = []byte("foo\n")
	want = []byte("foo\n")
	mockt = &testing.T{}
	DiffLines(mockt, got, want)
	if mockt.Failed() {
		t.Error("Two equal single line slices should succeed")
	}

	got = []byte("foo\nbar\nbaz\nenmime\n1\n2\n3\n4\n5\n")
	want = []byte("foo\nbar\nbaz\nenmime\n1\n2\n3\n4\n5\n")
	mockt = &testing.T{}
	DiffLines(mockt, got, want)
	if mockt.Failed() {
		t.Error("Two equal multiline slices should succeed")
	}
}

func TestDifferingLines(t *testing.T) {
	got := []byte("foo")
	want := []byte("bar")
	mockt := &testing.T{}
	DiffLines(mockt, got, want)
	if !mockt.Failed() {
		t.Error("Two differing single line slices should fail")
	}

	got = []byte("foo\n")
	want = []byte("bar\n")
	mockt = &testing.T{}
	DiffLines(mockt, got, want)
	if !mockt.Failed() {
		t.Error("Two differing single line slices should fail")
	}

	got = []byte("foo\nbar\nbaz\nenmime\n1\n2\n3\n4\n5\n")
	want = []byte("foo\nbar\nbaz\ninbucket\n1\n2\n3\n4\n5\n")
	mockt = &testing.T{}
	DiffLines(mockt, got, want)
	if !mockt.Failed() {
		t.Error("Two differing multiline slices should fail")
	}

	// Test missing EOL
	got = []byte("foo\nenmime")
	want = []byte("foo\ninbucket")
	mockt = &testing.T{}
	DiffLines(mockt, got, want)
	if !mockt.Failed() {
		t.Error("Two differing no-EOL slices should fail")
	}
}
