/**
 * @author Onur Akpolat <onurakpolat@gmail.com>
 */

package core

import . "gopkg.in/check.v1"

// AddAccount test to write empty entity
func (cs *CoreSuite) TestAddAccount_Empty(c *C) {
	// Write account
	savedAccount, err := AddAccount(emtpyAccount, true)

	// Perform tests
	c.Assert(savedAccount, IsNil)
	c.Assert(err, NotNil)
}

// AddAccount test to write account entity with just a name
func (cs *CoreSuite) TestAddAccount_Correct(c *C) {
	// Write account
	savedAccount, err := AddAccount(correctAccount, true)

	// Perform tests
	c.Assert(savedAccount, NotNil)
	c.Assert(err, IsNil)
	c.Assert(savedAccount.Name, Equals, correctAccount.Name)
	c.Assert(savedAccount.Enabled, Equals, true)
}

// GetAccountByID test to get an account by its id
func (cs *CoreSuite) TestGetAccountByID(c *C) {
	// Write correct account
	savedAccount := AddCorrectAccount()

	// Get account by id
	getAccount, err := GetAccountByID(savedAccount.ID)

	// Perform tests
	c.Assert(err, IsNil)
	c.Assert(getAccount, DeepEquals, savedAccount)
}

// BenchmarkAddAccount executes AddAccount 1000 times
func (cs *CoreSuite) BenchmarkAddAccount(c *C) {
	var i int64
	// Loop to create 1000 accounts
	for i = 1; i <= 1000; i++ {
		correctAccount.ID = i
		_, _ = AddAccount(correctAccount, false)
	}
}

// BenchmarkAddAccount executes GetAccount 1000 times
func (cs *CoreSuite) BenchmarkGetAccount(c *C) {
	var i int64
	// Loop to create 1000 accounts
	for i = 1; i <= 1000; i++ {
		_, _ = GetAccountByID(i)
	}
}