package kurrentdb_test

import (
	"context"
	"testing"
	"time"

	"github.com/kurrent-io/KurrentDB-Client-Go/v1/kurrentdb"
)

func SecureAuthenticationTests(t *testing.T, client *kurrentdb.Client) {
	t.Run("AuthenticationTests", func(t *testing.T) {
		t.Run("callWithTLSAndDefaultCredentials", callWithTLSAndDefaultCredentials(client))
		t.Run("callWithTLSAndOverrideCredentials", callWithTLSAndOverrideCredentials(client))
		t.Run("callWithTLSAndInvalidOverrideCredentials", callWithTLSAndInvalidOverrideCredentials(client))
	})
}

func callWithTLSAndDefaultCredentials(db *kurrentdb.Client) TestCall {
	return func(t *testing.T) {
		context, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		defer cancel()

		err := db.CreatePersistentSubscription(context, NAME_GENERATOR.Generate(), NAME_GENERATOR.Generate(), kurrentdb.PersistentStreamSubscriptionOptions{})

		if err != nil {
			t.Fatalf("Unexpected failure %+v", err)
		}
	}
}

func callWithTLSAndOverrideCredentials(db *kurrentdb.Client) TestCall {
	return func(t *testing.T) {
		context, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		defer cancel()

		opts := kurrentdb.PersistentStreamSubscriptionOptions{
			Authenticated: &kurrentdb.Credentials{
				Login:    "admin",
				Password: "changeit",
			},
		}
		err := db.CreatePersistentSubscription(context, NAME_GENERATOR.Generate(), NAME_GENERATOR.Generate(), opts)

		if err != nil {
			t.Fatalf("Unexpected failure %+v", err)
		}
	}
}

func callWithTLSAndInvalidOverrideCredentials(db *kurrentdb.Client) TestCall {
	return func(t *testing.T) {
		context, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		defer cancel()

		opts := kurrentdb.PersistentStreamSubscriptionOptions{
			Authenticated: &kurrentdb.Credentials{
				Login:    "invalid",
				Password: "invalid",
			},
		}
		err := db.CreatePersistentSubscription(context, NAME_GENERATOR.Generate(), NAME_GENERATOR.Generate(), opts)

		if err == nil {
			t.Fatalf("Unexpected succesfull completion!")
		}
	}
}
