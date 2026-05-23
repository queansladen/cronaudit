// Package tagfilter provides filtering of cron jobs by tag labels.
// Jobs may declare one or more tags in configuration; this package
// allows callers to select a subset of jobs matching a given tag set.
package tagfilter

import "strings"

// Job is the minimal interface required by the filter.
type Job struct {
	Name string
	Tags []string
}

// Filter holds a set of required tags used to match jobs.
type Filter struct {
	tags map[string]struct{}
}

// New creates a Filter that matches jobs containing ALL of the supplied tags.
// Tag matching is case-insensitive. An empty tag list matches every job.
func New(tags []string) *Filter {
	m := make(map[string]struct{}, len(tags))
	for _, t := range tags {
		m[strings.ToLower(strings.TrimSpace(t))] = struct{}{}
	}
	return &Filter{tags: m}
}

// Match reports whether the job satisfies the filter.
func (f *Filter) Match(job Job) bool {
	if len(f.tags) == 0 {
		return true
	}
	jobTags := make(map[string]struct{}, len(job.Tags))
	for _, t := range job.Tags {
		jobTags[strings.ToLower(strings.TrimSpace(t))] = struct{}{}
	}
	for required := range f.tags {
		if _, ok := jobTags[required]; !ok {
			return false
		}
	}
	return true
}

// Apply returns only those jobs from the slice that satisfy the filter.
func (f *Filter) Apply(jobs []Job) []Job {
	out := make([]Job, 0, len(jobs))
	for _, j := range jobs {
		if f.Match(j) {
			out = append(out, j)
		}
	}
	return out
}

// Tags returns the normalised set of tags held by the filter.
func (f *Filter) Tags() []string {
	result := make([]string, 0, len(f.tags))
	for t := range f.tags {
		result = append(result, t)
	}
	return result
}
