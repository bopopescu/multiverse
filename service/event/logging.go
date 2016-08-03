package event

import (
	"time"

	"github.com/go-kit/kit/log"

	"github.com/tapglue/multiverse/platform/metrics"
)

type logService struct {
	logger log.Logger
	next   Service
}

// LogServiceMiddleware given a Logger wraps the next Service with logging capabilities.
func LogServiceMiddleware(logger log.Logger, store string) ServiceMiddleware {
	return func(next Service) Service {
		logger = log.NewContext(logger).With(
			"service", "event",
			"store", store,
		)

		return &logService{logger: logger, next: next}
	}
}

func (s *logService) ActiveUserIDs(ns string, p Period) (ids []uint64, err error) {
	defer func(begin time.Time) {
		ps := []interface{}{
			"datapoints", len(ids),
			"duration_ns", time.Since(begin).Nanoseconds(),
			"method", "ActiveUserIDs",
			"namespace", ns,
			"period", p,
		}

		if err != nil {
			ps = append(ps, "err", err)
		}

		_ = s.logger.Log(ps...)
	}(time.Now())

	return s.next.ActiveUserIDs(ns, p)
}

func (s *logService) Count(ns string, opts QueryOptions) (count int, err error) {
	defer func(begin time.Time) {
		ps := []interface{}{
			"count", count,
			"duration_ns", time.Since(begin).Nanoseconds(),
			"method", "Count",
			"namespace", ns,
			"opts", opts,
		}

		if err != nil {
			ps = append(ps, "err", err)
		}

		_ = s.logger.Log(ps...)
	}(time.Now())

	return s.next.Count(ns, opts)
}

func (s *logService) CreatedByDay(
	ns string,
	start, end time.Time,
) (ts metrics.Timeseries, err error) {
	defer func(begin time.Time) {
		ps := []interface{}{
			"datapoints", len(ts),
			"duration_ns", time.Since(begin).Nanoseconds(),
			"end", end.Format(metrics.BucketFormat),
			"method", "CreatedByDay",
			"namespace", ns,
			"start", start.Format(metrics.BucketFormat),
		}

		if err != nil {
			ps = append(ps, "err", err)
		}

		_ = s.logger.Log(ps...)
	}(time.Now())

	return s.next.CreatedByDay(ns, start, end)
}

func (s *logService) Put(ns string, input *Event) (output *Event, err error) {
	defer func(begin time.Time) {
		ps := []interface{}{
			"duration_ns", time.Since(begin).Nanoseconds(),
			"input", input,
			"method", "Put",
			"namespace", ns,
			"output", output,
		}

		if err != nil {
			ps = append(ps, "err", err)
		}

		_ = s.logger.Log(ps...)
	}(time.Now())

	return s.next.Put(ns, input)
}

func (s *logService) Query(ns string, opts QueryOptions) (list List, err error) {
	defer func(begin time.Time) {
		ps := []interface{}{
			"datapoints", len(list),
			"duration_ns", time.Since(begin).Nanoseconds(),
			"method", "Query",
			"namespace", ns,
			"opts", opts,
		}

		if err != nil {
			ps = append(ps, "err", err)
		}

		_ = s.logger.Log(ps...)
	}(time.Now())

	return s.next.Query(ns, opts)
}

func (s *logService) Setup(ns string) (err error) {
	defer func(begin time.Time) {
		ps := []interface{}{
			"duration_ns", time.Since(begin).Nanoseconds(),
			"method", "Setup",
			"namespace", ns,
		}

		if err != nil {
			ps = append(ps, "err", err)
		}

		_ = s.logger.Log(ps...)
	}(time.Now())

	return s.next.Setup(ns)
}

func (s *logService) Teardown(ns string) (err error) {
	defer func(begin time.Time) {
		ps := []interface{}{
			"duration_ns", time.Since(begin).Nanoseconds(),
			"method", "Teardown",
			"namespace", ns,
		}

		if err != nil {
			ps = append(ps, "err", err)
		}

		_ = s.logger.Log(ps...)
	}(time.Now())

	return s.next.Teardown(ns)
}

type logSource struct {
	logger log.Logger
	next   Source
}

// LogSourceMiddleware given a Logger raps the next Source logging capabilities.
func LogSourceMiddleware(store string, logger log.Logger) SourceMiddleware {
	return func(next Source) Source {
		logger = log.NewContext(logger).With(
			"source", "event",
			"store", store,
		)

		return &logSource{
			logger: logger,
			next:   next,
		}
	}
}

func (s *logSource) Ack(id string) (err error) {
	defer func(begin time.Time) {
		ps := []interface{}{
			"ack_id", id,
			"method", "Ack",
		}

		if err != nil {
			ps = append(ps, "err", err)
		}

		_ = s.logger.Log(ps...)
	}(time.Now())

	return s.next.Ack(id)
}

func (s *logSource) Consume() (change *StateChange, err error) {
	defer func(begin time.Time) {
		ps := []interface{}{
			"duration_ns", time.Since(begin).Nanoseconds(),
			"method", "Consume",
		}

		if change != nil {
			ps = append(ps,
				"namespace", change.Namespace,
				"object_new", change.New,
				"object_old", change.Old,
			)
		}

		if err != nil {
			ps = append(ps, "err", err)
		}

		_ = s.logger.Log(ps...)
	}(time.Now())

	return s.next.Consume()
}

func (s *logSource) Propagate(ns string, old, new *Event) (id string, err error) {
	defer func(begin time.Time) {
		ps := []interface{}{
			"duration_ns", time.Since(begin).Nanoseconds(),
			"id", id,
			"method", "Propagate",
			"namespace", ns,
			"object_new", new,
			"object_old", old,
		}

		if err != nil {
			ps = append(ps, "err", err)
		}

		_ = s.logger.Log(ps...)
	}(time.Now())

	return s.next.Propagate(ns, old, new)
}
