/**
 * @author Onur Akpolat <onurakpolat@gmail.com>
 */

package redis

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/tapglue/backend/tgerrors"
	"github.com/tapglue/backend/utils"
	"github.com/tapglue/backend/v02/core"
	"github.com/tapglue/backend/v02/entity"
	storageHelper "github.com/tapglue/backend/v02/storage/helper"
	"github.com/tapglue/backend/v02/storage/redis"

	red "gopkg.in/redis.v2"
)

type (
	connection struct {
		storage *redis.Client
		redis   *red.Client
	}
)

func (c *connection) Create(connection *entity.Connection, retrieve bool) (con *entity.Connection, err tgerrors.TGError) {
	// We confirm the connection in the past forcefully so that we can update it at the confirmation time
	connection.ConfirmedAt = time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC)
	connection.Enabled = false
	connection.CreatedAt = time.Now()
	connection.UpdatedAt = connection.CreatedAt

	val, er := json.Marshal(connection)
	if er != nil {
		return nil, tgerrors.NewInternalError("failed to write the user connection (1)", er.Error())
	}

	key := storageHelper.Connection(connection.AccountID, connection.ApplicationID, connection.UserFromID, connection.UserToID)
	exist, er := c.redis.SetNX(key, string(val)).Result()
	if er != nil {
		return nil, tgerrors.NewInternalError("failed to write the user connection (2)", er.Error())
	}
	if !exist {
		return nil, tgerrors.NewInternalError("failed to write the user connection (3)", "connection does not exist")
	}

	return connection, nil
}

func (c *connection) Read(accountID, applicationID, userFromID, userToID int64) (connection *entity.Connection, err tgerrors.TGError) {
	key := storageHelper.Connection(accountID, applicationID, userFromID, userToID)

	result, er := c.redis.Get(key).Result()
	if er != nil {
		if er.Error() == "redis: nil" {
			return nil, nil
		}
		return nil, tgerrors.NewInternalError("failed to read the connection (1)", er.Error())
	}
	if result == "" {
		return nil, nil
	}

	connection = &entity.Connection{}
	er = json.Unmarshal([]byte(result), connection)
	if er == nil {
		return
	}
	return nil, tgerrors.NewInternalError("failed to read the connection (3)", er.Error())
}

func (c *connection) Update(existingConnection, updatedConnection entity.Connection, retrieve bool) (con *entity.Connection, err tgerrors.TGError) {
	updatedConnection.UpdatedAt = time.Now()
	var er error

	val, er := json.Marshal(updatedConnection)
	if err != nil {
		return nil, tgerrors.NewInternalError("failed to update the connection (1)", er.Error())
	}

	key := storageHelper.Connection(updatedConnection.AccountID, updatedConnection.ApplicationID, updatedConnection.UserFromID, updatedConnection.UserToID)
	exist, er := c.redis.Exists(key).Result()
	if !exist {
		return nil, tgerrors.NewNotFoundError("failed to update teh connection (2)", "connection not found")
	}
	if er != nil {
		return nil, tgerrors.NewInternalError("failed to update the connection (3)", er.Error())
	}

	if er = c.redis.Set(key, string(val)).Err(); er != nil {
		return nil, tgerrors.NewInternalError("failed to update the connection (4)", er.Error())
	}

	if !updatedConnection.Enabled {
		listKey := storageHelper.Connections(updatedConnection.AccountID, updatedConnection.ApplicationID, updatedConnection.UserFromID)
		if er = c.redis.LRem(listKey, 0, key).Err(); er != nil {
			return nil, tgerrors.NewInternalError("failed to update the connection (5)", er.Error())
		}
		userListKey := storageHelper.ConnectionUsers(updatedConnection.AccountID, updatedConnection.ApplicationID, updatedConnection.UserFromID)
		userKey := storageHelper.User(updatedConnection.AccountID, updatedConnection.ApplicationID, updatedConnection.UserToID)
		if er = c.redis.LRem(userListKey, 0, userKey).Err(); er != nil {
			return nil, tgerrors.NewInternalError("failed to update the connection (6)", er.Error())
		}
		followerListKey := storageHelper.FollowedByUsers(updatedConnection.AccountID, updatedConnection.ApplicationID, updatedConnection.UserToID)
		followerKey := storageHelper.User(updatedConnection.AccountID, updatedConnection.ApplicationID, updatedConnection.UserFromID)
		if er = c.redis.LRem(followerListKey, 0, followerKey).Err(); er != nil {
			return nil, tgerrors.NewInternalError("failed to update the connection (7)", er.Error())
		}
	}

	if !retrieve {
		return &updatedConnection, nil
	}

	return c.Read(updatedConnection.AccountID, updatedConnection.ApplicationID, updatedConnection.UserFromID, updatedConnection.UserToID)
}

func (c *connection) Delete(accountID, applicationID, userFromID, userToID int64) (err tgerrors.TGError) {
	key := storageHelper.Connection(accountID, applicationID, userFromID, userToID)
	result, er := c.redis.Del(key).Result()
	if er != nil {
		return tgerrors.NewInternalError("failed to delete the connection (1)", er.Error())
	}

	if result != 1 {
		return tgerrors.NewNotFoundError("failed to delete the connection (2)", "connection not found")
	}

	listKey := storageHelper.Connections(accountID, applicationID, userFromID)
	if er = c.redis.LRem(listKey, 0, key).Err(); er != nil {
		return tgerrors.NewInternalError("failed to delete the connection (3)", er.Error())
	}
	userListKey := storageHelper.ConnectionUsers(accountID, applicationID, userFromID)
	userKey := storageHelper.User(accountID, applicationID, userToID)
	if er = c.redis.LRem(userListKey, 0, userKey).Err(); er != nil {
		return tgerrors.NewInternalError("failed to delete the connection (4)", er.Error())
	}
	followerListKey := storageHelper.FollowedByUsers(accountID, applicationID, userToID)
	followerKey := storageHelper.User(accountID, applicationID, userFromID)
	if er = c.redis.LRem(followerListKey, 0, followerKey).Err(); er != nil {
		return tgerrors.NewInternalError("failed to delete the connection (5)", er.Error())
	}

	if err := c.DeleteEventsFromLists(accountID, applicationID, userFromID, userToID); err != nil {
		return tgerrors.NewInternalError("failed to delete the connection (6)", err.Error())
	}

	return nil
}

func (c *connection) List(accountID, applicationID, userID int64) (users []*entity.ApplicationUser, err tgerrors.TGError) {
	key := storageHelper.ConnectionUsers(accountID, applicationID, userID)
	result, er := c.redis.LRange(key, 0, -1).Result()
	if er != nil {
		return nil, tgerrors.NewInternalError("failed to read the connection list (1)", er.Error())
	}

	if len(result) == 0 {
		return []*entity.ApplicationUser{}, nil
	}

	return c.fetchAndDecodeMultipleUsers(result)
}

func (c *connection) FollowedBy(accountID, applicationID, userID int64) (users []*entity.ApplicationUser, err tgerrors.TGError) {
	key := storageHelper.FollowedByUsers(accountID, applicationID, userID)
	result, er := c.redis.LRange(key, 0, -1).Result()
	if er != nil {
		return nil, tgerrors.NewInternalError("failed to read the followers list (1)", er.Error())
	}

	if len(result) == 0 {
		return []*entity.ApplicationUser{}, nil
	}

	return c.fetchAndDecodeMultipleUsers(result)
}

func (c *connection) Confirm(connection *entity.Connection, retrieve bool) (con *entity.Connection, err tgerrors.TGError) {
	// We confirm the connection in the past forcefully so that we can update it at the confirmation time
	connection.Enabled = true
	connection.ConfirmedAt = time.Now()
	connection.UpdatedAt = connection.ConfirmedAt

	val, er := json.Marshal(connection)
	if er != nil {
		return nil, tgerrors.NewInternalError("failed to confirm the connection (1)", er.Error())
	}

	key := storageHelper.Connection(connection.AccountID, connection.ApplicationID, connection.UserFromID, connection.UserToID)

	cmd := red.NewStringCmd("SET", key, string(val), "XX")
	c.redis.Process(cmd)
	er = cmd.Err()
	if er != nil {
		return nil, tgerrors.NewInternalError("failed to confirm the connection (2)", er.Error())
	}

	listKey := storageHelper.Connections(connection.AccountID, connection.ApplicationID, connection.UserFromID)
	if er = c.redis.LPush(listKey, key).Err(); er != nil {
		return nil, tgerrors.NewInternalError("failed to confirm the connection (3)", er.Error())
	}

	userListKey := storageHelper.ConnectionUsers(connection.AccountID, connection.ApplicationID, connection.UserFromID)
	userKey := storageHelper.User(connection.AccountID, connection.ApplicationID, connection.UserToID)
	if er = c.redis.LPush(userListKey, userKey).Err(); er != nil {
		return nil, tgerrors.NewInternalError("failed to confirm the connection (4)", er.Error())
	}

	followerListKey := storageHelper.FollowedByUsers(connection.AccountID, connection.ApplicationID, connection.UserToID)
	followerKey := storageHelper.User(connection.AccountID, connection.ApplicationID, connection.UserFromID)
	if er = c.redis.LPush(followerListKey, followerKey).Err(); er != nil {
		return nil, tgerrors.NewInternalError("failed to confirm the connection (5)", er.Error())
	}

	if err = c.WriteEventsToList(connection); err != nil {
		return nil, err
	}

	if !retrieve {
		return connection, nil
	}

	return connection, nil
}

func (c *connection) WriteEventsToList(connection *entity.Connection) (err tgerrors.TGError) {
	connectionEventsKey := storageHelper.ConnectionEvents(connection.AccountID, connection.ApplicationID, connection.UserFromID)

	eventsKey := storageHelper.Events(connection.AccountID, connection.ApplicationID, connection.UserToID)

	events, er := c.redis.ZRevRangeWithScores(eventsKey, "0", "-1").Result()
	if er != nil {
		return tgerrors.NewInternalError("failed to write the event to the list", er.Error())
	}

	if len(events) >= 1 {
		var vals []red.Z

		for _, eventKey := range events {
			val := red.Z{Score: float64(eventKey.Score), Member: eventKey.Member}
			vals = append(vals, val)
		}

		if er = c.redis.ZAdd(connectionEventsKey, vals...).Err(); er != nil {
			return tgerrors.NewInternalError("failed to write the event to the list", er.Error())
		}
	}

	return
}

func (c *connection) DeleteEventsFromLists(accountID, applicationID, userFromID, userToID int64) (err tgerrors.TGError) {
	connectionEventsKey := storageHelper.ConnectionEvents(accountID, applicationID, userFromID)

	eventsKey := storageHelper.Events(accountID, applicationID, userToID)

	events, er := c.redis.ZRevRangeWithScores(eventsKey, "0", "-1").Result()
	if er != nil {
		return tgerrors.NewInternalError("failed to delete the event from connections (1)", er.Error())
	}

	if len(events) >= 1 {
		var members []string

		for _, eventKey := range events {
			member := eventKey.Member
			members = append(members, member)
		}

		if er = c.redis.ZRem(connectionEventsKey, members...).Err(); er != nil {
			return tgerrors.NewInternalError("failed to delete the event from the connections (2)", er.Error())
		}
	}

	return nil
}

func (c *connection) SocialConnect(user *entity.ApplicationUser, platform string, socialFriendsIDs []string) ([]*entity.ApplicationUser, tgerrors.TGError) {
	result := []*entity.ApplicationUser{}

	encodedSocialFriendsIDs := []string{}
	for idx := range socialFriendsIDs {
		encodedSocialFriendsIDs = append(encodedSocialFriendsIDs, storageHelper.SocialConnection(
			user.AccountID,
			user.ApplicationID,
			platform,
			utils.Base64Encode(socialFriendsIDs[idx])))
	}

	ourStoredUsersIDs, er := c.redis.MGet(encodedSocialFriendsIDs...).Result()
	if er != nil {
		return result, tgerrors.NewInternalError("social connection failed (1)", er.Error())
	}

	if len(ourStoredUsersIDs) == 0 {
		return result, nil
	}

	return c.AutoConnectSocialFriends(user, ourStoredUsersIDs)
}

func (c *connection) AutoConnectSocialFriends(user *entity.ApplicationUser, ourStoredUsersIDs []interface{}) (users []*entity.ApplicationUser, err tgerrors.TGError) {
	ourUserKeys := []string{}
	for idx := range ourStoredUsersIDs {
		userID, err := strconv.ParseInt(ourStoredUsersIDs[idx].(string), 10, 64)
		if err != nil {
			continue
		}

		key := storageHelper.Connection(user.AccountID, user.ApplicationID, user.ID, userID)
		if exists, err := c.redis.Exists(key).Result(); exists || err != nil {
			// We don't want to update existing connections as we don't know if the user disabled them willingly or not
			// TODO Figure out if this is the right thing to do
			continue
		}

		connection := &entity.Connection{
			AccountID:     user.AccountID,
			ApplicationID: user.ApplicationID,
			UserFromID:    user.ID,
			UserToID:      userID,
		}

		_, er := c.Create(connection, false)
		if er != nil {
			continue
		}

		_, er = c.Confirm(connection, false)
		if er != nil {
			continue
		}

		connection = &entity.Connection{
			AccountID:     user.AccountID,
			ApplicationID: user.ApplicationID,
			UserFromID:    userID,
			UserToID:      user.ID,
		}

		_, er = c.Create(connection, false)
		if er != nil {
			continue
		}

		_, er = c.Confirm(connection, false)
		if er != nil {
			continue
		}

		ourUserKeys = append(
			ourUserKeys,
			storageHelper.User(user.AccountID, user.ApplicationID, userID),
		)
	}

	return c.fetchAndDecodeMultipleUsers(ourUserKeys)
}

func (c *connection) fetchAndDecodeMultipleUsers(keys []string) (users []*entity.ApplicationUser, err tgerrors.TGError) {
	if len(keys) == 0 {
		return []*entity.ApplicationUser{}, nil
	}

	resultList, er := c.redis.MGet(keys...).Result()
	if er != nil {
		return nil, tgerrors.NewInternalError("failed to perform operation on user list (1)", er.Error())
	}

	user := &entity.ApplicationUser{}
	for _, result := range resultList {
		if er = json.Unmarshal([]byte(result.(string)), user); er != nil {
			return nil, tgerrors.NewInternalError("failed to perform operation on user list (2)", er.Error())
		}
		users = append(users, user)
		user = &entity.ApplicationUser{}
	}

	return
}

// NewConnection creates a new Connection
func NewConnection(storageClient *redis.Client) core.Connection {
	return &connection{
		storage: storageClient,
		redis:   storageClient.Datastore(),
	}
}
