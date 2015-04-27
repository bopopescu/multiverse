/**
 * @author Florin Patan <florinpatan@gmail.com>
 */

package core

import (
	"github.com/tapglue/backend/errors"
	"github.com/tapglue/backend/v02/entity"
)

type (
	// Account interface
	Account interface {
		// Create adds a new account to the database and returns the created account or an error
		Create(account *entity.Account, retrieve bool) (*entity.Account, errors.Error)

		// Read returns the account matching the ID or an error
		Read(accountID int64) (*entity.Account, errors.Error)

		// Update updates the account matching the ID or an error
		Update(existingAccount, updatedAccount entity.Account, retrieve bool) (*entity.Account, errors.Error)

		// Delete deletes the account matching the ID or an error
		Delete(*entity.Account) errors.Error

		// Exists validates if an account exists and returns the account or an error
		Exists(accountID int64) (bool, errors.Error)
	}
)
