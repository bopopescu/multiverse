/**
 * @author Onur Akpolat <onurakpolat@gmail.com>
 */

package core

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	. "github.com/tapglue/backend/utils"
	"github.com/tapglue/backend/v01/entity"
)

// ReadApplication returns the application matching the ID or an error
func ReadApplication(accountID, applicationID int64) (application *entity.Application, err error) {
	result, err := storageEngine.Get(storageClient.Application(accountID, applicationID)).Result()
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal([]byte(result), &application); err != nil {
		return nil, err
	}

	return
}

// UpdateApplication updates an application in the database and returns the created applicaton user or an error
func UpdateApplication(existingApplication, updatedApplication entity.Application, retrieve bool) (app *entity.Application, err error) {
	updatedApplication.UpdatedAt = time.Now()

	val, err := json.Marshal(updatedApplication)
	if err != nil {
		return nil, err
	}

	key := storageClient.Application(updatedApplication.AccountID, updatedApplication.ID)
	exist, err := storageEngine.Exists(key).Result()
	if !exist {
		return nil, fmt.Errorf("application does not exist")
	}
	if err != nil {
		return nil, err
	}

	if err = storageEngine.Set(key, string(val)).Err(); err != nil {
		return nil, err
	}

	if !updatedApplication.Enabled {
		listKey := storageClient.Applications(updatedApplication.AccountID)
		if err = storageEngine.LRem(listKey, 0, key).Err(); err != nil {
			return nil, err
		}
	}

	if !retrieve {
		return &updatedApplication, nil
	}

	return ReadApplication(updatedApplication.AccountID, updatedApplication.ID)
}

// DeleteApplication deletes the application matching the IDs or an error
func DeleteApplication(accountID, applicationID int64) (err error) {
	// TODO: Disable application users?
	// TODO: User connections?
	// TODO: Application lists?
	// TODO: Application events?

	key := storageClient.Application(accountID, applicationID)
	result, err := storageEngine.Del(key).Result()
	if err != nil {
		return err
	}

	if result != 1 {
		return fmt.Errorf("The resource for the provided id doesn't exist")
	}

	listKey := storageClient.Applications(accountID)
	if err = storageEngine.LRem(listKey, 0, key).Err(); err != nil {
		return err
	}

	return nil
}

// ReadApplicationList returns all applications from a certain account
func ReadApplicationList(accountID int64) (applications []*entity.Application, err error) {
	key := storageClient.Applications(accountID)

	result, err := storageEngine.LRange(key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return
	}

	resultList, err := storageEngine.MGet(result...).Result()
	if err != nil {
		return nil, err
	}

	application := &entity.Application{}
	for _, result := range resultList {
		if err = json.Unmarshal([]byte(result.(string)), application); err != nil {
			return nil, err
		}
		applications = append(applications, application)
		application = &entity.Application{}
	}

	return
}

// WriteApplication adds an application to the database and returns the created applicaton user or an error
func WriteApplication(application *entity.Application, retrieve bool) (app *entity.Application, err error) {
	if application.ID, err = storageClient.GenerateApplicationID(application.AccountID); err != nil {
		return nil, err
	}

	application.Enabled = true
	application.CreatedAt = time.Now()
	application.UpdatedAt = application.CreatedAt

	if application.AuthToken, err = storageClient.GenerateApplicationSecretKey(application); err != nil {
		return nil, err
	}

	val, err := json.Marshal(application)
	if err != nil {
		return nil, err
	}

	key := storageClient.Application(application.AccountID, application.ID)

	exist, err := storageEngine.SetNX(key, string(val)).Result()
	if !exist {
		return nil, fmt.Errorf("application already exists")
	}
	if err != nil {
		return nil, err
	}

	listKey := storageClient.Applications(application.AccountID)

	if err = storageEngine.LPush(listKey, key).Err(); err != nil {
		return nil, err
	}

	// Store the token details in redis
	_, err = storageEngine.HMSet(
		"tokens:"+Base64Encode(application.AuthToken),
		"acc", strconv.FormatInt(application.AccountID, 10),
		"app", strconv.FormatInt(application.ID, 10),
	).Result()
	if err != nil {
		return nil, err
	}

	if !retrieve {
		return application, nil
	}

	return ReadApplication(application.AccountID, application.ID)
}