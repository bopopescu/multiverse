/**
 * @author Florin Patan <florinpatan@gmail.com>
 */

// Package helper holds common functions regardless of the storage engine used
package helper

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"time"

	"github.com/tapglue/backend/utils"
	"github.com/tapglue/backend/v02/entity"
)

// Defining keys
const (
	IDAccount          = "ids:acs"
	IDAccountUser      = "ids:ac:%d:u"
	IDAccountApp       = "ids:ac:%d:a"
	IDApplicationUser  = "ids:a:%d:u"
	IDApplicationEvent = "ids:a:%d:e"

	account               = "acc:%d"
	accountUserByEmail    = "acc:byemail:%s"
	accountUserByUsername = "acc:byuname:%s"

	accountUser  = "acc:%d:user:%d"
	accountUsers = "acc:%d:users"

	application  = "acc:%d:app:%d"
	applications = "acc:%d:apps"

	applicationUser           = "acc:%d:app:%d:user:%d"
	applicationUsers          = "acc:%d:app:%d:user"
	applicationUserByEmail    = "acc:%d:app:%d:byemail:%s"
	applicationUserByUsername = "acc:%d:app:%d:byuname:%s"

	accountSession     = "acc:%d:sess:%d"
	applicationSession = "acc:%d:app:%d:sess:%d"

	connection       = "acc:%d:app:%d:user:%d:connection:%d"
	connections      = "acc:%d:app:%d:user:%d:connections"
	followsUsers     = "acc:%d:app:%d:user:%d:followsUsers"
	followedByUsers  = "acc:%d:app:%d:user:%d:follwedByUsers"
	socialConnection = "acc:%d:app:%d:social:%s:%s"

	event            = "acc:%d:app:%d:user:%d:event:%d"
	events           = "acc:%d:app:%d:user:%d:events"
	eventGeoKey      = "acc:%d:app:%d:events:geo"
	eventObjectKey   = "acc:%d:app:%d:events:object:%s"
	eventLocationKey = "acc:%d:app:%d:events:location:%s"

	connectionEvents     = "acc:%d:app:%d:user:%d:connectionEvents"
	connectionEventsLoop = "%s:connectionEvents"

	alpha1 = "ABCDEFGHIJKLMNOPQRSTUVWXYZ~!@#$%^&*()_+{}:\"|<>?"
	alpha2 = "abcdefghijklmnopqrstuvwxyz0123456789`-=[];'\\,./"
)

var (
	alpha1Len = rand.Intn(len(alpha1))
	alpha2Len = rand.Intn(len(alpha2))
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

// GenerateAccountSecretKey returns a token for the specified application of an account
func GenerateAccountSecretKey(account *entity.Account) string {
	// Generate a random salt for the token
	keySalt := generateTokenSalt(8)

	// Generate the token itself
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf(
		"%d%s%s",
		account.ID,
		keySalt,
		account.CreatedAt.Format(time.RFC3339),
	)))
	token := hasher.Sum(nil)

	return utils.Base64Encode(fmt.Sprintf(
		"%d:%s",
		account.ID,
		string(token),
	))
}

// GenerateApplicationSecretKey returns a token for the specified application of an account
func GenerateApplicationSecretKey(application *entity.Application) (string, error) {
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

	return utils.Base64Encode(fmt.Sprintf(
		"%d:%d:%s",
		application.AccountID,
		application.ID,
		string(token),
	)), nil
}

// GenerateAccountSessionID generated the session id for the specific
func GenerateAccountSessionID(user *entity.AccountUser) string {
	randomToken := generateTokenSalt(16)

	return utils.Base64Encode(fmt.Sprintf(
		"%d:%d:%s:%s",
		user.AccountID,
		user.ID,
		time.Now().Format(time.RFC3339),
		randomToken,
	))
}

// GenerateApplicationSessionID generated the session id for the specific
func GenerateApplicationSessionID(user *entity.ApplicationUser) string {
	randomToken := generateTokenSalt(16)

	return utils.Base64Encode(fmt.Sprintf(
		"%d:%d:%d:%s:%s",
		user.AccountID,
		user.ApplicationID,
		user.ID,
		time.Now().Format(time.RFC3339),
		randomToken,
	))
}

// GenerateEncryptedPassword generates and encrypted password using the specific salt and time
func GenerateEncryptedPassword(password, salt, time string) string {
	return utils.Base64Encode(
		utils.Sha256String(
			utils.Sha256String(
				utils.Sha256String(password+salt)+
					time) +
				"passwd"),
	)
}

// EncryptPassword will encrypt a string with the password encryption algorithm
func EncryptPassword(password string) string {
	salt := generateTokenSalt(32)
	timestamp := time.Now().Format(time.RFC3339)
	encryptedPassword := GenerateEncryptedPassword(password, salt, timestamp)

	return utils.Base64Encode(fmt.Sprintf("%s:%s:%s", utils.Base64Encode(salt), utils.Base64Encode(timestamp), encryptedPassword))
}

// Account returns the key for a specified account
func Account(accountID int64) string {
	return fmt.Sprintf(account, accountID)
}

// AccountUser returns the key for a specific user of an account
func AccountUser(accountID, accountUserID int64) string {
	return fmt.Sprintf(accountUser, accountID, accountUserID)
}

// AccountUsers returns the key for account users
func AccountUsers(accountID int64) string {
	return fmt.Sprintf(accountUsers, accountID)
}

// AccountUserByEmail returns the key for accounts by email
func AccountUserByEmail(email string) string {
	return fmt.Sprintf(accountUserByEmail, email)
}

// AccountUserByUsername returns the key for accounts by email
func AccountUserByUsername(username string) string {
	return fmt.Sprintf(accountUserByUsername, username)
}

// Application returns the key for one account app
func Application(accountID, applicationID int64) string {
	return fmt.Sprintf(application, accountID, applicationID)
}

// Applications returns the key for one account app
func Applications(accountID int64) string {
	return fmt.Sprintf(applications, accountID)
}

// Connection gets the key for the connection
func Connection(accountID, applicationID, userFromID, userToID int64) string {
	return fmt.Sprintf(connection, accountID, applicationID, userFromID, userToID)
}

// SocialConnection returns the key used to identify a user by the social platform of choice
func SocialConnection(accountID, applicationID int64, platformName, socialID string) string {
	return fmt.Sprintf(socialConnection, accountID, applicationID, platformName, socialID)
}

// Connections gets the key for the connections list
func Connections(accountID, applicationID, userFromID int64) string {
	return fmt.Sprintf(connections, accountID, applicationID, userFromID)
}

// ConnectionUsers gets the key for the connection users list
func ConnectionUsers(accountID, applicationID, userFromID int64) string {
	return fmt.Sprintf(followsUsers, accountID, applicationID, userFromID)
}

// FollowedByUsers gets the key for the list of followers
func FollowedByUsers(accountID, applicationID, userToID int64) string {
	return fmt.Sprintf(followedByUsers, accountID, applicationID, userToID)
}

// User gets the key for the user
func User(accountID, applicationID, userID int64) string {
	return fmt.Sprintf(applicationUser, accountID, applicationID, userID)
}

// ApplicationUserByEmail returns the key for accounts by email
func ApplicationUserByEmail(accountID, applicationID int64, email string) string {
	return fmt.Sprintf(applicationUserByEmail, accountID, applicationID, email)
}

// ApplicationUserByUsername returns the key for accounts by email
func ApplicationUserByUsername(accountID, applicationID int64, username string) string {
	return fmt.Sprintf(applicationUserByUsername, accountID, applicationID, username)
}

// Users gets the key the app users list
func Users(accountID, applicationID int64) string {
	return fmt.Sprintf(applicationUsers, accountID, applicationID)
}

// Event gets the key for an event
func Event(accountID, applicationID, userID, eventID int64) string {
	return fmt.Sprintf(event, accountID, applicationID, userID, eventID)
}

// EventGeoKey gets the key for geo events list
func EventGeoKey(accountID, applicationID int64) string {
	return fmt.Sprintf(eventGeoKey, accountID, applicationID)
}

// EventObjectKey gets the key for geo events list
func EventObjectKey(accountID, applicationID int64, objectKey string) string {
	return fmt.Sprintf(eventObjectKey, accountID, applicationID, objectKey)
}

// EventLocationKey gets the key for geo events list
func EventLocationKey(accountID, applicationID int64, location string) string {
	return fmt.Sprintf(eventLocationKey, accountID, applicationID, location)
}

// Events get the key for the events list
func Events(accountID, applicationID, userID int64) string {
	return fmt.Sprintf(events, accountID, applicationID, userID)
}

// ConnectionEvents get the key for the connections events list
func ConnectionEvents(accountID, applicationID, userID int64) string {
	return fmt.Sprintf(connectionEvents, accountID, applicationID, userID)
}

// ConnectionEventsLoop gets the key for looping through connections
func ConnectionEventsLoop(userID string) string {
	return fmt.Sprintf(connectionEventsLoop, userID)
}

// AccountSessionKey returns the key to be used for a certain session
func AccountSessionKey(accountID, userID int64) string {
	return fmt.Sprintf(accountSession, accountID, userID)
}

// ApplicationSessionKey returns the key to be used for a certain session
func ApplicationSessionKey(accountID, applicationID, userID int64) string {
	return fmt.Sprintf(applicationSession, accountID, applicationID, userID)
}

// SessionTimeoutDuration returns how much a session can be alive before it's auto-removed from the system
func SessionTimeoutDuration() time.Duration {
	return time.Duration(time.Hour * 24 * 356 * 10)
}