package test

import (
	"context"
	"testing"
	"time"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/suite"
)

func TestInsecureAuthenticationSuite(t *testing.T) {
	suite.Run(t, new(InsecureAuthSuite))
}

type InsecureAuthSuite struct {
	suite.Suite
	fixture *ClientFixture
}

func (s *InsecureAuthSuite) SetupSuite() {
	s.fixture = NewInsecureClientFixture(s.T())
}

func (s *InsecureAuthSuite) TearDownSuite() {
	s.fixture.Close(s.T())
}

func (s *InsecureAuthSuite) TestCallInsecureWithoutCredentials() {
	client := s.fixture.Client()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.CreatePersistentSubscription(ctx, s.fixture.NewStreamId(), s.fixture.NewGroupId(), kurrentdb.PersistentStreamSubscriptionOptions{})
	s.NoError(err)
}

func (s *InsecureAuthSuite) TestCallInsecureWithInvalidCredentials() {
	client := s.fixture.Client()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := kurrentdb.PersistentStreamSubscriptionOptions{
		Authenticated: &kurrentdb.Credentials{
			Login:    "invalid",
			Password: "invalid",
		},
	}
	err := client.CreatePersistentSubscription(ctx, s.fixture.NewStreamId(), s.fixture.NewGroupId(), opts)
	s.NoError(err)
}
