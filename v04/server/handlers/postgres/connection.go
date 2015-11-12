package postgres

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/tapglue/multiverse/errors"
	"github.com/tapglue/multiverse/tgflake"
	"github.com/tapglue/multiverse/v04/context"
	"github.com/tapglue/multiverse/v04/core"
	"github.com/tapglue/multiverse/v04/entity"
	"github.com/tapglue/multiverse/v04/errmsg"
	"github.com/tapglue/multiverse/v04/server/handlers"
	"github.com/tapglue/multiverse/v04/server/response"
	"github.com/tapglue/multiverse/v04/validator"
)

type connection struct {
	appUser core.ApplicationUser
	storage core.Connection
	event   core.Event
}

func (conn *connection) Update(ctx *context.Context) (err []errors.Error) {
	userFromID := ctx.ApplicationUserID

	accountID := ctx.OrganizationID
	applicationID := ctx.ApplicationID

	userToID, er := strconv.ParseUint(ctx.Vars["userToID"], 10, 64)
	if er != nil {
		return []errors.Error{errmsg.ErrApplicationUserIDInvalid.SetCurrentLocation()}
	}

	var connection entity.Connection
	if er := json.Unmarshal(ctx.Body, &connection); er != nil {
		return []errors.Error{errmsg.ErrServerReqBadJSONReceived.UpdateMessage(er.Error()).SetCurrentLocation()}
	}

	connection.UserFromID = userFromID
	connection.UserToID = userToID
	if !connection.IsValidType() {
		return []errors.Error{errmsg.ErrConnectionTypeIsWrong}
	}

	existingConnection, err := conn.storage.Read(accountID, applicationID, userFromID, userToID, connection.Type)
	if err != nil {
		return
	}
	if existingConnection == nil {
		return []errors.Error{errmsg.ErrConnectionUsersNotConnected.SetCurrentLocation()}
	}

	err = validator.UpdateConnection(conn.appUser, accountID, applicationID, existingConnection, &connection)
	if err != nil {
		return
	}

	updatedConnection, err := conn.storage.Update(accountID, applicationID, *existingConnection, connection, false)
	if err != nil {
		return
	}

	response.WriteResponse(ctx, updatedConnection, http.StatusCreated, 0)
	return
}

func (conn *connection) Delete(ctx *context.Context) (err []errors.Error) {
	accountID := ctx.OrganizationID
	applicationID := ctx.ApplicationID
	userFromID := ctx.ApplicationUserID

	userToID, er := strconv.ParseUint(ctx.Vars["applicationUserToID"], 10, 64)
	if er != nil {
		return []errors.Error{errmsg.ErrApplicationUserIDInvalid.SetCurrentLocation()}
	}

	connectionType := entity.ConnectionTypeType(ctx.Vars["connectionType"])
	if connectionType != entity.ConnectionTypeFollow &&
		connectionType != entity.ConnectionTypeFriend {
		return []errors.Error{errmsg.ErrConnectionTypeIsWrong.UpdateInternalMessage("got connection type: " + string(connectionType)).SetCurrentLocation()}
	}

	exists, err := conn.storage.Exists(accountID, applicationID, userFromID, userToID, connectionType)
	if err != nil {
		return
	}

	if !exists {
		return []errors.Error{errmsg.ErrConnectionNotFound.SetCurrentLocation()}
	}

	err = conn.storage.Delete(accountID, applicationID, userFromID, userToID, connectionType)
	if err != nil {
		return
	}

	response.WriteResponse(ctx, "", http.StatusNoContent, 10)
	return
}

func (conn *connection) Create(ctx *context.Context) (err []errors.Error) {
	var (
		connection = &entity.Connection{}
		er         error
	)

	if er = json.Unmarshal(ctx.Body, connection); er != nil {
		return []errors.Error{errmsg.ErrServerReqBadJSONReceived.UpdateMessage(er.Error()).SetCurrentLocation()}
	}

	connection.UserFromID = ctx.ApplicationUserID

	return conn.doCreateConnection(ctx, connection)
}

func (conn *connection) FollowingList(ctx *context.Context) (err []errors.Error) {
	accountID := ctx.OrganizationID
	applicationID := ctx.ApplicationID
	userID, er := strconv.ParseUint(ctx.Vars["applicationUserID"], 10, 64)
	if er != nil {
		return []errors.Error{errmsg.ErrApplicationUserIDInvalid.SetCurrentLocation()}
	}

	exists, err := conn.appUser.ExistsByID(accountID, applicationID, userID)
	if err != nil {
		return
	}

	if !exists {
		return []errors.Error{errmsg.ErrApplicationUserNotFound.SetCurrentLocation()}
	}

	userIDs, err := conn.storage.Following(accountID, applicationID, userID)
	if err != nil {
		return
	}

	users, err := conn.appUser.ReadMultiple(accountID, applicationID, userIDs)
	if err != nil {
		return
	}

	response.SanitizeApplicationUsers(users)

	resp := struct {
		Users      []*entity.ApplicationUser `json:"users"`
		UsersCount int                       `json:"users_count"`
	}{
		Users:      users,
		UsersCount: len(users),
	}

	status := http.StatusOK

	if resp.UsersCount == 0 {
		status = http.StatusNoContent
	}

	response.WriteResponse(ctx, resp, status, 10)
	return
}

func (conn *connection) CurrentUserFollowingList(ctx *context.Context) (err []errors.Error) {
	userIDs, err := conn.storage.Following(ctx.OrganizationID, ctx.ApplicationID, ctx.ApplicationUserID)
	if err != nil {
		return
	}

	users, err := conn.appUser.ReadMultiple(ctx.OrganizationID, ctx.ApplicationID, userIDs)
	if err != nil {
		return
	}

	response.SanitizeApplicationUsers(users)

	resp := struct {
		Users      []*entity.ApplicationUser `json:"users"`
		UsersCount int                       `json:"users_count"`
	}{
		Users:      users,
		UsersCount: len(users),
	}

	status := http.StatusOK
	if resp.UsersCount == 0 {
		status = http.StatusNoContent
	}

	response.WriteResponse(ctx, resp, status, 10)
	return
}

func (conn *connection) FollowedByList(ctx *context.Context) (err []errors.Error) {
	userID, er := strconv.ParseUint(ctx.Vars["applicationUserID"], 10, 64)
	if er != nil {
		return []errors.Error{errmsg.ErrApplicationUserIDInvalid.SetCurrentLocation()}
	}

	exists, err := conn.appUser.ExistsByID(ctx.OrganizationID, ctx.ApplicationID, userID)
	if err != nil {
		return
	}

	if !exists {
		return []errors.Error{errmsg.ErrApplicationUserNotFound.SetCurrentLocation()}
	}

	userIDs, err := conn.storage.FollowedBy(ctx.OrganizationID, ctx.ApplicationID, userID)
	if err != nil {
		return
	}

	users, err := conn.appUser.ReadMultiple(ctx.OrganizationID, ctx.ApplicationID, userIDs)
	if err != nil {
		return
	}

	response.SanitizeApplicationUsers(users)

	resp := struct {
		Users      []*entity.ApplicationUser `json:"users"`
		UsersCount int                       `json:"users_count"`
	}{
		Users:      users,
		UsersCount: len(users),
	}

	status := http.StatusOK
	if resp.UsersCount == 0 {
		status = http.StatusNoContent
	}

	response.WriteResponse(ctx, resp, status, 10)
	return
}

func (conn *connection) CurrentUserFollowedByList(ctx *context.Context) (err []errors.Error) {
	userIDs, err := conn.storage.FollowedBy(ctx.OrganizationID, ctx.ApplicationID, ctx.ApplicationUserID)
	if err != nil {
		return
	}

	users, err := conn.appUser.ReadMultiple(ctx.OrganizationID, ctx.ApplicationID, userIDs)
	if err != nil {
		return
	}

	response.SanitizeApplicationUsers(users)

	resp := struct {
		Users      []*entity.ApplicationUser `json:"users"`
		UsersCount int                       `json:"users_count"`
	}{
		Users:      users,
		UsersCount: len(users),
	}

	status := http.StatusOK

	if resp.UsersCount == 0 {
		status = http.StatusNoContent
	}

	response.WriteResponse(ctx, resp, status, 10)
	return
}

func (conn *connection) CreateSocial(ctx *context.Context) (err []errors.Error) {
	request := entity.CreateSocialConnectionRequest{}

	if er := json.Unmarshal(ctx.Body, &request); er != nil {
		return []errors.Error{errmsg.ErrServerReqBadJSONReceived.UpdateMessage(er.Error()).SetCurrentLocation()}
	}

	if request.ConnectionType != entity.ConnectionTypeFriend && request.ConnectionType != entity.ConnectionTypeFollow {
		return []errors.Error{errmsg.ErrConnectionTypeIsWrong.SetCurrentLocation()}
	}

	user := ctx.ApplicationUser

	if _, ok := user.SocialIDs[request.SocialPlatform]; !ok {
		if len(user.SocialIDs[request.SocialPlatform]) == 0 {
			user.SocialIDs = map[string]string{}
		}
		user.SocialIDs[request.SocialPlatform] = request.PlatformUserID
		_, err = conn.appUser.Update(
			ctx.OrganizationID,
			ctx.ApplicationID,
			*user,
			*user,
			false)
		if err != nil {
			return err
		}
	}

	if request.ConnectionState == "" {
		request.ConnectionState = entity.ConnectionStateConfirmed
	}

	userIDs, err := conn.storage.SocialConnect(
		ctx.OrganizationID,
		ctx.ApplicationID,
		user,
		request.SocialPlatform,
		request.ConnectionsIDs,
		request.ConnectionType,
		request.ConnectionState)
	if err != nil {
		return
	}

	users, err := conn.appUser.ReadMultiple(ctx.OrganizationID, ctx.ApplicationID, userIDs)
	if err != nil {
		return
	}

	response.SanitizeApplicationUsers(users)

	resp := struct {
		Users      []*entity.ApplicationUser `json:"users"`
		UsersCount int                       `json:"users_count"`
	}{
		Users:      users,
		UsersCount: len(users),
	}

	response.WriteResponse(ctx, resp, http.StatusCreated, 10)
	return
}

func (conn *connection) Friends(ctx *context.Context) (err []errors.Error) {
	accountID := ctx.OrganizationID
	applicationID := ctx.ApplicationID
	userID, er := strconv.ParseUint(ctx.Vars["applicationUserID"], 10, 64)
	if er != nil {
		return []errors.Error{errmsg.ErrApplicationUserIDInvalid.SetCurrentLocation()}
	}

	exists, err := conn.appUser.ExistsByID(accountID, applicationID, userID)
	if err != nil {
		return
	}

	if !exists {
		return []errors.Error{errmsg.ErrApplicationUserNotFound.SetCurrentLocation()}
	}

	userIDs, err := conn.storage.Friends(accountID, applicationID, userID)
	if err != nil {
		return
	}

	users, err := conn.appUser.ReadMultiple(ctx.OrganizationID, ctx.ApplicationID, userIDs)
	if err != nil {
		return
	}

	response.SanitizeApplicationUsers(users)

	resp := struct {
		Users      []*entity.ApplicationUser `json:"users"`
		UsersCount int                       `json:"users_count"`
	}{
		Users:      users,
		UsersCount: len(users),
	}

	status := http.StatusOK
	if resp.UsersCount == 0 {
		status = http.StatusNoContent
	}

	response.WriteResponse(ctx, resp, status, 10)
	return
}

func (conn *connection) CurrentUserFriends(ctx *context.Context) (err []errors.Error) {
	userIDs, err := conn.storage.Friends(ctx.OrganizationID, ctx.ApplicationID, ctx.ApplicationUserID)
	if err != nil {
		return
	}

	users, err := conn.appUser.ReadMultiple(ctx.OrganizationID, ctx.ApplicationID, userIDs)
	if err != nil {
		return
	}

	response.SanitizeApplicationUsers(users)

	resp := struct {
		Users      []*entity.ApplicationUser `json:"users"`
		UsersCount int                       `json:"users_count"`
	}{
		Users:      users,
		UsersCount: len(users),
	}

	status := http.StatusOK
	if resp.UsersCount == 0 {
		status = http.StatusNoContent
	}

	response.WriteResponse(ctx, resp, status, 10)
	return
}

func (conn *connection) CreateFriend(ctx *context.Context) []errors.Error {
	var (
		connection = &entity.Connection{}
		er         error
	)

	if er = json.Unmarshal(ctx.Body, connection); er != nil {
		return []errors.Error{errmsg.ErrServerReqBadJSONReceived.UpdateMessage(er.Error()).SetCurrentLocation()}
	}

	connection.Type = entity.ConnectionTypeFriend
	connection.UserFromID = ctx.ApplicationUserID
	return conn.doCreateConnection(ctx, connection)
}

func (conn *connection) CreateFollow(ctx *context.Context) []errors.Error {
	var (
		connection = &entity.Connection{}
		er         error
	)

	if er = json.Unmarshal(ctx.Body, connection); er != nil {
		return []errors.Error{errmsg.ErrServerReqBadJSONReceived.UpdateMessage(er.Error()).SetCurrentLocation()}
	}

	connection.Type = entity.ConnectionTypeFollow
	connection.UserFromID = ctx.ApplicationUserID
	return conn.doCreateConnection(ctx, connection)
}

func (conn *connection) doCreateConnection(ctx *context.Context, connection *entity.Connection) []errors.Error {
	existingConnection, err := conn.storage.Read(ctx.OrganizationID, ctx.ApplicationID, ctx.ApplicationUserID, connection.UserToID, connection.Type)
	if err != nil {
		if err[0].Code() != errmsg.ErrConnectionNotFound.Code() {
			return err
		}
	}

	if existingConnection != nil {
		if (connection.State == "" && existingConnection.State == entity.ConnectionStateConfirmed) ||
			(connection.State == existingConnection.State) {
			goto createResponse
		} else {
			if err := existingConnection.TransferState(connection.State, ctx.ApplicationUserID); err != nil {
				return err
			}
		}

		_, err = conn.storage.Update(ctx.OrganizationID, ctx.ApplicationID, *existingConnection, *existingConnection, false)
		if err != nil {
			return err
		}
	} else {
		if connection.State == "" {
			connection.TransferState(entity.ConnectionStateConfirmed, ctx.ApplicationUserID)
		}

		err = validator.CreateConnection(conn.appUser, ctx.OrganizationID, ctx.ApplicationID, connection)
		if err != nil {
			return err
		}

		err = conn.storage.Create(ctx.OrganizationID, ctx.ApplicationID, connection)
		if err != nil {
			return err
		}
	}

createResponse:
	response.WriteResponse(ctx, connection, http.StatusCreated, 0)
	return nil
}

func (conn *connection) CreateAutoConnectionEvent(ctx *context.Context, connection *entity.Connection) (*entity.Event, []errors.Error) {
	event := &entity.Event{
		UserID:     connection.UserFromID,
		Type:       "tg_" + string(connection.Type),
		Visibility: entity.EventPrivate,
		Target: &entity.Object{
			ID:   strconv.FormatUint(connection.UserToID, 10),
			Type: "tg_user",
		},
	}

	var err error
	event.ID, err = tgflake.FlakeNextID(ctx.ApplicationID, "events")
	if err != nil {
		return nil, []errors.Error{errmsg.ErrServerInternalError.UpdateInternalMessage(err.Error()).SetCurrentLocation()}
	}

	accountID := ctx.OrganizationID
	applicationID := ctx.ApplicationID

	er := conn.event.Create(accountID, applicationID, connection.UserFromID, event)
	return event, er
}

func (conn *connection) CreateAutoConnectionEvents(
	ctx *context.Context,
	user *entity.ApplicationUser, users []*entity.ApplicationUser,
	connectionType entity.ConnectionTypeType) ([]*entity.Event, []errors.Error) {

	events := []*entity.Event{}
	errs := []errors.Error{}
	for idx := range users {
		connection := &entity.Connection{
			UserFromID: user.ID,
			UserToID:   users[idx].ID,
			Type:       connectionType,
		}

		evt, err := conn.CreateAutoConnectionEvent(ctx, connection)

		events = append(events, evt)
		errs = append(errs, err...)
	}

	return events, errs
}

func (conn *connection) CurrentUserConnectionsByState(ctx *context.Context) []errors.Error {
	userID := ctx.ApplicationUserID
	connectionState := entity.ConnectionStateType(ctx.Vars["connectionState"])

	return conn.doGetUserConnectionsByState(ctx, userID, connectionState)
}

func (conn *connection) UserConnectionsByState(ctx *context.Context) []errors.Error {
	connectionState := entity.ConnectionStateType(ctx.Vars["connectionState"])

	userID, er := strconv.ParseUint(ctx.Vars["applicationUserID"], 10, 64)
	if er != nil {
		return []errors.Error{errmsg.ErrApplicationUserIDInvalid.SetCurrentLocation()}
	}

	exists, err := conn.appUser.ExistsByID(ctx.OrganizationID, ctx.ApplicationID, userID)
	if err != nil {
		return err
	}

	if !exists {
		return []errors.Error{errmsg.ErrApplicationUserNotFound.SetCurrentLocation()}
	}

	return conn.doGetUserConnectionsByState(ctx, userID, connectionState)
}

func (conn *connection) doGetUserConnectionsByState(ctx *context.Context, userID uint64, connectionState entity.ConnectionStateType) []errors.Error {
	orgID := ctx.OrganizationID
	appID := ctx.ApplicationID

	if !entity.IsValidConectionState(connectionState) {
		return []errors.Error{errmsg.ErrConnectionStateInvalid.SetCurrentLocation()}
	}

	connections, err := conn.storage.ConnectionsByState(orgID, appID, userID, connectionState)
	if err != nil {
		return err
	}

	incomingConnections := []*entity.Connection{}
	outgoingConnections := []*entity.Connection{}
	userIDs := []uint64{}
	for idx := range connections {
		connections[idx].Enabled = entity.PFalse
		if idx > 0 {
			if connections[idx-1].UserToID == connections[idx].UserFromID {
				continue
			}
		}

		if connections[idx].UserFromID == userID {
			userIDs = append(userIDs, connections[idx].UserToID)
			outgoingConnections = append(outgoingConnections, connections[idx])
		} else {
			userIDs = append(userIDs, connections[idx].UserFromID)
			incomingConnections = append(incomingConnections, connections[idx])
		}
	}

	users, err := conn.appUser.ReadMultiple(orgID, appID, userIDs)
	if err != nil {
		return err
	}

	response.SanitizeApplicationUsers(users)

	resp := entity.ConnectionsByStateResponse{
		IncomingConnections: incomingConnections,
		OutgoingConnections: outgoingConnections,
		Users:               users,
		IncomingConnectionsCount: len(incomingConnections),
		OutgoingConnectionsCount: len(outgoingConnections),
		UsersCount:               len(users),
	}

	status := http.StatusOK
	if resp.UsersCount == 0 {
		status = http.StatusNoContent
	}

	response.WriteResponse(ctx, resp, status, 10)
	return nil
}

// NewConnection initializes a new connection with an application user
func NewConnection(storage core.Connection, appUser core.ApplicationUser, evt core.Event) handlers.Connection {
	return &connection{
		storage: storage,
		appUser: appUser,
		event:   evt,
	}
}
