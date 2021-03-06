// Package server provides handling for all the requests towards this module
package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tapglue/multiverse/errors"
	"github.com/tapglue/multiverse/limiter"
	"github.com/tapglue/multiverse/logger"
	"github.com/tapglue/multiverse/v04/context"
	"github.com/tapglue/multiverse/v04/core"
	v03_postgres_core "github.com/tapglue/multiverse/v04/core/postgres"
	v03_redis_core "github.com/tapglue/multiverse/v04/core/redis"
	"github.com/tapglue/multiverse/v04/errmsg"
	"github.com/tapglue/multiverse/v04/server/response"
	v03_postgres "github.com/tapglue/multiverse/v04/storage/postgres"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
)

type (
	errorResponse struct {
		Code             int    `json:"code"`
		Message          string `json:"message"`
		DocumentationURL string `json:"documentation_url,omitempty"`
	}

	RouteMetricHandler func(routeName, routeVersion string, route http.HandlerFunc) http.HandlerFunc
)

// APIVersion holds which API Version does this module holds
const APIVersion = "0.4"

var (
	postgresOrganization                  core.Organization
	postgresAccountUser                   core.Member
	postgresApplication, redisApplication core.Application
	postgresApplicationUser               core.ApplicationUser
	postgresConnection                    core.Connection
	postgresEvent                         core.Event

	appRateLimiter limiter.Limiter

	appRateLimitProduction int64 = 20000
	appRateLimitStaging    int64 = 100
	appRateLimitSeconds          = 60 * time.Second
)

func init() {
	if os.Getenv("CI") == "true" {
		appRateLimitProduction = 50
		appRateLimitStaging = 10
		appRateLimitSeconds = 1 * time.Second
	}

	if os.Getenv("NO_LIMITS") == "true" {
		log.Println("WARNING: LAUNCHING WITH NO APP LIMITS!!!")
		log.Println("WARNING: LAUNCHING WITH NO APP LIMITS!!!")
		log.Println("WARNING: LAUNCHING WITH NO APP LIMITS!!!")
		appRateLimitProduction = 5000000
		appRateLimitStaging = 1000000
	}
}

// ValidateGetCommon runs a series of predefined, common, tests for GET requests
func ValidateGetCommon(ctx *context.Context) (err []errors.Error) {
	if ctx.R.Header.Get("User-Agent") != "" {
		return
	}
	return []errors.Error{errmsg.ErrServerReqBadUserAgent.SetCurrentLocation()}
}

// ValidatePutCommon runs a series of predefinied, common, tests for PUT requests
func ValidatePutCommon(ctx *context.Context) (err []errors.Error) {
	if ctx.SkipSecurity {
		return
	}

	if ctx.R.Header.Get("User-Agent") == "" {
		err = append(err, errmsg.ErrServerReqBadUserAgent.SetCurrentLocation())
	}

	if ctx.R.Header.Get("Content-Length") == "" {
		err = append(err, errmsg.ErrServerReqContentLengthMissing.SetCurrentLocation())
	}

	if ctx.R.Header.Get("Content-Type") == "" {
		err = append(err, errmsg.ErrServerReqContentTypeMissing.SetCurrentLocation())
	}

	if ctx.R.Header.Get("Content-Type") != "application/json" &&
		ctx.R.Header.Get("Content-Type") != "application/json; charset=UTF-8" {
		err = append(err, errmsg.ErrServerReqContentTypeMismatch.SetCurrentLocation())
	}

	reqCL, er := strconv.ParseInt(ctx.R.Header.Get("Content-Length"), 10, 64)
	if er != nil {
		err = append(err, errmsg.ErrServerReqContentLengthInvalid.SetCurrentLocation())
	}

	if reqCL != ctx.R.ContentLength {
		err = append(err, errmsg.ErrServerReqContentLengthSizeMismatch.SetCurrentLocation())
	} else {
		// TODO better handling here for limits, maybe make them customizable
		if reqCL > 2048 {
			err = append(err, errmsg.ErrServerReqPayloadTooBig.SetCurrentLocation())
		}
	}

	if ctx.R.Body == nil {
		err = append(err, errmsg.ErrServerReqBodyEmpty.SetCurrentLocation())
	}
	return
}

// ValidateDeleteCommon runs a series of predefinied, common, tests for DELETE requests
func ValidateDeleteCommon(ctx *context.Context) (err []errors.Error) {
	if ctx.R.Header.Get("User-Agent") == "" {
		err = append(err, errmsg.ErrServerReqBadUserAgent.SetCurrentLocation())
	}

	return
}

// ValidatePostCommon runs a series of predefined, common, tests for the POST requests
func ValidatePostCommon(ctx *context.Context) (err []errors.Error) {
	if ctx.SkipSecurity {
		return
	}

	if ctx.R.Header.Get("User-Agent") == "" {
		err = append(err, errmsg.ErrServerReqBadUserAgent.SetCurrentLocation())
	}

	if ctx.R.Header.Get("Content-Length") == "" {
		err = append(err, errmsg.ErrServerReqContentLengthMissing.SetCurrentLocation())
	}

	if ctx.R.Header.Get("Content-Type") == "" {
		err = append(err, errmsg.ErrServerReqContentTypeMissing.SetCurrentLocation())
	}

	if ctx.R.Header.Get("Content-Type") != "application/json" &&
		ctx.R.Header.Get("Content-Type") != "application/json; charset=UTF-8" {
		err = append(err, errmsg.ErrServerReqContentTypeMismatch.SetCurrentLocation())
	}

	reqCL, er := strconv.ParseInt(ctx.R.Header.Get("Content-Length"), 10, 64)
	if er != nil {
		err = append(err, errmsg.ErrServerReqContentLengthInvalid.SetCurrentLocation())
	}

	if reqCL != ctx.R.ContentLength {
		err = append(err, errmsg.ErrServerReqContentLengthSizeMismatch.SetCurrentLocation())
	} else {
		// TODO better handling here for limits, maybe make them customizable
		if reqCL > 2048 {
			err = append(err, errmsg.ErrServerReqPayloadTooBig.SetCurrentLocation())
		}
	}

	if ctx.R.Body == nil {
		err = append(err, errmsg.ErrServerReqBodyEmpty.SetCurrentLocation())
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
	if ctx.Application.InProduction {
		appRateLimit = appRateLimitProduction
	}

	limitee := &limiter.Limitee{
		Hash:       ctx.Application.AuthToken,
		Limit:      appRateLimit,
		WindowSize: appRateLimitSeconds,
	}

	limit, refreshTime, err := appRateLimiter.Request(limitee)
	if err != nil {
		return []errors.Error{errmsg.ErrServerInternalError.UpdateInternalMessage(err.Error()).SetCurrentLocation()}
	}

	ctx.Bag["rateLimit.enabled"] = true
	ctx.Bag["rateLimit.limit"] = limit
	ctx.Bag["rateLimit.refreshTime"] = refreshTime

	if limit == 0 {
		return []errors.Error{errors.New(429, 0, "Too Many Requests", "over quota", false)}
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
			if err := route.Filters[idx](ctx); err != nil {
				response.ErrorHappened(ctx, err)
				return
			}
		}

		for idx := range route.Handlers {
			if err := route.Handlers[idx](ctx); err != nil {
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
func Setup(
	v03PostgresClient v03_postgres.Client,
	appCache *redis.Pool,
	revision, hostname string) {

	if appRateLimiter == nil {
		panic("You must first initialize the rate limiter")
	}

	redisApplication = v03_redis_core.NewApplication(appCache)

	postgresOrganization = v03_postgres_core.NewOrganization(v03PostgresClient)
	postgresAccountUser = v03_postgres_core.NewMember(v03PostgresClient)
	postgresApplication = v03_postgres_core.NewApplication(v03PostgresClient, redisApplication)
	postgresApplicationUser = v03_postgres_core.NewApplicationUser(v03PostgresClient)
	postgresConnection = v03_postgres_core.NewConnection(v03PostgresClient)
	postgresEvent = v03_postgres_core.NewEvent(v03PostgresClient)

	if revision == "" {
		panic("omfg missing revision")
	}

	response.Setup(revision, hostname)
	InitHandlers()

	Routes = SetupRoutes()
}
