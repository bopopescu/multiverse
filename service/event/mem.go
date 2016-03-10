package event

import (
	"fmt"
	"time"

	"github.com/tapglue/multiverse/platform/flake"
	"github.com/tapglue/multiverse/platform/metrics"
)

type memService struct {
	events map[string]map[uint64]*Event
}

// NewMemService returns a memory backed implementation of Service.
func NewMemService() Service {
	return &memService{
		events: map[string]map[uint64]*Event{},
	}
}

func (s *memService) ActiveUserIDs(ns string, p Period) ([]uint64, error) {
	return nil, nil
}

func (s *memService) CreatedByDay(
	ns string,
	start, end time.Time,
) (metrics.Timeseries, error) {
	bucket, ok := s.events[ns]
	if !ok {
		return nil, ErrNamespaceNotFound
	}

	counts := map[string]int{}

	for _, event := range bucket {
		if event.CreatedAt.Before(start) || event.CreatedAt.After(end) {
			continue
		}

		b := event.CreatedAt.Format(metrics.BucketFormat)

		if _, ok := counts[b]; !ok {
			counts[b] = 0
		}

		counts[b]++
	}

	ts := metrics.Timeseries{}

	for bucket, value := range counts {
		ts = append(ts, metrics.Datapoint{
			Bucket: bucket,
			Value:  value,
		})
	}

	return ts, nil
}

func (s *memService) Put(ns string, event *Event) (*Event, error) {
	bucket, ok := s.events[ns]
	if !ok {
		return nil, ErrNamespaceNotFound
	}

	if event.ID == 0 {
		id, err := flake.NextID(flakeNamespace(ns))
		if err != nil {
			return nil, err
		}

		event.CreatedAt = time.Now()
		event.ID = id
	} else {
		keep := false

		for _, e := range bucket {
			if e.ID == event.ID {
				keep = true
				event.CreatedAt = e.CreatedAt
			}
		}

		if !keep {
			return nil, fmt.Errorf("event not found")
		}
	}

	event.UpdatedAt = time.Now()
	bucket[event.ID] = copy(event)

	return copy(event), nil
}

func (s *memService) Query(ns string, opts QueryOptions) (List, error) {
	bucket, ok := s.events[ns]
	if !ok {
		return nil, ErrNamespaceNotFound
	}

	es := List{}

	for id, event := range bucket {
		if opts.Enabled != nil && event.Enabled != *opts.Enabled {
			continue
		}

		if event.Object == nil && len(opts.ExternalObjectIDs) > 0 {
			continue
		}

		if event.Object == nil && len(opts.ExternalObjectTypes) > 0 {
			continue
		}

		if event.Object != nil && !inTypes(event.Object.ID, opts.ExternalObjectIDs) {
			continue
		}

		if event.Object != nil && !inTypes(event.Object.Type, opts.ExternalObjectTypes) {
			continue
		}

		if !inIDs(id, opts.IDs) {
			continue
		}

		if !inIDs(event.ObjectID, opts.ObjectIDs) {
			continue
		}

		if opts.Owned != nil {
			if event.Owned != *opts.Owned {
				continue
			}
		}

		if event.Target != nil && !inTypes(event.Target.ID, opts.TargetIDs) {
			continue
		}

		if event.Target != nil && !inTypes(event.Target.Type, opts.TargetTypes) {
			continue
		}

		if !inTypes(event.Type, opts.Types) {
			continue
		}

		if !inIDs(event.UserID, opts.UserIDs) {
			continue
		}

		if !inVisibilities(event.Visibility, opts.Visibilities) {
			continue
		}

		es = append(es, event)
	}

	return es, nil
}

func (s *memService) Setup(ns string) error {
	if _, ok := s.events[ns]; !ok {
		s.events[ns] = map[uint64]*Event{}
	}

	return nil
}

func (s *memService) Teardown(ns string) error {
	if _, ok := s.events[ns]; ok {
		delete(s.events, ns)
	}

	return nil
}

func copy(e *Event) *Event {
	old := *e
	return &old
}

func inIDs(id uint64, ids []uint64) bool {
	if len(ids) == 0 {
		return true
	}

	keep := false

	for _, i := range ids {
		if id == i {
			keep = true
			break
		}
	}

	return keep
}

func inTypes(ty string, ts []string) bool {
	if len(ts) == 0 {
		return true
	}

	keep := false

	for _, t := range ts {
		if ty == t {
			keep = true
			break
		}
	}

	return keep
}

func inVisibilities(visibility Visibility, vs []Visibility) bool {
	if len(vs) == 0 {
		return true
	}

	keep := false

	for _, v := range vs {
		if visibility == v {
			keep = true
			break
		}
	}

	return keep
}
