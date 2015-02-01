/**
 * @author Onur Akpolat <onurakpolat@gmail.com>
 */

package validator

import (
	"fmt"

	"github.com/tapglue/backend/core/entity"
)

const (
	userNameMin = 2
	userNameMax = 40
)

var (
	errorUserFirstNameSize = fmt.Errorf("user first name must be between %d and %d characters", userNameMin, userNameMax)
	errorUserFirstNameType = fmt.Errorf("user first name is not a valid alphanumeric sequence")

	errorUserLastNameSize = fmt.Errorf("user last name must be between %d and %d characters", userNameMin, userNameMax)
	errorUserLastNameType = fmt.Errorf("user last name is not a valid alphanumeric sequence")

	errorUserUsernameSize = fmt.Errorf("user username must be between %d and %d characters", userNameMin, userNameMax)
	errorUserUsernameType = fmt.Errorf("user username is not a valid alphanumeric sequence")

	errorApplicationIDZero = fmt.Errorf("application id can't be 0")
	errorApplicationIDType = fmt.Errorf("application id is not a valid integer")

	errorAuthTokenInvalid = fmt.Errorf("auth token is invalid")
	errorUserURLInvalid   = fmt.Errorf("user url is not a valid url")
	errorUserEmailInvalid = fmt.Errorf("user email is not valid")

	errorUserIDIsAlreadySet = fmt.Errorf("user id is already set")
)

// CreateUser validates a user
func CreateUser(user *entity.User) error {
	errs := []*error{}

	// Validate names
	if !stringBetween(user.FirstName, userNameMin, userNameMax) {
		errs = append(errs, &errorUserFirstNameSize)
	}

	if !stringBetween(user.LastName, userNameMin, userNameMax) {
		errs = append(errs, &errorUserLastNameSize)
	}

	if !stringBetween(user.Username, userNameMin, userNameMax) {
		errs = append(errs, &errorUserUsernameSize)
	}

	if !alphaNumExtraCharFirst.Match([]byte(user.FirstName)) {
		errs = append(errs, &errorUserFirstNameType)
	}

	if !alphaNumExtraCharFirst.Match([]byte(user.LastName)) {
		errs = append(errs, &errorUserLastNameType)
	}

	if !alphaNumExtraCharFirst.Match([]byte(user.Username)) {
		errs = append(errs, &errorUserUsernameType)
	}

	// Validate ApplicatonID
	if user.ApplicationID == 0 {
		errs = append(errs, &errorApplicationIDZero)
	}

	if numInt.Match([]byte(fmt.Sprintf("%d", user.ApplicationID))) {
		errs = append(errs, &errorApplicationIDType)
	}

	// Validate AuthToken
	if user.AuthToken == "" {
		errs = append(errs, &errorAuthTokenInvalid)
	}

	// Validate Email
	if user.Email == "" || !email.Match([]byte(user.Email)) {
		errs = append(errs, &errorUserEmailInvalid)
	}

	// Validate URL
	if user.URL != "" && !url.Match([]byte(user.URL)) {
		errs = append(errs, &errorUserURLInvalid)
	}

	// Validate Image
	if len(user.Image) > 0 {
		for _, image := range user.Image {
			if !url.Match([]byte(image.URL)) {
				errs = append(errs, &errorInvalidImageURL)
			}
		}
	}

	// TODO: Check if Application exists

	return packErrors(errs)
}
