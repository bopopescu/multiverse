/**
 * @author Onur Akpolat <onurakpolat@gmail.com>
 */

package db

import (
	"github.com/tapglue/backend/entity"

	. "gopkg.in/check.v1"
)

func (dbs *DatabaseSuite) TestAddAccount_Empty(c *C) {
	InitDatabases(cfg.DB())

	var account = &entity.Account{}

	savedAccount, err := AddAccount(account)

	c.Assert(savedAccount, IsNil)
	c.Assert(err, Not(IsNil))
}

func (dbs *DatabaseSuite) TestAddAccount_Normal(c *C) {
	InitDatabases(cfg.DB())

	var account = &entity.Account{
		Name: "Demo",
	}

	savedAccount, err := AddAccount(account)

	c.Assert(savedAccount, Not(IsNil))
	c.Assert(err, IsNil)
	c.Assert(savedAccount.Name, Equals, account.Name)
	c.Assert(savedAccount.Enabled, Equals, true)
}

// Test GetAccountByID
func (dbs *DatabaseSuite) TestGetAccountByID(c *C) {
	c.Skip("not implemented yet")
}
