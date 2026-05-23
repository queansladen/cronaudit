package tagfilter_test

import (
	"testing"

	"github.com/example/cronaudit/internal/tagfilter"
)

func jobs() []tagfilter.Job {
	return []tagfilter.Job{
		{Name: "backup", Tags: []string{"prod", "db"}},
		{Name: "report", Tags: []string{"prod", "reporting"}},
		{Name: "cleanup", Tags: []string{"staging", "db"}},
		{Name: "ping", Tags: []string{}},
	}
}

func TestMatch_EmptyFilterMatchesAll(t *testing.T) {
	f := tagfilter.New(nil)
	for _, j := range jobs() {
		if !f.Match(j) {
			t.Errorf("expected empty filter to match job %q", j.Name)
		}
	}
}

func TestMatch_SingleTag(t *testing.T) {
	f := tagfilter.New([]string{"prod"})
	want := map[string]bool{"backup": true, "report": true, "cleanup": false, "ping": false}
	for _, j := range jobs() {
		got := f.Match(j)
		if got != want[j.Name] {
			t.Errorf("Match(%q) = %v, want %v", j.Name, got, want[j.Name])
		}
	}
}

func TestMatch_MultipleTagsRequiresAll(t *testing.T) {
	f := tagfilter.New([]string{"prod", "db"})
	if !f.Match(tagfilter.Job{Name: "backup", Tags: []string{"prod", "db"}}) {
		t.Error("expected backup to match prod+db")
	}
	if f.Match(tagfilter.Job{Name: "report", Tags: []string{"prod", "reporting"}}) {
		t.Error("expected report NOT to match prod+db")
	}
}

func TestMatch_CaseInsensitive(t *testing.T) {
	f := tagfilter.New([]string{"PROD"})
	if !f.Match(tagfilter.Job{Name: "x", Tags: []string{"prod"}}) {
		t.Error("expected case-insensitive match")
	}
}

func TestApply_FiltersSlice(t *testing.T) {
	f := tagfilter.New([]string{"db"})
	result := f.Apply(jobs())
	if len(result) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(result))
	}
	names := map[string]bool{}
	for _, j := range result {
		names[j.Name] = true
	}
	if !names["backup"] || !names["cleanup"] {
		t.Errorf("unexpected jobs in result: %v", names)
	}
}

func TestApply_EmptyFilterReturnsAll(t *testing.T) {
	f := tagfilter.New([]string{})
	result := f.Apply(jobs())
	if len(result) != len(jobs()) {
		t.Errorf("expected all jobs, got %d", len(result))
	}
}

func TestTags_ReturnsNormalisedTags(t *testing.T) {
	f := tagfilter.New([]string{" Prod ", "DB"})
	tags := f.Tags()
	set := make(map[string]bool)
	for _, tag := range tags {
		set[tag] = true
	}
	if !set["prod"] || !set["db"] {
		t.Errorf("unexpected tags: %v", tags)
	}
}
