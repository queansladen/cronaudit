package differ_test

import (
	"strings"
	"testing"

	"github.com/example/cronaudit/internal/differ"
)

func TestCompare_NoDiff(t *testing.T) {
	r := differ.Compare("hello\nworld\n", "hello\nworld\n")
	if r.Changed {
		t.Fatal("expected Changed=false for identical outputs")
	}
	if r.Diff != "" {
		t.Fatalf("expected empty Diff, got %q", r.Diff)
	}
}

func TestCompare_EmptyToCurrent(t *testing.T) {
	r := differ.Compare("", "line1\nline2\n")
	if !r.Changed {
		t.Fatal("expected Changed=true")
	}
	if r.AddedLines == 0 {
		t.Fatal("expected added lines > 0")
	}
	if r.RemovedLines != 0 {
		t.Fatalf("expected RemovedLines=0, got %d", r.RemovedLines)
	}
}

func TestCompare_LineRemoved(t *testing.T) {
	prev := "alpha\nbeta\ngamma\n"
	curr := "alpha\ngamma\n"
	r := differ.Compare(prev, curr)
	if !r.Changed {
		t.Fatal("expected Changed=true")
	}
	if r.RemovedLines == 0 {
		t.Fatal("expected removed lines > 0")
	}
	if !strings.Contains(r.Diff, "- ") {
		t.Fatalf("expected diff to contain removal marker, got:\n%s", r.Diff)
	}
}

func TestCompare_LineAdded(t *testing.T) {
	prev := "alpha\n"
	curr := "alpha\nbeta\n"
	r := differ.Compare(prev, curr)
	if !r.Changed {
		t.Fatal("expected Changed=true")
	}
	if r.AddedLines == 0 {
		t.Fatal("expected added lines > 0")
	}
	if !strings.Contains(r.Diff, "+ ") {
		t.Fatalf("expected diff to contain addition marker, got:\n%s", r.Diff)
	}
}

func TestCompare_BothChanged(t *testing.T) {
	prev := "foo\nbar\n"
	curr := "foo\nbaz\n"
	r := differ.Compare(prev, curr)
	if !r.Changed {
		t.Fatal("expected Changed=true")
	}
	if r.AddedLines == 0 || r.RemovedLines == 0 {
		t.Fatalf("expected both added and removed lines, got added=%d removed=%d",
			r.AddedLines, r.RemovedLines)
	}
}
