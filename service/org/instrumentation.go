package org

import (
	"strings"
	"time"

	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/tapglue/multiverse/errors"
	v04_entity "github.com/tapglue/multiverse/v04/entity"
)

const (
	fieldMethod = "method"
	fieldStore  = "store"
)

type instrumentStrangleService struct {
	StrangleService

	errCount  metrics.Counter
	opCount   metrics.Counter
	opLatency metrics.TimeHistogram
	store     string
}

// InstrumentStrangleMiddleware observes key aspects of Service operations and
// exposes Prometheus metrics.
func InstrumentStrangleMiddleware(ns, store string) StrangleMiddleware {
	var (
		fieldKeys = []string{fieldMethod, fieldStore}
		namespace = strings.Replace(ns, "-", "_", -1)
		subsytem  = "service_org"

		errCount = kitprometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsytem,
			Name:      "err_count",
			Help:      "Number of failed operations",
		}, fieldKeys)
		opCount = kitprometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsytem,
			Name:      "op_count",
			Help:      "Number of operations performed",
		}, fieldKeys)
		opLatency = metrics.NewTimeHistogram(
			time.Second,
			kitprometheus.NewHistogram(
				prometheus.HistogramOpts{
					Namespace: namespace,
					Subsystem: subsytem,
					Name:      "op_latency_seconds",
					Help:      "Distribution of op duration in seconds",
				},
				fieldKeys,
			),
		)
	)

	return func(next StrangleService) StrangleService {
		return &instrumentStrangleService{
			errCount:        errCount,
			opCount:         opCount,
			opLatency:       opLatency,
			store:           store,
			StrangleService: next,
		}
	}
}

func (s *instrumentStrangleService) FindByKey(
	key string,
) (org *v04_entity.Organization, errs []errors.Error) {
	defer func(begin time.Time) {
		var (
			m = metrics.Field{
				Key:   fieldMethod,
				Value: "FindByKey",
			}
			store = metrics.Field{
				Key:   fieldStore,
				Value: s.store,
			}
		)

		if errs != nil {
			s.errCount.With(m).With(store).Add(1)
		}

		s.opCount.With(m).With(store).Add(1)
		s.opLatency.With(m).With(store).Observe(time.Since(begin))
	}(time.Now())

	return s.StrangleService.FindByKey(key)
}
