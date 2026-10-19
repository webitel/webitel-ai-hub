// Package query holds pure helpers for paging, sorting and field selection.
//
// Allowlist keys are lowercase API names. A returned name is one of those keys,
// never the raw caller text, so it can be placed in SQL only after the caller
// maps it to a fixed expression.
package query

// PageLimit is the SQL LIMIT for a page request.
//
// A positive size fetches one extra row so the caller can tell that another page
// exists. A non-positive size disables paging and returns 0, which means the
// caller omits LIMIT.
func PageLimit(size int) int {
	if size <= 0 {
		return 0
	}

	return size + 1
}

// ResolvePaging trims a lookahead page. When more than size rows arrived, the
// extra row is dropped and next reports a further page. A non-positive size
// means paging was disabled: items pass through unchanged.
func ResolvePaging[T any](size int, items []T) ([]T, bool) {
	if size > 0 && len(items) > size {
		return items[:size], true
	}

	return items, false
}
