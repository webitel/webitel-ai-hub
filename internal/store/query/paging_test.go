package query

import (
	"slices"
	"testing"
)

func TestPageLimit(t *testing.T) {
	tests := []struct {
		name string
		size int
		want int
	}{
		{name: "positive fetches one extra", size: 20, want: 21},
		{name: "one", size: 1, want: 2},
		{name: "zero disables", size: 0, want: 0},
		{name: "negative disables", size: -1, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PageLimit(tt.size); got != tt.want {
				t.Fatalf("PageLimit(%d) = %d, want %d", tt.size, got, tt.want)
			}
		})
	}
}

func TestResolvePaging(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		items    []int
		want     []int
		wantNext bool
	}{
		{name: "fewer than size", size: 5, items: []int{1, 2}, want: []int{1, 2}},
		{name: "exactly size", size: 2, items: []int{1, 2}, want: []int{1, 2}},
		{name: "lookahead row trimmed", size: 2, items: []int{1, 2, 3}, want: []int{1, 2}, wantNext: true},
		{name: "unlimited passes through", size: -1, items: []int{1, 2, 3}, want: []int{1, 2, 3}},
		{name: "zero size passes through", size: 0, items: []int{1, 2, 3}, want: []int{1, 2, 3}},
		{name: "empty", size: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, next := ResolvePaging(tt.size, tt.items)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("items = %v, want %v", got, tt.want)
			}

			if next != tt.wantNext {
				t.Fatalf("next = %v, want %v", next, tt.wantNext)
			}
		})
	}
}
