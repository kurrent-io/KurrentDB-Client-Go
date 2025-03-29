package kurrentdb_test

import (
	"context"
	"testing"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelete(t *testing.T) {
	t.Run("cluster", func(t *testing.T) {
		fixture := NewSecureClusterClientFixture(t)
		defer fixture.Close(t)
		runDeleteTests(t, fixture)
	})

	t.Run("singleNode", func(t *testing.T) {
		fixture := NewSecureSingleNodeClientFixture(t)
		defer fixture.Close(t)
		runDeleteTests(t, fixture)
	})
}

func runDeleteTests(t *testing.T, fixture *ClientFixture) {
	client := fixture.Client()

	t.Run("canDeleteStream", func(t *testing.T) {
		opts := kurrentdb.DeleteStreamOptions{
			StreamState: kurrentdb.Revision(0),
		}

		streamId := fixture.NewStreamId()

		_, err := client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{}, fixture.CreateTestEvent())
		assert.NoError(t, err)
		deleteResult, err := client.DeleteStream(context.Background(), streamId, opts)

		if err != nil {
			t.Fatalf("Unexpected failure %+v", err)
		}

		assert.True(t, deleteResult.Position.Commit > 0)
		assert.True(t, deleteResult.Position.Prepare > 0)
	})

	t.Run("canTombstoneStream", func(t *testing.T) {
		streamId := fixture.NewStreamId()

		_, err := client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{}, fixture.CreateTestEvent())
		deleteResult, err := client.TombstoneStream(context.Background(), streamId, kurrentdb.TombstoneStreamOptions{
			StreamState: kurrentdb.Revision(0),
		})

		if err != nil {
			t.Fatalf("Unexpected failure %+v", err)
		}

		assert.True(t, deleteResult.Position.Commit > 0)
		assert.True(t, deleteResult.Position.Prepare > 0)

		_, err = client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{}, fixture.CreateTestEvent())
		require.Error(t, err)
	})

	t.Run("detectStreamDeleted", func(t *testing.T) {
		streamId := fixture.NewStreamId()
		event := fixture.CreateTestEvent()

		_, err := client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{}, event)
		require.Nil(t, err)

		_, err = client.TombstoneStream(context.Background(), streamId, kurrentdb.TombstoneStreamOptions{})
		require.Nil(t, err)

		stream, err := client.ReadStream(context.Background(), streamId, kurrentdb.ReadStreamOptions{}, 1)
		require.NoError(t, err)
		defer stream.Close()

		evt, err := stream.Recv()
		require.Nil(t, evt)
		kurrentDbError, ok := kurrentdb.FromError(err)
		require.False(t, ok)
		require.Equal(t, kurrentDbError.Code(), kurrentdb.ErrorCodeStreamDeleted)
	})
}
