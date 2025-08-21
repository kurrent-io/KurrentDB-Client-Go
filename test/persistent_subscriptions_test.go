package test_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestPersistentSubscriptionSuite(t *testing.T) {
	suite.Run(t, new(PersistentSubscriptionSuite))
}

type PersistentSubscriptionSuite struct {
	suite.Suite
	fixture *ClientFixture
	client  *kurrentdb.Client
}

func (s *PersistentSubscriptionSuite) SetupTest() {
	s.fixture = NewSecureSingleNodeClientFixture(s.T())
	s.client = s.fixture.Client()
}

func (s *PersistentSubscriptionSuite) TearDownTest() {
	if s.fixture != nil {
		s.fixture.Close(s.T())
	}
}

func (s *PersistentSubscriptionSuite) TestCreatePersistentStreamSubscription() {
	streamID := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(streamID, 1)

	err := s.client.CreatePersistentSubscription(
		context.Background(),
		streamID,
		"Group 1",
		kurrentdb.PersistentStreamSubscriptionOptions{},
	)

	s.NoError(err)
}

func (s *PersistentSubscriptionSuite) TestCreatePersistentStreamSubscription_MessageTimeoutZero() {
	streamID := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(streamID, 1)

	settings := kurrentdb.SubscriptionSettingsDefault()
	settings.MessageTimeout = 0

	err := s.client.CreatePersistentSubscription(
		context.Background(),
		streamID,
		"Group 1",
		kurrentdb.PersistentStreamSubscriptionOptions{
			Settings: &settings,
		},
	)

	s.NoError(err)
}

func (s *PersistentSubscriptionSuite) TestCreatePersistentStreamSubscription_FailsIfAlreadyExists() {
	streamId := s.fixture.NewStreamId()
	groupId := s.fixture.NewGroupId()
	s.fixture.CreateTestEvents(streamId, 1)

	// First creation should succeed
	err := s.client.CreatePersistentSubscription(
		context.Background(),
		streamId,
		groupId,
		kurrentdb.PersistentStreamSubscriptionOptions{},
	)
	s.NoError(err)

	// Second creation should fail
	err = s.client.CreatePersistentSubscription(
		context.Background(),
		streamId,
		groupId,
		kurrentdb.PersistentStreamSubscriptionOptions{},
	)

	kurrentDbError, ok := kurrentdb.FromError(err)
	require.False(s.T(), ok)
	assert.Equal(s.T(), kurrentdb.ErrorCodeResourceAlreadyExists, kurrentDbError.Code())
}

func (s *PersistentSubscriptionSuite) TestUpdatePersistentStreamSubscription() {
	streamID := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(streamID, 1)

	// Initial creation
	err := s.client.CreatePersistentSubscription(
		context.Background(),
		streamID,
		"Group 1",
		kurrentdb.PersistentStreamSubscriptionOptions{},
	)
	s.NoError(err)

	// Update settings
	settings := kurrentdb.SubscriptionSettingsDefault()
	settings.HistoryBufferSize++

	err = s.client.UpdatePersistentSubscription(
		context.Background(),
		streamID,
		"Group 1",
		kurrentdb.PersistentStreamSubscriptionOptions{
			Settings: &settings,
		},
	)

	s.NoError(err)
}

func (s *PersistentSubscriptionSuite) TestDeletePersistentSubscription() {
	streamID := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(streamID, 1)

	// Create subscription first
	err := s.client.CreatePersistentSubscription(
		context.Background(),
		streamID,
		"Group 1",
		kurrentdb.PersistentStreamSubscriptionOptions{},
	)
	s.NoError(err)

	// Now delete it
	err = s.client.DeletePersistentSubscription(
		context.Background(),
		streamID,
		"Group 1",
		kurrentdb.DeletePersistentSubscriptionOptions{},
	)

	s.NoError(err)
}

func (s *PersistentSubscriptionSuite) TestPersistentSubscriptionLifecycle() {
	streamID := s.fixture.NewStreamId()
	groupName := "test-group"
	s.fixture.CreateTestEvents(streamID, 5)

	// Create
	err := s.client.CreatePersistentSubscription(
		context.Background(),
		streamID,
		groupName,
		kurrentdb.PersistentStreamSubscriptionOptions{},
	)
	s.NoError(err)

	// Update
	settings := kurrentdb.SubscriptionSettingsDefault()
	settings.MaxRetryCount = 10
	err = s.client.UpdatePersistentSubscription(
		context.Background(),
		streamID,
		groupName,
		kurrentdb.PersistentStreamSubscriptionOptions{
			Settings: &settings,
		},
	)
	s.NoError(err)

	// Delete
	err = s.client.DeletePersistentSubscription(
		context.Background(),
		streamID,
		groupName,
		kurrentdb.DeletePersistentSubscriptionOptions{},
	)
	s.NoError(err)
}

func (s *PersistentSubscriptionSuite) TestSubscriptionClosing() {
	streamId := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(streamId, 20)

	groupName := "Group 1"

	err := s.client.CreatePersistentSubscription(context.Background(), streamId, groupName, kurrentdb.PersistentStreamSubscriptionOptions{
		StartFrom: kurrentdb.Start{},
	})

	s.NoError(err)

	var receivedEvents sync.WaitGroup
	var droppedEvent sync.WaitGroup

	subscription, err := s.client.SubscribeToPersistentSubscription(
		context.Background(), streamId, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{
			BufferSize: 2,
		})

	s.NoError(err)

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

	s.NoError(err)
	receivedEvents.Add(10)
	droppedEvent.Add(1)
	timedOut := s.fixture.WaitWithTimeout(&receivedEvents, 5*time.Second)
	s.False(timedOut, "Timed out waiting for initial set of events")
	subscription.Close()
	timedOut = s.fixture.WaitWithTimeout(&droppedEvent, 5*time.Second)
	s.False(timedOut, "Timed out waiting for dropped event")
}

func (s *PersistentSubscriptionSuite) TestPersistentAllCreate() {
	groupName := s.fixture.NewGroupId()

	err := s.client.CreatePersistentSubscriptionToAll(
		context.Background(),
		groupName,
		kurrentdb.PersistentAllSubscriptionOptions{},
	)

	if err != nil {
		if kurrentDbError, ok := kurrentdb.FromError(err); ok && kurrentDbError.Code() == kurrentdb.ErrorCodeUnsupportedFeature {
			s.T().Skip("Feature not supported in this version")
		}
	}
	s.Require().NoError(err)
}

func (s *PersistentSubscriptionSuite) TestPersistentAllCreateWithStrategy() {
	groupName := s.fixture.NewGroupId()

	settings := kurrentdb.SubscriptionSettingsDefault()
	settings.ConsumerStrategyName = kurrentdb.ConsumerStrategyPinnedByCorrelation

	err := s.client.CreatePersistentSubscriptionToAll(
		context.Background(),
		groupName,
		kurrentdb.PersistentAllSubscriptionOptions{
			Settings: &settings,
		},
	)

	if err != nil {
		if kurrentDbError, ok := kurrentdb.FromError(err); ok && kurrentDbError.Code() == kurrentdb.ErrorCodeUnsupportedFeature {
			s.T().Skip("Feature not supported in this version")
		}
	}

	s.Require().NoError(err)

	info, err := s.client.GetPersistentSubscriptionInfoToAll(
		context.Background(),
		groupName,
		kurrentdb.GetPersistentSubscriptionOptions{},
	)
	s.Require().NoError(err)

	s.Require().Equal(kurrentdb.ConsumerStrategyPinnedByCorrelation, info.Settings.ConsumerStrategyName)
}

func (s *PersistentSubscriptionSuite) TestPersistentAllUpdate() {
	groupName := s.fixture.NewGroupId()

	// Create initial subscription
	err := s.client.CreatePersistentSubscriptionToAll(
		context.Background(),
		groupName,
		kurrentdb.PersistentAllSubscriptionOptions{},
	)
	if err != nil {
		if kurrentDbError, ok := kurrentdb.FromError(err); ok && kurrentDbError.Code() == kurrentdb.ErrorCodeUnsupportedFeature {
			s.T().Skip("Feature not supported in this version")
		}
	}
	s.Require().NoError(err)

	// Update settings
	settings := kurrentdb.SubscriptionSettingsDefault()
	settings.ResolveLinkTos = true

	err = s.client.UpdatePersistentSubscriptionToAll(
		context.Background(),
		groupName,
		kurrentdb.PersistentAllSubscriptionOptions{
			Settings: &settings,
		},
	)
	s.Require().NoError(err)
}

func (s *PersistentSubscriptionSuite) TestPersistentAllDelete() {
	groupName := s.fixture.NewGroupId()

	// Create first
	err := s.client.CreatePersistentSubscriptionToAll(
		context.Background(),
		groupName,
		kurrentdb.PersistentAllSubscriptionOptions{},
	)
	if err != nil {
		if kurrentDbError, ok := kurrentdb.FromError(err); ok && kurrentDbError.Code() == kurrentdb.ErrorCodeUnsupportedFeature {
			s.T().Skip("Feature not supported in this version")
		}
	}
	s.Require().NoError(err)

	// Delete
	err = s.client.DeletePersistentSubscriptionToAll(
		context.Background(),
		groupName,
		kurrentdb.DeletePersistentSubscriptionOptions{},
	)
	s.Require().NoError(err)
}

func (s *PersistentSubscriptionSuite) TestPersistentListAllSubs() {
	groupName := s.fixture.NewGroupId()
	streamName := s.fixture.NewStreamId()
	const eventCount = 2

	// Setup subscription
	err := s.client.CreatePersistentSubscription(
		context.Background(),
		streamName,
		groupName,
		kurrentdb.PersistentStreamSubscriptionOptions{},
	)
	s.Require().NoError(err)

	sub, err := s.client.SubscribeToPersistentSubscription(
		context.Background(),
		streamName,
		groupName,
		kurrentdb.SubscribeToPersistentSubscriptionOptions{},
	)
	s.Require().NoError(err)
	defer sub.Close()

	// Create test events
	s.fixture.CreateTestEvents(streamName, eventCount)

	// Process events
	var processed sync.WaitGroup
	processed.Add(eventCount)

	go func() {
		count := 0
		for {
			event := sub.Recv()
			if event.SubscriptionDropped != nil {
				s.Fail("Subscription dropped unexpectedly")
			}

			if event.EventAppeared != nil {
				err := sub.Nack("test reason", kurrentdb.NackActionPark, event.EventAppeared.Event)
				s.Require().NoError(err)
				processed.Done()
				count++
				if count >= eventCount {
					return
				}
			}
		}
	}()

	s.Eventually(func() bool {
		processed.Wait()
		return true
	}, 5*time.Second, 100*time.Millisecond, "Timed out processing events")

	// Replay parked messages
	err = s.client.ReplayParkedMessages(
		context.Background(),
		streamName,
		groupName,
		kurrentdb.ReplayParkedMessagesOptions{},
	)
	s.Require().NoError(err)

	// Verify replay
	var replayed sync.WaitGroup
	replayed.Add(eventCount)

	go func() {
		count := 0
		for {
			event := sub.Recv()
			if event.EventAppeared != nil {
				err := sub.Ack(event.EventAppeared.Event)
				s.Require().NoError(err)
				replayed.Done()
				count++
				if count >= eventCount {
					return
				}
			}
		}
	}()

	s.Eventually(func() bool {
		replayed.Wait()
		return true
	}, 5*time.Second, 100*time.Millisecond, "Timed out waiting for replayed events")
}

func (s *PersistentSubscriptionSuite) TestPersistentListSubsForStream() {
	streamName := s.fixture.NewStreamId()
	expectedGroups := make(map[string]struct{})
	const groupCount = 2

	// Create multiple subscriptions
	for i := 0; i < groupCount; i++ {
		groupName := s.fixture.NewGroupId()
		err := s.client.CreatePersistentSubscription(
			context.Background(),
			streamName,
			groupName,
			kurrentdb.PersistentStreamSubscriptionOptions{},
		)
		s.Require().NoError(err)
		expectedGroups[groupName] = struct{}{}
	}

	// List subscriptions
	subs, err := s.client.ListPersistentSubscriptionsForStream(
		context.Background(),
		streamName,
		kurrentdb.ListPersistentSubscriptionsOptions{},
	)
	s.Require().NoError(err)
	s.Require().NotEmpty(subs)

	// Verify results
	found := 0
	for _, sub := range subs {
		if _, exists := expectedGroups[sub.GroupName]; exists {
			found++
		}
	}
	s.Require().Equal(groupCount, found, "Should find all created subscriptions")
}

func (s *PersistentSubscriptionSuite) TestPersistentGetInfo() {
	streamName := s.fixture.NewStreamId()
	groupName := s.fixture.NewGroupId()

	// Create subscription with custom settings
	settings := kurrentdb.SubscriptionSettingsDefault()
	settings.CheckpointLowerBound = 1
	settings.CheckpointUpperBound = 1

	err := s.client.CreatePersistentSubscription(
		context.Background(),
		streamName,
		groupName,
		kurrentdb.PersistentStreamSubscriptionOptions{
			Settings: &settings,
		},
	)
	s.Require().NoError(err)

	// Get subscription info
	info, err := s.client.GetPersistentSubscriptionInfo(
		context.Background(),
		streamName,
		groupName,
		kurrentdb.GetPersistentSubscriptionOptions{},
	)
	s.Require().NoError(err)

	// Validate response
	s.Require().Equal(streamName, info.EventSource)
	s.Require().Equal(groupName, info.GroupName)
	s.Require().Equal(settings.CheckpointLowerBound, info.Settings.CheckpointLowerBound)
}

func (s *PersistentSubscriptionSuite) TestPersistentRestartSubsystem() {
	err := s.client.RestartPersistentSubscriptionSubsystem(
		context.Background(),
		kurrentdb.RestartPersistentSubscriptionSubsystemOptions{},
	)
	s.Require().NoError(err)
}
