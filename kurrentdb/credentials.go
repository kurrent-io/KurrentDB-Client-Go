package kurrentdb

import (
	"context"
	"encoding/base64"
)

// CredentialsProvider resolves the credentials for an outbound request. It is
// invoked once per request, so a provider can return a cached token until it
// nears expiry and then rotate it, and the new token is picked up transparently
// across reconnects. It is called on the request's goroutine and may run
// concurrently for in-flight requests, so it must be safe for concurrent use,
// reasonably fast (it blocks the request until it returns), and should honour
// the supplied context. Returning nil credentials sends the request
// unauthenticated.
type CredentialsProvider func(context.Context) (*Credentials, error)

// Credentials holds the authentication used for requests: either a
// login/password pair (HTTP Basic) or a bearer token (OAuth/JWT). A bearer
// token, when set, takes precedence over the login and password.
type Credentials struct {
	// User's login.
	Login string
	// User's password.
	Password string
	// BearerToken authenticates via an `Authorization: Bearer <token>` header,
	// e.g. an OAuth/OIDC access token. When set it takes precedence over Login
	// and Password. Bearer tokens are programmatic-only: they cannot be supplied
	// through a connection string.
	BearerToken string
}

// authorizationHeader renders the credentials as the value of an
// `Authorization` header (`"Bearer ..."` or `"Basic ..."`), or an empty string
// when no credentials are set.
func (creds *Credentials) authorizationHeader() string {
	if creds == nil {
		return ""
	}

	if creds.BearerToken != "" {
		return "Bearer " + creds.BearerToken
	}

	if creds.Login == "" && creds.Password == "" {
		return ""
	}

	return "Basic " + base64.StdEncoding.EncodeToString([]byte(creds.Login+":"+creds.Password))
}
