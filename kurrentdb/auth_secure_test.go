package kurrentdb_test

import (
	"context"
	"testing"
	"time"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

func TestSecureAuthentication(t *testing.T) {
	t.Run("cluster", func(t *testing.T) {
		fixture := NewSecureClusterClientFixture(t)
		defer fixture.Close(t)
		runAuthSecureTests(t, fixture)
	})

	t.Run("singleNode", func(t *testing.T) {
		fixture := NewSecureSingleNodeClientFixture(t)
		defer fixture.Close(t)
		runAuthSecureTests(t, fixture)
	})
}

func runAuthSecureTests(t *testing.T, fixture *ClientFixture) {
	client := fixture.Client()

	t.Run("callWithTLSAndDefaultCredentials", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		defer cancel()

		err := client.CreatePersistentSubscription(ctx, fixture.NewStreamId(), fixture.NewGroupId(), kurrentdb.PersistentStreamSubscriptionOptions{})

		if err != nil {
			t.Fatalf("Unexpected failure %+v", err)
		}
	})

	t.Run("callWithTLSAndOverrideCredentials", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		defer cancel()

		opts := kurrentdb.PersistentStreamSubscriptionOptions{
			Authenticated: &kurrentdb.Credentials{
				Login:    "admin",
				Password: "changeit",
			},
		}
		err := client.CreatePersistentSubscription(ctx, fixture.NewStreamId(), fixture.NewGroupId(), opts)

		if err != nil {
			t.Fatalf("Unexpected failure %+v", err)
		}
	})

	t.Run("callWithTLSAndInvalidOverrideCredentials", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		defer cancel()

		opts := kurrentdb.PersistentStreamSubscriptionOptions{
			Authenticated: &kurrentdb.Credentials{
				Login:    "invalid",
				Password: "invalid",
			},
		}
		err := client.CreatePersistentSubscription(ctx, fixture.NewStreamId(), fixture.NewGroupId(), opts)

		if err == nil {
			t.Fatalf("Unexpected succesfull completion!")
		}
	})
}
