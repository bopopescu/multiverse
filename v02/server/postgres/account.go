/**
 * @author Florin Patan <florinpatan@gmail.com>
 */

package postgres

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tapglue/backend/context"
	"github.com/tapglue/backend/errors"
	"github.com/tapglue/backend/v02/core"
	"github.com/tapglue/backend/v02/entity"
	"github.com/tapglue/backend/v02/server"
	"github.com/tapglue/backend/v02/validator"
)

type (
	account struct {
		storage core.Account
	}
)

func (acc *account) Read(ctx *context.Context) (err []errors.Error) {
	if ctx.Bag["account"] == nil {
		return []errors.Error{errors.NewInternalError("request is missing account context", "context missing")}
	}

	if ctx.Bag["account"].(*entity.Account).PublicID != ctx.Vars["accountID"] {
		return []errors.Error{errors.NewBadRequestError("account mismatch", "account mismatch")}
	}

	computeAccountLastModified(ctx, ctx.Bag["account"].(*entity.Account))

	server.WriteResponse(ctx, ctx.Bag["account"].(*entity.Account), http.StatusOK, 10)
	return
}

func (acc *account) Update(ctx *context.Context) (err []errors.Error) {
	account := *(ctx.Bag["account"].(*entity.Account))

	if account.PublicID != ctx.Vars["accountID"] {
		return []errors.Error{errors.New(errors.ConflictError, "failed to update the account (2)\naccount ID mismatch", "account ID mismatch", false)}
	}

	if er := json.Unmarshal(ctx.Body, &account); er != nil {
		return []errors.Error{errors.NewBadRequestError("failed to update the account (2)\n"+er.Error(), "malformed json received")}
	}

	account.ID = ctx.Bag["accountID"].(int64)

	if err := validator.UpdateAccount(ctx.Bag["account"].(*entity.Account), &account); err != nil {
		return err
	}

	updatedAccount, err := acc.storage.Update(*(ctx.Bag["account"].(*entity.Account)), account, true)
	if err != nil {
		return err
	}

	server.WriteResponse(ctx, updatedAccount, http.StatusCreated, 10)
	return nil
}

func (acc *account) Delete(ctx *context.Context) (err []errors.Error) {
	if ctx.R.Header.Get("X-Jarvis-Auth") != "ZTBmZjI3MGE2M2YzYzAzOWI1MjhiYTNi" {
		return []errors.Error{errors.NewNotFoundError("not found", "request does not contain a correct Jarvis auth")}
	}

	if ctx.Bag["account"].(*entity.Account).PublicID != ctx.Vars["accountID"] {
		return []errors.Error{errors.NewBadRequestError("account mismatch", "account mismatch")}
	}

	if err = acc.storage.Delete(ctx.Bag["account"].(*entity.Account)); err != nil {
		return err
	}

	server.WriteResponse(ctx, "", http.StatusNoContent, 10)
	return nil
}

func (acc *account) Create(ctx *context.Context) (err []errors.Error) {
	if ctx.R.Header.Get("X-Jarvis-Auth") != "ZTBmZjI3MGE2M2YzYzAzOWI1MjhiYTNi" {
		return []errors.Error{errors.NewNotFoundError("not found", "request does not contain a correct Jarvis auth")}
	}

	var account = &entity.Account{}

	if er := json.Unmarshal(ctx.Body, account); er != nil {
		return []errors.Error{errors.NewBadRequestError("failed to create the account (1)\n"+er.Error(), er.Error())}
	}

	if err = validator.CreateAccount(account); err != nil {
		return
	}

	if account, err = acc.storage.Create(account, true); err != nil {
		return
	}

	server.WriteResponse(ctx, account, http.StatusCreated, 0)
	return
}

func (acc *account) PopulateContext(ctx *context.Context) (err []errors.Error) {
	user, pass, ok := ctx.BasicAuth()
	if !ok {
		return []errors.Error{errors.NewBadRequestError("error while reading account credentials", fmt.Sprintf("got %s:%s", user, pass))}
	}
	account, err := acc.storage.FindByKey(user)
	if account == nil {
		return []errors.Error{errors.NewNotFoundError("account not found", "account not found")}
	}
	if err == nil {
		ctx.Bag["account"] = account
		ctx.Bag["accountID"] = account.ID
	}
	return
}

// NewAccount returns a new account handler tweaked specifically for Kinesis
func NewAccount(datastore core.Account) server.Account {
	return &account{
		storage: datastore,
	}
}
