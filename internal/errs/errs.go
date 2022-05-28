package errs

import "errors"

var ErrEmptyRegistrationData = errors.New("empty email / password in request body")
var ErrUserNotFound = errors.New("user not exist")
var ErrLoginMismatch = errors.New("login/password mismatch")
var ErrInvalidOrderNumber = errors.New("invalid order number")
var ErrOrderAlreadyExists = errors.New("order already exists")
var ErrOrderBelongsToAnotherUser = errors.New("order already uploaded by another user")
