/**
 * @author Florin Patan <florinpatan@gmail.com>
 */

package postgres

import (
	"github.com/tapglue/backend/context"
	"github.com/tapglue/backend/errors"
	"github.com/tapglue/backend/v02/core"
	"github.com/tapglue/backend/v02/server"
)

type (
	account struct {
		storage core.Account
	}
)

func (acc *account) Read(ctx *context.Context) (err errors.Error) {
	return errors.NewInternalError("not implemented yet", "not implemented yet")
}

func (acc *account) Update(ctx *context.Context) (err errors.Error) {
	return errors.NewInternalError("not implemented yet", "not implemented yet")
}

func (acc *account) Delete(ctx *context.Context) (err errors.Error) {
	return errors.NewInternalError("not implemented yet", "not implemented yet")
}

func (acc *account) Create(ctx *context.Context) (err errors.Error) {
	return errors.NewInternalError("not implemented yet", "not implemented yet")
}

func (acc *account) PopulateContext(ctx *context.Context) (err errors.Error) {
	return errors.NewInternalError("not implemented yet", "not implemented yet")
}

// NewAccount returns a new account handler tweaked specifically for Kinesis
func NewAccount(datastore core.Account) server.Account {
	return &account{
		storage: datastore,
	}
}
