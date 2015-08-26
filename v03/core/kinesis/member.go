package kinesis

import (
	"encoding/json"
	"fmt"

	"github.com/tapglue/backend/errors"
	"github.com/tapglue/backend/v03/core"
	"github.com/tapglue/backend/v03/entity"
	"github.com/tapglue/backend/v03/errmsg"
	"github.com/tapglue/backend/v03/storage/kinesis"

	ksis "github.com/sendgridlabs/go-kinesis"
)

type accountUser struct {
	a       core.Organization
	storage kinesis.Client
	ksis    *ksis.Kinesis
}

func (au *accountUser) Create(accountUser *entity.Member, retrieve bool) (*entity.Member, []errors.Error) {
	data, er := json.Marshal(accountUser)
	if er != nil {
		return nil, []errors.Error{errors.NewInternalError(0, "error while creating the account user (1)", er.Error())}
	}

	partitionKey := fmt.Sprintf("account-user-%d-%d", accountUser.OrgID, accountUser.ID)
	_, err := au.storage.PackAndPutRecord(kinesis.StreamAccountUserCreate, partitionKey, data)
	if err != nil {
		return nil, []errors.Error{err}
	}

	if retrieve {
		return accountUser, nil
	}

	return nil, nil
}

func (au *accountUser) Read(accountID, accountUserID int64) (accountUser *entity.Member, er []errors.Error) {
	return nil, []errors.Error{errmsg.ErrServerInvalidHandler}
}

func (au *accountUser) Update(existingAccountUser, updatedAccountUser entity.Member, retrieve bool) (*entity.Member, []errors.Error) {
	data, er := json.Marshal(updatedAccountUser)
	if er != nil {
		return nil, []errors.Error{errors.NewInternalError(0, "error while updating the account user (1)", er.Error())}
	}

	partitionKey := fmt.Sprintf("account-user-%d-%d", updatedAccountUser.OrgID, updatedAccountUser.ID)
	_, err := au.storage.PackAndPutRecord(kinesis.StreamAccountUserUpdate, partitionKey, data)
	if err != nil {
		return nil, []errors.Error{err}
	}

	if retrieve {
		return &updatedAccountUser, nil
	}

	return nil, nil
}

func (au *accountUser) Delete(accountUser *entity.Member) []errors.Error {
	data, er := json.Marshal(accountUser)
	if er != nil {
		return []errors.Error{errors.NewInternalError(0, "error while creating the event (1)", er.Error())}
	}

	partitionKey := fmt.Sprintf("account-user-%d-%d", accountUser.OrgID, accountUser.ID)
	_, err := au.storage.PackAndPutRecord(kinesis.StreamAccountUserDelete, partitionKey, data)
	if err != nil {
		return []errors.Error{err}
	}

	return nil
}

func (au *accountUser) List(accountID int64) (accountUsers []*entity.Member, er []errors.Error) {
	return accountUsers, []errors.Error{errmsg.ErrServerInvalidHandler}
}

func (au *accountUser) CreateSession(user *entity.Member) (string, []errors.Error) {
	return "", []errors.Error{errmsg.ErrServerInvalidHandler}
}

func (au *accountUser) RefreshSession(sessionToken string, user *entity.Member) (string, []errors.Error) {
	return "", []errors.Error{errmsg.ErrServerInvalidHandler}
}

func (au *accountUser) DestroySession(sessionToken string, user *entity.Member) []errors.Error {
	return []errors.Error{errmsg.ErrServerInvalidHandler}
}

func (au *accountUser) GetSession(user *entity.Member) (string, []errors.Error) {
	return "", []errors.Error{errmsg.ErrServerInvalidHandler}
}

func (au *accountUser) FindByEmail(email string) (*entity.Organization, *entity.Member, []errors.Error) {
	return nil, nil, []errors.Error{errmsg.ErrServerInvalidHandler}
}

func (au *accountUser) ExistsByEmail(email string) (bool, []errors.Error) {
	return false, []errors.Error{errmsg.ErrServerInvalidHandler}
}

func (au *accountUser) FindByUsername(username string) (*entity.Organization, *entity.Member, []errors.Error) {
	return nil, nil, []errors.Error{errmsg.ErrServerInvalidHandler}
}

func (au *accountUser) ExistsByUsername(username string) (bool, []errors.Error) {
	return false, []errors.Error{errmsg.ErrServerInvalidHandler}
}

func (au *accountUser) ExistsByID(accountID, userID int64) (bool, []errors.Error) {
	return false, []errors.Error{errmsg.ErrServerInvalidHandler}
}

func (au *accountUser) FindBySession(sessionKey string) (*entity.Member, []errors.Error) {
	return nil, []errors.Error{errmsg.ErrServerInvalidHandler}
}

func (au *accountUser) FindByPublicID(accountID int64, publicID string) (*entity.Member, []errors.Error) {
	return nil, []errors.Error{errmsg.ErrServerInvalidHandler}
}

// NewAccountUser creates a new AccountUser
func NewAccountUser(storageClient kinesis.Client) core.Member {
	return &accountUser{
		storage: storageClient,
		ksis:    storageClient.Datastore(),
		a:       NewAccount(storageClient),
	}
}

// NewAccountUserWithAccount creates a new AccountUser
func NewAccountUserWithAccount(storageClient kinesis.Client, a core.Organization) core.Member {
	return &accountUser{
		storage: storageClient,
		ksis:    storageClient.Datastore(),
		a:       a,
	}
}