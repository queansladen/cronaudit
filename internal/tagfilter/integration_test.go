package tagfilter_test

import (
	"testing"

	"github.com/example/cronaudit/internal/tagfilter"
)

// TestRoundtrip_NewFilterFromTagsAndApply exercises the full New → Apply
// pipeline with a realistic job list to guard against regressions.
func TestRoundtrip_NewFilterFromTagsAndApply(t *testing.T) {
	alljobs := []tagfilter.Job{
		{Name: "db-backup", Tags: []string{"prod", "db", "critical"}},
		{Name: "db-vacuum", Tags: []string{"prod", "db", "maintenance"}},
		{Name: "log-rotate", Tags: []string{"prod", "maintenance"}},
		{Name: "staging-sync", Tags: []string{"staging", "db"}},
		{Name: "health-ping", Tags: []string{"prod"}},
	}

	tests := []struct {
		name      string
		tags      []string
		wantNames []string
	}{
		{
			name:      "prod only",
			tags:      []string{"prod"},
			wantNames: []string{"db-backup", "db-vacuum", "log-rotate", "health-ping"},
		},
		{
			name:      "prod and db",
			tags:      []string{"prod", "db"},
			wantNames: []string{"db-backup", "db-vacuum"},
		},
		{
			name:      "maintenance",
			tags:      []string{"maintenance"},
			wantNames: []string{"db-vacuum", "log-rotate"},
		},
		{
			name:      "no filter",
			tags:      nil,
			wantNames: []string{"db-backup", "db-vacuum", "log-rotate", "staging-sync", "health-ping"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := tagfilter.New(tc.tags)
			result := f.Apply(alljobs)
			if len(result) != len(tc.wantNames) {
				t.Fatalf("got %d jobs, want %d", len(result), len(tc.wantNames))
			}
			got := make(map[string]bool, len(result))
			for _, j := range result {
				got[j.Name] = true
			}
			for _, wn := range tc.wantNames {
				if !got[wn] {
					t.Errorf("missing expected job %q in result", wn)
				}
			}
		})
	}
}
