package test

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/kurrent-io/KurrentDB-Client-Go/protos/kurrentdb/protocols/v1/gossip"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

func TestClusterSuite(t *testing.T) {
	suite.Run(t, new(ClusterTestSuite))
}

type ClusterTestSuite struct {
	suite.Suite
	fixture *ClientFixture
}

func (s *ClusterTestSuite) SetupSuite() {
	s.fixture = NewSecureClusterClientFixture(s.T())
}

func (s *ClusterTestSuite) TearDownSuite() {
	s.fixture.Close(s.T())
}

func (s *ClusterTestSuite) TestNotLeaderExceptionButWorkAfterRetry() {
	ctx := context.Background()
	retryCount := 10
	for count := 0; count < retryCount; count++ {
		config, err := kurrentdb.ParseConnectionString(fmt.Sprintf("esdb://admin:changeit@localhost:2111,localhost:2112,localhost:2113?nodepreference=follower&tlsverifycert=false"))
		s.Require().NoError(err, "Failed to parse connection string")

		db, err := kurrentdb.NewClient(config)
		s.Require().NoError(err, "Failed to create KurrentDB client")

		streamID := s.fixture.NewStreamId()
		group := s.fixture.NewGroupId()

		err = db.CreatePersistentSubscription(ctx, streamID, group, kurrentdb.PersistentStreamSubscriptionOptions{})
		if kurrentDbError, ok := kurrentdb.FromError(err); !ok {
			if kurrentDbError.IsErrorCode(kurrentdb.ErrorCodeResourceAlreadyExists) {
				// Retry if the stream name is not random enough.
				time.Sleep(1 * time.Second)
				db.Close()
				continue
			}
		}
		s.Assert().NotNil(err)
		// Try again. Should succeed as the db will reconnect to the leader.
		err = db.CreatePersistentSubscription(ctx, streamID, group, kurrentdb.PersistentStreamSubscriptionOptions{})
		if kurrentDbError, ok := kurrentdb.FromError(err); !ok {
			if kurrentDbError.IsErrorCode(kurrentdb.ErrorCodeResourceAlreadyExists) {
				db.Close()
				return
			}
			db.Close()
			s.T().Fatalf("Failed to create persistent subscription: %v", kurrentDbError)
		}
		s.Assert().Nil(err)
		db.Close()
		return
	}
	s.T().Fatalf("we retried %d times but the test is still failing", retryCount)
}

func (s *ClusterTestSuite) TestReadStreamAfterClusterRebalanced() {
	ctx := context.Background()
	config, err := kurrentdb.ParseConnectionString(fmt.Sprintf("esdb://admin:changeit@localhost:2111,localhost:2112,localhost:2113?nodepreference=leader&tlsverifycert=false"))
	s.Require().NoError(err, "Failed to parse connection string")

	db, err := kurrentdb.NewClient(config)
	s.Require().NoError(err, "Failed to create KurrentDB client")
	defer db.Close()

	streamID := s.fixture.NewStreamId()
	options := kurrentdb.ReadStreamOptions{From: kurrentdb.Start{}}
	stream, err := db.ReadStream(ctx, streamID, options, 10)
	if err != nil {
		s.T().Errorf("failed to read stream: %v", err)
		return
	}
	stream.Close()

	// Simulate leader node failure.
	members, err := db.Gossip(ctx)
	s.Require().NoError(err)

	for _, member := range members {
		if member.State != gossip.MemberInfo_Leader || !member.GetIsAlive() {
			continue
		}

		url := fmt.Sprintf("https://%s:%d/admin/shutdown",
			member.HttpEndPoint.Address, member.HttpEndPoint.Port)
		s.T().Log("Shutting down leader node: ", url)

		req, err := http.NewRequest("POST", url, nil)
		s.Require().NoError(err)

		req.SetBasicAuth("admin", "changeit")
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
		resp, err := client.Do(req)
		s.Require().NoError(err)
		resp.Body.Close()
		break
	}

	// Wait for the cluster to rebalance.
	time.Sleep(5 * time.Second)

	retryCount := 10
	for count := 0; count < retryCount; count++ {
		stream, err = db.ReadStream(ctx, streamID, options, 10)
		if err == nil {
			stream.Close()
			s.T().Logf("Successfully read stream after %d retries", count+1)
			return
		}
	}
	s.T().Fatalf("we retried %d times but the test is still failing", retryCount)
}
