package postgres

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tapglue/multiverse/context"
	"github.com/tapglue/multiverse/errors"
	"github.com/tapglue/multiverse/v02/core"
	"github.com/tapglue/multiverse/v02/entity"
	"github.com/tapglue/multiverse/v02/errmsg"
	"github.com/tapglue/multiverse/v02/server/handlers"
	"github.com/tapglue/multiverse/v02/server/response"
	"github.com/tapglue/multiverse/v02/validator"
)

type (
	application struct {
		storage core.Application
	}
)

func (app *application) Read(ctx *context.Context) (err []errors.Error) {
	return []errors.Error{errmsg.ErrServerAPIVersionRemoved}
	// TODO This one read only the current application maybe we want to have something to read any application?
	response.ComputeApplicationLastModified(ctx, ctx.Bag["application"].(*entity.Application))
	response.WriteResponse(ctx, ctx.Bag["application"].(*entity.Application), http.StatusOK, 10)
	return
}

func (app *application) Update(ctx *context.Context) (err []errors.Error) {
	return []errors.Error{errmsg.ErrServerAPIVersionRemoved}
	application := *(ctx.Bag["application"].(*entity.Application))
	if er := json.Unmarshal(ctx.Body, &application); er != nil {
		return []errors.Error{errmsg.ErrServerReqBadJSONReceived.UpdateMessage(er.Error())}
	}

	application.ID = ctx.Bag["applicationID"].(int64)
	application.AccountID = ctx.Bag["accountID"].(int64)
	application.PublicID = ctx.Bag["account"].(*entity.Account).PublicID

	if err = validator.UpdateApplication(ctx.Bag["application"].(*entity.Application), &application); err != nil {
		return
	}

	updatedApplication, err := app.storage.Update(*ctx.Bag["application"].(*entity.Application), application, true)
	if err != nil {
		return
	}

	response.WriteResponse(ctx, updatedApplication, http.StatusCreated, 0)
	return
}

func (app *application) Delete(ctx *context.Context) (err []errors.Error) {
	return []errors.Error{errmsg.ErrServerAPIVersionRemoved}
	if err = app.storage.Delete(ctx.Bag["application"].(*entity.Application)); err != nil {
		return
	}

	response.WriteResponse(ctx, "", http.StatusNoContent, 10)
	return
}

func (app *application) Create(ctx *context.Context) (err []errors.Error) {
	return []errors.Error{errmsg.ErrServerAPIVersionRemoved}
	var (
		application = &entity.Application{}
	)

	if er := json.Unmarshal(ctx.Body, application); er != nil {
		return []errors.Error{errmsg.ErrServerReqBadJSONReceived.UpdateMessage(er.Error())}
	}

	application.AccountID = ctx.Bag["accountID"].(int64)
	application.PublicAccountID = ctx.Bag["account"].(*entity.Account).PublicID

	if err = validator.CreateApplication(application); err != nil {
		return
	}

	if application, err = app.storage.Create(application, true); err != nil {
		return
	}

	response.WriteResponse(ctx, application, http.StatusCreated, 0)
	return
}

func (app *application) List(ctx *context.Context) (err []errors.Error) {
	return []errors.Error{errmsg.ErrServerAPIVersionRemoved}
	var (
		applications []*entity.Application
	)

	if applications, err = app.storage.List(ctx.Bag["accountID"].(int64)); err != nil {
		return
	}

	resp := &struct {
		Applications []*entity.Application `json:"applications"`
	}{
		Applications: applications,
	}

	response.ComputeApplicationsLastModified(ctx, resp.Applications)

	response.WriteResponse(ctx, resp, http.StatusOK, 10)
	return
}

func (app *application) PopulateContext(ctx *context.Context) (err []errors.Error) {
	user, pass, ok := ctx.BasicAuth()
	if !ok {
		return []errors.Error{errmsg.ErrAuthInvalidApplicationCredentials.UpdateInternalMessage(fmt.Sprintf("got %s:%s", user, pass))}
	}
	ctx.Bag["application"], err = app.storage.FindByKey(user)
	if err == nil {
		ctx.Bag["accountID"] = ctx.Bag["application"].(*entity.Application).AccountID
		ctx.Bag["applicationID"] = ctx.Bag["application"].(*entity.Application).ID
	}
	return
}

func (app *application) PopulateContextFromID(ctx *context.Context) (err []errors.Error) {
	applicationID := ctx.Vars["applicationID"]
	if !validator.IsValidUUID5(applicationID) {
		return []errors.Error{errmsg.ErrApplicationIDInvalid}
	}

	ctx.Bag["application"], err = app.storage.FindByPublicID(applicationID)
	if err == nil {
		if ctx.Bag["application"].(*entity.Application) == nil {
			return []errors.Error{errmsg.ErrApplicationNotFound}
		}

		ctx.Bag["accountID"] = ctx.Bag["application"].(*entity.Application).AccountID
		ctx.Bag["applicationID"] = ctx.Bag["application"].(*entity.Application).ID
	}
	return
}

// NewApplication returns a new application route handler
func NewApplication(storage core.Application) handlers.Application {
	return &application{
		storage: storage,
	}
}
