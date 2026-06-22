package auth

import "errors"

// ErrInvalidCredentials is returned when login credentials do not match.
var ErrInvalidCredentials = errors.New("auth: invalid email or password")

// ErrDuplicateEmail is returned when a user tries to register with an email already in use.
var ErrDuplicateEmail = errors.New("auth: email already registered")

// ErrEmailNotVerified is returned when a user tries to log in before verifying their email.
var ErrEmailNotVerified = errors.New("auth: email not verified")

// ErrOAuthAccount is returned when a user with no password tries to log in with email/password.
var ErrOAuthAccount = errors.New("auth: account uses oauth login")

// ErrInvalidToken is returned when a verification token is invalid or already used.
var ErrInvalidToken = errors.New("auth: invalid verification token")

// ErrTokenExpired is returned when a verification token has expired.
var ErrTokenExpired = errors.New("auth: verification token expired")
