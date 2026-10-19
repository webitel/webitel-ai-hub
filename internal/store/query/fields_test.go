package query

import (
	"slices"
	"testing"
)

func TestSelectFields(t *testing.T) {
	allow := allowNames("id", "name", "created_at")
	identity := []string{"id"}

	tests := []struct {
		name      string
		requested []string
		identity  []string
		want      []string
	}{
		{name: "empty request skips defaults"},
		{name: "blanks are empty", requested: []string{"", "  "}},
		{name: "one per element", requested: []string{"id", "name"}, want: []string{"id", "name"}},
		{name: "comma and space packed", requested: []string{"id, name created_at"}, want: []string{"id", "name", "created_at"}},
		{name: "lowercased", requested: []string{"ID", "Name"}, want: []string{"id", "name"}},
		{name: "unknown dropped", requested: []string{"name,title"}, identity: identity, want: []string{"name", "id"}},
		{name: "duplicates dropped", requested: []string{"name", "name", "id"}, want: []string{"name", "id"}},
		{name: "identity already selected", requested: []string{"id", "name"}, identity: identity, want: []string{"id", "name"}},
		{name: "only unknown still keeps identity", requested: []string{"title"}, identity: identity, want: []string{"id"}},
		{name: "unknown without identity", requested: []string{"title"}},
		{name: "identity outside allow dropped", requested: []string{"name"}, identity: []string{"secret"}, want: []string{"name"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SelectFields(tt.requested, allow, tt.identity)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("SelectFields() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestSelectFieldsNilAllow(t *testing.T) {
	got := SelectFields([]string{"id"}, nil, []string{"id"})
	if got != nil {
		t.Fatalf("SelectFields() = %#v, want nil", got)
	}
}
