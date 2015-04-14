/**
 * @author Onur Akpolat <onurakpolat@gmail.com>
 */

package core

import (
	"github.com/tapglue/backend/tgerrors"
	"github.com/tapglue/backend/v02/entity"
)

type (
	// AccountUser interface
	AccountUser interface {
		// Create adds a new account user to the database and returns the created account user or an error
		Create(accountUser *entity.AccountUser, retrieve bool) (*entity.AccountUser, tgerrors.TGError)

		// Read returns the account matching the ID or an error
		Read(accountID, accountUserID int64) (accountUser *entity.AccountUser, er tgerrors.TGError)

		// Update update an account user in the database and returns the updated account user or an error
		Update(existingAccountUser, updatedAccountUser entity.AccountUser, retrieve bool) (*entity.AccountUser, tgerrors.TGError)

		// Delete deletes the account user matching the IDs or an error
		Delete(accountID, userID int64) tgerrors.TGError

		// List returns all the users from a certain account
		List(accountID int64) (accountUsers []*entity.AccountUser, er tgerrors.TGError)

		// CreateSession handles the creation of a user session and returns the session token
		CreateSession(user *entity.AccountUser) (string, tgerrors.TGError)

		// RefreshSession generates a new session token for the user session
		RefreshSession(sessionToken string, user *entity.AccountUser) (string, tgerrors.TGError)

		// DestroySession removes the user session
		DestroySession(sessionToken string, user *entity.AccountUser) tgerrors.TGError

		// GetSession retrieves the account user session token
		GetSession(user *entity.AccountUser) (string, tgerrors.TGError)

		// FindByEmail returns the account and account user for a certain e-mail address
		FindByEmail(email string) (*entity.Account, *entity.AccountUser, tgerrors.TGError)

		// ExistsByEmail checks if the account exists for a certain email
		ExistsByEmail(email string) (bool, tgerrors.TGError)

		// FindByUsername returns the account and account user for a certain username
		FindByUsername(username string) (*entity.Account, *entity.AccountUser, tgerrors.TGError)

		// ExistsByUsername checks if the account exists for a certain username
		ExistsByUsername(username string) (bool, tgerrors.TGError)

		// ExistsByID checks if an account user exists by ID or not
		ExistsByID(accountID, accountUserID int64) bool
	}
)
