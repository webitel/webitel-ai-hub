package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	ProfileKindLLM       = "llm"
	ProfileKindTTS       = "tts"
	ProfileKindSTT       = "stt"
	ProfileKindEmbedding = "embedding"
	ProfileKindRealtime  = "realtime"
	ProfileKindReranker  = "reranker"

	ProfileStateActive   = "active"
	ProfileStateDisabled = "disabled"
	ProfileStateArchived = "archived"
)

// DefaultProfileConfig is the empty profile config shape from schema.
var DefaultProfileConfig = json.RawMessage(`{"schema_version":1,"params":{},"limits":{}}`)

// DefaultProviderOptions is the empty provider_options default from schema.
var DefaultProviderOptions = json.RawMessage(`{}`)

// Profile binds a catalog model to a connection for one tenant.
type Profile struct {
	ID                  uuid.UUID       `json:"id" db:"id"`
	DomainID            int64           `json:"domain_id" db:"domain_id"`
	ConnectionID        uuid.UUID       `json:"connection_id" db:"connection_id"`
	Name                string          `json:"name" db:"name"`
	Kind                string          `json:"kind" db:"kind"`
	ModelID             string          `json:"model_id" db:"model_id"`
	ModelClass          string          `json:"model_class" db:"model_class"`
	State               string          `json:"state" db:"state"`
	IsDefault           bool            `json:"is_default" db:"is_default"`
	FallbackProfileID   *uuid.UUID      `json:"fallback_profile_id,omitempty" db:"fallback_profile_id"`
	Config              json.RawMessage `json:"config" db:"config"`
	EmbeddingDimensions *int            `json:"embedding_dimensions,omitempty" db:"embedding_dimensions"`
	ProviderOptions     json.RawMessage `json:"provider_options" db:"provider_options"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at" db:"updated_at"`
	CreatedBy           *int64          `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy           *int64          `json:"updated_by,omitempty" db:"updated_by"`
}
