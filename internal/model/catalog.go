package model

import "encoding/json"

const (
	CatalogSourceSeed    = "seed"
	CatalogSourceHubSync = "hub_sync"
	CatalogSourceManual  = "manual"
)

// CatalogKey identifies one row in the global model catalog.
type CatalogKey struct {
	Provider string
	Channel  string
	Kind     string
	ModelID  string
}

// ModelCatalogEntry is a platform-wide model catalog row.
type ModelCatalogEntry struct {
	Provider     string          `json:"provider" db:"provider"`
	Channel      string          `json:"channel" db:"channel"`
	Kind         string          `json:"kind" db:"kind"`
	ModelID      string          `json:"model_id" db:"model_id"`
	Class        string          `json:"class" db:"class"`
	UnitType     string          `json:"unit_type" db:"unit_type"`
	Capabilities json.RawMessage `json:"capabilities" db:"capabilities"`
	Source       string          `json:"source" db:"source"`
}
