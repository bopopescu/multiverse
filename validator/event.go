/**
 * @author Onur Akpolat <onurakpolat@gmail.com>
 */

package validator

import (
	"fmt"

	"github.com/tapglue/backend/core/entity"
)

const (
	verbMin = 1
	verbMax = 20
)

var (
	errorVerbSize = fmt.Errorf("verb must be between %d and %d characters", verbMin, verbMax)
	errorVerbType = fmt.Errorf("verb is not a valid alphanumeric sequence")

	errorUserIDZero = fmt.Errorf("user id can't be 0")
	errorUserIDType = fmt.Errorf("user id is not a valid integer")

	errorEventIDIsAlreadySet = fmt.Errorf("event id is already set")
)

// CreateEvent validates an event
func CreateEvent(event *entity.Event) error {
	errs := []*error{}

	// Validate ApplicationID
	if event.ApplicationID == 0 {
		errs = append(errs, &errorApplicationIDZero)
	}

	// Validate UserID
	if event.UserID == 0 {
		errs = append(errs, &errorUserIDZero)
	}

	// Validate Verb
	if !stringBetween(event.Verb, verbMin, verbMax) {
		errs = append(errs, &errorVerbSize)
	}

	if !alphaNumExtraCharFirst.Match([]byte(event.Verb)) {
		errs = append(errs, &errorVerbType)
	}

	// Validate ID
	if event.ID != 0 {
		errs = append(errs, &errorEventIDIsAlreadySet)
	}

	// Validate User
	if !userExists(event.ApplicationID, event.UserID) {
		errs = append(errs, &errorUserDoesNotExists)
	}

	return packErrors(errs)
}
