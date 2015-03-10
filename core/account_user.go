/**
 * @author Onur Akpolat <onurakpolat@gmail.com>
 */

package core

import (
	"encoding/json"
	"fmt"
	"time"

	"strconv"

	"github.com/tapglue/backend/core/entity"
	"github.com/tapglue/backend/utils"
)

// ReadAccountUser returns the account matching the ID or an error
func ReadAccountUser(accountID, accountUserID int64) (accountUser *entity.AccountUser, err error) {
	result, err := storageEngine.Get(storageClient.AccountUser(accountID, accountUserID)).Result()
	if err != nil {
		return nil, err
	}

	// Parse JSON
	if err = json.Unmarshal([]byte(result), &accountUser); err != nil {
		return nil, err
	}

	return
}

// UpdateAccountUser update an account user in the database and returns the updated account user or an error
func UpdateAccountUser(accountUser *entity.AccountUser, retrieve bool) (accUser *entity.AccountUser, err error) {
	accountUser.UpdatedAt = time.Now()

	val, err := json.Marshal(accountUser)
	if err != nil {
		return nil, err
	}

	key := storageClient.AccountUser(accountUser.AccountID, accountUser.ID)
	exist, err := storageEngine.Exists(key).Result()
	if !exist {
		return nil, fmt.Errorf("account user does not exist")
	}
	if err != nil {
		return nil, err
	}

	if err = storageEngine.Set(key, string(val)).Err(); err != nil {
		return nil, err
	}

	if !accountUser.Enabled {
		listKey := storageClient.AccountUsers(accountUser.AccountID)
		if err = storageEngine.LRem(listKey, 0, key).Err(); err != nil {
			return nil, err
		}
	}

	if !retrieve {
		return accountUser, nil
	}

	return ReadAccountUser(accountUser.AccountID, accountUser.ID)
}

// DeleteAccountUser deletes the account user matching the IDs or an error
func DeleteAccountUser(accountID, userID int64) (err error) {
	// TODO: Make not deletable if its the only account user of an account
	key := storageClient.AccountUser(accountID, userID)
	result, err := storageEngine.Del(key).Result()
	if err != nil {
		return err
	}

	if result != 1 {
		return fmt.Errorf("The resource for the provided id doesn't exist")
	}

	listKey := storageClient.AccountUsers(accountID)
	if err = storageEngine.LRem(listKey, 0, key).Err(); err != nil {
		return err
	}

	return nil
}

// ReadAccountUserList returns all the users from a certain account
func ReadAccountUserList(accountID int64) (accountUsers []*entity.AccountUser, err error) {
	result, err := storageEngine.LRange(storageClient.Account(accountID), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	resultList, err := storageEngine.MGet(result...).Result()
	if err != nil {
		return nil, err
	}

	accountUser := &entity.AccountUser{}
	for _, result := range resultList {
		if err = json.Unmarshal([]byte(result.(string)), accountUser); err != nil {
			return nil, err
		}
		accountUsers = append(accountUsers, accountUser)
		accountUser = &entity.AccountUser{}
	}

	return
}

// WriteAccountUser adds a new account user to the database and returns the created account user or an error
func WriteAccountUser(accountUser *entity.AccountUser, retrieve bool) (accUser *entity.AccountUser, err error) {
	if accountUser.ID, err = storageClient.GenerateAccountUserID(accountUser.AccountID); err != nil {
		return nil, err
	}

	accountUser.Enabled = true
	accountUser.CreatedAt = time.Now()
	accountUser.UpdatedAt, accountUser.ReceivedAt = accountUser.CreatedAt, accountUser.CreatedAt

	// Encrypt password
	accountUser.Password = storageClient.EncryptPassword(accountUser.Password)

	val, err := json.Marshal(accountUser)
	if err != nil {
		return nil, err
	}

	key := storageClient.AccountUser(accountUser.AccountID, accountUser.ID)
	exist, err := storageEngine.SetNX(key, string(val)).Result()
	if !exist {
		return nil, fmt.Errorf("account user does not exists")
	}
	if err != nil {
		return nil, err
	}

	idListKey := storageClient.AccountUsers(accountUser.AccountID)
	if err = storageEngine.LPush(idListKey, key).Err(); err != nil {
		return nil, err
	}

	emailListKey := storageClient.AccountUserByEmail(utils.Base64Encode(accountUser.Email))
	err = storageEngine.HMSet(
		emailListKey,
		"acc", fmt.Sprintf("%d", accountUser.AccountID),
		"usr", fmt.Sprintf("%d", accountUser.ID),
	).Err()

	if err != nil {
		return nil, err
	}

	if !retrieve {
		return accountUser, nil
	}

	return ReadAccountUser(accountUser.AccountID, accountUser.ID)
}

// CreateUserSession handles the creation of a user session and returns the session token
func CreateAccountUserSession(user *entity.AccountUser) (string, error) {
	// TODO support multiple sessions?
	// TODO rate limit this to x / per day?
	// TODO rate limit this to be at least x minutes after the logout
	// TODO do we customize the key session timeout per app

	sessionKey := storageClient.AccountSessionKey(user.AccountID, user.ID)
	token := storageClient.GenerateAccountSessionID(user)

	_, err := storageEngine.Set(sessionKey, token).Result()
	if err != nil {
		return "", err
	}

	expired, err := storageEngine.Expire(sessionKey, storageClient.SessionTimeoutDuration()).Result()
	if err != nil {
		return "", err
	}
	if !expired {
		return "", fmt.Errorf("could not set expire time")
	}

	return token, nil
}

// RefreshUserSession generates a new session token for the user session
func RefreshAccountUserSession(sessionToken string, user *entity.AccountUser) (string, error) {
	// TODO support multiple sessions?
	// TODO rate limit this to x / per day?
	// TODO rate limit this to be at least x minutes after the logout
	// TODO do we customize the key session timeout per app

	sessionKey := storageClient.AccountSessionKey(user.AccountID, user.ID)

	storedToken, err := storageEngine.Get(sessionKey).Result()
	if err != nil {
		return "", err
	}

	if storedToken != sessionToken {
		return "", fmt.Errorf("session token mismatch")
	}

	token := storageClient.GenerateAccountSessionID(user)

	if err := storageEngine.Set(sessionKey, token).Err(); err != nil {
		return "", err
	}

	expired, err := storageEngine.Expire(sessionKey, storageClient.SessionTimeoutDuration()).Result()
	if err != nil {
		return "", err
	}
	if !expired {
		return "", fmt.Errorf("could not set expire time")
	}

	return token, nil
}

// DestroyUserSession removes the user session
func DestroyAccountUserSession(sessionToken string, user *entity.AccountUser) error {
	// TODO support multiple sessions?
	// TODO rate limit this to x / per day?
	sessionKey := storageClient.AccountSessionKey(user.AccountID, user.ID)

	storedToken, err := storageEngine.Get(sessionKey).Result()
	if err != nil {
		return err
	}

	if storedToken != sessionToken {
		return fmt.Errorf("session token mismatch")
	}

	result, err := storageEngine.Del(sessionKey).Result()
	if err != nil {
		return err
	}

	if result != 1 {
		return fmt.Errorf("invalid session")
	}

	return nil
}

// FindAccountAndUserByEmail returns the account and account user for a certain e-mail address
func FindAccountAndUserByEmail(email string) (*entity.Account, *entity.AccountUser, error) {
	emailListKey := storageClient.AccountUserByEmail(utils.Base64Encode(email))

	details, err := storageEngine.HMGet(emailListKey, "acc", "usr").Result()
	if err != nil {
		return nil, nil, err
	}

	accountID, err := strconv.ParseInt(details[0].(string), 10, 64)
	if err != nil {
		return nil, nil, err
	}

	userID, err := strconv.ParseInt(details[1].(string), 10, 64)
	if err != nil {
		return nil, nil, err
	}

	account, err := ReadAccount(accountID)
	if err != nil {
		return nil, nil, err
	}

	accountUser, err := ReadAccountUser(accountID, userID)
	if err != nil {
		return nil, nil, err
	}

	return account, accountUser, nil
}
