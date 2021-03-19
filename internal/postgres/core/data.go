package core

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/models/operation"
)

// EventOperation -
type EventOperation struct {
	Network          string    `json:"network"`
	Hash             string    `json:"hash"`
	Internal         bool      `json:"internal"`
	Status           string    `json:"status"`
	Timestamp        time.Time `json:"timestamp"`
	Kind             string    `json:"kind"`
	Fee              int64     `json:"fee,omitempty"`
	Amount           int64     `json:"amount,omitempty"`
	Entrypoint       string    `json:"entrypoint,omitempty"`
	Source           string    `json:"source"`
	SourceAlias      string    `json:"source_alias,omitempty"`
	Destination      string    `json:"destination,omitempty"`
	DestinationAlias string    `json:"destination_alias,omitempty"`
	Delegate         string    `json:"delegate,omitempty"`
	DelegateAlias    string    `json:"delegate_alias,omitempty"`

	Result *operation.Result  `json:"result,omitempty"`
	Errors []*tezerrors.Error `json:"errors,omitempty"`
	Burned int64              `json:"burned,omitempty"`
}

// EventMigration -
type EventMigration struct {
	Network      string    `json:"network"`
	Protocol     string    `json:"protocol"`
	PrevProtocol string    `json:"prev_protocol,omitempty"`
	Hash         string    `json:"hash,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	Level        int64     `json:"level"`
	Address      string    `json:"address"`
	Kind         string    `json:"kind"`
}

// EventContract -
type EventContract struct {
	Network   string    `json:"network"`
	Address   string    `json:"address"`
	Hash      string    `json:"hash"`
	ProjectID string    `json:"project_id"`
	Timestamp time.Time `json:"timestamp"`
}
