package server_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tapglue/backend/config"
	"github.com/tapglue/backend/errors"
	ratelimiter_redis "github.com/tapglue/backend/limiter/redis"
	"github.com/tapglue/backend/logger"
	. "github.com/tapglue/backend/utils"
	"github.com/tapglue/backend/v03/core"
	v03_postgres_core "github.com/tapglue/backend/v03/core/postgres"
	"github.com/tapglue/backend/v03/entity"
	"github.com/tapglue/backend/v03/server"
	v03_kinesis "github.com/tapglue/backend/v03/storage/kinesis"
	v03_postgres "github.com/tapglue/backend/v03/storage/postgres"
	v03_redis "github.com/tapglue/backend/v03/storage/redis"

	redigo "github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	t.Parallel()
	TestingT(t)
}

type (
	ServerSuite          struct{}
	OrganizationSuite    struct{}
	MemberSuite          struct{}
	ApplicationSuite     struct{}
	ApplicationUserSuite struct{}
	ConnectionSuite      struct{}
	EventSuite           struct{}
)

var (
	_ = Suite(&ServerSuite{})
	_ = Suite(&OrganizationSuite{})
	_ = Suite(&MemberSuite{})
	_ = Suite(&ApplicationSuite{})
	_ = Suite(&ApplicationUserSuite{})
	_ = Suite(&ConnectionSuite{})
	_ = Suite(&EventSuite{})

	conf               *config.Config
	doLogTest          = flag.Bool("lt", false, "Set flag in order to get logs output from the tests")
	doCurlLogs         = flag.Bool("ct", false, "Set flag in order to get logs output from the tests as curl requests, sets -lt=true")
	doLogResponseTimes = flag.Bool("rt", false, "Set flag in order to get logs with response times only")
	doLogResponses     = flag.Bool("rl", false, "Set flag in order to get logs with response headers and bodies")
	quickBenchmark     = flag.Bool("qb", false, "Set flag in order to run only the benchmarks and skip all tests")
	mainLogChan        = make(chan *logger.LogMsg)
	errorLogChan       = make(chan *logger.LogMsg)

	coreAcc     core.Organization
	coreAccUser core.Member
	coreApp     core.Application
	coreAppUser core.ApplicationUser
	coreConn    core.Connection
	coreEvt     core.Event

	v03KinesisClient  v03_kinesis.Client
	v03PostgresClient v03_postgres.Client

	nilTime             *time.Time
	redigoRateLimitPool *redigo.Pool
)

func init() {
	flag.Parse()

	if *doCurlLogs {
		*doLogTest = true
	}

	runtime.GOMAXPROCS(runtime.NumCPU())
	conf = config.NewConf("")

	errors.Init(conf.Environment != "prod")

	if *doLogResponseTimes {
		go logger.TGLogResponseTimes(mainLogChan)
		go logger.TGLogResponseTimes(errorLogChan)
	} else if *doLogTest {
		if *doCurlLogs {
			go logger.TGCurlLog(mainLogChan)
		} else {
			go logger.TGLog(mainLogChan)
		}
		go logger.TGLog(errorLogChan)
	} else {
		go logger.TGSilentLog(mainLogChan)
		go logger.TGSilentLog(errorLogChan)
	}

	if conf.Environment == "prod" {
		v03KinesisClient = v03_kinesis.New(conf.Kinesis.AuthKey, conf.Kinesis.SecretKey, conf.Kinesis.Region, conf.Environment)
	} else {
		if conf.Kinesis.Endpoint != "" {
			v03KinesisClient = v03_kinesis.NewTest(conf.Kinesis.AuthKey, conf.Kinesis.SecretKey, conf.Kinesis.Region, conf.Kinesis.Endpoint, conf.Environment)
		} else {
			v03KinesisClient = v03_kinesis.New(conf.Kinesis.AuthKey, conf.Kinesis.SecretKey, conf.Kinesis.Region, conf.Environment)
		}
	}

	//v02KinesisClient.SetupStreams(v02_kinesis.Streams)
	switch conf.Environment {
	case "dev":
		v03KinesisClient.SetupStreams([]string{v03_kinesis.PackedStreamNameDev})
	case "test":
		v03KinesisClient.SetupStreams([]string{v03_kinesis.PackedStreamNameTest})
	case "prod":
		v03KinesisClient.SetupStreams([]string{v03_kinesis.PackedStreamNameProduction})
	}

	v03PostgresClient = v03_postgres.New(conf.Postgres)

	redigoRateLimitPool = v03_redis.NewRedigoPool(conf.Redis.Hosts[0], "")
	applicationRateLimiter := ratelimiter_redis.NewLimiter(redigoRateLimitPool, "ratelimiter.app.")

	coreAcc = v03_postgres_core.NewOrganization(v03PostgresClient)
	coreAccUser = v03_postgres_core.NewMember(v03PostgresClient)
	coreApp = v03_postgres_core.NewApplication(v03PostgresClient)
	coreAppUser = v03_postgres_core.NewApplicationUser(v03PostgresClient)
	coreConn = v03_postgres_core.NewConnection(v03PostgresClient)
	coreEvt = v03_postgres_core.NewEvent(v03PostgresClient)

	server.SetupRateLimit(applicationRateLimiter)
	server.Setup(v03KinesisClient, v03PostgresClient, "HEAD", "CI-Machine")

	testBootup(conf.Postgres)

	createdAt := struct {
		CreatedAt *time.Time
	}{}
	er := json.Unmarshal([]byte(`{"created_at": null}`), &createdAt)
	if er != nil {
		panic(er)
	}
	nilTime = createdAt.CreatedAt
}

func (s *ServerSuite) SetUpSuite(c *C) {
	if *quickBenchmark {
		c.Skip("Running in quick benchmark mode")
	}
}

func (s *OrganizationSuite) SetUpSuite(c *C) {
	if *quickBenchmark {
		c.Skip("Running in quick benchmark mode")
	}
}

func (s *MemberSuite) SetUpSuite(c *C) {
	if *quickBenchmark {
		c.Skip("Running in quick benchmark mode")
	}
}

func (s *ApplicationSuite) SetUpSuite(c *C) {
	if *quickBenchmark {
		c.Skip("Running in quick benchmark mode")
	}
}

func (s *ApplicationUserSuite) SetUpSuite(c *C) {
	if *quickBenchmark {
		c.Skip("Running in quick benchmark mode")
	}
}

func (s *ConnectionSuite) SetUpSuite(c *C) {
	if *quickBenchmark {
		c.Skip("Running in quick benchmark mode")
	}
}

func (s *EventSuite) SetUpSuite(c *C) {
	if *quickBenchmark {
		c.Skip("Running in quick benchmark mode")
	}
}

// createCommonRequestHeaders create a correct request header
func createCommonRequestHeaders(req *http.Request) {
	payload := PeakBody(req).Bytes()

	//req.Header.Add("x-tapglue-date", time.Now().Format(time.RFC3339))
	req.Header.Add("User-Agent", "Tapglue Test UA")

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Length", strconv.FormatInt(int64(len(payload)), 10))
}

// getComposedRoute takes a routeName and parameter and returns the route including the version
func getComposedRoute(routeName string, params ...interface{}) string {
	if routeName == "index" {
		return "/"
	} else if routeName == "humans" {
		return "/humans.txt"
	} else if routeName == "robots" {
		return "/robots.txt"
	}

	pattern := getRoute(routeName).TestPattern()

	if len(params) == 0 {
		return pattern
	}

	return fmt.Sprintf(pattern, params...)
}

// getQueryRoute does something
func getQueryRoute(routeName, query string, params ...interface{}) string {
	if routeName == "index" {
		return "/"
	} else if routeName == "humans" {
		return "/humans.txt"
	} else if routeName == "robots" {
		return "/robots.txt"
	}

	pattern := getRoute(routeName).TestPattern()
	pattern += "?" + query

	if len(params) == 0 {
		return pattern
	}

	return fmt.Sprintf(pattern, params...)
}

// getComposedRouteString takes the route and stringyfies all the params
func getComposedRouteString(routeName string, params ...interface{}) string {
	if routeName == "index" {
		return "/"
	} else if routeName == "humans" {
		return "/humans.txt"
	} else if routeName == "robots" {
		return "/robots.txt"
	}

	pattern := getRoute(routeName).TestPattern()

	if len(params) == 0 {
		return pattern
	}

	return fmt.Sprintf(pattern, params...)
}

// getComposedRouteString does something
func getQueryRouteString(routeName, query string, params ...interface{}) string {
	if routeName == "index" {
		return "/"
	} else if routeName == "humans" {
		return "/humans.txt"
	} else if routeName == "robots" {
		return "/robots.txt"
	}

	pattern := getRoute(routeName).TestPattern()

	if len(params) == 0 {
		return pattern
	}

	pattern = strings.Replace(pattern, "%d", "%s", -1)
	pattern = strings.Replace(pattern, "%.f", "%s", -1)
	pattern = strings.Replace(pattern, "%.7f", "%s", -1)
	pattern += "?" + query
	return fmt.Sprintf(pattern, params...)
}

// runRequest takes a route, path, payload and token, performs a request and return a response recorder
func runRequest(routeName, routePath, payload string, signFunc func(*http.Request)) (int, string, errors.Error) {
	var (
		requestRoute *server.Route
		routePattern string
	)

	if routeName == "index" {
		requestRoute = getRoute(routeName)
		routePattern = "/"
	} else if routeName == "humans" {
		requestRoute = getRoute(routeName)
		routePattern = "/humans.txt"
	} else if routeName == "robots" {
		requestRoute = getRoute(routeName)
		routePattern = "/robots.txt"
	} else {
		requestRoute = getRoute(routeName)
		routePattern = requestRoute.RoutePattern()
	}

	req, err := http.NewRequest(
		requestRoute.Method,
		routePath,
		strings.NewReader(payload),
	)
	if err != nil {
		panic(err)
	}

	createCommonRequestHeaders(req)

	signFunc(req)

	w := httptest.NewRecorder()
	m := mux.NewRouter()
	m.
		HandleFunc(routePattern, server.CustomHandler(requestRoute, mainLogChan, errorLogChan, "test", false, true)).
		Methods(requestRoute.Method)
	m.ServeHTTP(w, req)

	body := w.Body.String()

	if *doLogResponses {
		fmt.Printf("Got response: %#v with body %s\n", w, body)
	}

	return w.Code, body, nil
}

// runRequestWithHeaders is like runRequest but with Headerzz!!!
func runRequestWithHeaders(routeName, routePath, payload string, headerz, signFunc func(*http.Request)) (int, string, http.Header, errors.Error) {
	var (
		requestRoute *server.Route
		routePattern string
	)

	if routeName == "index" {
		requestRoute = getRoute(routeName)
		routePattern = "/"
	} else if routeName == "humans" {
		requestRoute = getRoute(routeName)
		routePattern = "/humans.txt"
	} else if routeName == "robots" {
		requestRoute = getRoute(routeName)
		routePattern = "/robots.txt"
	} else {
		requestRoute = getRoute(routeName)
		routePattern = requestRoute.RoutePattern()
	}

	req, err := http.NewRequest(
		requestRoute.Method,
		routePath,
		strings.NewReader(payload),
	)
	if err != nil {
		panic(err)
	}

	createCommonRequestHeaders(req)
	signFunc(req)
	headerz(req)

	w := httptest.NewRecorder()
	m := mux.NewRouter()
	m.
		HandleFunc(routePattern, server.CustomHandler(requestRoute, mainLogChan, errorLogChan, "test", false, true)).
		Methods(requestRoute.Method)
	m.ServeHTTP(w, req)

	body := w.Body.String()

	if *doLogResponses {
		fmt.Printf("Got response: %#v with body %s\n", w, body)
	}

	return w.Code, body, w.Header(), nil
}

// getMemberSessionToken retrieves the session token for a certain user
func getMemberSessionToken(user *entity.Member) string {
	sessionToken, err := coreAccUser.CreateSession(user)
	if err != nil {
		panic(err)
	}

	return sessionToken
}

func nilSigner(*http.Request) {

}

func signOrganizationRequest(organization *entity.Organization, member *entity.Member, goodOrganizationToken, goodMemberToken bool) func(*http.Request) {
	return func(r *http.Request) {
		user := ""
		pass := ""

		if goodOrganizationToken && organization != nil {
			user = organization.AuthToken
		}
		if goodOrganizationToken && organization == nil {
			user = ""
		}
		if !goodOrganizationToken && organization != nil {
			user = organization.AuthToken + "a"
		}
		if !goodOrganizationToken && organization == nil {
			user = "a"
		}

		if goodMemberToken && member != nil {
			pass = member.SessionToken
		}
		if goodMemberToken && member == nil {
			pass = ""
		}
		if !goodMemberToken && member != nil {
			pass = member.SessionToken + "a"
		}
		if !goodMemberToken && member == nil {
			pass = "a"
		}

		if user == "" && pass == "" {
			return
		}

		encodedAuth := Base64Encode(user + ":" + pass)

		r.Header.Add("Authorization", "Basic "+encodedAuth)
	}
}

func signApplicationRequest(application *entity.Application, applicationUser *entity.ApplicationUser, goodApplicationToken, goodApplicationUserToken bool) func(*http.Request) {
	return func(r *http.Request) {
		user := ""
		pass := ""

		if goodApplicationToken && application != nil {
			user = application.AuthToken
		}
		if goodApplicationToken && application == nil {
			user = ""
		}
		if !goodApplicationToken && application != nil {
			user = application.AuthToken + "a"
		}
		if !goodApplicationToken && application == nil {
			user = "a"
		}

		if goodApplicationUserToken && applicationUser != nil {
			pass = applicationUser.SessionToken
		}
		if goodApplicationUserToken && applicationUser == nil {
			pass = ""
		}
		if !goodApplicationUserToken && applicationUser != nil {
			pass = applicationUser.SessionToken + "a"
		}
		if !goodApplicationUserToken && applicationUser == nil {
			pass = "a"
		}

		encodedAuth := Base64Encode(user + ":" + pass)

		r.Header.Add("Authorization", "Basic "+encodedAuth)
	}
}

func getRoute(routeName string) *server.Route {
	routes := server.SetupRoutes()
	for idx := range routes {
		if routes[idx].Name == routeName {
			return routes[idx]
		}
	}

	panic(fmt.Sprintf("route %q not found", routeName))
}
