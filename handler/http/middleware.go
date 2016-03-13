package http

import (
	"compress/gzip"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"

	"github.com/tapglue/multiverse/errors"
	"github.com/tapglue/multiverse/limiter"
	"github.com/tapglue/multiverse/service/app"
	"github.com/tapglue/multiverse/service/member"
	"github.com/tapglue/multiverse/service/org"
	"github.com/tapglue/multiverse/service/user"
	v04_entity "github.com/tapglue/multiverse/v04/entity"
)

// CORS adds the standard set of CORS headers.
func CORS() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "User-Agent, Content-Type, Content-Length, Accept-Encoding, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next(ctx, w, r)
		}
	}
}

// CtxApp extracts the App from the Authentication header.
func CtxApp(apps app.StrangleService) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			token, _, ok := r.BasicAuth()
			if !ok {
				respondError(w, 1001, wrapError(ErrUnauthorized, "application user not found"))
				return
			}

			var (
				app  *v04_entity.Application
				errs []errors.Error
			)

			if len(token) == 32 {
				app, errs = apps.FindByApplicationToken(token)
				if errs != nil {
					respondError(w, 0, errs[0])
					return
				}

				ctx = tokenTypeInContext(ctx, tokenApplication)
			} else if len(token) == 44 {
				app, errs = apps.FindByBackendToken(token)
				if errs != nil {
					respondError(w, 0, errs[0])
					return
				}

				ctx = tokenTypeInContext(ctx, tokenBackend)
			}

			if app == nil {
				respondError(w, 1001, wrapError(ErrUnauthorized, "application not found"))
				return
			}

			next(appInContext(ctx, app), w, r)
		}
	}
}

// CtxMember extracts the member from the Authentication header and adds it to the
// Context.
func CtxMember(members member.StrangleService) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			_, token, ok := r.BasicAuth()
			if !ok {
				respondError(w, 4005, wrapError(ErrUnauthorized, "error while reading org credentials"))
				return
			}

			member, errs := members.FindBySession(token)
			if errs != nil {
				respondError(w, 0, errs[0])
				return
			}

			if member == nil {
				respondError(w, 7004, wrapError(ErrUnauthorized, "member not found"))
				return
			}

			next(memberInContext(ctx, member), w, r)
		}
	}
}

// CtxOrg extracts the Org from the Authentication header.
func CtxOrg(orgs org.StrangleService) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			token, _, ok := r.BasicAuth()
			if !ok {
				respondError(w, 4004, wrapError(ErrUnauthorized, "error while reading org credentials"))
				return
			}

			org, errs := orgs.FindByKey(token)
			if errs != nil {
				respondError(w, 0, errs[0])
				return
			}

			if org == nil {
				respondError(w, 6008, wrapError(ErrUnauthorized, "org not found"))
				return
			}

			next(orgInContext(ctx, org), w, r)
		}
	}
}

// CtxPrepare adds a baseline of information to the Context currently:
// * api version
// * route name
func CtxPrepare(version string) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			route := "unknown"

			if current := mux.CurrentRoute(r); current != nil {
				route = current.GetName()
			}

			ctx = routeInContext(ctx, route)
			ctx = versionInContext(ctx, version)

			next(ctx, w, r)
		}
	}
}

// CtxUser extracts the user from the Authentication header and adds it to the
// Context.
func CtxUser(users user.StrangleService) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			var (
				app       = appFromContext(ctx)
				tokenType = tokenFromContext(ctx)
			)

			_, token, ok := r.BasicAuth()
			if !ok {
				respondError(w, 4007, wrapError(ErrUnauthorized, "error while reading user credentials"))
				return
			}

			if token == "" {
				respondError(w, 4013, wrapError(ErrUnauthorized, "session token missing from request"))
				return
			}

			var user *v04_entity.ApplicationUser

			switch tokenType {
			case tokenApplication:
				var errs []errors.Error

				user, errs = users.FindBySession(app.OrgID, app.ID, token)
				if errs != nil {
					respondError(w, 0, errs[0])
					return
				}
			case tokenBackend:
				var errs []errors.Error

				id, err := strconv.ParseUint(token, 10, 64)
				if err != nil {
					respondError(w, 0, err)
					return
				}

				user, errs = users.Read(app.OrgID, app.ID, id, false)
				if errs != nil {
					respondError(w, 0, errs[0])
					return
				}
			default:
				respondError(w, 4007, wrapError(ErrUnauthorized, "error while reading user credentials"))
				return
			}

			next(userInContext(ctx, user), w, r)
		}
	}
}

// DebugHeaders adds extra information encoded in a custom header namespace for
// potential tracing and debugging post-mortem.
func DebugHeaders(rev, host string) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Tapglue-Host", host)
			w.Header().Set("X-Tapglue-Revision", rev)

			next(ctx, w, r)
		}
	}
}

// Gzip ensures proper encoding of the response if the client accepts it.
func Gzip() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				w.Header().Set("Content-Encoding", "gzip")

				gz := gzip.NewWriter(w)
				defer gz.Close()

				w = gzipResponseWriter{w, gz}
			}

			next(ctx, w, r)
		}
	}
}

// HasUserAgent ensures a valid User-Agent is set.
func HasUserAgent() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("User-Agent") == "" {
				respondError(w, 5002, wrapError(ErrBadRequest, "User-Agent header must be set"))
				return
			}

			next(ctx, w, r)
		}
	}
}

// Instrument observes key aspects of a request/response and exposes Prometheus
// metrics.
func Instrument(
	ns string,
) Middleware {
	var (
		namespace    = strings.Replace(ns, "-", "_", -1)
		subsystem    = "handler"
		fieldKeys    = []string{"version", "route", "status"}
		requestCount = kitprometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "request_count",
			Help:      "Number of requests received",
		}, fieldKeys)
		requestLatency = metrics.NewTimeHistogram(
			time.Second,
			kitprometheus.NewHistogram(
				prometheus.HistogramOpts{
					Namespace: namespace,
					Subsystem: subsystem,
					Name:      "request_latency_seconds",
					Help:      "Total duration of requests in seconds",
				},
				fieldKeys,
			),
		)
		responseBytes = kitprometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "response_bytes",
			Help:      "Bytes returned as response bodies",
		}, fieldKeys)
	)

	return func(next Handler) Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			var (
				begin     = time.Now()
				resr      = &responseRecorder{ResponseWriter: w}
				routeName = routeFromContext(ctx)
				v         = versionFromContext(ctx)
			)

			next(ctx, resr, r)

			var (
				route = metrics.Field{
					Key:   "route",
					Value: routeName,
				}
				status = metrics.Field{
					Key:   "status",
					Value: strconv.Itoa(resr.StatusCode),
				}
				version = metrics.Field{
					Key:   "version",
					Value: v,
				}
			)

			requestCount.With(route).With(status).With(version).Add(1)
			requestLatency.With(route).With(status).With(version).Observe(time.Since(begin))
			responseBytes.With(route).With(status).With(version).Add(uint64(resr.ContentLength))
		}
	}
}

// Log logs information per single request-response.
func Log(logger log.Logger) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			var (
				begin   = time.Now()
				reqr    = newRequestRecorder(r)
				resr    = &responseRecorder{ResponseWriter: w}
				route   = routeFromContext(ctx)
				version = versionFromContext(ctx)
			)

			next(ctx, resr, r)

			resr.Headers = w.Header()

			logger.Log(
				"duration_ns", time.Since(begin),
				"query", r.URL.Query(),
				"request", reqr,
				"response", resr,
				"route", route,
				"version", version,
			)
		}
	}
}

// RateLimit enforces request limits per application.
func RateLimit(limits limiter.Limiter) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			var (
				app = appFromContext(ctx)
				l   = &limiter.Limitee{
					Hash:       app.AuthToken,
					Limit:      app.Limit(),
					WindowSize: time.Minute,
				}
			)

			quota, expires, err := limits.Request(l)
			if err != nil {
				respondError(w, 0, err)
				return
			}

			w.Header().Set("X-Ratelimit-Quota", strconv.FormatInt(app.Limit(), 10))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(quota, 10))
			w.Header().Set("X-Ratelimit-Reset", strconv.FormatInt(expires.Unix(), 10))

			if quota < 0 {
				respondError(w, 0, wrapError(ErrLimitExceeded, "request quota exceeded"))
				return
			}

			next(ctx, w, r)
		}
	}
}

// SecureHeaders adds a list of commonly recgonised best-pratice security
// headers.
// Source: https://www.owasp.org/index.php/List_of_useful_HTTP_headers
func SecureHeaders() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Strict-Transport-Security", "max-age=63072000")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")

			next(ctx, w, r)
		}
	}
}

// ValidateContent checks if content-length and content-type are set for
// requests with paylaod and adhere to our required limits and values.
func ValidateContent() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" && r.Method != "PUT" {
				next(ctx, w, r)
				return
			}

			if cl := r.Header.Get("Content-Length"); cl == "" {
				respondError(w, 5004, wrapError(ErrBadRequest, "Content-Length header missing"))
				return
			} else if l, err := strconv.ParseInt(cl, 10, 64); err != nil {
				respondError(w, 5003, wrapError(ErrBadRequest, "Content-Length header is invalid"))
				return
			} else if l != r.ContentLength {
				respondError(w, 5005, wrapError(ErrBadRequest, "Content-Length header size mismatch"))
				return
			} else if r.ContentLength > 32768 {
				respondError(w, 5011, wrapError(ErrBadRequest, "payload too big"))
				return
			}

			// Enforece content type checking for requests with payload.
			if r.ContentLength > 0 {
				if ct := r.Header.Get("Content-Type"); ct == "" {
					respondError(w, 5007, wrapError(ErrBadRequest, "Content-Type header missing"))
					return
				} else if ct != "application/json" && ct != "application/json; charset=UTF-8" {
					respondError(w, 5006, wrapError(ErrBadRequest, "Content-Type header missmatch"))
					return
				}
			}

			if r.Body == nil {
				respondError(w, 5012, wrapError(ErrBadRequest, "empty request body"))
				return
			}

			next(ctx, w, r)
		}
	}
}

type gzipResponseWriter struct {
	http.ResponseWriter
	io.Writer
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

type requestRecorder struct {
	Header           map[string][]string `json:"header"`
	Host             string              `json:"host"`
	Method           string              `json:"method"`
	Proto            string              `json:"proto"`
	RemoteAddr       string              `json:"remoteAddr"`
	RequestURI       string              `json:"requestURI"`
	TransferEncoding []string            `json:"transferEncoding"`
	URL              string              `json:"url"`
}

func newRequestRecorder(r *http.Request) *requestRecorder {
	return &requestRecorder{
		Header:           r.Header,
		Host:             r.Host,
		Method:           strings.ToLower(r.Method),
		Proto:            r.Proto,
		RemoteAddr:       r.RemoteAddr,
		RequestURI:       r.RequestURI,
		TransferEncoding: r.TransferEncoding,
		URL:              r.URL.String(),
	}
}

type responseRecorder struct {
	http.ResponseWriter `json:"-"`

	Headers       map[string][]string `json:"header"`
	ContentLength int                 `json:"contentLength"`
	StatusCode    int                 `json:"statusCode"`
}

func (rc *responseRecorder) Write(b []byte) (int, error) {
	n, err := rc.ResponseWriter.Write(b)

	rc.ContentLength += n

	return n, err
}

func (rc *responseRecorder) WriteHeader(code int) {
	rc.StatusCode = code
	rc.ResponseWriter.WriteHeader(code)
}
