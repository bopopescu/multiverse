package org

import (
	"time"

	kitmetrics "github.com/go-kit/kit/metrics"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/tapglue/multiverse/errors"
	"github.com/tapglue/multiverse/platform/metrics"
	v04_entity "github.com/tapglue/multiverse/v04/entity"
)

const serviceName = "org"

type instrumentStrangleService struct {
	StrangleService

	component string
	errCount  kitmetrics.Counter
	opCount   kitmetrics.Counter
	opLatency *prometheus.HistogramVec
	store     string
}

// InstrumentStrangleMiddleware observes key aspects of Service operations and
// exposes Prometheus metrics.
func InstrumentStrangleMiddleware(
	component, store string,
	errCount kitmetrics.Counter,
	opCount kitmetrics.Counter,
	opLatency *prometheus.HistogramVec,
) StrangleMiddleware {
	return func(next StrangleService) StrangleService {
		return &instrumentStrangleService{
			component:       component,
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
		if errs != nil {
			s.errCount.With(
				metrics.FieldComponent, s.component,
				metrics.FieldMethod, "FindByKey",
				metrics.FieldNamespace, "",
				metrics.FieldService, serviceName,
				metrics.FieldStore, s.store,
			).Add(1)
		}

		s.opCount.With(
			metrics.FieldComponent, s.component,
			metrics.FieldMethod, "FindByKey",
			metrics.FieldNamespace, "",
			metrics.FieldService, serviceName,
			metrics.FieldStore, s.store,
		).Add(1)

		s.opLatency.With(prometheus.Labels{
			metrics.FieldComponent: s.component,
			metrics.FieldMethod:    "FindByKey",
			metrics.FieldNamespace: "",
			metrics.FieldService:   serviceName,
			metrics.FieldStore:     s.store,
		}).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.StrangleService.FindByKey(key)
}
