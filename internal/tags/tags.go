// Package tags detects and processes SRT style tags in subtitle text.
package tags

import "regexp"

// Handler transforms the inner text of a matched tag.
type Handler func(inner string) string

// registry maps lowercase tag names to their handlers.
// Tags not in the registry are stripped (inner text preserved).
var registry = map[string]Handler{
	"i": func(s string) string { return "_" + s + "_" },
}

var (
	// Matches <tagname ...>inner</tagname> (case-insensitive, non-greedy).
	htmlTag = regexp.MustCompile(`(?i)<([a-z]+)[^>]*>(.*?)</[a-z]+>`)
	// Matches SSA override blocks like {\an8} or {\pos(10,20)}.
	ssaTag = regexp.MustCompile(`\{\\[^}]*\}`)
)

// Process removes or maps all known SRT style tags in text.
func Process(text string) string {
	// Strip SSA tags entirely.
	text = ssaTag.ReplaceAllString(text, "")

	// Apply HTML-style tag handlers iteratively (handles nesting).
	for htmlTag.MatchString(text) {
		text = htmlTag.ReplaceAllStringFunc(text, func(match string) string {
			sub := htmlTag.FindStringSubmatch(match)
			if sub == nil {
				return match
			}
			name, inner := sub[1], sub[2]
			if h, ok := registry[name]; ok {
				return h(inner)
			}
			return inner // strip unknown/unsupported tags
		})
	}
	return text
}
