package auth

import "errors"

var ErrInvalidUsernameOrPassword = errors.New("invalid username or password")
var ErrUserNotFound = errors.New("user not found")
var ErrCurrentPasswordRequired = errors.New("current password is required to change password")
var ErrCurrentPasswordIncorrect = errors.New("current password is incorrect")
var ErrPasswordTooShort = errors.New("password must be at least 6 characters long")

var ErrInvalidToken = errors.New("invalid token")
var ErrTokenExpired = errors.New("token has expired")
var ErrTokenNotFound = errors.New("token not found")
var ErrEmailAlreadyExists = errors.New("email already exists")
