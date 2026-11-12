package query

import "strings"

// SelectFields returns the requested names that exist in allow, in request order.
//
// Callers may pass one field per element or pack several into one element with
// commas or spaces. Names are lowercased and duplicates are dropped. An empty
// request returns nil so the caller can apply its own defaults. When the request
// is not empty, identity names that are themselves allowlisted are appended if
// they were not already selected.
func SelectFields(requested []string, allow map[string]struct{}, identity []string) []string {
	names := normalizeFields(requested)
	if len(names) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(names)+len(identity))

	selected := make([]string, 0, len(names)+len(identity))
	for _, name := range names {
		selected = appendAllowed(selected, seen, name, allow)
	}

	for _, name := range identity {
		selected = appendAllowed(selected, seen, name, allow)
	}

	if len(selected) == 0 {
		return nil
	}

	return selected
}

func normalizeFields(fields []string) []string {
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		for _, name := range strings.FieldsFunc(field, isFieldSeparator) {
			name = strings.ToLower(strings.TrimSpace(name))
			if name == "" {
				continue
			}

			out = append(out, name)
		}
	}

	if len(out) == 0 {
		return nil
	}

	return out
}

func isFieldSeparator(r rune) bool {
	return r == ',' || r == ' '
}

func appendAllowed(out []string, seen map[string]struct{}, name string, allow map[string]struct{}) []string {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return out
	}

	if _, ok := allow[name]; !ok {
		return out
	}

	if _, dup := seen[name]; dup {
		return out
	}

	seen[name] = struct{}{}

	return append(out, name)
}
