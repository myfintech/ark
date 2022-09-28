package kube

import (
	"regexp"
	"strings"
)

var match = regexp.MustCompile(`\W+`)

// NormalizeNamespace cleans a namespace to ensure kubernetes compatibility
func NormalizeNamespace(namespace string) string {
	limit := 63
	normalized := strings.TrimSpace(
		strings.ToLower(
			match.ReplaceAllString(namespace, "-"),
		),
	)

	if len(normalized) < limit {
		return normalized
	} else {
		return normalized[:limit]
	}
}
