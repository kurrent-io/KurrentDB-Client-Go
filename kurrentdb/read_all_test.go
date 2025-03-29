package kurrentdb_test

import (
	"context"
	"testing"
	"time"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"

	"github.com/stretchr/testify/assert"
)

func TestReadAll(t *testing.T) {
	fixture := NewSecureClusterClientFixture(t)
	defer fixture.Close(t)

	client := fixture.Client()

	t.Run("readAllEventsForwardsFromZeroPosition(", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		defer cancel()

		numberOfEvents := uint64(10)

		opts := kurrentdb.ReadAllOptions{
			Direction:      kurrentdb.Forwards,
			From:           kurrentdb.Start{},
			ResolveLinkTos: true,
		}
		stream, err := client.ReadAll(ctx, opts, numberOfEvents)
		if err != nil {
			t.Fatalf("Unexpected failure %+v", err)
		}

		defer stream.Close()

		events, err := fixture.CollectEvents(stream)
		if err != nil {
			t.Fatalf("Unexpected failure %+v", err)
		}

		assert.Equal(t, numberOfEvents, uint64(len(events)), "Expected the correct number of messages to be returned")
	})

	t.Run("readAllEventsForwardsFromNonZeroPosition", func(t *testing.T) {
		streamId := fixture.NewStreamId()

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		defer cancel()

		initialEvent := fixture.CreateTestEvent()
		result, err := fixture.client.AppendToStream(ctx, streamId, kurrentdb.AppendToStreamOptions{}, initialEvent)

		fixture.CreateTestEvents(streamId, 1000)

		assert.NoError(t, err)
		opts := kurrentdb.ReadAllOptions{
			From:           kurrentdb.Position{Commit: result.CommitPosition, Prepare: result.PreparePosition},
			ResolveLinkTos: true,
		}

		stream, err := client.ReadAll(ctx, opts, 1000)
		if err != nil {
			t.Fatalf("Unexpected failure %+v", err)
		}

		defer stream.Close()

		events, err := fixture.CollectEvents(stream)
		assert.NoError(t, err)

		assert.Equal(t, events[0].OriginalEvent().Position.Commit, result.CommitPosition)
		assert.Equal(t, events[0].OriginalEvent().EventID, initialEvent.EventID)
		assert.True(t, events[1].OriginalEvent().Position.Commit > result.CommitPosition)
	})

	t.Run("readAllEventsBackwardsFromZeroPosition", func(t *testing.T) {
		streamId := fixture.NewStreamId()
		testEvents := fixture.CreateTestEvents(streamId, 1000)

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		defer cancel()

		opts := kurrentdb.ReadAllOptions{
			From:           kurrentdb.End{},
			Direction:      kurrentdb.Backwards,
			ResolveLinkTos: true,
		}

		stream, err := client.ReadAll(ctx, opts, 1000)
		if err != nil {
			t.Fatalf("Unexpected failure %+v", err)
		}

		defer stream.Close()

		collectedEvents, err := fixture.CollectEvents(stream)
		if err != nil {
			t.Fatalf("Unexpected failure %+v", err)
		}

		assert.NotEmpty(t, collectedEvents)
		assert.Len(t, collectedEvents, 1000)

		found := false
		for _, collectedEvent := range collectedEvents {
			for _, testEvent := range testEvents {
				if collectedEvent.OriginalEvent().EventID == testEvent.EventID {
					found = true
					break
				}
			}
		}
		assert.True(t, found, "Expected to find at least one event from the testEvents")
	})

	t.Run("readAllEventsBackwardsFromNonZeroPosition", func(t *testing.T) {
		streamId := fixture.NewStreamId()

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		defer cancel()

		testEvents := fixture.CreateTestEvents(streamId, 1000)

		initialEvent := fixture.CreateTestEvent()
		result, err := fixture.client.AppendToStream(ctx, streamId, kurrentdb.AppendToStreamOptions{}, initialEvent)
		assert.NoError(t, err)

		fixture.CreateTestEvents(streamId, 10)

		opts := kurrentdb.ReadAllOptions{
			From:           kurrentdb.Position{Commit: result.CommitPosition, Prepare: result.PreparePosition},
			Direction:      kurrentdb.Backwards,
			ResolveLinkTos: true,
		}

		stream, err := client.ReadAll(ctx, opts, 1000)
		if err != nil {
			t.Fatalf("Unexpected failure %+v", err)
		}

		defer stream.Close()

		collectedEvents, err := fixture.CollectEvents(stream)
		if err != nil {
			t.Fatalf("Unexpected failure %+v", err)
		}

		found := false
		for _, collectedEvent := range collectedEvents {
			for _, testEvent := range testEvents {
				if collectedEvent.OriginalEvent().EventID == testEvent.EventID {
					found = true
					break
				}
			}
		}
		assert.True(t, found, "Expected to find at least one event from the testEvents")
	})

	t.Run("readAllEventsWithCredentialsOverride", func(t *testing.T) {
		streamId := fixture.NewStreamId()

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		defer cancel()

		result, err := fixture.client.AppendToStream(ctx, streamId, kurrentdb.AppendToStreamOptions{}, fixture.CreateTestEvent())
		assert.NoError(t, err)

		fixture.CreateTestEvents(streamId, 10)

		opts := kurrentdb.ReadAllOptions{
			Authenticated: &kurrentdb.Credentials{
				Login:    "admin",
				Password: "changeit",
			},
			From:           kurrentdb.Position{Commit: result.CommitPosition, Prepare: result.PreparePosition},
			Direction:      kurrentdb.Forwards,
			ResolveLinkTos: false,
		}

		stream, err := client.ReadAll(ctx, opts, 10)
		if err != nil {
			t.Fatalf("Unexpected failure %+v", err)
		}

		defer stream.Close()

		// collect all events to see if no error occurs
		_, err = fixture.CollectEvents(stream)
		if err != nil {
			t.Fatalf("Unexpected failure %+v", err)
		}
	})
}
