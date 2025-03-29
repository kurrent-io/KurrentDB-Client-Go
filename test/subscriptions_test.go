package test_test

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

func TestSubscriptionSuite(t *testing.T) {
	suite.Run(t, new(SubscriptionTestSuite))
}

type SubscriptionTestSuite struct {
	suite.Suite
	fixture *ClientFixture
}

func (s *SubscriptionTestSuite) SetupSuite() {
	s.fixture = NewSecureSingleNodeClientFixture(s.T())
}

func (s *SubscriptionTestSuite) TearDownSuite() {
	s.fixture.Close(s.T())
}

func (s *SubscriptionTestSuite) TestStreamSubscriptionDeliversAllEventsInStreamAndListensForNewEvents() {
	fixture := s.fixture
	client := fixture.Client()

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

	s.NoError(err)
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
				s.Equal(appendEvent.EventID, event.OriginalEvent().EventID)
				s.Equal(uint64(10), event.OriginalEvent().EventNumber)
				s.Equal(streamId, event.OriginalEvent().StreamID)
				s.Equal(appendEvent.Data, event.OriginalEvent().Data)
				s.Equal(appendEvent.Metadata, event.OriginalEvent().UserMetadata)
				break
			}
		}
		appendedEvents.Done()
	}()

	timedOut := fixture.WaitWithTimeout(&receivedEvents, time.Duration(10)*time.Second)
	s.False(timedOut, "Timed out waiting for initial set of events")

	// Write a new event
	opts2 := kurrentdb.AppendToStreamOptions{
		StreamState: kurrentdb.Revision(9),
	}
	writeResult, err := client.AppendToStream(context.Background(), streamId, opts2, appendEvent)
	s.NoError(err)
	s.Equal(uint64(10), writeResult.NextExpectedVersion)

	// Assert event was forwarded to the subscription
	timedOut = fixture.WaitWithTimeout(&appendedEvents, time.Duration(10)*time.Second)
	s.False(timedOut, "Timed out waiting for the appended events")
}

func (s *SubscriptionTestSuite) TestAllSubscriptionWithFilterDeliversCorrectEvents() {
	fixture := s.fixture
	client := fixture.Client()

	// Arrange
	var receivedEvents sync.WaitGroup
	receivedEvents.Add(1)

	prefix := uuid.New().String() + "-"

	streamId := fixture.NewStreamId()

	for i := 0; i < 10; i++ {
		event := fixture.CreateTestEvent()
		event.EventType = prefix + event.EventType
		_, err := client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{}, event)
		s.NoError(err)
	}

	// Act
	subscription, err := client.SubscribeToAll(context.Background(), kurrentdb.SubscribeToAllOptions{
		From: kurrentdb.Start{},
		Filter: &kurrentdb.SubscriptionFilter{
			Type:     kurrentdb.EventFilterType,
			Prefixes: []string{prefix},
		},
	})

	defer subscription.Close()

	// Assert
	go func() {
		current := 0
		for {
			subEvent := subscription.Recv()

			if subEvent.SubscriptionDropped != nil {
				break

			} else if subEvent.EventAppeared != nil {
				event := subEvent.EventAppeared
				s.True(strings.HasPrefix(event.OriginalEvent().EventType, prefix))
				current++
				if current == 10 {
					receivedEvents.Done()
					break

				}
			}
		}
	}()

	s.NoError(err)
	timedOut := fixture.WaitWithTimeout(&receivedEvents, time.Duration(10)*time.Second)
	s.False(timedOut, "Timed out while waiting for events via the subscription")
}

func (s *SubscriptionTestSuite) TestSubscribeToAllFilter() {
	fixture := s.fixture
	client := fixture.Client()

	sub, err := client.SubscribeToAll(context.Background(), kurrentdb.SubscribeToAllOptions{
		From:   kurrentdb.Start{},
		Filter: kurrentdb.ExcludeSystemEventsFilter(),
	})

	s.NoError(err)

	var completed sync.WaitGroup
	completed.Add(1)

	go func() {
		for {
			event := sub.Recv()

			if event.EventAppeared != nil {
				if strings.HasPrefix(event.EventAppeared.OriginalEvent().EventType, "$") {
					s.Fail("We should not have system events!")
					return
				}

				completed.Done()
				break
			}
		}
	}()

	timedOut := fixture.WaitWithTimeout(&completed, time.Duration(30)*time.Second)
	s.False(timedOut, "Timed out waiting for filtered subscription completion")
}

func (s *SubscriptionTestSuite) TestConnectionClosing() {
	fixture := s.fixture
	client := fixture.Client()

	var droppedEvent sync.WaitGroup

	streamId := fixture.NewStreamId()
	fixture.CreateTestEvents(streamId, 20)

	subscription, err := client.SubscribeToStream(context.Background(), streamId, kurrentdb.SubscribeToStreamOptions{
		From: kurrentdb.Start{},
	})
	s.NoError(err)

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

	droppedEvent.Add(1)
	timedOut := fixture.WaitWithTimeout(&droppedEvent, time.Duration(10)*time.Second)
	s.False(timedOut, "Timed out waiting for dropped event")
}

func (s *SubscriptionTestSuite) TestSubscriptionToAllWithCredentialsOverride() {
	fixture := s.fixture
	client := fixture.Client()

	opts := kurrentdb.SubscribeToAllOptions{
		Authenticated: &kurrentdb.Credentials{
			Login:    "admin",
			Password: "changeit",
		},
		From:   kurrentdb.Start{},
		Filter: kurrentdb.ExcludeSystemEventsFilter(),
	}
	_, err := client.SubscribeToAll(context.Background(), opts)

	s.NoError(err)
}

func (s *SubscriptionTestSuite) TestSubscriptionToStreamCaughtUp() {
	fixture := s.fixture
	client := fixture.Client()

	const minSupportedVersion = 23
	const expectedEventCount = 10
	const testTimeout = 1 * time.Minute

	kurrentDbVersion, err := client.GetServerVersion()
	s.NoError(err, "Error getting server version")

	if kurrentDbVersion.Major < minSupportedVersion {
		s.T().Skip("CaughtUp message is not supported in this version of KurrentDB")
	}

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	streamId := fixture.NewStreamId()
	fixture.CreateTestEvents(streamId, 10)

	subscription, err := client.SubscribeToStream(ctx, streamId, kurrentdb.SubscribeToStreamOptions{From: kurrentdb.Start{}})
	s.NoError(err)
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
				s.T().Error("Context timed out before receiving CaughtUp message")
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
					s.True(count >= expectedEventCount, "Did not receive the exact number of expected events before CaughtUp")
					return
				}
			}
		}
	}()

	caughtUpTimedOut := fixture.WaitWithTimeout(&caughtUpReceived, testTimeout)
	s.False(caughtUpTimedOut, "Timed out waiting for CaughtUp message")
}
