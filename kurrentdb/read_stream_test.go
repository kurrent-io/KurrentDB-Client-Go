package kurrentdb_test

import (
	"context"
	"errors"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
	"time"
)

func TestReadStream(t *testing.T) {
	t.Run("cluster", func(t *testing.T) {
		fixture := NewSecureClusterClientFixture(t)
		defer fixture.Close(t)
		runReadStreamTests(t, fixture)
	})

	t.Run("singleNode", func(t *testing.T) {
		fixture := NewSecureSingleNodeClientFixture(t)
		defer fixture.Close(t)
		runReadStreamTests(t, fixture)
	})
}

func runReadStreamTests(t *testing.T, fixture *ClientFixture) {
	client := fixture.Client()

	t.Run("read stream events forward from zero", func(t *testing.T) {
		// Arrange
		streamId := fixture.NewStreamId()

		testEvents := fixture.CreateTestEvents(streamId, 10)

		opts := kurrentdb.ReadStreamOptions{
			Direction:      kurrentdb.Forwards,
			ResolveLinkTos: true,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		defer cancel()

		// Act
		stream, err := client.ReadStream(ctx, streamId, opts, uint64(len(testEvents)))

		if err != nil {
			t.Fatalf("Unexpected failure %+v", err)
		}

		defer stream.Close()

		// Assert
		events, err := fixture.CollectEvents(stream)

		if err != nil {
			t.Fatalf("Unexpected failure %+v", err)
		}

		assert.Equal(t, 10, len(events), "Expected the correct number of messages to be returned")
		assert.Equal(t, testEvents[0].EventID, events[0].OriginalEvent().EventID)
	})

	t.Run("read stream events backward from end", func(t *testing.T) {
		// Arrange
		streamId := fixture.NewStreamId()

		testEvents := fixture.CreateTestEvents(streamId, 10)

		opts := kurrentdb.ReadStreamOptions{
			Direction:      kurrentdb.Backwards,
			From:           kurrentdb.End{},
			ResolveLinkTos: true,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		defer cancel()

		// Act
		stream, err := client.ReadStream(ctx, streamId, opts, uint64(len(testEvents)))

		if err != nil {
			t.Fatalf("Unexpected failure %+v", err)
		}

		defer stream.Close()

		// Assert
		events, err := fixture.CollectEvents(stream)

		if err != nil {
			t.Fatalf("Unexpected failure %+v", err)
		}

		assert.Equal(t, 10, len(events), "Expected the correct number of messages to be returned")
		assert.Equal(t, testEvents[len(testEvents)-1].EventID, events[0].OriginalEvent().EventID, "Last event appended should be the first event read")
	})

	t.Run("read stream returns EOF after completion", func(t *testing.T) {
		// Arrange
		streamId := fixture.NewStreamId()

		fixture.CreateTestEvents(streamId, 10)

		// Act
		stream, err := client.ReadStream(context.Background(), streamId, kurrentdb.ReadStreamOptions{}, 1_024)

		// Assert
		require.NoError(t, err)
		defer stream.Close()
		_, err = fixture.CollectEvents(stream)
		require.NoError(t, err)

		_, err = stream.Recv()
		require.Error(t, err)
		require.True(t, err == io.EOF)
	})

	t.Run("read stream not found", func(t *testing.T) {
		// Arrange
		streamId := fixture.NewStreamId()

		// Act
		stream, err := client.ReadStream(context.Background(), streamId, kurrentdb.ReadStreamOptions{}, 1)

		// Assert
		require.NoError(t, err)
		defer stream.Close()

		evt, err := stream.Recv()
		require.Nil(t, evt)

		kurrentDbError, ok := kurrentdb.FromError(err)

		require.False(t, ok)
		require.Equal(t, kurrentDbError.Code(), kurrentdb.ErrorCodeResourceNotFound)
	})

	t.Run("read stream with MaxAge", func(t *testing.T) {
		// Arrange
		streamId := fixture.NewStreamId()
		fixture.CreateTestEvents(streamId, 10)

		// Act
		metadata := kurrentdb.StreamMetadata{}
		metadata.SetMaxAge(time.Second)

		_, err := client.SetStreamMetadata(context.Background(), streamId, kurrentdb.AppendToStreamOptions{}, metadata)

		assert.NoError(t, err)

		time.Sleep(2 * time.Second)

		// Assert
		stream, err := client.ReadStream(context.Background(), streamId, kurrentdb.ReadStreamOptions{}, 10)
		require.NoError(t, err)
		defer stream.Close()

		evt, err := stream.Recv()
		require.Nil(t, evt)
		require.Error(t, err)
		require.True(t, errors.Is(err, io.EOF))
	})

	t.Run("read stream with credentials override", func(t *testing.T) {
		// Arrange
		streamId := fixture.NewStreamId()
		opts := kurrentdb.AppendToStreamOptions{
			Authenticated: &kurrentdb.Credentials{
				Login:    "admin",
				Password: "changeit",
			},
		}

		// Act & Assert
		_, err := client.AppendToStream(context.Background(), streamId, opts, fixture.CreateTestEvent())

		assert.NoError(t, err)

		streamId = fixture.NewStreamId()
		opts.Authenticated.Password = "invalid"
		_, err = client.AppendToStream(context.Background(), streamId, opts, fixture.CreateTestEvent())

		assert.Error(t, err)
	})
}
