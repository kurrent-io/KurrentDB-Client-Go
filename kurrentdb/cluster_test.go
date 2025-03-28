package kurrentdb_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/kurrent-io/KurrentDB-Client-Go/protos/gossip"
	"net/http"
	"testing"
	"time"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"

	"github.com/stretchr/testify/assert"
)

func TestCluster(t *testing.T) {
	fixture := NewSecureClusterClientFixture(t)
	defer fixture.Close(t)

	t.Run("notLeaderExceptionButWorkAfterRetry", func(t *testing.T) {
		// Seems on GHA, we need to try more that once because the name generator is not random enough.
		for count := 0; count < 10; count++ {
			ctx := context.Background()

			// We purposely connect to a follower node so we can trigger on not leader exception.
			config, err := kurrentdb.ParseConnectionString(fmt.Sprintf("esdb://admin:changeit@localhost:2111,localhost:2112,localhost:2113?nodepreference=follower&tlsverifycert=false"))
			assert.NoError(t, err, "Failed to parse connection string")

			db, err := kurrentdb.NewClient(config)
			assert.NoError(t, err, "Failed to create KurrentDB client")

			defer db.Close()
			streamID := fixture.NewStreamId()

			group := fixture.NewGroupId()

			err = db.CreatePersistentSubscription(ctx, streamID, group, kurrentdb.PersistentStreamSubscriptionOptions{})

			if esdbError, ok := kurrentdb.FromError(err); !ok {
				if esdbError.IsErrorCode(kurrentdb.ErrorCodeResourceAlreadyExists) {
					// Name generator is not random enough.
					time.Sleep(1 * time.Second)
					continue
				}
			}

			assert.NotNil(t, err)

			// It should work now as the db automatically reconnected to the leader node.
			err = db.CreatePersistentSubscription(ctx, streamID, group, kurrentdb.PersistentStreamSubscriptionOptions{})

			if esdbErr, ok := kurrentdb.FromError(err); !ok {
				if esdbErr.IsErrorCode(kurrentdb.ErrorCodeResourceAlreadyExists) {
					// Freak accident considering we use random stream name, safe to assume the test was
					// successful.
					return
				}

				t.Fatalf("Failed to create persistent subscription: %v", esdbErr)
			}

			assert.Nil(t, err)
			return
		}

		t.Fatalf("we retried long enough but the test is still failing")
	})

	t.Run("readStreamAfterClusterRebalanced", func(t *testing.T) {
		// We purposely connect to a leader node.
		config, err := kurrentdb.ParseConnectionString(fmt.Sprintf("esdb://admin:changeit@localhost:2111,localhost:2112,localhost:2113?nodepreference=leader&tlsverifycert=false"))
		assert.NoError(t, err, "Failed to parse connection string")

		db, err := kurrentdb.NewClient(config)
		assert.NoError(t, err, "Failed to create KurrentDB client")

		ctx := context.Background()
		streamID := fixture.NewStreamId()

		// Start reading the stream
		options := kurrentdb.ReadStreamOptions{From: kurrentdb.Start{}}

		stream, err := db.ReadStream(ctx, streamID, options, 10)
		if err != nil {
			t.Errorf("failed to read stream: %v", err)
			return
		}

		stream.Close()

		// Simulate leader node failure
		members, err := db.Gossip(ctx)

		assert.Nil(t, err)

		for _, member := range members {
			if member.State != gossip.MemberInfo_Leader || !member.GetIsAlive() {
				continue
			}

			// Shutdown the leader node
			url := fmt.Sprintf("https://%s:%d/admin/shutdown", member.HttpEndPoint.Address, member.HttpEndPoint.Port)
			t.Log("Shutting down leader node: ", url)

			req, err := http.NewRequest("POST", url, nil)
			assert.Nil(t, err)

			req.SetBasicAuth("admin", "changeit")
			client := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			}
			resp, err := client.Do(req)

			assert.Nil(t, err)
			resp.Body.Close()

			break
		}

		// Wait for the cluster to rebalance
		time.Sleep(5 * time.Second)

		// Try reading the stream again
		for count := 0; count < 10; count++ {
			stream, err = db.ReadStream(ctx, streamID, options, 10)
			if err != nil {
				continue
			}

			stream.Close()

			t.Logf("Successfully read stream after %d retries", count+1)
			return
		}

		t.Fatalf("we retried long enough but the test is still failing")
	})
}
