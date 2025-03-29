package kurrentdb_test

import (
	"context"
	"testing"
	"time"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

func TestInsecureAuthentication(t *testing.T) {
	fixture := NewInsecureClientFixture(t)
	defer fixture.Close(t)

	client := fixture.Client()

	t.Run("callInsecureWithoutCredentials", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		defer cancel()

		err := client.CreatePersistentSubscription(ctx, fixture.NewStreamId(), fixture.NewGroupId(), kurrentdb.PersistentStreamSubscriptionOptions{})

		if err != nil {
			t.Fatalf("Unexpected failure %+v", err)
		}
	})

	t.Run("callInsecureWithInvalidCredentials", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		defer cancel()

		opts := kurrentdb.PersistentStreamSubscriptionOptions{
			Authenticated: &kurrentdb.Credentials{
				Login:    "invalid",
				Password: "invalid",
			},
		}
		err := client.CreatePersistentSubscription(ctx, fixture.NewStreamId(), fixture.NewGroupId(), opts)

		if err != nil {
			t.Fatalf("Unexpected failure %+v", err)
		}
	})
}
