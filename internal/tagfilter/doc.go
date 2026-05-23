// Package tagfilter provides tag-based filtering of cron jobs.
//
// Jobs may be labelled with one or more string tags in the cronaudit
// configuration file. The Filter type allows callers — such as the
// scheduler, reporter, and exporter — to restrict processing to only
// those jobs whose tag set is a superset of the required tags.
//
// Matching is case-insensitive and trims surrounding whitespace so that
// tag values like " Prod" and "prod" are treated as equivalent.
package tagfilter
