package test

import "testing"

func TestEqualStrings(t *testing.T) {
	got := make([]string, 0)
	want := make([]string, 0)
	mockt := &testing.T{}
	DiffStrings(mockt, got, want)
	if mockt.Failed() {
		t.Error("Two empty slices should succeed")
	}

	got = []string{"foo"}
	want = []string{"foo"}
	mockt = &testing.T{}
	DiffStrings(mockt, got, want)
	if mockt.Failed() {
		t.Error("Two equal single line slices should succeed")
	}

	got = []string{"foo", "bar", "baz", "enmime", "1", "2", "3", "4", "5", ""}
	want = []string{"foo", "bar", "baz", "enmime", "1", "2", "3", "4", "5", ""}
	mockt = &testing.T{}
	DiffStrings(mockt, got, want)
	if mockt.Failed() {
		t.Error("Two equal multiline slices should succeed")
	}
}

func TestDifferingStrings(t *testing.T) {
	got := []string{"foo"}
	want := []string{"bar"}
	mockt := &testing.T{}
	DiffStrings(mockt, got, want)
	if !mockt.Failed() {
		t.Error("Two differing single line slices should fail")
	}

	got = []string{"foo", "bar", "baz", "enmime", "1", "2", "3", "4", "5", ""}
	want = []string{"foo", "bar", "baz", "inbucket", "1", "2", "3", "4", "5", ""}
	mockt = &testing.T{}
	DiffStrings(mockt, got, want)
	if !mockt.Failed() {
		t.Error("Two differing multiline slices should fail")
	}
}

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

func TestGoldenMissing(t *testing.T) {
	mockt := &testing.T{}
	done := make(chan struct{})
	go func() {
		// t.Fatal exits current goroutine, calling deferrals
		defer close(done)
		DiffGolden(mockt, []byte{}, "zzzDOESNTEXIST")
	}()
	<-done
	if !mockt.Failed() {
		t.Error("Missing golden file should fail test")
	}
}

func TestGolden(t *testing.T) {
	mockt := &testing.T{}
	DiffGolden(mockt, []byte("one\n"), "testdata", "test.golden")
	if !mockt.Failed() {
		t.Error("Differing bytes in golden file should fail test")
	}

	mockt = &testing.T{}
	DiffGolden(mockt, []byte("one\ntwo\nthree\n"), "testdata", "test.golden")
	if mockt.Failed() {
		t.Error("Same bytes in golden file should not fail test")
	}
}
