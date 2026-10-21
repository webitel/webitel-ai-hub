package query

import (
	"slices"
	"testing"
)

func TestParseSort(t *testing.T) {
	allow := allowNames("name", "id", "created_at")

	tests := []struct {
		name string
		expr string
		want []Order
	}{
		{name: "empty"},
		{name: "bare name ascends", expr: "name", want: []Order{{Field: "name"}}},
		{name: "explicit ascending", expr: "+name", want: []Order{{Field: "name"}}},
		{name: "descending", expr: "-created_at", want: []Order{{Field: "created_at", Desc: true}}},
		{
			name: "multiple and unknown dropped",
			expr: "name,title,-id",
			want: []Order{{Field: "name"}, {Field: "id", Desc: true}},
		},
		{name: "spaces trimmed", expr: " name , -id ", want: []Order{{Field: "name"}, {Field: "id", Desc: true}}},
		{name: "case folded to allowlist key", expr: "NAME", want: []Order{{Field: "name"}}},
		{name: "empty criteria dropped", expr: ",,name,", want: []Order{{Field: "name"}}},
		{name: "marker without field dropped", expr: "+,-"},
		{name: "nil allow drops everything", expr: "+name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAllow := allow
			if tt.name == "nil allow drops everything" {
				gotAllow = nil
			}

			got := ParseSort(tt.expr, gotAllow)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("ParseSort(%q) = %#v, want %#v", tt.expr, got, tt.want)
			}
		})
	}
}

func allowNames(names ...string) map[string]struct{} {
	allow := make(map[string]struct{}, len(names))
	for _, name := range names {
		allow[name] = struct{}{}
	}

	return allow
}
