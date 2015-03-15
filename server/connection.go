/**
 * @author Florin Patan <florinpatan@gmail.com>
 */

package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"fmt"

	"github.com/tapglue/backend/context"
	"github.com/tapglue/backend/core"
	"github.com/tapglue/backend/core/entity"
	"github.com/tapglue/backend/validator"
)

// updateConnection handles requests to update a user connection
// Request: PUT account/:AccountID/application/:ApplicationID/user/:UserFromID/connection/:UserToID
func updateConnection(ctx *context.Context) {
	var (
		userToID int64
		err      error
	)

	if userToID, err = strconv.ParseInt(ctx.Vars["userToId"], 10, 64); err != nil {
		errorHappened(ctx, "userToId is not set or the value is incorrect", http.StatusBadRequest, err)
		return
	}

	existingConnection, err := core.ReadConnection(ctx.AccountID, ctx.ApplicationID, ctx.ApplicationUserID, userToID)
	if err != nil {
		errorHappened(ctx, "unexpected error happened", http.StatusInternalServerError, err)
		return
	}

	if existingConnection == nil {
		errorHappened(ctx, "users are not connected", http.StatusBadRequest, fmt.Errorf("users are not connected"))
		return
	}

	connection := *existingConnection
	decoder := json.NewDecoder(ctx.Body)
	if err = decoder.Decode(&connection); err != nil {
		errorHappened(ctx, err.Error(), http.StatusBadRequest, err)
		return
	}

	connection.AccountID = ctx.AccountID
	connection.ApplicationID = ctx.ApplicationID
	connection.UserFromID = ctx.ApplicationUserID
	connection.UserToID = userToID

	if err = validator.UpdateConnection(existingConnection, &connection); err != nil {
		errorHappened(ctx, err.Error(), http.StatusBadRequest, err)
		return
	}

	updatedConnection, err := core.UpdateConnection(*existingConnection, connection, false)
	if err != nil {
		errorHappened(ctx, err.Error(), http.StatusInternalServerError, err)
		return
	}

	writeResponse(ctx, updatedConnection, http.StatusCreated, 0)
}

// deleteConnection handles requests to delete a single connection
// Request: DELETE account/:AccountID/application/:ApplicationID/user/:UserFromID/connection/:UserToID
func deleteConnection(ctx *context.Context) {
	var (
		userToID int64
		err      error
	)

	if userToID, err = strconv.ParseInt(ctx.Vars["userToId"], 10, 64); err != nil {
		errorHappened(ctx, "userToId is not set or the value is incorrect", http.StatusBadRequest, err)
		return
	}

	if err = core.DeleteConnection(ctx.AccountID, ctx.ApplicationID, ctx.ApplicationUserID, userToID); err != nil {
		errorHappened(ctx, err.Error(), http.StatusInternalServerError, err)
		return
	}

	writeResponse(ctx, "", http.StatusNoContent, 10)
}

// createConnection handles requests to create a user connection
// Request: POST /application/:applicationId/user/:UserID/connections
func createConnection(ctx *context.Context) {
	var (
		connection = &entity.Connection{}
		err        error
	)

	if err = json.NewDecoder(ctx.Body).Decode(connection); err != nil {
		errorHappened(ctx, err.Error(), http.StatusBadRequest, err)
		return
	}

	connection.AccountID = ctx.AccountID
	connection.ApplicationID = ctx.ApplicationID
	connection.UserFromID = ctx.ApplicationUserID

	if err = validator.CreateConnection(connection); err != nil {
		errorHappened(ctx, err.Error(), http.StatusBadRequest, err)
		return
	}

	if connection, err = core.WriteConnection(connection, false); err != nil {
		errorHappened(ctx, err.Error(), http.StatusInternalServerError, err)
		return
	}

	if connection, err = core.ConfirmConnection(connection, true); err != nil {
		errorHappened(ctx, err.Error(), http.StatusInternalServerError, err)
		return
	}

	writeResponse(ctx, connection, http.StatusCreated, 0)
}

// getConnectionList handles requests to list a users connections
// Request: GET account/:AccountID/application/:ApplicationID/user/:UserID/connections
func getConnectionList(ctx *context.Context) {
	var (
		users []*entity.User
		err   error
	)

	if users, err = core.ReadConnectionList(ctx.AccountID, ctx.ApplicationID, ctx.ApplicationUserID); err != nil {
		errorHappened(ctx, err.Error(), http.StatusInternalServerError, err)
		return
	}

	for idx := range users {
		users[idx].Password = ""
	}

	response := &struct {
		ApplicationID int64 `json:"applicationId"`
		Users         []*entity.User
	}{
		ApplicationID: ctx.ApplicationID,
		Users:         users,
	}

	writeResponse(ctx, response, http.StatusOK, 10)
}

// confirmConnection handles requests to confirm a user connection
// Request: POST account/:AccountID/application/:ApplicationID/user/:UserID/connection/confirm
func confirmConnection(ctx *context.Context) {
	var (
		connection = &entity.Connection{}
		userFromID int64
		err        error
	)

	if userFromID, err = strconv.ParseInt(ctx.Vars["userId"], 10, 64); err != nil {
		errorHappened(ctx, "userId is not set or the value is incorrect", http.StatusBadRequest, err)
		return
	}

	if err = json.NewDecoder(ctx.Body).Decode(connection); err != nil {
		errorHappened(ctx, err.Error(), http.StatusBadRequest, err)
		return
	}

	connection.AccountID = ctx.AccountID
	connection.ApplicationID = ctx.ApplicationID
	connection.UserFromID = userFromID

	if err = validator.ConfirmConnection(connection); err != nil {
		errorHappened(ctx, err.Error(), http.StatusBadRequest, err)
		return
	}

	if connection, err = core.ConfirmConnection(connection, false); err != nil {
		errorHappened(ctx, err.Error(), http.StatusInternalServerError, err)
		return
	}

	writeResponse(ctx, connection, http.StatusCreated, 0)
}

// TODO: GET FOLLOWER LIST (followedbyid users)
// TODO: GET FOLLOWING USERS LIST
