/**
 * @author Onur Akpolat <onurakpolat@gmail.com>
 */

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/tapglue/backend/entity"
	. "gopkg.in/check.v1"
)

// Test CLHeader
func test_CLHeader(req *http.Request) {
	req.Header.Add("Content-Length", strconv.FormatInt(int64(len("{name:'Demo'}")), 10))
}

// Test create acccount request with a wrong key
func (s *ServerSuite) TestCreateAccount_WrongKey(c *C) {
	req, err := http.NewRequest(
		"POST",
		"http://localhost:8089/account",
		strings.NewReader("{namae:''}"),
	)
	test_CLHeader(req)
	c.Assert(err, IsNil)

	w := httptest.NewRecorder()
	createAccount(w, req)

	c.Assert(w.Code, Equals, http.StatusBadRequest)
	c.Assert(w.Body.String(), Not(Equals), "")
}

// Test a correct create account request
func (s *ServerSuite) TestCreateAccount_Correct(c *C) {
	payload := "{\"name\":\"Demo\"}"
	req, err := http.NewRequest(
		"POST",
		"http://localhost:8089/account",
		strings.NewReader(payload),
	)
	test_CLHeader(req)
	c.Assert(err, IsNil)

	w := httptest.NewRecorder()
	createAccount(w, req)

	c.Assert(w.Code, Equals, http.StatusCreated)
	response := w.Body.String()
	c.Assert(response, Not(Equals), "")

	account := &entity.Account{}
	err = json.Unmarshal([]byte(response), account)
	c.Assert(err, IsNil)
	if account.ID < 1 {
		c.Fail()
	}
	c.Assert(account.Name, Equals, "Demo")
	c.Assert(account.Enabled, Equals, true)
}

// Test getAccount
func (s *ServerSuite) TestGetAccount_OK(c *C) {
	// This (section until GET) should probably be reduced to setting up some common dummy data
	payload := "{\"name\":\"Demo\"}"
	req, err := http.NewRequest(
		"POST",
		"http://localhost:8089/account",
		strings.NewReader(payload),
	)
	test_CLHeader(req)
	c.Assert(err, IsNil)

	w := httptest.NewRecorder()
	createAccount(w, req)

	c.Assert(w.Code, Equals, http.StatusCreated)
	response := w.Body.String()
	c.Assert(response, Not(Equals), "")

	account := &entity.Account{}
	err = json.Unmarshal([]byte(response), account)
	c.Assert(err, IsNil)
	if account.ID < 1 {
		c.Fail()
	}
	c.Assert(account.Name, Equals, "Demo")
	c.Assert(account.Enabled, Equals, true)

	// Now we test the GET part

	req, err = http.NewRequest(
		"GET",
		fmt.Sprintf("http://localhost:8089/account/%d", account.ID),
		nil,
	)
	fmt.Println(fmt.Sprintf("http://localhost:8089/account/%d", account.ID))
	c.Assert(err, IsNil)

	w = httptest.NewRecorder()
	m := mux.NewRouter()
	route := routes["getAccount"]
	m.HandleFunc(route.pattern, route.handlerFunc).Methods(route.method)
	m.ServeHTTP(w, req)

	c.Assert(w.Code, Equals, http.StatusOK)
	response = w.Body.String()
	c.Assert(response, Not(Equals), "")

	accountGet := &entity.Account{}
	err = json.Unmarshal([]byte(response), accountGet)
	c.Assert(err, IsNil)
	c.Assert(accountGet.ID, Equals, account.ID)
	c.Assert(accountGet.Name, Equals, account.Name)
	c.Assert(accountGet.Enabled, Equals, true)
}
