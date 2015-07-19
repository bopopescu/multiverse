package server_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/tapglue/backend/v02/entity"
	"github.com/tapglue/backend/v02/server"

	"github.com/gorilla/mux"
	. "gopkg.in/check.v1"
)

// Test createEvent request with a wrong key
func (s *EventSuite) TestCreateEvent_WrongKey(c *C) {
	accounts := CorrectDeploy(1, 0, 1, 1, 0, true, true)
	application := accounts[0].Applications[0]
	user := application.Users[0]

	payload := "{typeea:''}"

	routeName := "createCurrentUserEvent"
	route := getComposedRoute(routeName)
	code, body, err := runRequest(routeName, route, payload, signApplicationRequest(application, user, true, true))
	c.Assert(err, IsNil)

	c.Assert(code, Equals, http.StatusBadRequest)
	c.Assert(body, Not(Equals), "")
}

// Test createEvent request with an wrong name
func (s *EventSuite) TestCreateEvent_WrongValue(c *C) {
	accounts := CorrectDeploy(1, 0, 1, 1, 0, true, true)
	application := accounts[0].Applications[0]
	user := application.Users[0]

	payload := `{"type":"","language":""}`

	routeName := "createCurrentUserEvent"
	route := getComposedRoute(routeName)
	code, body, err := runRequest(routeName, route, payload, signApplicationRequest(application, user, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusBadRequest)
	c.Assert(body, Not(Equals), "")
}

// Test a correct createEvent request
func (s *EventSuite) TestCreateEvent_OK(c *C) {
	accounts := CorrectDeploy(1, 0, 1, 1, 0, true, true)
	application := accounts[0].Applications[0]
	user := application.Users[0]
	event := CorrectEvent(application.ID)

	payload := fmt.Sprintf(
		`{"type":%q, "language":%q, "visibility": %d}`,
		event.Type,
		event.Language,
		entity.EventPublic,
	)

	routeName := "createCurrentUserEvent"
	route := getComposedRoute(routeName)
	code, body, err := runRequest(routeName, route, payload, signApplicationRequest(application, user, true, true))
	c.Assert(err, IsNil)

	c.Assert(code, Equals, http.StatusCreated)

	c.Assert(body, Not(Equals), "")

	receivedEvent := &entity.Event{}
	er := json.Unmarshal([]byte(body), receivedEvent)
	c.Assert(er, IsNil)
	c.Assert(receivedEvent.ID, Not(Equals), "")
	c.Assert(receivedEvent.UserID, Equals, user.ID)
	c.Assert(receivedEvent.Enabled, Equals, true)
	c.Assert(receivedEvent.Type, Equals, event.Type)
	c.Assert(receivedEvent.Language, Equals, event.Language)
	c.Assert(int(receivedEvent.Visibility), Equals, entity.EventPublic)

	payload = fmt.Sprintf(
		`{"type":%q, "language":%q}`,
		event.Type,
		event.Language,
	)

	code, body, err = runRequest(routeName, route, payload, signApplicationRequest(application, user, true, true))
	c.Assert(err, IsNil)

	c.Assert(code, Equals, http.StatusCreated)

	c.Assert(body, Not(Equals), "")

	receivedEvent = &entity.Event{}
	er = json.Unmarshal([]byte(body), receivedEvent)
	c.Assert(er, IsNil)
	c.Assert(receivedEvent.UserID, Equals, user.ID)
	c.Assert(receivedEvent.Enabled, Equals, true)
	c.Assert(receivedEvent.Type, Equals, event.Type)
	c.Assert(receivedEvent.Language, Equals, event.Language)
	c.Assert(int(receivedEvent.Visibility), Equals, entity.EventPublic)

	payload = fmt.Sprintf(
		`{"type":%q, "language":%q, "visibility": %d}`,
		event.Type,
		event.Language,
		entity.EventGlobal,
	)

	code, body, err = runRequest(routeName, route, payload, signApplicationRequest(application, user, true, true))
	c.Assert(err, IsNil)

	c.Assert(code, Equals, http.StatusCreated)

	c.Assert(body, Not(Equals), "")

	receivedEvent = &entity.Event{}
	er = json.Unmarshal([]byte(body), receivedEvent)
	c.Assert(er, IsNil)
	c.Assert(receivedEvent.UserID, Equals, user.ID)
	c.Assert(receivedEvent.Enabled, Equals, true)
	c.Assert(receivedEvent.Type, Equals, event.Type)
	c.Assert(receivedEvent.Language, Equals, event.Language)
	c.Assert(int(receivedEvent.Visibility), Equals, entity.EventGlobal)
}

// Test a correct updateEvent request
func (s *EventSuite) TestUpdateEvent_OK(c *C) {
	accounts := CorrectDeploy(1, 0, 1, 1, 1, true, true)
	application := accounts[0].Applications[0]
	user := application.Users[0]
	event := user.Events[0]

	payload := fmt.Sprintf(
		`{"type":"%s", "language":"%s"}`,
		event.Type,
		event.Language+"aaaa",
	)

	routeName := "updateCurrentUserEvent"
	route := getComposedRoute(routeName, event.ID)
	code, body, err := runRequest(routeName, route, payload, signApplicationRequest(application, user, true, true))
	c.Assert(err, IsNil)

	c.Assert(code, Equals, http.StatusCreated)

	c.Assert(body, Not(Equals), "")

	receivedEvent := &entity.Event{}
	er := json.Unmarshal([]byte(body), receivedEvent)
	c.Assert(er, IsNil)
	c.Assert(receivedEvent.UserID, Equals, user.ID)
	c.Assert(receivedEvent.Type, Equals, event.Type)
	c.Assert(receivedEvent.Language, Equals, event.Language+"aaaa")
}

// Test updateEvent request with a wrong id
func (s *EventSuite) TestUpdateEvent_WrongID(c *C) {
	accounts := CorrectDeploy(1, 0, 1, 1, 1, true, true)
	application := accounts[0].Applications[0]
	user := application.Users[0]
	event := user.Events[0]

	payload := fmt.Sprintf(
		`{"type":%q, "language":%q, "enabled":false}`,
		event.Type,
		event.Language,
	)

	routeName := "updateCurrentUserEvent"
	route := getComposedRoute(routeName, event.ID+1)
	code, _, err := runRequest(routeName, route, payload, signApplicationRequest(application, user, true, true))
	c.Assert(err, IsNil)

	c.Assert(code, Equals, http.StatusNotFound)
}

func (s *EventSuite) TestUpdateEventMalformedIDFails(c *C) {
	c.Skip("we can't have malformed ids for now in the tests")
	accounts := CorrectDeploy(1, 0, 1, 1, 1, true, true)
	user := accounts[0].Applications[0].Users[0]

	payload := fmt.Sprintf(
		`{"type":%q, "language":%q, "enabled":false}`,
		user.Events[0].Type,
		user.Events[0].Language,
	)

	routeName := "updateCurrentUserEvent"
	route := getComposedRouteString(routeName, 9087654321123456789)
	code, body, err := runRequest(routeName, route, payload, signApplicationRequest(accounts[0].Applications[0], user, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusBadRequest)
	c.Assert(body, Equals, `{"errors":[{"code":3002,"message":"event id is not valid"}]}`+"\n")
}

func (s *EventSuite) TestUpdateEventMalformedPayloadFails(c *C) {
	accounts := CorrectDeploy(1, 0, 1, 1, 1, true, true)
	user := accounts[0].Applications[0].Users[0]

	payload := fmt.Sprintf(
		`{"type":%q, "language":%q, "enabled":false`,
		user.Events[0].Type,
		user.Events[0].Language,
	)

	routeName := "updateCurrentUserEvent"
	route := getComposedRoute(routeName, user.Events[0].ID)
	code, body, err := runRequest(routeName, route, payload, signApplicationRequest(accounts[0].Applications[0], user, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusBadRequest)
	c.Assert(body, Equals, `{"errors":[{"code":5001,"message":"unexpected end of JSON input"}]}`+"\n")
}

// Test updateEvent request with a wrong value
func (s *EventSuite) TestUpdateEvent_WrongValue(c *C) {
	accounts := CorrectDeploy(1, 0, 1, 1, 1, true, true)
	application := accounts[0].Applications[0]
	user := application.Users[0]
	event := user.Events[0]

	payload := fmt.Sprintf(
		`{"type":"", "language":"%s", "enabled":false}`,
		event.Language,
	)

	routeName := "updateCurrentUserEvent"
	route := getComposedRoute(routeName, event.ID)
	code, body, err := runRequest(routeName, route, payload, signApplicationRequest(application, user, true, true))
	c.Assert(err, IsNil)

	c.Assert(code, Equals, http.StatusBadRequest)
	c.Assert(body, Not(Equals), "")
}

// Test a correct deleteEvent request
func (s *EventSuite) TestDeleteEvent_OK(c *C) {
	accounts := CorrectDeploy(1, 0, 1, 1, 1, true, true)
	application := accounts[0].Applications[0]
	user := application.Users[0]
	event := user.Events[0]

	routeName := "deleteCurrentUserEvent"
	route := getComposedRoute(routeName, event.ID)
	code, _, err := runRequest(routeName, route, "", signApplicationRequest(application, user, true, true))
	c.Assert(err, IsNil)

	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusNoContent)
}

// Test deleteEvent request with a wrong id
func (s *EventSuite) TestDeleteEvent_WrongID(c *C) {
	accounts := CorrectDeploy(1, 0, 1, 1, 1, true, true)
	application := accounts[0].Applications[0]
	user := application.Users[0]
	event := user.Events[0]

	routeName := "deleteCurrentUserEvent"
	route := getComposedRoute(routeName, event.ID+1)
	code, _, err := runRequest(routeName, route, "", signApplicationRequest(application, user, true, true))
	c.Assert(err, IsNil)

	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusNotFound)
}

func (s *EventSuite) TestDeleteEventMalformedIDFails(c *C) {
	c.Skip("we can't have malformed ids for now in the tests")
	accounts := CorrectDeploy(1, 0, 1, 1, 1, true, true)
	user := accounts[0].Applications[0].Users[0]

	routeName := "deleteCurrentUserEvent"
	route := getComposedRouteString(routeName, "90876543211234567890")
	code, _, err := runRequest(routeName, route, "", signApplicationRequest(accounts[0].Applications[0], user, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusBadRequest)
}

// Test a correct getEvent request
func (s *EventSuite) TestGetEvent_OK(c *C) {
	accounts := CorrectDeploy(1, 0, 1, 1, 1, false, true)
	application := accounts[0].Applications[0]
	user := application.Users[0]
	event := user.Events[rand.Intn(1)]

	routeName := "getEvent"
	route := getComposedRoute(routeName, user.ID, event.ID)
	code, body, err := runRequest(routeName, route, "", signApplicationRequest(application, user, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusOK)

	c.Assert(body, Not(Equals), "")

	receivedEvent := &entity.Event{}
	er := json.Unmarshal([]byte(body), receivedEvent)
	c.Assert(er, IsNil)
	c.Assert(receivedEvent.ID, Equals, event.ID)
	c.Assert(receivedEvent.UserID, Equals, user.ID)
	c.Assert(receivedEvent.Enabled, Equals, true)
}

// Test a correct getEventList request
func (s *EventSuite) TestGetEventList_OK(c *C) {
	accounts := CorrectDeploy(1, 0, 1, 2, 5, false, true)
	application := accounts[0].Applications[0]
	user := application.Users[0]

	routeName := "getEventList"
	route := getComposedRoute(routeName, user.ID)
	code, body, err := runRequest(routeName, route, "", signApplicationRequest(application, user, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusOK)
	c.Assert(body, Not(Equals), "")

	response := struct {
		Events      []*entity.Event `json:"events"`
		EventsCount int             `json:"events_count"`
	}{}
	er := json.Unmarshal([]byte(body), &response)
	c.Assert(er, IsNil)
	c.Assert(response.EventsCount, Equals, len(user.Events))
	for idx := range response.Events {
		c.Logf("pass %d", idx)
		response.Events[idx].UpdatedAt = user.Events[4-idx].UpdatedAt
		c.Assert(response.Events[idx], DeepEquals, user.Events[4-idx])
	}
}

// Test getEvent request with a wrong id
func (s *EventSuite) TestGetEventWrongIDFails(c *C) {
	accounts := CorrectDeploy(1, 0, 1, 1, 1, true, true)
	application := accounts[0].Applications[0]
	user := application.Users[0]
	event := user.Events[0]

	routeName := "getEvent"
	route := getComposedRoute(routeName, user.ID, event.ID+1)
	code, _, err := runRequest(routeName, route, "", signApplicationRequest(application, user, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusNotFound)
}

func (s *EventSuite) TestGetEventMalformedIDFails(c *C) {
	c.Skip("we can't have malformed ids for now in the tests")
	accounts := CorrectDeploy(1, 0, 1, 1, 1, true, true)
	user := accounts[0].Applications[0].Users[0]

	routeName := "getEvent"
	route := getComposedRouteString(routeName, user.ID, "90876543211234567890")
	code, _, err := runRequest(routeName, route, "", signApplicationRequest(accounts[0].Applications[0], user, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusBadRequest)
}

func (s *EventSuite) TestGeoLocationSearch(c *C) {
	accounts := CorrectDeploy(1, 0, 1, 1, 7, false, true)
	application := accounts[0].Applications[0]
	user := application.Users[0]

	routeName := "searchEvents"
	scenarios := []struct {
		URL           string
		StatusCode    int
		ExpectedOrder []string
	}{
		{
			URL:           getQueryRoute(routeName, "lat=%.7f&lon=%.7f&rad=%.7f", user.Events[0].Latitude, user.Events[0].Longitude, 25000.0),
			StatusCode:    http.StatusOK,
			ExpectedOrder: []string{"onur", "mercedes", "cinestar", "ziko", "palace", "gas", "dlsniper"},
		},
		{
			URL:           getQueryRoute(routeName, "lat=%.7f&lon=%.7f&rad=%.7f", user.Events[0].Latitude, user.Events[0].Longitude, 1250.0),
			StatusCode:    http.StatusOK,
			ExpectedOrder: []string{"ziko", "palace", "gas", "dlsniper"},
		},
		{
			URL:           getQueryRoute(routeName, "lat=%.7f&lon=%.7f&nearest=%.0f", user.Events[0].Latitude, user.Events[0].Longitude, 3.0),
			StatusCode:    http.StatusOK,
			ExpectedOrder: []string{"dlsniper", "gas", "ziko"},
		},
	}

	for idx, scenario := range scenarios {
		code, body, err := runRequest(routeName, scenario.URL, "", signApplicationRequest(application, user, true, true))
		c.Assert(err, IsNil)
		c.Assert(code, Equals, http.StatusOK)
		c.Assert(body, Not(Equals), "")

		response := struct {
			Events      []*entity.Event `json:"events"`
			EventsCount int             `json:"events_count"`
		}{}
		er := json.Unmarshal([]byte(body), &response)
		c.Assert(er, IsNil)

		c.Assert(response.EventsCount, Equals, len(scenario.ExpectedOrder))
		for id := range response.Events {
			c.Logf("#%d pass: %d", idx, id)
			c.Assert(response.Events[id].Location, Equals, scenario.ExpectedOrder[id])
		}
	}
}

func (s *EventSuite) TestGeoLocationInvalidSearchDataFails(c *C) {
	accounts := CorrectDeploy(1, 0, 1, 1, 6, false, true)
	application := accounts[0].Applications[0]
	user := application.Users[0]

	routeName := "searchEvents"

	scenarios := []struct {
		Latitude     string
		Longitude    string
		Radius       string
		StatusCode   int
		ResponseBody string
	}{
		// 0
		{
			Latitude:     fmt.Sprintf("%.7f", user.Events[0].Latitude),
			Longitude:    fmt.Sprintf("%.7f", user.Events[0].Longitude),
			Radius:       "-25000.0",
			StatusCode:   http.StatusBadRequest,
			ResponseBody: "{\"errors\":[{\"code\":3001,\"message\":\"Location radius can't be smaller than 2 meters\"}]}\n",
		},
		// 1
		{
			Latitude:     "0.0.0",
			Longitude:    fmt.Sprintf("%.7f", user.Events[0].Longitude),
			Radius:       "25000",
			StatusCode:   http.StatusBadRequest,
			ResponseBody: "{\"errors\":[{\"code\":5010,\"message\":\"strconv.ParseFloat: parsing \\\"0.0.0\\\": invalid syntax\"}]}\n",
		},
		// 2
		{
			Latitude:     fmt.Sprintf("%.7f", user.Events[0].Latitude),
			Longitude:    "0.0.0",
			Radius:       "25000",
			StatusCode:   http.StatusBadRequest,
			ResponseBody: "{\"errors\":[{\"code\":5010,\"message\":\"strconv.ParseFloat: parsing \\\"0.0.0\\\": invalid syntax\"}]}\n",
		},
		// 3
		{
			Latitude:     fmt.Sprintf("%.7f", user.Events[0].Latitude),
			Longitude:    fmt.Sprintf("%.7f", user.Events[0].Longitude),
			Radius:       "0.0.0",
			StatusCode:   http.StatusBadRequest,
			ResponseBody: "{\"errors\":[{\"code\":5010,\"message\":\"strconv.ParseFloat: parsing \\\"0.0.0\\\": invalid syntax\"}]}\n",
		},
	}

	for idx := range scenarios {
		route := getQueryRouteString(routeName, "lat=%s&lon=%s&rad=%s", scenarios[idx].Latitude, scenarios[idx].Longitude, scenarios[idx].Radius)
		code, body, err := runRequest(routeName, route, "", signApplicationRequest(application, user, true, true))
		c.Logf("pass: %d", idx)
		c.Assert(err, IsNil)
		c.Assert(code, Equals, scenarios[idx].StatusCode)
		c.Assert(body, Equals, scenarios[idx].ResponseBody)
	}
}

func (s *EventSuite) TestGetLocation(c *C) {
	accounts := CorrectDeploy(1, 0, 1, 2, 7, true, true)
	application := accounts[0].Applications[0]
	user1 := application.Users[0]

	routeName := "searchEvents"
	route := getQueryRoute(routeName, "location=%s", user1.Events[0].Location)
	code, body, err := runRequest(routeName, route, "", signApplicationRequest(application, user1, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusOK)
	c.Assert(body, Not(Equals), "")

	response := struct {
		Events      []*entity.Event `json:"events"`
		EventsCount int             `json:"events_count"`
	}{}
	er := json.Unmarshal([]byte(body), &response)
	c.Assert(er, IsNil)
	c.Assert(response.EventsCount, Equals, 1)
	c.Assert(response.Events[0], DeepEquals, user1.Events[0])
}

func (s *EventSuite) TestGetObjectEvents(c *C) {
	c.Skip("routes removed for now")
	accounts := CorrectDeploy(1, 0, 1, 2, 7, true, true)
	application := accounts[0].Applications[0]
	user1 := application.Users[0]
	user2 := application.Users[1]

	routeName := "getObjectEventList"
	route := getComposedRoute(routeName, application.AccountID, application.ID, user1.Events[0].Object.ID)
	code, body, err := runRequest(routeName, route, "", signApplicationRequest(application, user1, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusOK)
	c.Assert(body, Not(Equals), "")

	events := []*entity.Event{}
	er := json.Unmarshal([]byte(body), &events)
	c.Assert(er, IsNil)
	c.Assert(len(events), Equals, 2)
	c.Assert(events[0], DeepEquals, user2.Events[0])
	c.Assert(events[1], DeepEquals, user1.Events[0])
}

func (s *EventSuite) TestGetFeed(c *C) {
	accounts := CorrectDeploy(1, 0, 1, 10, 10, true, true)
	application := accounts[0].Applications[0]
	userFrom := application.Users[0]

	// Check activity feed events
	routeName := "getFeed"
	route := getComposedRoute(routeName)
	code, body, err := runRequest(routeName, route, "", signApplicationRequest(application, userFrom, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusOK)
	c.Assert(body, Not(Equals), "")

	response := struct {
		Count  int                               `json:"unread_events_count"`
		Events []entity.Event                    `json:"events"`
		Users  map[string]entity.ApplicationUser `json:"users"`
	}{}
	er := json.Unmarshal([]byte(body), &response)
	c.Assert(er, IsNil)

	c.Assert(response.Count, Equals, 33)
	c.Assert(len(response.Events), Equals, 33)
	c.Assert(len(response.Users), Equals, 9)

	time.Sleep(10 * time.Millisecond)

	routeName = "getFeed"
	route = getComposedRoute(routeName)
	code, body, err = runRequest(routeName, route, "", signApplicationRequest(application, userFrom, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusOK)
	c.Assert(body, Not(Equals), "")

	response = struct {
		Count  int                               `json:"unread_events_count"`
		Events []entity.Event                    `json:"events"`
		Users  map[string]entity.ApplicationUser `json:"users"`
	}{}
	er = json.Unmarshal([]byte(body), &response)
	c.Assert(er, IsNil)
	c.Assert(response.Count, Equals, 0)
	c.Assert(len(response.Events), Equals, 33)
	c.Assert(len(response.Users), Equals, 9)
}

func (s *EventSuite) TestGetFeedWithCacheHeaders(c *C) {
	accounts := CorrectDeploy(1, 0, 1, 10, 10, true, true)
	application := accounts[0].Applications[0]
	userFrom := application.Users[0]

	// Check activity feed events
	routeName := "getFeed"
	route := getComposedRoute(routeName)
	code, body, headers, err := runRequestWithHeaders(routeName, route, "", func(r *http.Request) {}, signApplicationRequest(application, userFrom, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusOK)
	c.Assert(body, Not(Equals), "")
	c.Assert(headers.Get("ETag"), Not(Equals), "")
	c.Assert(headers.Get("Last-Modified"), Not(Equals), "")
	etag := headers.Get("ETag")
	lastModified := headers.Get("Last-Modified")

	response := struct {
		Count  int                               `json:"unread_events_count"`
		Events []entity.Event                    `json:"events"`
		Users  map[string]entity.ApplicationUser `json:"users"`
	}{}
	er := json.Unmarshal([]byte(body), &response)
	c.Assert(er, IsNil)

	c.Assert(response.Count, Equals, 33)
	c.Assert(len(response.Events), Equals, 33)
	c.Assert(len(response.Users), Equals, 9)

	time.Sleep(10 * time.Millisecond)

	// This request should return a different etag and last modified since
	// when we retrieved the feed we've also changed what we send next time, unread_count = 0
	code, body, headers, err = runRequestWithHeaders(routeName, route, "", func(r *http.Request) { r.Header.Set("ETag", etag) }, signApplicationRequest(application, userFrom, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusOK)
	c.Assert(body, Not(Equals), "")
	c.Assert(headers.Get("ETag"), Not(Equals), "")
	c.Assert(headers.Get("Last-Modified"), Not(Equals), "")
	c.Assert(headers.Get("ETag"), Not(Equals), etag)
	c.Assert(headers.Get("Last-Modified"), Equals, lastModified)
	etag = headers.Get("ETag")

	// Now we do our real tests
	code, body, headers, err = runRequestWithHeaders(routeName, route, "", func(r *http.Request) { r.Header.Set("If-None-Match", etag) }, signApplicationRequest(application, userFrom, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusNotModified)
	c.Assert(body, Equals, "")
	c.Assert(headers.Get("ETag"), Equals, etag)
	c.Assert(headers.Get("Last-Modified"), Equals, lastModified)

	code, body, headers, err = runRequestWithHeaders(routeName, route, "", func(r *http.Request) { r.Header.Set("If-Modified-Since", lastModified) }, signApplicationRequest(application, userFrom, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusNotModified)
	c.Assert(body, Equals, "")
	c.Assert(headers.Get("ETag"), Equals, etag)
	c.Assert(headers.Get("Last-Modified"), Equals, lastModified)

	code, body, headers, err = runRequestWithHeaders(routeName, route, "", func(r *http.Request) {
		r.Header.Set("If-None-Match", etag)
		r.Header.Set("If-Modified-Since", lastModified)
	}, signApplicationRequest(application, userFrom, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusNotModified)
	c.Assert(body, Equals, "")
	c.Assert(headers.Get("ETag"), Equals, etag)
	c.Assert(headers.Get("Last-Modified"), Equals, lastModified)
}

func (s *EventSuite) TestGetUnreadFeed(c *C) {
	accounts := CorrectDeploy(1, 0, 1, 10, 10, true, true)
	application := accounts[0].Applications[0]
	userFrom := application.Users[0]

	// Check activity feed events
	routeName := "getFeed"
	route := getComposedRoute(routeName)
	code, body, err := runRequest(routeName, route, "", signApplicationRequest(application, userFrom, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusOK)
	c.Assert(body, Not(Equals), "")

	response := struct {
		Count  int            `json:"unread_events_count"`
		Events []entity.Event `json:"events"`
	}{}
	er := json.Unmarshal([]byte(body), &response)
	c.Assert(er, IsNil)

	c.Assert(response.Count, Equals, 33)
	c.Assert(len(response.Events), Equals, 33)

	time.Sleep(10 * time.Millisecond)

	routeName = "getUnreadFeed"
	route = getComposedRoute(routeName)
	code, body, err = runRequest(routeName, route, "", signApplicationRequest(application, userFrom, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusNoContent)
	c.Assert(body, Not(Equals), "")

	response = struct {
		Count  int            `json:"unread_events_count"`
		Events []entity.Event `json:"events"`
	}{}
	er = json.Unmarshal([]byte(body), &response)
	c.Assert(er, IsNil)
	c.Assert(response.Count, Equals, 0)
	c.Assert(len(response.Events), Equals, 0)
}

func (s *EventSuite) TestGetUnreadFeedCount(c *C) {
	accounts := CorrectDeploy(1, 0, 1, 10, 10, true, true)
	application := accounts[0].Applications[0]
	userFrom := application.Users[0]

	// Check activity feed events
	routeName := "getUnreadFeedCount"
	route := getComposedRoute(routeName)
	code, body, err := runRequest(routeName, route, "", signApplicationRequest(application, userFrom, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusOK)
	c.Assert(body, Not(Equals), "")
	response := struct {
		Count  int            `json:"unread_events_count"`
		Events []entity.Event `json:"events"`
	}{}
	er := json.Unmarshal([]byte(body), &response)
	c.Assert(er, IsNil)
	c.Assert(response.Count, Equals, 33)

	routeName = "getUnreadFeed"
	route = getComposedRoute(routeName)
	code, body, err = runRequest(routeName, route, "", signApplicationRequest(application, userFrom, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusOK)
	c.Assert(body, Not(Equals), "")
	response = struct {
		Count  int            `json:"unread_events_count"`
		Events []entity.Event `json:"events"`
	}{}
	er = json.Unmarshal([]byte(body), &response)
	c.Assert(er, IsNil)
	c.Assert(response.Count, Equals, 33)
	c.Assert(len(response.Events), Equals, 33)

	time.Sleep(1 * time.Second)

	routeName = "getUnreadFeedCount"
	route = getComposedRoute(routeName)
	code, body, err = runRequest(routeName, route, "", signApplicationRequest(application, userFrom, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusOK)
	c.Assert(body, Not(Equals), "")
	er = json.Unmarshal([]byte(body), &response)
	c.Assert(er, IsNil)
	c.Assert(response.Count, Equals, 0)
}

func BenchmarkCreateEvent1_Write(b *testing.B) {
	accounts := CorrectDeploy(1, 0, 1, 1, 0, false, true)
	account := accounts[0]
	application := account.Applications[0]
	user := application.Users[0]

	payload := `{"type":"like", "language":"en"}`

	routeName := "createCurrentUserEvent"
	routePath := getComposedRoute(routeName)
	requestRoute := getRoute(routeName)
	req, er := http.NewRequest(
		requestRoute.Method,
		routePath,
		strings.NewReader(payload),
	)
	if er != nil {
		panic(er)
	}

	createCommonRequestHeaders(req)
	signApplicationRequest(application, user, true, true)(req)

	w := httptest.NewRecorder()
	m := mux.NewRouter()

	m.
		HandleFunc(requestRoute.RoutePattern(), server.CustomHandler(requestRoute, mainLogChan, errorLogChan, "test", false, true)).
		Methods(requestRoute.Method)

	for i := 1; i <= b.N; i++ {
		m.ServeHTTP(w, req)
		if w.Code != 201 {
			b.Errorf("Received non 201 code, %d", w.Code)
		}
	}
}

func BenchmarkCreateEvent2_Read(b *testing.B) {
	accounts := CorrectDeploy(1, 0, 1, 1, 1, false, true)
	account := accounts[0]
	application := account.Applications[0]
	user := application.Users[0]
	event := user.Events[0]

	routeName := "getEvent"
	routePath := getComposedRoute(routeName, user.ID, event.ID)

	requestRoute := getRoute(routeName)

	req, er := http.NewRequest(
		requestRoute.Method,
		routePath,
		nil,
	)
	if er != nil {
		panic(er)
	}

	createCommonRequestHeaders(req)
	signApplicationRequest(application, user, true, true)(req)

	w := httptest.NewRecorder()
	m := mux.NewRouter()

	m.
		HandleFunc(requestRoute.RoutePattern(), server.CustomHandler(requestRoute, mainLogChan, errorLogChan, "test", false, true)).
		Methods(requestRoute.Method)

	for i := 1; i <= b.N; i++ {
		m.ServeHTTP(w, req)
		if w.Code != 200 {
			b.Errorf("Received non 200 code, %d", w.Code)
		}
	}
}
