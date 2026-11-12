package query

import "strings"

// Order is one allowlisted sort key.
type Order struct {
	Field string
	Desc  bool
}

// ParseSort reads a comma-separated "+field,-field" expression.
//
// A bare name sorts ascending. Names missing from allow, or empty after the
// direction marker, are dropped. allow keys must be lowercase.
func ParseSort(expr string, allow map[string]struct{}) []Order {
	parts := strings.Split(expr, ",")
	orders := make([]Order, 0, len(parts))

	for _, part := range parts {
		order, ok := parseSortPart(part, allow)
		if !ok {
			continue
		}

		orders = append(orders, order)
	}

	if len(orders) == 0 {
		return nil
	}

	return orders
}

func parseSortPart(part string, allow map[string]struct{}) (Order, bool) {
	part = strings.TrimSpace(part)
	if part == "" {
		return Order{}, false
	}

	desc := false

	switch part[0] {
	case '+':
		part = part[1:]
	case '-':
		desc = true
		part = part[1:]
	}

	field := strings.ToLower(strings.TrimSpace(part))
	if _, ok := allow[field]; !ok {
		return Order{}, false
	}

	return Order{Field: field, Desc: desc}, true
}
