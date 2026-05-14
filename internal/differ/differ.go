// Package differ compares two run outputs and produces a human-readable diff.
package differ

import (
	"fmt"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// Result holds the outcome of a diff comparison.
type Result struct {
	// Changed is true when the two outputs differ.
	Changed bool
	// Diff is a unified-style text representation of the changes.
	Diff string
	// AddedLines is the count of lines present in current but not previous.
	AddedLines int
	// RemovedLines is the count of lines present in previous but not current.
	RemovedLines int
}

// Compare diffs previous and current output strings and returns a Result.
// When the two strings are identical Changed is false and Diff is empty.
func Compare(previous, current string) Result {
	if previous == current {
		return Result{Changed: false}
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(previous, current, true)
	diffs = dmp.DiffCleanupSemantic(diffs)

	var sb strings.Builder
	added, removed := 0, 0

	for _, d := range diffs {
		lines := strings.Count(d.Text, "\n")
		if lines == 0 && d.Text != "" {
			lines = 1
		}
		switch d.Type {
		case diffmatchpatch.DiffInsert:
			for _, line := range splitLines(d.Text) {
				fmt.Fprintf(&sb, "+ %s\n", line)
			}
			added += lines
		case diffmatchpatch.DiffDelete:
			for _, line := range splitLines(d.Text) {
				fmt.Fprintf(&sb, "- %s\n", line)
			}
			removed += lines
		case diffmatchpatch.DiffEqual:
			for _, line := range splitLines(d.Text) {
				fmt.Fprintf(&sb, "  %s\n", line)
			}
		}
	}

	return Result{
		Changed:      true,
		Diff:         sb.String(),
		AddedLines:   added,
		RemovedLines: removed,
	}
}

// splitLines splits text into individual lines, trimming a trailing empty
// element that results from a trailing newline.
func splitLines(s string) []string {
	parts := strings.Split(s, "\n")
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}
