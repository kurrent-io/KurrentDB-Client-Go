package test

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/suite"
)

// These tests run against the OAuth stack in docker-compose.oauth.yml
// (Keycloak + a KurrentDB node configured for OAuth). Bring it up with
// `make start-oauth` (requires KURRENTDB_LICENSE_KEY); the suite skips when it
// is not running.
const (
	oauthTokenEndpoint  = "https://localhost:8443/realms/kurrent/protocol/openid-connect/token"
	oauthNodeConnString = "kurrentdb://localhost:2116?tls=true&tlscafile=../certs/ca/ca.crt"
)

// fetchOAuthToken obtains an access token from Keycloak using the resource
// owner password grant (the realm enables direct access grants for tests).
func fetchOAuthToken(username, password string) (string, error) {
	httpClient := &http.Client{
		Timeout: 5 * time.Second,
		// Test-only: Keycloak serves the self-signed test CA. Do not copy this
		// into non-test code - a real token source must validate the IdP cert.
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}

	resp, err := httpClient.PostForm(oauthTokenEndpoint, url.Values{
		"grant_type": {"password"},
		"client_id":  {"kurrentdb-client"},
		"username":   {username},
		"password":   {password},
		"scope":      {"openid"},
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, body)
	}

	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("no access_token in response")
	}

	return out.AccessToken, nil
}

// bearerProvider returns a provider that supplies a freshly minted token for
// the given user on each call.
func bearerProvider(username, password string) kurrentdb.CredentialsProvider {
	return func(context.Context) (*kurrentdb.Credentials, error) {
		token, err := fetchOAuthToken(username, password)
		if err != nil {
			return nil, err
		}
		return &kurrentdb.Credentials{BearerToken: token}, nil
	}
}

func TestOAuthAuthenticationSuite(t *testing.T) {
	suite.Run(t, new(OAuthAuthTestSuite))
}

type OAuthAuthTestSuite struct {
	suite.Suite
}

// SetupSuite skips the suite unless the whole OAuth stack is up and working:
// Keycloak issues a token AND the OAuth node accepts it. This guarantees that
// a failure in the negative tests below is about the credentials under test,
// not an unavailable or unhealthy stack.
func (s *OAuthAuthTestSuite) SetupSuite() {
	token, err := fetchOAuthToken("admin", "changeit")
	if err != nil {
		s.T().Skipf("OAuth stack not available (run `make start-oauth`): %v", err)
	}

	client := s.newClient(staticBearer(token))
	defer client.Close()
	if err := s.createSubscription(client, kurrentdb.PersistentStreamSubscriptionOptions{}); err != nil {
		s.T().Skipf("OAuth node not ready (run `make start-oauth`): %v", err)
	}
}

func staticBearer(token string) kurrentdb.CredentialsProvider {
	return func(context.Context) (*kurrentdb.Credentials, error) {
		return &kurrentdb.Credentials{BearerToken: token}, nil
	}
}

func (s *OAuthAuthTestSuite) newClient(provider kurrentdb.CredentialsProvider) *kurrentdb.Client {
	config, err := kurrentdb.ParseConnectionString(oauthNodeConnString)
	s.Require().NoError(err, "Failed to parse connection string")
	config.CredentialsProvider = provider

	client, err := kurrentdb.NewClient(config)
	s.Require().NoError(err, "Failed to create KurrentDB client")

	return client
}

// createSubscription performs an admin-authorized operation, used as the probe
// for whether authentication and authorization succeeded.
func (s *OAuthAuthTestSuite) createSubscription(client *kurrentdb.Client, opts kurrentdb.PersistentStreamSubscriptionOptions) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return client.CreatePersistentSubscription(ctx, uuid.NewString(), uuid.NewString(), opts)
}

func (s *OAuthAuthTestSuite) TestProviderWithValidTokenSucceeds() {
	client := s.newClient(bearerProvider("admin", "changeit"))
	defer client.Close()

	err := s.createSubscription(client, kurrentdb.PersistentStreamSubscriptionOptions{})
	s.NoError(err, "a valid bearer token from the provider should authenticate")
}

func (s *OAuthAuthTestSuite) TestPerCallBearerTokenSucceeds() {
	token, err := fetchOAuthToken("admin", "changeit")
	s.Require().NoError(err)

	client := s.newClient(nil)
	defer client.Close()

	err = s.createSubscription(client, kurrentdb.PersistentStreamSubscriptionOptions{
		Authenticated: &kurrentdb.Credentials{BearerToken: token},
	})
	s.NoError(err, "a valid per-call bearer token should authenticate")
}

func (s *OAuthAuthTestSuite) TestValidTokenWithoutRequiredRoleIsDenied() {
	// The "noroles" user authenticates but lacks the role to manage persistent
	// subscriptions: the token is accepted (authentication) but the operation
	// is rejected (authorization).
	client := s.newClient(bearerProvider("noroles", "changeit"))
	defer client.Close()

	err := s.createSubscription(client, kurrentdb.PersistentStreamSubscriptionOptions{})
	kErr, _ := kurrentdb.FromError(err)
	s.Require().NotNil(kErr)
	s.True(kErr.IsErrorCode(kurrentdb.ErrorCodeAccessDenied), "expected access denied, got %v", err)
}

func (s *OAuthAuthTestSuite) TestProviderErrorFails() {
	client := s.newClient(func(context.Context) (*kurrentdb.Credentials, error) {
		return nil, fmt.Errorf("token source unavailable")
	})
	defer client.Close()

	err := s.createSubscription(client, kurrentdb.PersistentStreamSubscriptionOptions{})
	s.Require().Error(err)
	// The provider's error propagates through to the caller, proving the failure
	// originated in credential resolution rather than the transport.
	s.ErrorContains(err, "token source unavailable")
}
