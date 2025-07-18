package test

import (
	"context"
	"testing"
	"time"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/suite"
)

func TestSecureAuthenticationSuite(t *testing.T) {
	suite.Run(t, new(SecureAuthTestSuite))
}

type SecureAuthTestSuite struct {
	suite.Suite
	fixture *ClientFixture
}

func (s *SecureAuthTestSuite) SetupSuite() {
	s.fixture = NewSecureSingleNodeClientFixture(s.T())
}

func (s *SecureAuthTestSuite) TearDownSuite() {
	s.fixture.Close(s.T())
}

func (s *SecureAuthTestSuite) TestCallWithTLSAndDefaultCredentials() {
	fixture := s.fixture

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.fixture.client.CreatePersistentSubscription(ctx, fixture.NewStreamId(), fixture.NewGroupId(), kurrentdb.PersistentStreamSubscriptionOptions{})
	s.NoError(err, "Unexpected failure")
}

func (s *SecureAuthTestSuite) TestCallWithTLSAndOverrideCredentials() {
	fixture := s.fixture

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := kurrentdb.PersistentStreamSubscriptionOptions{
		Authenticated: &kurrentdb.Credentials{
			Login:    "admin",
			Password: "changeit",
		},
	}
	err := s.fixture.client.CreatePersistentSubscription(ctx, fixture.NewStreamId(), fixture.NewGroupId(), opts)

	s.NoError(err, "Unexpected failure")
}

func (s *SecureAuthTestSuite) TestCallWithTLSAndInvalidOverrideCredentials() {
	fixture := s.fixture

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := kurrentdb.PersistentStreamSubscriptionOptions{
		Authenticated: &kurrentdb.Credentials{
			Login:    "invalid",
			Password: "invalid",
		},
	}
	err := s.fixture.client.CreatePersistentSubscription(ctx, fixture.NewStreamId(), fixture.NewGroupId(), opts)

	s.Error(err, "Expected an error but got success")
}
