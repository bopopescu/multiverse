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
	application struct {
		storage redis.Client
		redis   *red.Client
	}
)

func (app *application) Create(application *entity.Application, retrieve bool) (*entity.Application, tgerrors.TGError) {
	var er error
	if application.ID, er = app.storage.GenerateApplicationID(application.AccountID); er != nil {
		return nil, tgerrors.NewInternalError("failed to create the application (1)", er.Error())
	}

	application.Enabled = true
	application.CreatedAt = time.Now()
	application.UpdatedAt = application.CreatedAt

	if application.AuthToken, er = storageHelper.GenerateApplicationSecretKey(application); er != nil {
		return nil, tgerrors.NewInternalError("failed to create the application (2)", er.Error())
	}

	val, er := json.Marshal(application)
	if er != nil {
		return nil, tgerrors.NewInternalError("failed to create the application (3)", er.Error())
	}

	key := storageHelper.Application(application.AccountID, application.ID)
	exist, er := app.redis.SetNX(key, string(val)).Result()
	if !exist {
		return nil, tgerrors.NewInternalError("failed to create the application (3)", "duplicate app")
	}
	if er != nil {
		return nil, tgerrors.NewInternalError("failed to create the application (4)", er.Error())
	}

	listKey := storageHelper.Applications(application.AccountID)
	if er = app.redis.LPush(listKey, key).Err(); er != nil {
		return nil, tgerrors.NewInternalError("failed to create the application (5)", er.Error())
	}

	// Store the token details in redis
	_, er = app.redis.HMSet(
		"tokens:"+utils.Base64Encode(application.AuthToken),
		"acc", strconv.FormatInt(application.AccountID, 10),
		"app", strconv.FormatInt(application.ID, 10),
	).Result()
	if er != nil {
		return nil, tgerrors.NewInternalError("failed to create the application (6)", er.Error())
	}

	if !retrieve {
		return application, nil
	}

	return app.Read(application.AccountID, application.ID)
}

func (app *application) Read(accountID, applicationID int64) (*entity.Application, tgerrors.TGError) {
	result, er := app.redis.Get(storageHelper.Application(accountID, applicationID)).Result()
	if er != nil {
		return nil, tgerrors.NewInternalError("failed to read the application (1)", er.Error())
	}

	application := &entity.Application{}
	if er := json.Unmarshal([]byte(result), application); er != nil {
		return nil, tgerrors.NewInternalError("failed to read the application (2)", er.Error())
	}

	return application, nil
}

func (app *application) Update(existingApplication, updatedApplication entity.Application, retrieve bool) (*entity.Application, tgerrors.TGError) {
	updatedApplication.UpdatedAt = time.Now()

	val, er := json.Marshal(updatedApplication)
	if er != nil {
		return nil, tgerrors.NewInternalError("failed to update the application (1)\n"+er.Error(), er.Error())
	}

	key := storageHelper.Application(updatedApplication.AccountID, updatedApplication.ID)
	exist, er := app.redis.Exists(key).Result()
	if !exist {
		return nil, tgerrors.NewNotFoundError("failed to update the application (2)", "app not found")
	}
	if er != nil {
		return nil, tgerrors.NewInternalError("failed to update the application (3)", er.Error())
	}

	if er = app.redis.Set(key, string(val)).Err(); er != nil {
		return nil, tgerrors.NewInternalError("failed to update the application (4)", er.Error())
	}

	if !updatedApplication.Enabled {
		listKey := storageHelper.Applications(updatedApplication.AccountID)
		if er = app.redis.LRem(listKey, 0, key).Err(); er != nil {
			return nil, tgerrors.NewInternalError("failed to update the application (5)", er.Error())
		}
	}

	if !retrieve {
		return &updatedApplication, nil
	}

	return app.Read(updatedApplication.AccountID, updatedApplication.ID)
}

func (app *application) Delete(application *entity.Application) tgerrors.TGError {
	// TODO: Disable application users?
	// TODO: User connections?
	// TODO: Application lists?
	// TODO: Application events?

	key := storageHelper.Application(application.AccountID, application.ID)
	result, er := app.redis.Del(key).Result()
	if er != nil {
		return tgerrors.NewInternalError("failed to delete the application (1)", er.Error())
	}

	if result != 1 {
		return tgerrors.NewInternalError("failed to delete the application (2)", "app not found")
	}

	listKey := storageHelper.Applications(application.AccountID)
	if er := app.redis.LRem(listKey, 0, key).Err(); er != nil {
		return tgerrors.NewInternalError("failed to delete the application (3)", er.Error())
	}

	return nil
}

func (app *application) List(accountID int64) ([]*entity.Application, tgerrors.TGError) {
	key := storageHelper.Applications(accountID)

	result, er := app.redis.LRange(key, 0, -1).Result()
	if er != nil {
		return nil, tgerrors.NewInternalError("failed to read the applications list (1)", er.Error())
	}

	applications := []*entity.Application{}
	if len(result) == 0 {
		return applications, nil
	}

	resultList, er := app.redis.MGet(result...).Result()
	if er != nil {
		return nil, tgerrors.NewInternalError("failed to read the applications list (2)", er.Error())
	}

	application := &entity.Application{}
	for _, result := range resultList {
		if er = json.Unmarshal([]byte(result.(string)), application); er != nil {
			return nil, tgerrors.NewInternalError("failed to read the applications list (3)", er.Error())
		}
		applications = append(applications, application)
		application = &entity.Application{}
	}

	return applications, nil
}

func (app *application) Exists(accountID, applicationID int64) bool {
	application, err := app.Read(accountID, applicationID)
	if err != nil {
		return false
	}

	return application.Enabled
}

// NewApplication creates a new Application
func NewApplication(storageClient redis.Client) core.Application {
	return &application{
		storage: storageClient,
		redis:   storageClient.Datastore(),
	}
}