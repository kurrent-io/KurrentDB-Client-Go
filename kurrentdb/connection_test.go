package kurrentdb_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"

	"github.com/stretchr/testify/assert"
)

func TestConnection(t *testing.T) {
	t.Run("cluster", func(t *testing.T) {
		fixture := NewSecureClusterClientFixture(t)
		defer fixture.Close(t)
		runConnectionTests(t, fixture)
	})

	t.Run("singleNode", func(t *testing.T) {
		fixture := NewSecureSingleNodeClientFixture(t)
		defer fixture.Close(t)
		runConnectionTests(t, fixture)
	})
}

func runConnectionTests(t *testing.T, fixture *ClientFixture) {
	client := fixture.Client()

	t.Run("closeConnection", func(t *testing.T) {
		testEvent := fixture.CreateTestEvent()

		streamID := uuid.New()
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		defer cancel()
		opts := kurrentdb.AppendToStreamOptions{
			StreamState: kurrentdb.NoStream{},
		}
		_, err := client.AppendToStream(ctx, streamID.String(), opts, testEvent)

		if err != nil {
			t.Fatalf("Unexpected failure %+v", err)
		}

		client.Close()
		opts.StreamState = kurrentdb.Any{}
		_, err = client.AppendToStream(ctx, streamID.String(), opts, testEvent)

		kurrentDbError, ok := kurrentdb.FromError(err)
		assert.False(t, ok)
		assert.Equal(t, kurrentDbError.Code(), kurrentdb.ErrorCodeConnectionClosed)
	})
}
