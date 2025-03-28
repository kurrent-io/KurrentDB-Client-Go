package kurrentdb_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersistentSub(t *testing.T) {
	t.Run("cluster", func(t *testing.T) {
		fixture := NewSecureClusterClientFixture(t)
		defer fixture.Close(t)
		runPersistentSubscriptionsTests(t, fixture)
	})

	t.Run("singleNode", func(t *testing.T) {
		fixture := NewSecureSingleNodeClientFixture(t)
		defer fixture.Close(t)
		runPersistentSubscriptionsTests(t, fixture)
	})
}

func runPersistentSubscriptionsTests(t *testing.T, fixture *ClientFixture) {
	client := fixture.Client()

	t.Run("createPersistentStreamSubscription", func(t *testing.T) {
		streamId := fixture.NewStreamId()
		fixture.CreateTestEvents(streamId, 1)

		err := client.CreatePersistentSubscription(
			context.Background(),
			streamId,
			"Group 1",
			kurrentdb.PersistentStreamSubscriptionOptions{},
		)

		require.NoError(t, err)
	})

	t.Run("createPersistentStreamSubscription_MessageTimeoutZero", func(t *testing.T) {
		streamId := fixture.NewStreamId()
		fixture.CreateTestEvents(streamId, 1)

		settings := kurrentdb.SubscriptionSettingsDefault()
		settings.MessageTimeout = 0

		err := client.CreatePersistentSubscription(
			context.Background(),
			streamId,
			"Group 1",
			kurrentdb.PersistentStreamSubscriptionOptions{
				Settings: &settings,
			},
		)

		require.NoError(t, err)
	})

	t.Run("createPersistentStreamSubscription_StreamNotExits", func(t *testing.T) {
		streamId := fixture.NewStreamId()
		fixture.CreateTestEvents(streamId, 1)

		err := client.CreatePersistentSubscription(
			context.Background(),
			streamId,
			"Group 1",
			kurrentdb.PersistentStreamSubscriptionOptions{},
		)

		require.NoError(t, err)
	})

	t.Run("createPersistentStreamSubscription_FailsIfAlreadyExists", func(t *testing.T) {
		streamId := fixture.NewStreamId()
		fixture.CreateTestEvents(streamId, 1)

		err := client.CreatePersistentSubscription(
			context.Background(),
			streamId,
			"Group 1",
			kurrentdb.PersistentStreamSubscriptionOptions{},
		)

		require.NoError(t, err)

		err = client.CreatePersistentSubscription(
			context.Background(),
			streamId,
			"Group 1",
			kurrentdb.PersistentStreamSubscriptionOptions{},
		)

		esdbErr, ok := kurrentdb.FromError(err)
		assert.False(t, ok)
		assert.Equal(t, esdbErr.Code(), kurrentdb.ErrorCodeResourceAlreadyExists)
	})

	t.Run("createPersistentStreamSubscription_AfterDeleting", func(t *testing.T) {
		streamId := fixture.NewStreamId()
		fixture.CreateTestEvents(streamId, 1)

		err := client.CreatePersistentSubscription(
			context.Background(),
			streamId,
			"Group 1",
			kurrentdb.PersistentStreamSubscriptionOptions{},
		)

		require.NoError(t, err)

		err = client.DeletePersistentSubscription(context.Background(), streamId, "Group 1", kurrentdb.DeletePersistentSubscriptionOptions{})

		require.NoError(t, err)

		err = client.CreatePersistentSubscription(
			context.Background(),
			streamId,
			"Group 1",
			kurrentdb.PersistentStreamSubscriptionOptions{},
		)

		require.NoError(t, err)
	})

	t.Run("updatePersistentStreamSubscription", func(t *testing.T) {
		streamId := fixture.NewStreamId()
		fixture.CreateTestEvents(streamId, 1)

		err := client.CreatePersistentSubscription(
			context.Background(),
			streamId,
			"Group 1",
			kurrentdb.PersistentStreamSubscriptionOptions{},
		)

		require.NoError(t, err)

		settings := kurrentdb.SubscriptionSettingsDefault()
		settings.HistoryBufferSize++
		settings.ConsumerStrategyName = kurrentdb.ConsumerStrategyDispatchToSingle
		settings.MaxSubscriberCount++
		settings.ReadBatchSize++
		settings.CheckpointAfter++
		settings.CheckpointUpperBound++
		settings.CheckpointLowerBound++
		settings.LiveBufferSize++
		settings.MaxRetryCount++
		settings.MessageTimeout++
		settings.ExtraStatistics = !settings.ExtraStatistics
		settings.ResolveLinkTos = !settings.ResolveLinkTos

		err = client.UpdatePersistentSubscription(context.Background(), streamId, "Group 1", kurrentdb.PersistentStreamSubscriptionOptions{
			Settings: &settings,
		})

		require.NoError(t, err)
	})

	t.Run("updatePersistentStreamSubscription_ErrIfSubscriptionDoesNotExist", func(t *testing.T) {
		streamId := fixture.NewStreamId()
		fixture.CreateTestEvents(streamId, 1)

		err := client.UpdatePersistentSubscription(context.Background(), streamId, "Group 1", kurrentdb.PersistentStreamSubscriptionOptions{})

		require.Error(t, err)
	})

	t.Run("deletePersistentStreamSubscription", func(t *testing.T) {
		streamId := fixture.NewStreamId()
		fixture.CreateTestEvents(streamId, 1)

		err := client.CreatePersistentSubscription(
			context.Background(),
			streamId,
			"Group 1",
			kurrentdb.PersistentStreamSubscriptionOptions{},
		)

		require.NoError(t, err)

		err = client.DeletePersistentSubscription(
			context.Background(),
			streamId,
			"Group 1",
			kurrentdb.DeletePersistentSubscriptionOptions{},
		)

		require.NoError(t, err)
	})

	t.Run("deletePersistentSubscription_ErrIfSubscriptionDoesNotExist", func(t *testing.T) {
		streamId := fixture.NewStreamId()

		err := client.DeletePersistentSubscription(
			context.Background(),
			streamId,
			"a",
			kurrentdb.DeletePersistentSubscriptionOptions{},
		)

		kurrentDbError, ok := kurrentdb.FromError(err)
		assert.False(t, ok)
		assert.Equal(t, kurrentDbError.Code(), kurrentdb.ErrorCodeResourceNotFound)
	})

	t.Run("testPersistentSubscriptionClosing", func(t *testing.T) {
		streamId := fixture.NewStreamId()
		fixture.CreateTestEvents(streamId, 20)

		groupName := "Group 1"

		err := client.CreatePersistentSubscription(context.Background(), streamId, groupName, kurrentdb.PersistentStreamSubscriptionOptions{
			StartFrom: kurrentdb.Start{},
		})

		require.NoError(t, err)

		var receivedEvents sync.WaitGroup
		var droppedEvent sync.WaitGroup

		subscription, err := client.SubscribeToPersistentSubscription(
			context.Background(), streamId, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{
				BufferSize: 2,
			})

		require.NoError(t, err)

		go func() {
			current := 1

			for {
				subEvent := subscription.Recv()

				if subEvent.EventAppeared != nil {
					if current <= 10 {
						receivedEvents.Done()
						current++
					}

					subscription.Ack(subEvent.EventAppeared.Event)

					continue
				}

				if subEvent.SubscriptionDropped != nil {
					droppedEvent.Done()
					break
				}
			}
		}()

		require.NoError(t, err)
		receivedEvents.Add(10)
		droppedEvent.Add(1)
		timedOut := fixture.WaitWithTimeout(&receivedEvents, 5*time.Second)
		require.False(t, timedOut, "Timed out waiting for initial set of events")
		subscription.Close()
		timedOut = fixture.WaitWithTimeout(&droppedEvent, 5*time.Second)
		require.False(t, timedOut, "Timed out waiting for dropped event")
	})

	t.Run("persistentAllCreate", func(t *testing.T) {
		groupName := fixture.NewGroupId()

		err := client.CreatePersistentSubscriptionToAll(
			context.Background(),
			groupName,
			kurrentdb.PersistentAllSubscriptionOptions{},
		)

		if err, ok := kurrentdb.FromError(err); !ok {
			if err.Code() == kurrentdb.ErrorCodeUnsupportedFeature && fixture.IsKurrentDbVersion20() {
				t.Skip()
			}
		}

		require.NoError(t, err)
	})

	t.Run("persistentAllUpdate", func(t *testing.T) {
		groupName := fixture.NewGroupId()

		err := client.CreatePersistentSubscriptionToAll(
			context.Background(),
			groupName,
			kurrentdb.PersistentAllSubscriptionOptions{},
		)

		if err, ok := kurrentdb.FromError(err); !ok {
			if err.Code() == kurrentdb.ErrorCodeUnsupportedFeature && fixture.IsKurrentDbVersion20() {
				t.Skip()
			}
		}

		require.NoError(t, err)

		setts := kurrentdb.SubscriptionSettingsDefault()
		setts.ResolveLinkTos = true

		err = client.UpdatePersistentSubscriptionToAll(context.Background(), groupName, kurrentdb.PersistentAllSubscriptionOptions{
			Settings: &setts,
		})

		require.NoError(t, err)
	})

	t.Run("persistentAllDelete", func(t *testing.T) {
		groupName := fixture.NewGroupId()

		err := client.CreatePersistentSubscriptionToAll(
			context.Background(),
			groupName,
			kurrentdb.PersistentAllSubscriptionOptions{},
		)

		if err, ok := kurrentdb.FromError(err); !ok {
			if err.Code() == kurrentdb.ErrorCodeUnsupportedFeature && fixture.IsKurrentDbVersion20() {
				t.Skip()
			}
		}

		require.NoError(t, err)

		err = client.DeletePersistentSubscriptionToAll(context.Background(), groupName, kurrentdb.DeletePersistentSubscriptionOptions{})

		require.NoError(t, err)
	})

	t.Run("persistentListAllSubs", func(t *testing.T) {
		groupName := fixture.NewGroupId()
		streamName := fixture.NewStreamId()
		eventCount := uint32(2)

		err := client.CreatePersistentSubscription(context.Background(), streamName, groupName, kurrentdb.PersistentStreamSubscriptionOptions{})

		require.NoError(t, err)

		sub, err := client.SubscribeToPersistentSubscription(context.Background(), streamName, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{})
		defer sub.Close()

		require.NoError(t, err)

		fixture.CreateTestEvents(streamName, eventCount)

		require.NoError(t, err)

		i := uint32(0)
		for i < eventCount {
			event := sub.Recv()

			if event.SubscriptionDropped != nil {
				t.FailNow()
			}

			if event.EventAppeared != nil {
				err = sub.Nack("because reasons", kurrentdb.NackActionPark, event.EventAppeared.Event)
				require.NoError(t, err)
				i++
			}
		}

		time.Sleep(5 * time.Second)

		err = client.ReplayParkedMessages(context.Background(), streamName, groupName, kurrentdb.ReplayParkedMessagesOptions{})

		require.NoError(t, err)

		i = 0
		for i < eventCount {
			event := sub.Recv()

			if event.SubscriptionDropped != nil {
				t.FailNow()
			}

			if event.EventAppeared != nil {
				err = sub.Ack(event.EventAppeared.Event)
				require.NoError(t, err)
				i++
			}
		}

		require.NoError(t, err)
	})

	t.Run("persistentListAllSubsWithCredentialsOverride", func(t *testing.T) {
		groupName := fixture.NewGroupId()
		streamName := fixture.NewStreamId()

		err := client.CreatePersistentSubscription(context.Background(), streamName, groupName, kurrentdb.PersistentStreamSubscriptionOptions{})

		require.NoError(t, err)

		opts := kurrentdb.ListPersistentSubscriptionsOptions{
			Authenticated: &kurrentdb.Credentials{
				Login:    "admin",
				Password: "changeit",
			},
		}
		_, err = client.ListAllPersistentSubscriptions(context.Background(), opts)

		require.NoError(t, err)
	})

	t.Run("persistentReplayParkedMessages", func(t *testing.T) {
		streamName := fixture.NewStreamId()
		groupName := fixture.NewGroupId()
		eventCount := uint32(2)

		err := client.CreatePersistentSubscriptionToAll(context.Background(), groupName, kurrentdb.PersistentAllSubscriptionOptions{})

		if err, ok := kurrentdb.FromError(err); !ok {
			if err.Code() == kurrentdb.ErrorCodeUnsupportedFeature {
				t.Skip()
			}
		}

		require.NoError(t, err)

		sub, err := client.SubscribeToPersistentSubscriptionToAll(context.Background(), groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{})
		defer sub.Close()

		require.NoError(t, err)

		fixture.CreateTestEvents(streamName, eventCount)

		require.NoError(t, err)

		i := uint32(0)
		for i < eventCount {
			event := sub.Recv()

			if event.SubscriptionDropped != nil {
				t.FailNow()
			}

			if event.EventAppeared != nil {
				if event.EventAppeared.Event.OriginalEvent().StreamID == streamName {
					err = sub.Nack("because reasons", kurrentdb.NackActionPark, event.EventAppeared.Event)
					require.NoError(t, err)
					i++
				} else {
					err = sub.Ack(event.EventAppeared.Event)
					require.NoError(t, err)
				}
			}
		}

		time.Sleep(5 * time.Second)

		err = client.ReplayParkedMessagesToAll(context.Background(), groupName, kurrentdb.ReplayParkedMessagesOptions{})

		require.NoError(t, err)

		i = 0
		for i < eventCount {
			event := sub.Recv()

			if event.SubscriptionDropped != nil {
				t.FailNow()
			}

			if event.EventAppeared != nil {
				err = sub.Ack(event.EventAppeared.Event)
				require.NoError(t, err)

				if event.EventAppeared.Event.Event.StreamID == streamName {
					i++
				}
			}
		}

		require.NoError(t, err)
	})

	t.Run("persistentReplayParkedMessagesWithCredentialsOverride", func(t *testing.T) {
		streamId := fixture.NewStreamId()
		fixture.CreateTestEvents(streamId, 1)
	})

	t.Run("persistentReplayParkedMessagesToAll", func(t *testing.T) {
		streamId := fixture.NewStreamId()
		fixture.CreateTestEvents(streamId, 1)
	})

	t.Run("persistentListSubsForStream", func(t *testing.T) {
		streamName := fixture.NewStreamId()
		groupNames := make(map[string]int)
		count := 2

		for i := 0; i < count; i++ {
			name := fixture.NewGroupId()
			err := client.CreatePersistentSubscription(context.Background(), streamName, name, kurrentdb.PersistentStreamSubscriptionOptions{})
			require.NoError(t, err)
			groupNames[name] = 0
		}

		subs, err := client.ListPersistentSubscriptionsForStream(context.Background(), streamName, kurrentdb.ListPersistentSubscriptionsOptions{})

		require.NoError(t, err)
		require.NotNil(t, subs)
		require.NotEmpty(t, subs)

		found := 0
		for i := range subs {
			if subs[i].EventSource == streamName {
				if _, exists := groupNames[subs[i].GroupName]; exists {
					found++
				}
			}
		}

		require.Equal(t, 2, found)
	})

	t.Run("persistentListSubsToAll", func(t *testing.T) {
		groupNames := make(map[string]int)
		count := 2

		for i := 0; i < count; i++ {
			name := fixture.NewGroupId()
			err := client.CreatePersistentSubscriptionToAll(context.Background(), name, kurrentdb.PersistentAllSubscriptionOptions{})

			if err, ok := kurrentdb.FromError(err); !ok {
				if err.Code() == kurrentdb.ErrorCodeUnsupportedFeature {
					t.Skip()
				}
			}

			require.NoError(t, err)
			groupNames[name] = 0
		}

		subs, err := client.ListPersistentSubscriptionsToAll(context.Background(), kurrentdb.ListPersistentSubscriptionsOptions{})

		require.NoError(t, err)
		require.NotNil(t, subs)
		require.NotEmpty(t, subs)

		found := 0
		for i := range subs {
			if _, exists := groupNames[subs[i].GroupName]; exists {
				found++
			}
		}

		require.Equal(t, 2, found)
	})

	t.Run("persistentGetInfo", func(t *testing.T) {
		streamName := fixture.NewStreamId()
		groupName := fixture.NewGroupId()

		setts := kurrentdb.SubscriptionSettingsDefault()
		setts.CheckpointLowerBound = 1
		setts.CheckpointUpperBound = 1

		var events []kurrentdb.EventData

		for i := 0; i < 50; i++ {
			events = append(events, fixture.CreateTestEvent())
		}

		_, err := client.AppendToStream(context.Background(), streamName, kurrentdb.AppendToStreamOptions{}, events...)

		require.NoError(t, err)

		err = client.CreatePersistentSubscription(context.Background(), streamName, groupName, kurrentdb.PersistentStreamSubscriptionOptions{
			StartFrom: kurrentdb.Start{},
			Settings:  &setts,
		})

		require.NoError(t, err)

		var receivedEvents sync.WaitGroup

		subscription, err := client.SubscribeToPersistentSubscription(
			context.Background(), streamName, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{
				BufferSize: 2,
			})

		require.NoError(t, err)

		go func() {
			current := 1

			for {
				subEvent := subscription.Recv()

				if subEvent.EventAppeared != nil {
					current++

					subscription.Ack(subEvent.EventAppeared.Event)

					if current >= 10 {
						receivedEvents.Done()
						break
					}

					continue
				}
			}
		}()

		require.NoError(t, err)
		receivedEvents.Add(1)
		timedOut := fixture.WaitWithTimeout(&receivedEvents, 5*time.Second)
		require.False(t, timedOut, "Timed out waiting for initial set of events")
		info, err := client.GetPersistentSubscriptionInfo(context.Background(), streamName, groupName, kurrentdb.GetPersistentSubscriptionOptions{})
		require.NoError(t, err)
		require.Equal(t, streamName, info.EventSource)
		require.Equal(t, groupName, info.GroupName)
		subscription.Close()
	})

	t.Run("persistentGetInfoWithCredentialsOverride", func(t *testing.T) {
		streamName := fixture.NewStreamId()
		groupName := fixture.NewGroupId()

		err := client.CreatePersistentSubscription(context.Background(), streamName, groupName, kurrentdb.PersistentStreamSubscriptionOptions{})

		require.NoError(t, err)

		opts := kurrentdb.GetPersistentSubscriptionOptions{
			Authenticated: &kurrentdb.Credentials{
				Login:    "admin",
				Password: "changeit",
			},
		}
		sub, err := client.GetPersistentSubscriptionInfo(context.Background(), streamName, groupName, opts)

		require.NoError(t, err)

		require.Equal(t, streamName, sub.EventSource)
		require.Equal(t, groupName, sub.GroupName)
	})

	t.Run("persistentGetInfoToAll", func(t *testing.T) {
		groupName := fixture.NewGroupId()

		err := client.CreatePersistentSubscriptionToAll(context.Background(), groupName, kurrentdb.PersistentAllSubscriptionOptions{})

		if err, ok := kurrentdb.FromError(err); !ok {
			if err.Code() == kurrentdb.ErrorCodeUnsupportedFeature {
				t.Skip()
			}
		}

		require.NoError(t, err)

		sub, err := client.GetPersistentSubscriptionInfoToAll(context.Background(), groupName, kurrentdb.GetPersistentSubscriptionOptions{})

		require.NoError(t, err)

		require.Equal(t, groupName, sub.GroupName)
	})

	t.Run("persistentGetInfoEncoding", func(t *testing.T) {
		streamName := fmt.Sprintf("/%s/foo", fixture.NewStreamId())
		groupName := fmt.Sprintf("/%s/foo", fixture.NewGroupId())

		err := client.CreatePersistentSubscription(context.Background(), streamName, groupName, kurrentdb.PersistentStreamSubscriptionOptions{})

		require.NoError(t, err)

		sub, err := client.GetPersistentSubscriptionInfo(context.Background(), streamName, groupName, kurrentdb.GetPersistentSubscriptionOptions{})

		require.NoError(t, err)

		require.Equal(t, streamName, sub.EventSource)
		require.Equal(t, groupName, sub.GroupName)
	})

	t.Run("persistentRestartSubsystem", func(t *testing.T) {
		err := client.RestartPersistentSubscriptionSubsystem(context.Background(), kurrentdb.RestartPersistentSubscriptionSubsystemOptions{})
		require.NoError(t, err)
	})
}
