package event

import (
	"fmt"
	"strconv"
	"time"

	"github.com/tapglue/multiverse/platform/metrics"
	"github.com/tapglue/multiverse/platform/service"
	v04_entity "github.com/tapglue/multiverse/v04/entity"
)

// Predefined time periods to use for aggregates.
const (
	ByDay   Period = "1 day"
	ByWeek  Period = "1 week"
	ByMonth Period = "1 month"
)

// Visibility variants available for Events.
const (
	VisibilityPrivate Visibility = (iota + 1) * 10
	VisibilityConnection
	VisibilityPublic
	VisibilityGlobal
)

// TG reserved keywords for types.
const (
	TargetUser = "tg_user"
	TypeFollow = "tg_follow"
	TypeFriend = "tg_friend"
)

// AggregateService for event interactions.
type AggregateService interface {
	ActiveUserIDs(string, Period) ([]uint64, error)
}

// Event is the buidling block to express interaction on internal/external
// objects.
type Event struct {
	Enabled    bool       `json:"enabled"`
	ID         uint64     `json:"id"`
	Language   string     `json:"language,omitempty"`
	Object     *Object    `json:"object,omitempty"`
	ObjectID   uint64     `json:"object_id"`
	Owned      bool       `json:"owned"`
	Target     *Target    `json:"target"`
	Type       string     `json:"type"`
	UserID     uint64     `json:"user_id"`
	Visibility Visibility `json:"visibility"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// Validate performs semantic checks on the passed Event values for correctness.
func (e Event) Validate() error {
	if e.Type == "" {
		return wrapError(ErrInvalidEvent, "missing type")
	}

	if e.UserID == 0 {
		return wrapError(ErrInvalidEvent, "missing owner")
	}

	if e.Visibility < 10 || e.Visibility > 50 {
		return wrapError(ErrInvalidEvent, "visibility not supported")
	}

	return nil
}

// Mape is an event collection with the id as index.
type Map map[uint64]*Event

// List is an Event collection.
type List []*Event

// IDs returns ID for every Event.
func (es List) IDs() []uint64 {
	ids := []uint64{}

	for _, e := range es {
		ids = append(ids, e.ID)
	}

	return ids
}

func (es List) Len() int {
	return len(es)
}

func (es List) Less(i, j int) bool {
	return es[i].CreatedAt.After(es[j].CreatedAt)
}

func (es List) Swap(i, j int) {
	es[i], es[j] = es[j], es[i]
}

// UserIDs returns UserID for every Event.
func (es List) UserIDs() []uint64 {
	ids := []uint64{}

	for _, e := range es {
		ids = append(ids, e.UserID)

		// Extract user ids from target as well.
		if e.Target != nil && e.Target.Type == v04_entity.TypeTargetUser {
			id, err := strconv.ParseUint(e.Target.ID, 10, 64)
			if err != nil {
				// We fail silently here for now until we find a way to log this. As the
				// only effect is that we don't add a potential user to the map
				continue
			}

			ids = append(ids, id)
		}
	}

	return ids
}

// Object describes an external entity whcih can have a type and an id.
type Object struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	URL  string `json:"url"`
}

// QueryOptions are used to narrow down Event queries.
type QueryOptions struct {
	Enabled             *bool
	ExternalObjectIDs   []string
	ExternalObjectTypes []string
	IDs                 []uint64
	ObjectIDs           []uint64
	Owned               *bool
	TargetIDs           []string
	TargetTypes         []string
	Types               []string
	UserIDs             []uint64
	Visibilities        []Visibility
}

// Target describes the person addressed in an event. To be phased out.
type Target struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// Period is a pre-defined time duration.
type Period string

// Service for event interactions.
type Service interface {
	AggregateService
	metrics.BucketByDay
	service.Lifecycle

	Count(namespace string, opts QueryOptions) (int, error)
	Put(namespace string, event *Event) (*Event, error)
	Query(namespace string, opts QueryOptions) (List, error)
}

// ServiceMiddleware is a chainable behaviour modifier for Service.
type ServiceMiddleware func(Service) Service

// Visibility determines the visibility of Objects when consumed.
type Visibility uint8

func flakeNamespace(ns string) string {
	return fmt.Sprintf("%s_%s", ns, "events")
}
