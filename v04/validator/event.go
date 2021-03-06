package validator

import (
	"github.com/tapglue/multiverse/errors"
	"github.com/tapglue/multiverse/v04/core"
	"github.com/tapglue/multiverse/v04/entity"
	"github.com/tapglue/multiverse/v04/errmsg"
)

const (
	typeMin = 1
	typeMax = 30
)

// CreateEvent validates an event on create
func CreateEvent(datastore core.ApplicationUser, accountID, applicationID int64, event *entity.Event) (errs []errors.Error) {
	if !StringLengthBetween(event.Type, typeMin, typeMax) {
		errs = append(errs, errmsg.ErrEventTypeSize.SetCurrentLocation())
	}

	if event.ID != 0 {
		errs = append(errs, errmsg.ErrEventIDIsAlreadySet.SetCurrentLocation())
	}

	if event.Visibility == 0 {
		errs = append(errs, errmsg.ErrEventMissingVisiblity.SetCurrentLocation())
	} else if event.Visibility != entity.EventPrivate &&
		event.Visibility != entity.EventConnections &&
		event.Visibility != entity.EventPublic &&
		event.Visibility != entity.EventGlobal {
		errs = append(errs, errmsg.ErrEventInvalidVisiblity.SetCurrentLocation())
	}

	if len(errs) == 0 {
		// Run expensive check only if there are no existing errors
		if exists, err := datastore.ExistsByID(accountID, applicationID, event.UserID); !exists || err != nil {
			if err != nil {
				errs = append(errs, err...)
			} else {
				errs = append(errs, errmsg.ErrApplicationUserNotFound.SetCurrentLocation())
			}
		}
	}

	return
}

// UpdateEvent validates an event on update
func UpdateEvent(existingEvent, event *entity.Event) (errs []errors.Error) {
	if !StringLengthBetween(event.Type, typeMin, typeMax) {
		errs = append(errs, errmsg.ErrEventTypeSize.SetCurrentLocation())
	}

	if event.Visibility == 0 {
		errs = append(errs, errmsg.ErrEventMissingVisiblity.SetCurrentLocation())
	} else if event.Visibility != entity.EventPrivate &&
		event.Visibility != entity.EventConnections &&
		event.Visibility != entity.EventPublic &&
		event.Visibility != entity.EventGlobal {
		errs = append(errs, errmsg.ErrEventInvalidVisiblity.SetCurrentLocation())
	}

	// TODO define more rules for updating an event

	return
}
