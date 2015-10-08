// Package server provides handling for all the requests towards this module
package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tapglue/multiverse/context"
	"github.com/tapglue/multiverse/errors"
	"github.com/tapglue/multiverse/limiter"
	"github.com/tapglue/multiverse/logger"
	"github.com/tapglue/multiverse/v02/core"
	v02_kinesis_core "github.com/tapglue/multiverse/v02/core/kinesis"
	v02_postgres_core "github.com/tapglue/multiverse/v02/core/postgres"
	"github.com/tapglue/multiverse/v02/entity"
	"github.com/tapglue/multiverse/v02/errmsg"
	"github.com/tapglue/multiverse/v02/server/response"
	v02_kinesis "github.com/tapglue/multiverse/v02/storage/kinesis"
	v02_postgres "github.com/tapglue/multiverse/v02/storage/postgres"

	"github.com/gorilla/mux"
)

type (
	errorResponse struct {
		Code             int    `json:"code"`
		Message          string `json:"message"`
		DocumentationURL string `json:"documentation_url,omitempty"`
	}
)

// APIVersion holds which API Version does this module holds
const APIVersion = "0.2"

var (
	postgresAccount, kinesisAccount                 core.Account
	postgresAccountUser, kinesisAccountUser         core.AccountUser
	postgresApplication, kinesisApplication         core.Application
	postgresApplicationUser, kinesisApplicationUser core.ApplicationUser
	postgresConnection, kinesisConnection           core.Connection
	postgresEvent, kinesisEvent                     core.Event

	appRateLimiter limiter.Limiter

	appRateLimitProduction int64 = 10000
	appRateLimitStaging    int64 = 100
	appRateLimitSeconds          = 60 * time.Second
)

func init() {
	if os.Getenv("CI") == "true" {
		appRateLimitProduction = 50
		appRateLimitStaging = 10
	}
}

// ValidateGetCommon runs a series of predefined, common, tests for GET requests
func ValidateGetCommon(ctx *context.Context) (err []errors.Error) {
	if ctx.R.Header.Get("User-Agent") != "" {
		return
	}
	return []errors.Error{errmsg.ErrServerReqBadUserAgent}
}

// ValidatePutCommon runs a series of predefinied, common, tests for PUT requests
func ValidatePutCommon(ctx *context.Context) (err []errors.Error) {
	if ctx.SkipSecurity {
		return
	}

	if ctx.R.Header.Get("User-Agent") == "" {
		err = append(err, errmsg.ErrServerReqBadUserAgent)
	}

	if ctx.R.Header.Get("Content-Length") == "" {
		err = append(err, errmsg.ErrServerReqContentLengthMissing)
	}

	if ctx.R.Header.Get("Content-Type") == "" {
		err = append(err, errmsg.ErrServerReqContentTypeMissing)
	}

	if ctx.R.Header.Get("Content-Type") != "application/json" &&
		ctx.R.Header.Get("Content-Type") != "application/json; charset=UTF-8" {
		err = append(err, errmsg.ErrServerReqContentTypeMismatch)
	}

	reqCL, er := strconv.ParseInt(ctx.R.Header.Get("Content-Length"), 10, 64)
	if er != nil {
		err = append(err, errmsg.ErrServerReqContentLengthInvalid)
	}

	if reqCL != ctx.R.ContentLength {
		err = append(err, errmsg.ErrServerReqContentLengthSizeMismatch)
	} else {
		// TODO better handling here for limits, maybe make them customizable
		if reqCL > 2048 {
			err = append(err, errmsg.ErrServerReqPayloadTooBig)
		}
	}

	if ctx.R.Body == nil {
		err = append(err, errmsg.ErrServerReqBodyEmpty)
	}
	return
}

// ValidateDeleteCommon runs a series of predefinied, common, tests for DELETE requests
func ValidateDeleteCommon(ctx *context.Context) (err []errors.Error) {
	if ctx.R.Header.Get("User-Agent") == "" {
		err = append(err, errmsg.ErrServerReqBadUserAgent)
	}

	return
}

// ValidatePostCommon runs a series of predefined, common, tests for the POST requests
func ValidatePostCommon(ctx *context.Context) (err []errors.Error) {
	if ctx.SkipSecurity {
		return
	}

	if ctx.R.Header.Get("User-Agent") == "" {
		err = append(err, errmsg.ErrServerReqBadUserAgent)
	}

	if ctx.R.Header.Get("Content-Length") == "" {
		err = append(err, errmsg.ErrServerReqContentLengthMissing)
	}

	if ctx.R.Header.Get("Content-Type") == "" {
		err = append(err, errmsg.ErrServerReqContentTypeMissing)
	}

	if ctx.R.Header.Get("Content-Type") != "application/json" &&
		ctx.R.Header.Get("Content-Type") != "application/json; charset=UTF-8" {
		err = append(err, errmsg.ErrServerReqContentTypeMismatch)
	}

	reqCL, er := strconv.ParseInt(ctx.R.Header.Get("Content-Length"), 10, 64)
	if er != nil {
		err = append(err, errmsg.ErrServerReqContentLengthInvalid)
	}

	if reqCL != ctx.R.ContentLength {
		err = append(err, errmsg.ErrServerReqContentLengthSizeMismatch)
	} else {
		// TODO better handling here for limits, maybe make them customizable
		if reqCL > 2048 {
			err = append(err, errmsg.ErrServerReqPayloadTooBig)
		}
	}

	if ctx.R.Body == nil {
		err = append(err, errmsg.ErrServerReqBodyEmpty)
	}
	return
}

// GetRoute takes a route name and returns the route including the version
func GetRoute(routeName string) *Route {
	for idx := range Routes {
		if Routes[idx].Name == routeName {
			return Routes[idx]
		}
	}

	panic(fmt.Sprintf("route %q not found", routeName))
}

// RateLimitApplication takes care of app the rate limits for the application
func RateLimitApplication(ctx *context.Context) []errors.Error {
	if ctx.SkipSecurity {
		return nil
	}

	appRateLimit := appRateLimitStaging
	if ctx.Bag["application"].(*entity.Application).InProduction {
		appRateLimit = appRateLimitProduction
	}

	limitee := &limiter.Limitee{
		Hash:       ctx.Bag["application"].(*entity.Application).AuthToken,
		Limit:      appRateLimit,
		WindowSize: appRateLimitSeconds,
	}

	limit, refreshTime, err := appRateLimiter.Request(limitee)
	if err != nil {
		return []errors.Error{errmsg.ErrServerInternalError.UpdateInternalMessage(err.Error())}
	}

	ctx.Bag["rateLimit.enabled"] = true
	ctx.Bag["rateLimit.limit"] = limit
	ctx.Bag["rateLimit.refreshTime"] = refreshTime

	if limit == 0 {
		return []errors.Error{errmsg.ErrTooManyRequests}
	}

	return nil
}

// CustomHandler generates the handler for a certain route
func CustomHandler(route *Route, mainLogChan, errorLogChan chan *logger.LogMsg, env string, skipSecurity, debug bool) http.HandlerFunc {
	extraHandlers := []RouteFunc{}
	switch route.Method {
	case "DELETE":
		{
			extraHandlers = append(extraHandlers, ValidateDeleteCommon)
		}
	case "GET":
		{
			extraHandlers = append(extraHandlers, ValidateGetCommon)
		}
	case "PUT":
		{
			extraHandlers = append(extraHandlers, ValidatePutCommon)
		}
	case "POST":
		{
			extraHandlers = append(extraHandlers, ValidatePostCommon)
		}
	}
	route.Handlers = append(extraHandlers, route.Handlers...)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, err := NewContext(w, r, mux.Vars(r), mainLogChan, errorLogChan, route, env, debug)
		if err != nil {
			response.ErrorHappened(ctx, err)
			return
		}
		ctx.SkipSecurity = skipSecurity

		for idx := range route.Filters {
			if err = route.Filters[idx](ctx); err != nil {
				response.ErrorHappened(ctx, err)
				return
			}
		}

		for idx := range route.Handlers {
			if err = route.Handlers[idx](ctx); err != nil {
				response.ErrorHappened(ctx, err)
				return
			}
		}

		ua := strings.ToLower(ctx.R.Header.Get("User-Agent"))
		switch true {
		case strings.HasPrefix(ua, "elb"):
			fallthrough
		case strings.HasPrefix(ua, "updown"):
			fallthrough
		case strings.HasPrefix(ua, "pingdom"):
			return
		}

		go ctx.LogRequest(ctx.StatusCode, -1)
	}
}

// CustomOptionsHandler handles all the OPTIONS requests for us
func CustomOptionsHandler(route *Route, mainLogChan, errorLogChan chan *logger.LogMsg, env string, skipSecurity, debug bool) http.HandlerFunc {
	// Override the route method to what we need
	route.Method = "OPTIONS"
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, err := NewContext(w, r, mux.Vars(r), mainLogChan, errorLogChan, route, env, debug)
		ctx.SkipSecurity = skipSecurity
		if err != nil {
			go response.ErrorHappened(ctx, err)
			return
		}
		ctx.StatusCode = 200
		if err = response.CORSHandler(ctx); err != nil {
			go response.ErrorHappened(ctx, err)
			return
		}

		ua := strings.ToLower(ctx.R.Header.Get("User-Agent"))
		switch true {
		case strings.HasPrefix(ua, "elb"):
			fallthrough
		case strings.HasPrefix(ua, "updown"):
			fallthrough
		case strings.HasPrefix(ua, "pingdom"):
			return
		}
		go ctx.LogRequest(ctx.StatusCode, -1)
	}
}

// SetupRateLimit initializes the rate limiters
func SetupRateLimit(applicationRateLimiter limiter.Limiter) {
	appRateLimiter = applicationRateLimiter
}

// Setup initializes the route handlers
// Must be called after initializing the cores
func Setup(v02KinesisClient v02_kinesis.Client, v02PostgresClient v02_postgres.Client, revision, hostname string) {
	if appRateLimiter == nil {
		panic("You must first initialize the rate limiter")
	}

	kinesisAccount = v02_kinesis_core.NewAccount(v02KinesisClient)
	kinesisAccountUser = v02_kinesis_core.NewAccountUser(v02KinesisClient)
	kinesisApplication = v02_kinesis_core.NewApplication(v02KinesisClient)
	kinesisApplicationUser = v02_kinesis_core.NewApplicationUser(v02KinesisClient)
	kinesisConnection = v02_kinesis_core.NewConnection(v02KinesisClient)
	kinesisEvent = v02_kinesis_core.NewEvent(v02KinesisClient)

	postgresAccount = v02_postgres_core.NewAccount(v02PostgresClient)
	postgresAccountUser = v02_postgres_core.NewAccountUser(v02PostgresClient)
	postgresApplication = v02_postgres_core.NewApplication(v02PostgresClient)
	postgresApplicationUser = v02_postgres_core.NewApplicationUser(v02PostgresClient)
	postgresConnection = v02_postgres_core.NewConnection(v02PostgresClient)
	postgresEvent = v02_postgres_core.NewEvent(v02PostgresClient)

	if revision == "" {
		panic("omfg missing revision")
	}

	response.Setup(revision, hostname)
	InitHandlers()

	Routes = SetupRoutes()
}
