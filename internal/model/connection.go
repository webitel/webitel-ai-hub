package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	ModeResold = "resold"
	ModeByok   = "byok"
	ModeByom   = "byom"

	ChannelConsumer   = "consumer"
	ChannelEnterprise = "enterprise"

	AuthKindAPIKey            = "api_key"
	AuthKindGCPServiceAccount = "gcp_service_account"
	AuthKindAzureAD           = "azure_ad"
	AuthKindNone              = "none"
)

// Connection is a tenant-scoped provider endpoint configuration.
//
// Credentials is write-only: store reads never populate it.
type Connection struct {
	ID            uuid.UUID `json:"id" db:"id"`
	DomainID      int64     `json:"domain_id" db:"domain_id"`
	Name          string    `json:"name" db:"name"`
	Provider      string    `json:"provider" db:"provider"`
	Channel       string    `json:"channel" db:"channel"`
	Mode          string    `json:"mode" db:"mode"`
	AuthKind      string    `json:"auth_kind" db:"auth_kind"`
	Credentials   []byte    `json:"-" db:"credentials"`
	BaseURL       *string   `json:"base_url,omitempty" db:"base_url"`
	Region        *string   `json:"region,omitempty" db:"region"`
	ZeroRetention bool      `json:"zero_retention" db:"zero_retention"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy     *int64    `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy     *int64    `json:"updated_by,omitempty" db:"updated_by"`
}
