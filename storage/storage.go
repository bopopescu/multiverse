/**
 * @author Florin Patan <florinpatan@gmail.com>
 */

// Package storage holds common functions regardless of the storage engine used
package storage

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"time"

	"github.com/tapglue/backend/core/entity"
	. "github.com/tapglue/backend/utils"

	red "gopkg.in/redis.v2"
)

type (
	// Client structure holds the storage engine and functions needed to operate the backend
	Client struct {
		engine *red.Client
	}
)

// Defining keys
const (
	idAccount          = "ids:acs"
	idAccountUser      = "ids:ac:%d:u"
	idAccountApp       = "ids:ac:%d:a"
	idApplicationUser  = "ids:a:%d:u"
	idApplicationEvent = "ids:a:%d:e"

	account = "acc:%d"

	accountUser  = "acc:%d:user:%d"
	accountUsers = "acc:%d:users"

	application  = "acc:%d:app:%d"
	applications = "acc:%d:apps"

	user  = "acc:%d:app:%d:user:%d"
	users = "acc:%d:app:%d:user"

	session = "acc:%d:app:%d:sess:%d"

	connection      = "acc:%d:app:%d:user:%d:connection:%d"
	connections     = "acc:%d:app:%d:user:%d:connections"
	followsUsers    = "acc:%d:app:%d:user:%d:followsUsers"
	followedByUsers = "acc:%d:app:%d:user:%d:follwedByUsers"

	event  = "acc:%d:app:%d:user:%d:event:%d"
	events = "acc:%d:app:%d:user:%d:events"

	connectionEvents     = "acc:%d:app:%d:user:%d:connectionEvents"
	connectionEventsLoop = "%s:connectionEvents"

	alpha1 = "ABCDEFGHIJKLMNOPQRSTUVWXYZ~!@#$%^&*()_+{}:\"|<>?"
	alpha2 = "abcdefghijklmnopqrstuvwxyz0123456789`-=[];'\\,./"
)

var (
	alpha1Len = rand.Intn(len(alpha1))
	alpha2Len = rand.Intn(len(alpha2))
	instance  *Client
)

func generateTokenSalt(size int) string {
	rand.Seed(time.Now().UnixNano())
	salt := ""

	for i := 0; i < size/2; i++ {
		salt += string(alpha1[rand.Intn(alpha1Len)])
		salt += string(alpha2[rand.Intn(alpha2Len)])
	}

	return salt
}

// GenerateAccountID generates a new account ID
func (client *Client) GenerateAccountID() (int64, error) {
	return client.engine.Incr(idAccount).Result()
}

// GenerateAccountUserID generates a new account user id for a specified account
func (client *Client) GenerateAccountUserID(accountID int64) (int64, error) {
	return client.engine.Incr(fmt.Sprintf(idAccountUser, accountID)).Result()
}

// GenerateApplicationID generates a new application ID
func (client *Client) GenerateApplicationID(accountID int64) (int64, error) {
	return client.engine.Incr(fmt.Sprintf(idAccountApp, accountID)).Result()
}

// GenerateApplicationSecretKey returns a token for the specified application of an account
func (client *Client) GenerateApplicationSecretKey(application *entity.Application) (string, error) {
	// Generate a random salt for the token
	keySalt := generateTokenSalt(8)

	// Generate the token itself
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf(
		"%d%d%s%s",
		application.AccountID,
		application.ID,
		keySalt,
		application.CreatedAt.Format(time.RFC3339),
	)))
	token := hasher.Sum(nil)

	return Base64Encode(fmt.Sprintf(
		"%d:%d:%s",
		application.AccountID,
		application.ID,
		string(token),
	)), nil
}

// GenerateApplicationUserID generates the user id in the specified app
func (client *Client) GenerateApplicationUserID(applicationID int64) (int64, error) {
	return client.engine.Incr(fmt.Sprintf(idApplicationUser, applicationID)).Result()
}

// GenerateApplicationEventID generates the event id in the specified app
func (client *Client) GenerateApplicationEventID(applicationID int64) (int64, error) {
	return client.engine.Incr(fmt.Sprintf(idApplicationEvent, applicationID)).Result()
}

// GenerateSessionID generated the session id for the specific
func (client *Client) GenerateSessionID(user *entity.User) string {
	randomToken := generateTokenSalt(16)

	return Base64Encode(fmt.Sprintf(
		"%d:%d:%d:%s:%s",
		user.AccountID,
		user.ApplicationID,
		user.ID,
		time.Now().Format(time.RFC3339),
		randomToken,
	))
}

// Account returns the key for a specified account
func (client *Client) Account(accountID int64) string {
	return fmt.Sprintf(account, accountID)
}

// AccountUser returns the key for a specific user of an account
func (client *Client) AccountUser(accountID, accountUserID int64) string {
	return fmt.Sprintf(accountUser, accountID, accountUserID)
}

// AccountUsers returns the key for account users
func (client *Client) AccountUsers(accountID int64) string {
	return fmt.Sprintf(accountUsers, accountID)
}

// Application returns the key for one account app
func (client *Client) Application(accountID, applicationID int64) string {
	return fmt.Sprintf(application, accountID, applicationID)
}

// Applications returns the key for one account app
func (client *Client) Applications(accountID int64) string {
	return fmt.Sprintf(applications, accountID)
}

// Connection gets the key for the connection
func (client *Client) Connection(accountID, applicationID, userFromID, userToID int64) string {
	return fmt.Sprintf(connection, accountID, applicationID, userFromID, userToID)
}

// Connections gets the key for the connections list
func (client *Client) Connections(accountID, applicationID, userFromID int64) string {
	return fmt.Sprintf(connections, accountID, applicationID, userFromID)
}

// ConnectionUsers gets the key for the connectioned users list
func (client *Client) ConnectionUsers(accountID, applicationID, userFromID int64) string {
	return fmt.Sprintf(followsUsers, accountID, applicationID, userFromID)
}

// FollowedByUsers gets the key for the list of followers
func (client *Client) FollowedByUsers(accountID, applicationID, userToID int64) string {
	return fmt.Sprintf(followedByUsers, accountID, applicationID, userToID)
}

// User gets the key for the user
func (client *Client) User(accountID, applicationID, userID int64) string {
	return fmt.Sprintf(user, accountID, applicationID, userID)
}

// Users gets the key the app users list
func (client *Client) Users(accountID, applicationID int64) string {
	return fmt.Sprintf(users, accountID, applicationID)
}

// Event gets the key for an event
func (client *Client) Event(accountID, applicationID, userID, eventID int64) string {
	return fmt.Sprintf(event, accountID, applicationID, userID, eventID)
}

// Events get the key for the events list
func (client *Client) Events(accountID, applicationID, userID int64) string {
	return fmt.Sprintf(events, accountID, applicationID, userID)
}

// ConnectionEvents get the key for the connections events list
func (client *Client) ConnectionEvents(accountID, applicationID, userID int64) string {
	return fmt.Sprintf(connectionEvents, accountID, applicationID, userID)
}

// ConnectionEventsLoop gets the key for looping through connections
func (client *Client) ConnectionEventsLoop(userID string) string {
	return fmt.Sprintf(connectionEventsLoop, userID)
}

// SessionKey returns the key to be used for a certain session
func (client *Client) SessionKey(accountID, applicationID, userID int64) string {
	return fmt.Sprintf(session, accountID, applicationID, userID)
}

// SessionTimeoutDuration returns how much a session can be alive before it's auto-removed from the system
func (client *Client) SessionTimeoutDuration() time.Duration {
	return time.Duration(time.Hour * 24 * 356 * 10)
}

// Engine returns the storage engine used
func (client *Client) Engine() *red.Client {
	return client.engine
}

// Init initializes the storage package with the required storage engine
func Init(engine *red.Client) *Client {
	if instance == nil {
		instance = &Client{
			engine: engine,
		}
	}

	return instance
}
