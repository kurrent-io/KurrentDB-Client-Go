package kurrentdb_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"

	"github.com/stretchr/testify/require"
)

func TestSubscription(t *testing.T) {
	t.Run("cluster", func(t *testing.T) {
		fixture := NewSecureClusterClientFixture(t)
		defer fixture.Close(t)
		runSubscriptionTests(t, fixture)
	})

	t.Run("singleNode", func(t *testing.T) {
		fixture := NewSecureSingleNodeClientFixture(t)
		defer fixture.Close(t)
		runSubscriptionTests(t, fixture)
	})
}

func runSubscriptionTests(t *testing.T, fixture *ClientFixture) {
	client := fixture.Client()

	t.Run("streamSubscriptionDeliversAllEventsInStreamAndListensForNewEvents", func(t *testing.T) {
		// Arrange
		streamId := fixture.NewStreamId()
		fixture.CreateTestEvents(streamId, 10)
		appendEvent := fixture.CreateTestEvent()

		var receivedEvents sync.WaitGroup
		var appendedEvents sync.WaitGroup

		// Act & Assert
		subscription, err := client.SubscribeToStream(context.Background(), streamId, kurrentdb.SubscribeToStreamOptions{
			From: kurrentdb.Start{},
		})

		require.NoError(t, err)
		defer subscription.Close()
		receivedEvents.Add(10)
		appendedEvents.Add(1)

		go func() {
			current := 0
			for {
				subEvent := subscription.Recv()

				if subEvent.EventAppeared != nil {
					current++
					if current <= 10 {
						receivedEvents.Done()
						continue
					}

					event := subEvent.EventAppeared
					require.Equal(t, appendEvent.EventID, event.OriginalEvent().EventID)
					require.Equal(t, uint64(10), event.OriginalEvent().EventNumber)
					require.Equal(t, streamId, event.OriginalEvent().StreamID)
					require.Equal(t, appendEvent.Data, event.OriginalEvent().Data)
					require.Equal(t, appendEvent.Metadata, event.OriginalEvent().UserMetadata)
					break
				}
			}
			appendedEvents.Done()
		}()

		timedOut := fixture.WaitWithTimeout(&receivedEvents, time.Duration(10)*time.Second)
		require.False(t, timedOut, "Timed out waiting for initial set of events")

		// Write a new event
		opts2 := kurrentdb.AppendToStreamOptions{
			StreamState: kurrentdb.Revision(9),
		}
		writeResult, err := client.AppendToStream(context.Background(), streamId, opts2, appendEvent)
		require.NoError(t, err)
		require.Equal(t, uint64(10), writeResult.NextExpectedVersion)

		// Assert event was forwarded to the subscription
		timedOut = fixture.WaitWithTimeout(&appendedEvents, time.Duration(10)*time.Second)
		require.False(t, timedOut, "Timed out waiting for the appended events")
	})

	//t.Run("allSubscriptionWithFilterDeliversCorrectEvents", func(t *testing.T) {
	//	// Arrange
	//	var receivedEvents sync.WaitGroup
	//	receivedEvents.Add(1)
	//
	//	prefix := uuid.New().String() + "-"
	//
	//	streamId := fixture.NewStreamId()
	//
	//	for i := 0; i < 10; i++ {
	//		event := fixture.CreateTestEvent()
	//		event.EventType = prefix + event.EventType
	//		_, err := client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{}, event)
	//		require.NoError(t, err)
	//	}
	//
	//	// Act
	//	subscription, err := client.SubscribeToAll(context.Background(), kurrentdb.SubscribeToAllOptions{
	//		From: kurrentdb.Start{},
	//		Filter: &kurrentdb.SubscriptionFilter{
	//			Type:     kurrentdb.EventFilterType,
	//			Prefixes: []string{prefix},
	//		},
	//	})
	//
	//	defer subscription.Close()
	//
	//	// Assert
	//	go func() {
	//		current := 0
	//		for {
	//			subEvent := subscription.Recv()
	//
	//			if subEvent.SubscriptionDropped != nil {
	//				break
	//
	//			} else if subEvent.EventAppeared != nil {
	//				event := subEvent.EventAppeared
	//				require.True(t, strings.HasPrefix(event.OriginalEvent().EventType, prefix))
	//				current++
	//				if current == 10 {
	//					receivedEvents.Done()
	//					break
	//
	//				}
	//			}
	//		}
	//	}()
	//
	//	require.NoError(t, err)
	//	timedOut := fixture.WaitWithTimeout(&receivedEvents, time.Duration(10)*time.Second)
	//	require.False(t, timedOut, "Timed out while waiting for events via the subscription")
	//})

	t.Run("subscribeToAllFilter", func(t *testing.T) {
		sub, err := client.SubscribeToAll(context.Background(), kurrentdb.SubscribeToAllOptions{From: kurrentdb.Start{}, Filter: kurrentdb.ExcludeSystemEventsFilter()})

		if err != nil {
			t.Error(err)
		}

		var completed sync.WaitGroup
		completed.Add(1)

		go func() {
			for {
				event := sub.Recv()

				if event.EventAppeared != nil {
					if strings.HasPrefix(event.EventAppeared.OriginalEvent().EventType, "$") {
						t.Errorf("We should not have system events!")
						return
					}

					completed.Done()
					break
				}
			}
		}()

		timedOut := fixture.WaitWithTimeout(&completed, time.Duration(30)*time.Second)
		require.False(t, timedOut, "Timed out waiting for filtered subscription completion")
	})

	t.Run("connectionClosing", func(t *testing.T) {
		var droppedEvent sync.WaitGroup

		streamId := fixture.NewStreamId()
		fixture.CreateTestEvents(streamId, 20)

		subscription, err := client.SubscribeToStream(context.Background(), streamId, kurrentdb.SubscribeToStreamOptions{
			From: kurrentdb.Start{},
		})

		go func() {
			current := 1

			for {
				subEvent := subscription.Recv()

				if subEvent.EventAppeared != nil {
					if current <= 10 {
						current++
					} else {
						subscription.Close()
					}

					continue
				}

				if subEvent.SubscriptionDropped != nil {
					break
				}
			}

			droppedEvent.Done()
		}()

		require.NoError(t, err)
		droppedEvent.Add(1)
		timedOut := fixture.WaitWithTimeout(&droppedEvent, time.Duration(10)*time.Second)
		require.False(t, timedOut, "Timed out waiting for dropped event")
	})

	t.Run("subscriptionToAllWithCredentialsOverride", func(t *testing.T) {
		opts := kurrentdb.SubscribeToAllOptions{
			Authenticated: &kurrentdb.Credentials{
				Login:    "admin",
				Password: "changeit",
			},
			From:   kurrentdb.Start{},
			Filter: kurrentdb.ExcludeSystemEventsFilter(),
		}
		_, err := client.SubscribeToAll(context.Background(), opts)

		if err != nil {
			t.Error(err)
		}
	})

	t.Run("subscriptionToStreamCaughtUp", func(t *testing.T) {
		const minSupportedVersion = 23
		const expectedEventCount = 10
		const testTimeout = 1 * time.Minute

		kurrentDbVersion, err := client.GetServerVersion()
		require.NoError(t, err, "Error getting server version")

		if kurrentDbVersion.Major < minSupportedVersion {
			t.Skip("CaughtUp message is not supported in this version of KurrentDB")
		}

		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		streamId := fixture.NewStreamId()
		fixture.CreateTestEvents(streamId, 10)

		subscription, err := client.SubscribeToStream(ctx, streamId, kurrentdb.SubscribeToStreamOptions{From: kurrentdb.Start{}})
		require.NoError(t, err)
		defer subscription.Close()

		var caughtUpReceived sync.WaitGroup
		caughtUpReceived.Add(1)

		go func() {
			var count uint64 = 0
			defer caughtUpReceived.Done()
			allEventsAcknowledged := false

			for {
				select {
				case <-ctx.Done():
					t.Error("Context timed out before receiving CaughtUp message")
					return
				default:
					event := subscription.Recv()

					if event.EventAppeared != nil {
						count++

						if count == expectedEventCount {
							allEventsAcknowledged = true
						}

						continue
					}

					if allEventsAcknowledged && event.CaughtUp != nil {
						require.True(t, count >= expectedEventCount, "Did not receive the exact number of expected events before CaughtUp")
						return
					}
				}
			}
		}()

		caughtUpTimedOut := fixture.WaitWithTimeout(&caughtUpReceived, testTimeout)
		require.False(t, caughtUpTimedOut, "Timed out waiting for CaughtUp message")
	})
}
