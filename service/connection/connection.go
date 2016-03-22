package connection

import (
	"time"

	"github.com/tapglue/multiverse/platform/metrics"
	"github.com/tapglue/multiverse/platform/service"
)

// Supported states for connections.
const (
	StateConfirmed State = "confirmed"
	StatePending   State = "pending"
	StateRejected  State = "rejected"
)

// Supported types for connections.
const (
	TypeFollow Type = "follow"
	TypeFriend Type = "friend"
)

// Connection represents a relation between two users.
type Connection struct {
	Enabled   bool      `json:"enabled"`
	FromID    uint64    `json:"user_from_id"`
	State     State     `json:"state"`
	ToID      uint64    `json:"user_to_id"`
	Type      Type      `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate performs checks on the Connection values for completeness and
// correctness.
func (c Connection) Validate() error {
	if c.FromID == 0 {
		return wrapError(ErrInvalidConnection, "from id not set")
	}

	if c.ToID == 0 {
		return wrapError(ErrInvalidConnection, "to id not set")
	}

	switch c.State {
	case StateConfirmed, StatePending, StateRejected:
		// valid
	default:
		return wrapError(ErrInvalidConnection, "invalid state")
	}

	switch c.Type {
	case TypeFollow, TypeFriend:
		// valid
	default:
		return wrapError(ErrInvalidConnection, "invalid type")
	}

	return nil
}

// List is a collection of Connections.
type List []*Connection

// FromIDs returns the extracted FromID of all connections as list.
func (l List) FromIDs() []uint64 {
	ids := []uint64{}

	for _, c := range l {
		ids = append(ids, c.ToID)
	}

	return ids
}

// ToIDs returns the extracted ToID of all connections as list.
func (l List) ToIDs() []uint64 {
	ids := []uint64{}

	for _, c := range l {
		ids = append(ids, c.ToID)
	}

	return ids
}

// QueryOptions are used to narrow down Connection queries.
type QueryOptions struct {
	Enabled *bool
	FromIDs []uint64
	States  []State
	ToIDs   []uint64
	Types   []Type
}

// Service for connection interactions.
type Service interface {
	metrics.BucketByDay
	service.Lifecycle

	Put(namespace string, connection *Connection) (*Connection, error)
	Query(namespace string, opts QueryOptions) (List, error)
}

// ServiceMiddleware is a chainable behaviour modifier for Service.
type ServiceMiddleware func(Service) Service

// State of a connection request.
type State string

// Type of a user relation.
type Type string
