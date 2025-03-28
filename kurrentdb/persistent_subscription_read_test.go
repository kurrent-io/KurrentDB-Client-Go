package kurrentdb_test

import (
	"context"
	"testing"
	"time"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"

	"github.com/stretchr/testify/require"
)

func TestPersistentSubRead(t *testing.T) {
	t.Run("cluster", func(t *testing.T) {
		fixture := NewSecureClusterClientFixture(t)
		defer fixture.Close(t)
		runPersistentSubReadTests(t, fixture)
	})

	t.Run("singleNode", func(t *testing.T) {
		fixture := NewSecureSingleNodeClientFixture(t)
		defer fixture.Close(t)
		runPersistentSubReadTests(t, fixture)
	})
}

func runPersistentSubReadTests(t *testing.T, fixture *ClientFixture) {
	client := fixture.Client()

	t.Run("ReadExistingStream_AckToReceiveNewEvents", func(t *testing.T) {
		streamId := fixture.NewStreamId()
		firstEvent := fixture.CreateTestEvent()
		secondEvent := fixture.CreateTestEvent()
		thirdEvent := fixture.CreateTestEvent()
		events := []kurrentdb.EventData{firstEvent, secondEvent, thirdEvent}

		_, err := client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{}, events...)
		require.NoError(t, err)

		groupName := fixture.NewGroupId()
		err = client.CreatePersistentSubscription(
			context.Background(),
			streamId,
			groupName,
			kurrentdb.PersistentStreamSubscriptionOptions{
				StartFrom: kurrentdb.Start{},
			},
		)
		require.NoError(t, err)

		readConn, err := client.SubscribeToPersistentSubscription(
			context.Background(), streamId, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{
				BufferSize: 2,
			})
		require.NoError(t, err)
		defer readConn.Close()

		firstRead := readConn.Recv().EventAppeared.Event
		require.NotNil(t, firstRead)
		require.Equal(t, firstEvent.EventID, firstRead.OriginalEvent().EventID)

		secondRead := readConn.Recv().EventAppeared.Event
		require.NotNil(t, secondRead)
		require.Equal(t, secondEvent.EventID, secondRead.OriginalEvent().EventID)

		err = readConn.Ack(firstRead)
		require.NoError(t, err)

		thirdRead := readConn.Recv()
		require.NotNil(t, thirdRead.EventAppeared)
		require.Equal(t, thirdEvent.EventID, thirdRead.EventAppeared.Event.OriginalEvent().EventID)
	})

	t.Run("ToExistingStream_StartFromBeginning_AndEventsInIt", func(t *testing.T) {
		streamId := fixture.NewStreamId()
		events := make([]kurrentdb.EventData, 10)
		for i := 0; i < 10; i++ {
			events[i] = fixture.CreateTestEvent()
		}

		_, err := client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
			StreamState: kurrentdb.NoStream{},
		}, events...)
		require.NoError(t, err)

		groupName := fixture.NewGroupId()
		err = client.CreatePersistentSubscription(
			context.Background(),
			streamId,
			groupName,
			kurrentdb.PersistentStreamSubscriptionOptions{
				StartFrom: kurrentdb.Start{},
			},
		)
		require.NoError(t, err)

		readConn, err := client.SubscribeToPersistentSubscription(
			context.Background(), streamId, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{})
		require.NoError(t, err)
		defer readConn.Close()

		recv := readConn.Recv()
		require.NotNil(t, t, recv.EventAppeared, "EventAppeared should not be nil")
		readEvent := recv.EventAppeared.Event
		require.Equal(t, events[0].EventID, readEvent.OriginalEvent().EventID)
		require.EqualValues(t, 0, readEvent.OriginalEvent().EventNumber)
	})

	t.Run("ToNonExistingStream_StartFromBeginning_AppendEventsAfterwards", func(t *testing.T) {
		streamId := fixture.NewStreamId()
		groupName := fixture.NewGroupId()
		err := client.CreatePersistentSubscription(
			context.Background(),
			streamId,
			groupName,
			kurrentdb.PersistentStreamSubscriptionOptions{
				StartFrom: kurrentdb.Start{},
			},
		)
		require.NoError(t, err)

		events := make([]kurrentdb.EventData, 10)
		for i := 0; i < 10; i++ {
			events[i] = fixture.CreateTestEvent()
		}

		_, err = client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
			StreamState: kurrentdb.NoStream{},
		}, events...)
		require.NoError(t, err)

		readConn, err := client.SubscribeToPersistentSubscription(
			context.Background(), streamId, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{})
		require.NoError(t, err)
		defer readConn.Close()

		recv := readConn.Recv()
		require.NotNil(t, t, recv.EventAppeared, "EventAppeared should not be nil")
		readEvent := recv.EventAppeared.Event
		require.Equal(t, events[0].EventID, readEvent.OriginalEvent().EventID)
		require.EqualValues(t, 0, readEvent.OriginalEvent().EventNumber)
	})

	t.Run("ToExistingStream_StartFromEnd_EventsInItAndAppendEventsAfterwards", func(t *testing.T) {
		streamId := fixture.NewStreamId()
		events := make([]kurrentdb.EventData, 11)
		for i := 0; i < 11; i++ {
			events[i] = fixture.CreateTestEvent()
		}

		_, err := client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
			StreamState: kurrentdb.NoStream{},
		}, events[:10]...)
		require.NoError(t, err)

		groupName := fixture.NewGroupId()
		err = client.CreatePersistentSubscription(
			context.Background(),
			streamId,
			groupName,
			kurrentdb.PersistentStreamSubscriptionOptions{
				StartFrom: kurrentdb.End{},
			},
		)
		require.NoError(t, err)

		_, err = client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
			StreamState: kurrentdb.Revision(9),
		}, events[10])
		require.NoError(t, err)

		readConn, err := client.SubscribeToPersistentSubscription(
			context.Background(), streamId, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{})
		require.NoError(t, err)
		defer readConn.Close()

		recv := readConn.Recv()
		require.NotNil(t, t, recv.EventAppeared, "EventAppeared should not be nil")
		readEvent := recv.EventAppeared.Event
		require.Equal(t, events[10].EventID, readEvent.OriginalEvent().EventID)
		require.EqualValues(t, 10, readEvent.OriginalEvent().EventNumber)
	})

	t.Run("ToExistingStream_StartFromEnd_EventsInIt", func(t *testing.T) {
		streamId := fixture.NewStreamId()
		events := make([]kurrentdb.EventData, 10)
		for i := 0; i < 10; i++ {
			events[i] = fixture.CreateTestEvent()
		}

		_, err := client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
			StreamState: kurrentdb.NoStream{},
		}, events...)
		require.NoError(t, err)

		groupName := fixture.NewGroupId()
		err = client.CreatePersistentSubscription(
			context.Background(),
			streamId,
			groupName,
			kurrentdb.PersistentStreamSubscriptionOptions{
				StartFrom: kurrentdb.End{},
			},
		)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		readConn, err := client.SubscribeToPersistentSubscription(
			ctx, streamId, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{})
		require.NoError(t, err)
		defer readConn.Close()

		done := make(chan struct{})
		go func() {
			event := readConn.Recv()
			if event.EventAppeared != nil {
				done <- struct{}{}
			}
		}()

		select {
		case <-ctx.Done():
			// Expected no events
		case <-done:
			t.Fatal("Received unexpected event")
		}
	})

	t.Run("ToNonExistingStream_StartFromTwo_AppendEventsAfterwards", func(t *testing.T) {
		streamId := fixture.NewStreamId()
		groupName := fixture.NewGroupId()
		err := client.CreatePersistentSubscription(
			context.Background(),
			streamId,
			groupName,
			kurrentdb.PersistentStreamSubscriptionOptions{
				StartFrom: kurrentdb.Revision(2),
			},
		)
		require.NoError(t, err)

		events := make([]kurrentdb.EventData, 4)
		for i := 0; i < 4; i++ {
			events[i] = fixture.CreateTestEvent()
		}

		_, err = client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
			StreamState: kurrentdb.NoStream{},
		}, events...)
		require.NoError(t, err)

		readConn, err := client.SubscribeToPersistentSubscription(
			context.Background(), streamId, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{})
		require.NoError(t, err)
		defer readConn.Close()

		recv := readConn.Recv()
		require.NotNil(t, t, recv.EventAppeared, "EventAppeared should not be nil")
		readEvent := recv.EventAppeared.Event
		require.Equal(t, events[2].EventID, readEvent.OriginalEvent().EventID)
		require.EqualValues(t, 2, readEvent.OriginalEvent().EventNumber)
	})

	t.Run("ToExistingStream_StartFrom10_EventsInItAppendEventsAfterwards", func(t *testing.T) {
		streamId := fixture.NewStreamId()
		events := make([]kurrentdb.EventData, 11)
		for i := 0; i < 11; i++ {
			events[i] = fixture.CreateTestEvent()
		}

		_, err := client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
			StreamState: kurrentdb.NoStream{},
		}, events[:10]...)
		require.NoError(t, err)

		groupName := fixture.NewGroupId()
		err = client.CreatePersistentSubscription(
			context.Background(),
			streamId,
			groupName,
			kurrentdb.PersistentStreamSubscriptionOptions{
				StartFrom: kurrentdb.Revision(10),
			},
		)
		require.NoError(t, err)

		_, err = client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
			StreamState: kurrentdb.Revision(9),
		}, events[10])
		require.NoError(t, err)

		readConn, err := client.SubscribeToPersistentSubscription(
			context.Background(), streamId, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{})
		require.NoError(t, err)
		defer readConn.Close()

		recv := readConn.Recv()
		require.NotNil(t, t, recv.EventAppeared, "EventAppeared should not be nil")
		readEvent := recv.EventAppeared.Event
		require.Equal(t, events[10].EventID, readEvent.OriginalEvent().EventID)
		require.EqualValues(t, 10, readEvent.OriginalEvent().EventNumber)
	})

	t.Run("ToExistingStream_StartFrom4_EventsInIt", func(t *testing.T) {
		streamId := fixture.NewStreamId()
		events := make([]kurrentdb.EventData, 11)
		for i := 0; i < 11; i++ {
			events[i] = fixture.CreateTestEvent()
		}

		_, err := client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
			StreamState: kurrentdb.NoStream{},
		}, events[:10]...)
		require.NoError(t, err)

		groupName := fixture.NewGroupId()
		err = client.CreatePersistentSubscription(
			context.Background(),
			streamId,
			groupName,
			kurrentdb.PersistentStreamSubscriptionOptions{
				StartFrom: kurrentdb.Revision(4),
			},
		)
		require.NoError(t, err)

		_, err = client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
			StreamState: kurrentdb.Revision(9),
		}, events[10])
		require.NoError(t, err)

		readConn, err := client.SubscribeToPersistentSubscription(
			context.Background(), streamId, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{})
		require.NoError(t, err)
		defer readConn.Close()

		recv := readConn.Recv()
		require.NotNil(t, t, recv.EventAppeared, "EventAppeared should not be nil")
		readEvent := recv.EventAppeared.Event
		require.Equal(t, events[4].EventID, readEvent.OriginalEvent().EventID)
		require.EqualValues(t, 4, readEvent.OriginalEvent().EventNumber)
	})

	t.Run("ToExistingStream_StartFromHigherRevisionThenEventsInStream_EventsInItAppendEventsAfterwards", func(t *testing.T) {
		streamId := fixture.NewStreamId()
		events := make([]kurrentdb.EventData, 12)
		for i := 0; i < 12; i++ {
			events[i] = fixture.CreateTestEvent()
		}

		_, err := client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
			StreamState: kurrentdb.NoStream{},
		}, events[:11]...)
		require.NoError(t, err)

		groupName := fixture.NewGroupId()
		err = client.CreatePersistentSubscription(
			context.Background(),
			streamId,
			groupName,
			kurrentdb.PersistentStreamSubscriptionOptions{
				StartFrom: kurrentdb.Revision(11),
			},
		)
		require.NoError(t, err)

		_, err = client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
			StreamState: kurrentdb.Revision(10),
		}, events[11])
		require.NoError(t, err)

		readConn, err := client.SubscribeToPersistentSubscription(
			context.Background(), streamId, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{})
		require.NoError(t, err)
		defer readConn.Close()

		recv := readConn.Recv()
		require.NotNil(t, t, recv.EventAppeared, "EventAppeared should not be nil")
		readEvent := recv.EventAppeared.Event
		require.Equal(t, events[11].EventID, readEvent.OriginalEvent().EventID)
		require.EqualValues(t, 11, readEvent.OriginalEvent().EventNumber)
	})

	t.Run("ReadExistingStream_NackToReceiveNewEvents", func(t *testing.T) {
		streamId := fixture.NewStreamId()
		firstEvent := fixture.CreateTestEvent()
		secondEvent := fixture.CreateTestEvent()
		thirdEvent := fixture.CreateTestEvent()
		events := []kurrentdb.EventData{firstEvent, secondEvent, thirdEvent}

		_, err := client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{}, events...)
		require.NoError(t, err)

		groupName := fixture.NewGroupId()
		err = client.CreatePersistentSubscription(
			context.Background(),
			streamId,
			groupName,
			kurrentdb.PersistentStreamSubscriptionOptions{
				StartFrom: kurrentdb.Start{},
			},
		)
		require.NoError(t, err)

		readConn, err := client.SubscribeToPersistentSubscription(
			context.Background(), streamId, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{
				BufferSize: 2,
			})
		require.NoError(t, err)
		defer readConn.Close()

		firstRead := readConn.Recv().EventAppeared.Event
		require.NotNil(t, firstRead)
		require.Equal(t, firstEvent.EventID, firstRead.OriginalEvent().EventID)

		secondRead := readConn.Recv().EventAppeared.Event
		require.NotNil(t, secondRead)
		require.Equal(t, secondEvent.EventID, secondRead.OriginalEvent().EventID)

		err = readConn.Nack("test reason", kurrentdb.NackActionPark, firstRead)
		require.NoError(t, err)

		thirdRead := readConn.Recv()
		require.NotNil(t, thirdRead.EventAppeared)
		require.Equal(t, thirdEvent.EventID, thirdRead.EventAppeared.Event.OriginalEvent().EventID)
	})

	t.Run("persistentSubscriptionToAll_Read", func(t *testing.T) {
		groupName := fixture.NewGroupId()
		err := client.CreatePersistentSubscriptionToAll(
			context.Background(),
			groupName,
			kurrentdb.PersistentAllSubscriptionOptions{
				StartFrom: kurrentdb.Start{},
			},
		)

		if err, ok := kurrentdb.FromError(err); !ok {
			if err.Code() == kurrentdb.ErrorCodeUnsupportedFeature && fixture.IsKurrentDbVersion20() {
				t.Skip()
			}
		}

		require.NoError(t, err)

		readConnectionClient, err := client.SubscribeToPersistentSubscriptionToAll(
			context.Background(), groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{
				BufferSize: 2,
			},
		)
		require.NoError(t, err)
		defer readConnectionClient.Close()

		firstReadEvent := readConnectionClient.Recv().EventAppeared.Event
		require.NoError(t, err)
		require.NotNil(t, firstReadEvent)

		secondReadEvent := readConnectionClient.Recv().EventAppeared.Event
		require.NoError(t, err)
		require.NotNil(t, secondReadEvent)

		// since buffer size is two, after reading two outstanding messages
		// we must acknowledge a message in order to receive third one
		err = readConnectionClient.Ack(firstReadEvent)
		require.NoError(t, err)

		thirdReadEvent := readConnectionClient.Recv()
		require.NoError(t, err)
		require.NotNil(t, thirdReadEvent)
		err = readConnectionClient.Ack(thirdReadEvent.EventAppeared.Event)
		require.NoError(t, err)
	})
}
