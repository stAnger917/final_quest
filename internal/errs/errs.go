package errs

import "errors"

var ErrEmptyRegistrationData = errors.New("empty email / password in request body")
var ErrUserNotFound = errors.New("user not exist")
var ErrLoginMismatch = errors.New("login/password mismatch")
var ErrInvalidOrderNumber = errors.New("invalid order number")
var ErrOrderAlreadyExists = errors.New("order already exists")
var ErrOrderBelongsToAnotherUser = errors.New("order already uploaded by another user")
var ErrUserAlreadyExists = errors.New("user already exists")
var ErrIncorrectOrderBody = errors.New("invalid request Content-type")
var ErrIncorrectWithdrawReqBody = errors.New("invalid order number / order sum")
var ErrNotEnoughFounds = errors.New("not enough founds")
var ErrOrderNotFound = errors.New("order not found")
var ErrToManyRequests = errors.New("no more than N requests per minute allowed")
