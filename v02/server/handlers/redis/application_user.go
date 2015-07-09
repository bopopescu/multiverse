/**
 * @author Florin Patan <florinpatan@gmail.com>
 */

package redis

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/tapglue/backend/context"
	"github.com/tapglue/backend/errors"
	"github.com/tapglue/backend/v02/core"
	"github.com/tapglue/backend/v02/entity"
	"github.com/tapglue/backend/v02/errmsg"
	"github.com/tapglue/backend/v02/server/handlers"
	"github.com/tapglue/backend/v02/server/response"
	"github.com/tapglue/backend/v02/validator"
)

type (
	applicationUser struct {
		storage core.ApplicationUser
	}
)

func (appUser *applicationUser) Read(ctx *context.Context) (err []errors.Error) {
	return []errors.Error{errmsg.ErrServerDeprecatedStorage}
	response.WriteResponse(ctx, ctx.Bag["applicationUser"].(*entity.ApplicationUser), http.StatusOK, 10)
	return
}

func (appUser *applicationUser) ReadCurrent(ctx *context.Context) (err []errors.Error) {
	return []errors.Error{errmsg.ErrServerDeprecatedStorage}
	response.WriteResponse(ctx, ctx.Bag["applicationUser"].(*entity.ApplicationUser), http.StatusOK, 10)
	return
}

func (appUser *applicationUser) UpdateCurrent(ctx *context.Context) (err []errors.Error) {
	return []errors.Error{errmsg.ErrServerDeprecatedStorage}
	user := *(ctx.Bag["applicationUser"].(*entity.ApplicationUser))
	var er error
	if er = json.Unmarshal(ctx.Body, &user); er != nil {
		return []errors.Error{errors.NewBadRequestError(0, "failed to update the user (1)\n"+er.Error(), er.Error())}
	}

	user.ID = ctx.Bag["applicationUserID"].(string)

	if err = validator.UpdateUser(
		appUser.storage,
		ctx.Bag["accountID"].(int64),
		ctx.Bag["applicationID"].(int64),
		ctx.Bag["applicationUser"].(*entity.ApplicationUser),
		&user); err != nil {
		return
	}

	updatedUser, err := appUser.storage.Update(
		ctx.Bag["accountID"].(int64),
		ctx.Bag["applicationID"].(int64),
		*(ctx.Bag["applicationUser"].(*entity.ApplicationUser)),
		user,
		true)
	if err != nil {
		return
	}

	updatedUser.Password = ""

	response.WriteResponse(ctx, updatedUser, http.StatusCreated, 0)
	return
}

func (appUser *applicationUser) DeleteCurrent(ctx *context.Context) (err []errors.Error) {
	return []errors.Error{errmsg.ErrServerDeprecatedStorage}
	if err = appUser.storage.Delete(
		ctx.Bag["accountID"].(int64),
		ctx.Bag["applicationID"].(int64),
		ctx.Bag["applicationUser"].(*entity.ApplicationUser)); err != nil {
		return
	}

	response.WriteResponse(ctx, "", http.StatusNoContent, 10)
	return
}

func (appUser *applicationUser) Create(ctx *context.Context) (err []errors.Error) {
	return []errors.Error{errmsg.ErrServerDeprecatedStorage}
	var (
		user = &entity.ApplicationUser{}
		er   error
	)

	if er = json.Unmarshal(ctx.Body, user); er != nil {
		return []errors.Error{errors.NewBadRequestError(0, "failed to create the application user (1)\n"+er.Error(), er.Error())}
	}

	if err = validator.CreateUser(
		appUser.storage,
		ctx.Bag["accountID"].(int64),
		ctx.Bag["applicationID"].(int64),
		user); err != nil {
		return
	}

	if user, err = appUser.storage.Create(
		ctx.Bag["accountID"].(int64),
		ctx.Bag["applicationID"].(int64),
		user,
		true); err != nil {
		return
	}

	user.Password = ""

	response.WriteResponse(ctx, user, http.StatusCreated, 0)
	return
}

func (appUser *applicationUser) Login(ctx *context.Context) (err []errors.Error) {
	return []errors.Error{errmsg.ErrServerDeprecatedStorage}
	var (
		loginPayload = &entity.LoginPayload{}
		user         *entity.ApplicationUser
		sessionToken string
		er           error
	)

	if er = json.Unmarshal(ctx.Body, loginPayload); er != nil {
		return []errors.Error{errors.NewBadRequestError(0, "failed to login the user (1)\n"+er.Error(), er.Error())}
	}

	if err = validator.IsValidLoginPayload(loginPayload); err != nil {
		return
	}

	if loginPayload.Email != "" {
		user, err = appUser.storage.FindByEmail(ctx.Bag["accountID"].(int64), ctx.Bag["applicationID"].(int64), loginPayload.Email)
		if err != nil {
			return
		}
	}

	if loginPayload.Username != "" {
		user, err = appUser.storage.FindByUsername(ctx.Bag["accountID"].(int64), ctx.Bag["applicationID"].(int64), loginPayload.Username)
		if err != nil {
			return
		}
	}

	if user == nil {
		return []errors.Error{errors.NewInternalError(0, "failed to login the application user (2)\n", "user is nil")}
	}

	if !user.Enabled {
		return []errors.Error{errors.NewNotFoundError(0, "failed to login the user (3)\nuser is disabled", "user is disabled")}
	}

	if err = validator.ApplicationUserCredentialsValid(loginPayload.Password, user); err != nil {
		return
	}

	if sessionToken, err = appUser.storage.CreateSession(
		ctx.Bag["accountID"].(int64),
		ctx.Bag["applicationID"].(int64),
		user); err != nil {
		return
	}

	timeNow := time.Now()
	user.LastLogin = &timeNow
	_, err = appUser.storage.Update(
		ctx.Bag["accountID"].(int64),
		ctx.Bag["applicationID"].(int64),
		*user,
		*user,
		false)
	if err != nil {
		return
	}

	response.WriteResponse(ctx, struct {
		UserID string `json:"id"`
		Token  string `json:"session_token"`
	}{
		UserID: user.ID,
		Token:  sessionToken,
	}, http.StatusCreated, 0)
	return
}

func (appUser *applicationUser) RefreshSession(ctx *context.Context) (err []errors.Error) {
	return []errors.Error{errmsg.ErrServerDeprecatedStorage}
	var (
		tokenPayload struct {
			Token string `json:"session_token"`
		}
		sessionToken string
		er           error
	)

	if er = json.Unmarshal(ctx.Body, &tokenPayload); er != nil {
		return []errors.Error{errors.NewBadRequestError(0, "failed to refresh the session token (1)\n"+er.Error(), er.Error())}
	}

	if tokenPayload.Token != ctx.SessionToken {
		return []errors.Error{errors.NewBadRequestError(0, "failed to refresh the session token (2)\nsession token mismatch", "session token mismatch")}
	}

	if sessionToken, err = appUser.storage.RefreshSession(
		ctx.Bag["accountID"].(int64),
		ctx.Bag["applicationID"].(int64),
		ctx.SessionToken,
		ctx.Bag["applicationUser"].(*entity.ApplicationUser)); err != nil {
		return
	}

	response.WriteResponse(ctx, struct {
		Token string `json:"session_token"`
	}{Token: sessionToken}, http.StatusCreated, 0)
	return
}

func (appUser *applicationUser) Logout(ctx *context.Context) (err []errors.Error) {
	return []errors.Error{errmsg.ErrServerDeprecatedStorage}
	var (
		tokenPayload struct {
			Token string `json:"session_token"`
		}
		er error
	)

	if er = json.Unmarshal(ctx.Body, &tokenPayload); er != nil {
		return []errors.Error{errors.NewBadRequestError(0, "failed to logout the user (1)\n"+er.Error(), er.Error())}
	}

	if tokenPayload.Token != ctx.SessionToken {
		return []errors.Error{errors.NewBadRequestError(0, "failed to logout the user (2)\nsession token mismatch", "session token mismatch")}
	}

	if err = appUser.storage.DestroySession(
		ctx.Bag["accountID"].(int64),
		ctx.Bag["applicationID"].(int64),
		ctx.SessionToken,
		ctx.Bag["applicationUser"].(*entity.ApplicationUser)); err != nil {
		return
	}

	response.WriteResponse(ctx, "logged out", http.StatusOK, 0)
	return
}

func (appUser *applicationUser) Search(*context.Context) []errors.Error {
	return []errors.Error{errors.NewInternalError(0, "not implemented yet", "not implemented yet")}
}

func (appUser *applicationUser) PopulateContext(ctx *context.Context) (err []errors.Error) {
	return []errors.Error{errmsg.ErrServerDeprecatedStorage}
	ctx.Bag["applicationUser"], err = appUser.storage.Read(ctx.Bag["accountID"].(int64), ctx.Bag["applicationID"].(int64), ctx.Bag["applicationUserID"].(string))
	return
}

// NewApplicationUser returns a new application user routes handler
func NewApplicationUser(storage core.ApplicationUser) handlers.ApplicationUser {
	return &applicationUser{
		storage: storage,
	}
}