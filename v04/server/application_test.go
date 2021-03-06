// +build !bench

package server_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/tapglue/multiverse/v04/entity"

	. "gopkg.in/check.v1"
)

// Test createApplication request with a wrong key
func (s *ApplicationSuite) TestCreateApplication_WrongKey(c *C) {
	account, err := AddCorrectOrganization(true)
	c.Assert(err, IsNil)

	accountUser, err := AddCorrectMember(account.ID, true)
	c.Assert(err, IsNil)

	LoginMember(accountUser)

	payload := "{namae:''}"

	routeName := "createApplication"
	route := getComposedRoute(routeName, account.PublicID)
	code, body, err := runRequest(routeName, route, payload, signOrganizationRequest(account, accountUser, true, true))
	c.Assert(err, IsNil)

	c.Assert(code, Equals, http.StatusBadRequest)
	c.Assert(body, Not(Equals), "")
}

// Test createApplication request with an wrong name
func (s *ApplicationSuite) TestCreateApplication_WrongValue(c *C) {
	account, err := AddCorrectOrganization(true)
	c.Assert(err, IsNil)

	accountUser, err := AddCorrectMember(account.ID, true)
	c.Assert(err, IsNil)

	LoginMember(accountUser)

	payload := `{"name":""}`

	routeName := "createApplication"
	route := getComposedRoute(routeName, account.PublicID)
	code, body, err := runRequest(routeName, route, payload, signOrganizationRequest(account, accountUser, true, true))
	c.Assert(err, IsNil)

	c.Assert(code, Equals, http.StatusBadRequest)
	c.Assert(body, Not(Equals), "")
}

// Test a correct createApplication request
func (s *ApplicationSuite) TestCreateApplication_OK(c *C) {
	account, err := AddCorrectOrganization(true)
	c.Assert(err, IsNil)

	accountUser, err := AddCorrectMember(account.ID, true)
	c.Assert(err, IsNil)

	LoginMember(accountUser)

	application := CorrectApplication()

	payload := fmt.Sprintf(
		`{"name":"%s", "description":"%s", "url": "%s"}`,
		application.Name,
		application.Description,
		application.URL,
	)
	c.Assert(err, IsNil)

	routeName := "createApplication"
	route := getComposedRoute(routeName, account.PublicID)
	code, body, err := runRequest(routeName, route, payload, signOrganizationRequest(account, accountUser, true, true))
	c.Assert(err, IsNil)

	c.Assert(code, Equals, http.StatusCreated)

	c.Assert(body, Not(Equals), "")

	receivedApplication := &entity.Application{}
	er := json.Unmarshal([]byte(body), receivedApplication)
	c.Assert(er, IsNil)
	if receivedApplication.PublicID == "" {
		c.Fail()
	}
	c.Assert(receivedApplication.ID, Not(Equals), "")
	c.Assert(receivedApplication.Name, Equals, application.Name)
	c.Assert(receivedApplication.Description, Equals, application.Description)
	c.Assert(receivedApplication.URL, Equals, application.URL)
	c.Assert(receivedApplication.Enabled, Equals, true)
}

// Test a correct updateApplication request
func (s *ApplicationSuite) TestUpdateApplication_OK(c *C) {
	accounts := CorrectDeploy(1, 1, 1, 0, 0, false, true)
	account := accounts[0]
	accountUser := account.Members[0]
	application := account.Applications[0]

	payload := fmt.Sprintf(
		`{"name":"%s", "description":"i changed the description", "url": "%s", "enabled": true}`,
		application.Name,
		application.URL,
	)

	routeName := "updateApplication"
	route := getComposedRoute(routeName, account.PublicID, application.PublicID)
	code, body, err := runRequest(routeName, route, payload, signOrganizationRequest(account, accountUser, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusCreated)
	c.Assert(body, Not(Equals), "")

	receivedApplication := &entity.Application{}
	er := json.Unmarshal([]byte(body), receivedApplication)
	c.Assert(er, IsNil)
	if receivedApplication.PublicID == "" {
		c.Fail()
	}

	c.Assert(receivedApplication.PublicID, Equals, application.PublicID)
	c.Assert(receivedApplication.PublicOrgID, Equals, application.PublicOrgID)
	c.Assert(receivedApplication.Name, Equals, application.Name)
	c.Assert(receivedApplication.URL, Equals, application.URL)
	c.Assert(receivedApplication.Enabled, Equals, true)
}

// Test a correct updateApplication request with a wrong id
func (s *ApplicationSuite) TestUpdateApplication_WrongID(c *C) {
	accounts := CorrectDeploy(1, 1, 1, 0, 0, false, true)
	account := accounts[0]
	accountUser := account.Members[0]
	application := account.Applications[0]

	payload := fmt.Sprintf(
		`{"name":"%s", "description":"i changed the description", "url": "%s", "enabled": true}`,
		application.Name,
		application.URL,
	)

	routeName := "updateApplication"
	route := getComposedRoute(routeName, application.PublicOrgID, application.PublicID+"a")
	code, _, er := runRequest(routeName, route, payload, signOrganizationRequest(account, accountUser, true, true))
	c.Assert(er, IsNil)
	c.Assert(code, Equals, http.StatusNotFound)
}

// Test a correct updateApplication request with an invalid description
func (s *ApplicationSuite) TestUpdateApplication_WrongValue(c *C) {
	accounts := CorrectDeploy(1, 1, 1, 0, 0, false, true)
	account := accounts[0]
	accountUser := account.Members[0]
	application := account.Applications[0]

	payload := fmt.Sprintf(
		`{"name":"%s", "description":"", "url": "%s", "enabled": true}`,
		application.Name,
		application.URL,
	)

	routeName := "updateApplication"
	route := getComposedRoute(routeName, application.PublicOrgID, application.PublicID+"a")
	code, _, err := runRequest(routeName, route, payload, signOrganizationRequest(account, accountUser, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusNotFound)
}

// Test a correct updateApplication request with a wrong token
func (s *ApplicationSuite) TestUpdateApplication_WrongToken(c *C) {
	accounts := CorrectDeploy(1, 1, 1, 0, 0, false, true)
	account := accounts[0]
	accountUser := account.Members[0]
	application := account.Applications[0]

	payload := fmt.Sprintf(
		`{"name":"%s", "description":"i changed the description", "url": "%s", "enabled": true}`,
		application.Name,
		application.URL,
	)

	routeName := "updateApplication"
	route := getComposedRoute(routeName, application.PublicOrgID, application.PublicID)
	code, _, err := runRequest(routeName, route, payload, signOrganizationRequest(account, accountUser, false, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusNotFound)
}

// Test a correct deleteApplication request
func (s *ApplicationSuite) TestDeleteApplication_OK(c *C) {
	accounts := CorrectDeploy(1, 1, 1, 0, 0, false, true)
	account := accounts[0]
	accountUser := account.Members[0]
	application := account.Applications[0]

	routeName := "deleteApplication"
	route := getComposedRoute(routeName, account.PublicID, application.PublicID)
	code, _, err := runRequest(routeName, route, "", signOrganizationRequest(account, accountUser, true, true))

	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusNoContent)
}

// Test a correct deleteApplication request with a wrong id
func (s *ApplicationSuite) TestDeleteApplication_WrongID(c *C) {
	accounts := CorrectDeploy(1, 1, 1, 0, 0, false, true)
	account := accounts[0]
	accountUser := account.Members[0]
	application := account.Applications[0]

	routeName := "deleteApplication"
	route := getComposedRoute(routeName, account.PublicID, application.PublicID+"1")
	code, _, err := runRequest(routeName, route, "", signOrganizationRequest(account, accountUser, true, true))
	c.Assert(err, IsNil)

	c.Assert(code, Equals, http.StatusNotFound)
}

// Test a correct deleteApplication request with a wrong token
func (s *ApplicationSuite) TestDeleteApplication_WrongToken(c *C) {
	accounts := CorrectDeploy(1, 1, 1, 0, 0, false, true)
	account := accounts[0]
	accountUser := account.Members[0]
	application := account.Applications[0]

	routeName := "deleteApplication"
	route := getComposedRoute(routeName, account.PublicID, application.PublicID+"1")
	code, _, err := runRequest(routeName, route, "", signOrganizationRequest(account, accountUser, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusNotFound)
}

// Test a correct getApplication request
func (s *ApplicationSuite) TestGetApplication_OK(c *C) {
	accounts := CorrectDeploy(1, 1, 1, 0, 0, false, true)
	account := accounts[0]
	accountUser := account.Members[0]
	application := account.Applications[rand.Intn(1)]

	routeName := "getApplication"
	route := getComposedRoute(routeName, account.PublicID, application.PublicID)
	code, body, err := runRequest(routeName, route, "", signOrganizationRequest(account, accountUser, true, true))
	c.Assert(err, IsNil)

	c.Assert(code, Equals, http.StatusOK)

	c.Assert(body, Not(Equals), "")

	receivedApplication := &entity.Application{}
	er := json.Unmarshal([]byte(body), receivedApplication)
	c.Assert(er, IsNil)
	c.Assert(receivedApplication.PublicID, Equals, application.PublicID)
	c.Assert(receivedApplication.Name, Equals, application.Name)
	c.Assert(receivedApplication.Description, Equals, application.Description)
	c.Assert(receivedApplication.Enabled, Equals, true)
}

// Test a correct getApplication request with a wrong id
func (s *ApplicationSuite) TestGetApplication_WrongID(c *C) {
	accounts := CorrectDeploy(1, 1, 1, 0, 0, false, true)
	account := accounts[0]
	accountUser := account.Members[0]
	application := account.Applications[0]

	routeName := "getApplication"
	route := getComposedRoute(routeName, account.PublicID, application.PublicID+"a")
	code, _, err := runRequest(routeName, route, "", signOrganizationRequest(account, accountUser, true, true))
	c.Assert(err, IsNil)

	c.Assert(code, Equals, http.StatusNotFound)
}

// Test a correct getApplication request with a wrong token
func (s *ApplicationSuite) TestGetApplication_WrongToken(c *C) {
	accounts := CorrectDeploy(1, 1, 1, 0, 0, false, true)
	account := accounts[0]
	accountUser := account.Members[0]
	application := account.Applications[0]

	routeName := "getApplication"
	route := getComposedRoute(routeName, account.PublicID, application.PublicID)
	code, _, err := runRequest(routeName, route, "", signOrganizationRequest(account, accountUser, true, false))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusNotFound)
}

func (s *ApplicationSuite) TestGetApplicationListWorks(c *C) {
	accounts := CorrectDeploy(2, 1, 1, 0, 0, false, true)
	account := accounts[0]
	accountUser := account.Members[0]
	expected := account.Applications[0]

	routeName := "getApplications"
	route := getComposedRoute(routeName, account.PublicID)
	code, body, err := runRequest(routeName, route, "", signOrganizationRequest(account, accountUser, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusOK)
	c.Assert(body, Not(Equals), "")

	response := &struct {
		Applications []*entity.Application `json:"applications"`
	}{}

	er := json.Unmarshal([]byte(body), response)
	c.Assert(er, IsNil)
	c.Assert(len(response.Applications), Equals, 1)
	received := response.Applications[0]
	received.CreatedAt = expected.CreatedAt
	received.UpdatedAt = expected.UpdatedAt
	expected.Users = nil
	expected.ID = 0
	expected.OrgID = 0

	c.Assert(received, DeepEquals, expected)
}

func (s *ApplicationSuite) TestApplicationMalformedPayloadsFails(c *C) {
	accounts := CorrectDeploy(1, 1, 1, 0, 0, false, true)
	account := accounts[0]
	accountUser := account.Members[0]
	application := account.Applications[0]

	scenarios := []struct {
		Payload      string
		RouteName    string
		Route        string
		StatusCode   int
		ResponseBody string
	}{
		{
			Payload:      "{",
			RouteName:    "updateApplication",
			Route:        getComposedRoute("updateApplication", account.PublicID, application.PublicID),
			StatusCode:   http.StatusBadRequest,
			ResponseBody: `{"errors":[{"code":5001,"message":"unexpected end of JSON input"}]}` + "\n",
		},
	}

	for idx := range scenarios {
		code, body, err := runRequest(scenarios[idx].RouteName, scenarios[idx].Route, scenarios[idx].Payload, signOrganizationRequest(account, accountUser, true, true))
		c.Logf("pass: %d", idx)
		c.Assert(err, IsNil)
		c.Assert(code, Equals, scenarios[idx].StatusCode)
		c.Assert(body, Equals, scenarios[idx].ResponseBody)
	}
}

func (s *ApplicationSuite) TestUpdateApplicationState(c *C) {
	account := CorrectDeploy(1, 1, 1, 0, 0, false, true)[0]
	accountUser := account.Members[0]
	application := account.Applications[0]

	payload := fmt.Sprintf(
		`{"name":"%s", "description":"i changed the description", "url": "%s", "enabled": true, "in_production": true}`,
		application.Name,
		application.URL,
	)

	routeName := "updateApplication"
	route := getComposedRoute(routeName, account.PublicID, application.PublicID)
	code, body, err := runRequest(routeName, route, payload, signOrganizationRequest(account, accountUser, true, true))
	c.Assert(err, IsNil)
	c.Assert(code, Equals, http.StatusCreated)
	c.Assert(body, Not(Equals), "")

	receivedApplication := &entity.Application{}
	er := json.Unmarshal([]byte(body), receivedApplication)
	c.Assert(er, IsNil)
	if receivedApplication.PublicID == "" {
		c.Fail()
	}

	c.Assert(receivedApplication.Name, Equals, application.Name)
	c.Assert(receivedApplication.URL, Equals, application.URL)
	c.Assert(receivedApplication.Enabled, Equals, true)
	c.Assert(receivedApplication.InProduction, Equals, true)
}
