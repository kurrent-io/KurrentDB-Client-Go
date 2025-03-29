package kurrentdb_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/assert"
)

func TestAppendEvents(t *testing.T) {
	t.Run("cluster", func(t *testing.T) {
		fixture := NewSecureClusterClientFixture(t)
		defer fixture.Close(t)
		runAppendEventsTests(t, fixture)
	})

	t.Run("singleNode", func(t *testing.T) {
		fixture := NewSecureSingleNodeClientFixture(t)
		defer fixture.Close(t)
		runAppendEventsTests(t, fixture)
	})
}

func runAppendEventsTests(t *testing.T, fixture *ClientFixture) {
	client := fixture.Client()

	t.Run("appendToStreamSingleEventNoStream", func(t *testing.T) {
		// Arrange
		testEvent := fixture.CreateTestEvent()
		testEvent.EventID = uuid.New()
		streamId := fixture.NewStreamId()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		opts := kurrentdb.AppendToStreamOptions{StreamState: kurrentdb.NoStream{}}

		// Act
		_, err := client.AppendToStream(ctx, streamId, opts, testEvent)
		assert.Nil(t, err, "Unexpected failure when appending to stream")

		stream, err := client.ReadStream(ctx, streamId, kurrentdb.ReadStreamOptions{}, 1)
		assert.Nil(t, err, "Unexpected failure when reading stream")
		defer stream.Close()

		events, err := fixture.CollectEvents(stream)
		assert.Nil(t, err, "Unexpected failure when collecting events")

		// Assert
		assert.Len(t, events, 1, "Expected a single event")
		assert.Equal(t, testEvent.EventID, events[0].OriginalEvent().EventID)
		assert.Equal(t, testEvent.EventType, events[0].OriginalEvent().EventType)
		assert.Equal(t, streamId, events[0].OriginalEvent().StreamID)
		assert.Equal(t, testEvent.Data, events[0].OriginalEvent().Data)
		assert.Equal(t, testEvent.Metadata, events[0].OriginalEvent().UserMetadata)
	})

	t.Run("appendToStreamFailsOnNonExistentStream", func(t *testing.T) {
		// Arrange
		streamId := fixture.NewStreamId()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		opts := kurrentdb.AppendToStreamOptions{StreamState: kurrentdb.StreamExists{}}

		// Act
		_, err := client.AppendToStream(ctx, streamId, opts, fixture.CreateTestEvent())
		kurrentDbError, ok := kurrentdb.FromError(err)

		// Assert
		assert.False(t, ok)
		assert.Equal(t, kurrentdb.ErrorCodeWrongExpectedVersion, kurrentDbError.Code())
	})

	t.Run("metadataOperation", func(t *testing.T) {
		// Arrange
		streamId := fixture.NewStreamId()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		opts := kurrentdb.AppendToStreamOptions{StreamState: kurrentdb.Any{}}

		_, err := client.AppendToStream(ctx, streamId, opts, fixture.CreateTestEvent())
		assert.Nil(t, err, "Error writing event")

		acl := kurrentdb.Acl{}
		acl.AddReadRoles("admin")
		meta := kurrentdb.StreamMetadata{}
		meta.SetMaxAge(2 * time.Second)
		meta.SetAcl(acl)

		result, err := client.SetStreamMetadata(ctx, streamId, opts, meta)
		assert.Nil(t, err, "Error writing stream metadata")
		assert.NotNil(t, result, "Metadata write result should not be nil")

		metaActual, err := client.GetStreamMetadata(ctx, streamId, kurrentdb.ReadStreamOptions{})
		assert.Nil(t, err, "Error reading stream metadata")
		assert.Equal(t, meta, *metaActual, "Metadata should match")
	})
}
