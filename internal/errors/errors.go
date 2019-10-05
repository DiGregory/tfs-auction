package errors

import "errors"

var ErrBadReq = errors.New("Bad Request")
var ErrEmailExist = errors.New("email already exists")
var ErrInvalidEmailOrPass = errors.New("invalid email or password")
var ErrUnAuth = errors.New("unauthorized request")
var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict req")