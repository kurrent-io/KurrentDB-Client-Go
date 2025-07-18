package test

import (
	"context"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

func TestPersistentSubscriptionReadSuite(t *testing.T) {
	suite.Run(t, new(PersistentSubReadTestSuite))
}

type PersistentSubReadTestSuite struct {
	suite.Suite
	fixture *ClientFixture
}

func (s *PersistentSubReadTestSuite) SetupTest() {
	s.fixture = NewSecureSingleNodeClientFixture(s.T())
}

func (s *PersistentSubReadTestSuite) TearDownTest() {
	s.fixture.Close(s.T())
}

func (s *PersistentSubReadTestSuite) TestReadExistingStream_AckToReceiveNewEvents() {
	streamId := s.fixture.NewStreamId()
	firstEvent := s.fixture.CreateTestEvent()
	secondEvent := s.fixture.CreateTestEvent()
	thirdEvent := s.fixture.CreateTestEvent()
	events := []kurrentdb.EventData{firstEvent, secondEvent, thirdEvent}

	_, err := s.fixture.client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{}, events...)
	s.Require().NoError(err)

	groupName := s.fixture.NewGroupId()
	err = s.fixture.client.CreatePersistentSubscription(
		context.Background(),
		streamId,
		groupName,
		kurrentdb.PersistentStreamSubscriptionOptions{
			StartFrom: kurrentdb.Start{},
		},
	)
	s.Require().NoError(err)

	readConn, err := s.fixture.client.SubscribeToPersistentSubscription(
		context.Background(), streamId, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{
			BufferSize: 2,
		})
	s.Require().NoError(err)
	defer readConn.Close()

	firstRead := readConn.Recv().EventAppeared.Event
	s.Require().NotNil(firstRead)
	s.Require().Equal(firstEvent.EventID, firstRead.OriginalEvent().EventID)

	secondRead := readConn.Recv().EventAppeared.Event
	s.Require().NotNil(secondRead)
	s.Require().Equal(secondEvent.EventID, secondRead.OriginalEvent().EventID)

	err = readConn.Ack(firstRead)
	s.Require().NoError(err)

	thirdRead := readConn.Recv()
	s.Require().NotNil(thirdRead.EventAppeared)
	s.Require().Equal(thirdEvent.EventID, thirdRead.EventAppeared.Event.OriginalEvent().EventID)
}

func (s *PersistentSubReadTestSuite) TestToExistingStream_StartFromBeginning_AndEventsInIt() {
	streamId := s.fixture.NewStreamId()
	events := make([]kurrentdb.EventData, 10)
	for i := 0; i < 10; i++ {
		events[i] = s.fixture.CreateTestEvent()
	}

	_, err := s.fixture.client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
		StreamState: kurrentdb.NoStream{},
	}, events...)
	s.Require().NoError(err)

	groupName := s.fixture.NewGroupId()
	err = s.fixture.client.CreatePersistentSubscription(
		context.Background(),
		streamId,
		groupName,
		kurrentdb.PersistentStreamSubscriptionOptions{
			StartFrom: kurrentdb.Start{},
		},
	)
	s.Require().NoError(err)

	readConn, err := s.fixture.client.SubscribeToPersistentSubscription(
		context.Background(), streamId, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{})
	s.Require().NoError(err)
	defer readConn.Close()

	recv := readConn.Recv()
	s.Require().NotNil(recv.EventAppeared, "EventAppeared should not be nil")
	readEvent := recv.EventAppeared.Event
	s.Require().Equal(events[0].EventID, readEvent.OriginalEvent().EventID)
	s.Require().EqualValues(0, readEvent.OriginalEvent().EventNumber)
}

func (s *PersistentSubReadTestSuite) TestToNonExistingStream_StartFromBeginning_AppendEventsAfterwards() {
	streamId := s.fixture.NewStreamId()
	groupName := s.fixture.NewGroupId()
	err := s.fixture.client.CreatePersistentSubscription(
		context.Background(),
		streamId,
		groupName,
		kurrentdb.PersistentStreamSubscriptionOptions{
			StartFrom: kurrentdb.Start{},
		},
	)
	s.Require().NoError(err)

	events := make([]kurrentdb.EventData, 10)
	for i := 0; i < 10; i++ {
		events[i] = s.fixture.CreateTestEvent()
	}

	_, err = s.fixture.client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
		StreamState: kurrentdb.NoStream{},
	}, events...)
	s.Require().NoError(err)

	readConn, err := s.fixture.client.SubscribeToPersistentSubscription(
		context.Background(), streamId, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{})
	s.Require().NoError(err)
	defer readConn.Close()

	recv := readConn.Recv()
	s.Require().NotNil(recv.EventAppeared, "EventAppeared should not be nil")
	readEvent := recv.EventAppeared.Event
	s.Require().Equal(events[0].EventID, readEvent.OriginalEvent().EventID)
	s.Require().EqualValues(0, readEvent.OriginalEvent().EventNumber)
}

func (s *PersistentSubReadTestSuite) TestToExistingStream_StartFromEnd_EventsInItAndAppendEventsAfterwards() {
	streamId := s.fixture.NewStreamId()
	events := make([]kurrentdb.EventData, 11)
	for i := 0; i < 11; i++ {
		events[i] = s.fixture.CreateTestEvent()
	}

	_, err := s.fixture.client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
		StreamState: kurrentdb.NoStream{},
	}, events[:10]...)
	s.Require().NoError(err)

	groupName := s.fixture.NewGroupId()
	err = s.fixture.client.CreatePersistentSubscription(
		context.Background(),
		streamId,
		groupName,
		kurrentdb.PersistentStreamSubscriptionOptions{
			StartFrom: kurrentdb.End{},
		},
	)
	s.Require().NoError(err)

	_, err = s.fixture.client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
		StreamState: kurrentdb.Revision(9),
	}, events[10])
	s.Require().NoError(err)

	readConn, err := s.fixture.client.SubscribeToPersistentSubscription(
		context.Background(), streamId, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{})
	s.Require().NoError(err)
	defer readConn.Close()

	recv := readConn.Recv()
	s.Require().NotNil(recv.EventAppeared, "EventAppeared should not be nil")
	readEvent := recv.EventAppeared.Event
	s.Require().Equal(events[10].EventID, readEvent.OriginalEvent().EventID)
	s.Require().EqualValues(10, readEvent.OriginalEvent().EventNumber)
}

func (s *PersistentSubReadTestSuite) TestToExistingStream_StartFromEnd_EventsInIt() {
	streamId := s.fixture.NewStreamId()
	events := make([]kurrentdb.EventData, 10)
	for i := 0; i < 10; i++ {
		events[i] = s.fixture.CreateTestEvent()
	}

	_, err := s.fixture.client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
		StreamState: kurrentdb.NoStream{},
	}, events...)
	s.Require().NoError(err)

	groupName := s.fixture.NewGroupId()
	err = s.fixture.client.CreatePersistentSubscription(
		context.Background(),
		streamId,
		groupName,
		kurrentdb.PersistentStreamSubscriptionOptions{
			StartFrom: kurrentdb.End{},
		},
	)
	s.Require().NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	readConn, err := s.fixture.client.SubscribeToPersistentSubscription(
		ctx, streamId, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{})
	s.Require().NoError(err)
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
		s.Require().Fail("Received unexpected event")
	}
}

func (s *PersistentSubReadTestSuite) TestToNonExistingStream_StartFromTwo_AppendEventsAfterwards() {
	streamId := s.fixture.NewStreamId()
	groupName := s.fixture.NewGroupId()
	err := s.fixture.client.CreatePersistentSubscription(
		context.Background(),
		streamId,
		groupName,
		kurrentdb.PersistentStreamSubscriptionOptions{
			StartFrom: kurrentdb.Revision(2),
		},
	)
	s.Require().NoError(err)

	events := make([]kurrentdb.EventData, 4)
	for i := 0; i < 4; i++ {
		events[i] = s.fixture.CreateTestEvent()
	}

	_, err = s.fixture.client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
		StreamState: kurrentdb.NoStream{},
	}, events...)
	s.Require().NoError(err)

	readConn, err := s.fixture.client.SubscribeToPersistentSubscription(
		context.Background(), streamId, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{})
	s.Require().NoError(err)
	defer readConn.Close()

	recv := readConn.Recv()
	s.Require().NotNil(recv.EventAppeared, "EventAppeared should not be nil")
	readEvent := recv.EventAppeared.Event
	s.Require().Equal(events[2].EventID, readEvent.OriginalEvent().EventID)
	s.Require().EqualValues(2, readEvent.OriginalEvent().EventNumber)
}

func (s *PersistentSubReadTestSuite) TestToExistingStream_StartFrom10_EventsInItAppendEventsAfterwards() {
	streamId := s.fixture.NewStreamId()
	events := make([]kurrentdb.EventData, 11)
	for i := 0; i < 11; i++ {
		events[i] = s.fixture.CreateTestEvent()
	}

	_, err := s.fixture.client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
		StreamState: kurrentdb.NoStream{},
	}, events[:10]...)
	s.Require().NoError(err)

	groupName := s.fixture.NewGroupId()
	err = s.fixture.client.CreatePersistentSubscription(
		context.Background(),
		streamId,
		groupName,
		kurrentdb.PersistentStreamSubscriptionOptions{
			StartFrom: kurrentdb.Revision(10),
		},
	)
	s.Require().NoError(err)

	_, err = s.fixture.client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
		StreamState: kurrentdb.Revision(9),
	}, events[10])
	s.Require().NoError(err)

	readConn, err := s.fixture.client.SubscribeToPersistentSubscription(
		context.Background(), streamId, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{})
	s.Require().NoError(err)
	defer readConn.Close()

	recv := readConn.Recv()
	s.Require().NotNil(recv.EventAppeared, "EventAppeared should not be nil")
	readEvent := recv.EventAppeared.Event
	s.Require().Equal(events[10].EventID, readEvent.OriginalEvent().EventID)
	s.Require().EqualValues(10, readEvent.OriginalEvent().EventNumber)
}

func (s *PersistentSubReadTestSuite) TestToExistingStream_StartFrom4_EventsInIt() {
	streamId := s.fixture.NewStreamId()
	events := make([]kurrentdb.EventData, 11)
	for i := 0; i < 11; i++ {
		events[i] = s.fixture.CreateTestEvent()
	}

	_, err := s.fixture.client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
		StreamState: kurrentdb.NoStream{},
	}, events[:10]...)
	s.Require().NoError(err)

	groupName := s.fixture.NewGroupId()
	err = s.fixture.client.CreatePersistentSubscription(
		context.Background(),
		streamId,
		groupName,
		kurrentdb.PersistentStreamSubscriptionOptions{
			StartFrom: kurrentdb.Revision(4),
		},
	)
	s.Require().NoError(err)

	_, err = s.fixture.client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
		StreamState: kurrentdb.Revision(9),
	}, events[10])
	s.Require().NoError(err)

	readConn, err := s.fixture.client.SubscribeToPersistentSubscription(
		context.Background(), streamId, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{})
	s.Require().NoError(err)
	defer readConn.Close()

	recv := readConn.Recv()
	s.Require().NotNil(recv.EventAppeared, "EventAppeared should not be nil")
	readEvent := recv.EventAppeared.Event
	s.Require().Equal(events[4].EventID, readEvent.OriginalEvent().EventID)
	s.Require().EqualValues(4, readEvent.OriginalEvent().EventNumber)
}

func (s *PersistentSubReadTestSuite) TestToExistingStream_StartFromHigherRevisionThenEventsInStream_EventsInItAppendEventsAfterwards() {
	streamId := s.fixture.NewStreamId()
	events := make([]kurrentdb.EventData, 12)
	for i := 0; i < 12; i++ {
		events[i] = s.fixture.CreateTestEvent()
	}

	_, err := s.fixture.client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
		StreamState: kurrentdb.NoStream{},
	}, events[:11]...)
	s.Require().NoError(err)

	groupName := s.fixture.NewGroupId()
	err = s.fixture.client.CreatePersistentSubscription(
		context.Background(),
		streamId,
		groupName,
		kurrentdb.PersistentStreamSubscriptionOptions{
			StartFrom: kurrentdb.Revision(11),
		},
	)
	s.Require().NoError(err)

	_, err = s.fixture.client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{
		StreamState: kurrentdb.Revision(10),
	}, events[11])
	s.Require().NoError(err)

	readConn, err := s.fixture.client.SubscribeToPersistentSubscription(
		context.Background(), streamId, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{})
	s.Require().NoError(err)
	defer readConn.Close()

	recv := readConn.Recv()
	s.Require().NotNil(recv.EventAppeared, "EventAppeared should not be nil")
	readEvent := recv.EventAppeared.Event
	s.Require().Equal(events[11].EventID, readEvent.OriginalEvent().EventID)
	s.Require().EqualValues(11, readEvent.OriginalEvent().EventNumber)
}

func (s *PersistentSubReadTestSuite) TestReadExistingStream_NackToReceiveNewEvents() {
	streamId := s.fixture.NewStreamId()
	firstEvent := s.fixture.CreateTestEvent()
	secondEvent := s.fixture.CreateTestEvent()
	thirdEvent := s.fixture.CreateTestEvent()
	events := []kurrentdb.EventData{firstEvent, secondEvent, thirdEvent}

	_, err := s.fixture.client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{}, events...)
	s.Require().NoError(err)

	groupName := s.fixture.NewGroupId()
	err = s.fixture.client.CreatePersistentSubscription(
		context.Background(),
		streamId,
		groupName,
		kurrentdb.PersistentStreamSubscriptionOptions{
			StartFrom: kurrentdb.Start{},
		},
	)
	s.Require().NoError(err)

	readConn, err := s.fixture.client.SubscribeToPersistentSubscription(
		context.Background(), streamId, groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{
			BufferSize: 2,
		})
	s.Require().NoError(err)
	defer readConn.Close()

	firstRead := readConn.Recv().EventAppeared.Event
	s.Require().NotNil(firstRead)
	s.Require().Equal(firstEvent.EventID, firstRead.OriginalEvent().EventID)

	secondRead := readConn.Recv().EventAppeared.Event
	s.Require().NotNil(secondRead)
	s.Require().Equal(secondEvent.EventID, secondRead.OriginalEvent().EventID)

	err = readConn.Nack("test reason", kurrentdb.NackActionPark, firstRead)
	s.Require().NoError(err)

	thirdRead := readConn.Recv()
	s.Require().NotNil(thirdRead.EventAppeared)
	s.Require().Equal(thirdEvent.EventID, thirdRead.EventAppeared.Event.OriginalEvent().EventID)
}

func (s *PersistentSubReadTestSuite) TestPersistentSubscriptionToAll_Read() {
	streamId := s.fixture.NewStreamId()
	_ = s.fixture.CreateTestEvents(streamId, 3)

	groupName := s.fixture.NewGroupId()
	err := s.fixture.client.CreatePersistentSubscriptionToAll(
		context.Background(),
		groupName,
		kurrentdb.PersistentAllSubscriptionOptions{
			StartFrom: kurrentdb.Start{},
		},
	)

	if err != nil {
		if convertedErr, ok := kurrentdb.FromError(err); ok && convertedErr.Code() == kurrentdb.ErrorCodeUnsupportedFeature && s.fixture.IsKurrentDbVersion20() {
			s.T().Skip()
		}
	}
	s.Require().NoError(err)

	readConnectionClient, err := s.fixture.client.SubscribeToPersistentSubscriptionToAll(
		context.Background(), groupName, kurrentdb.SubscribeToPersistentSubscriptionOptions{
			BufferSize: 2,
		},
	)
	s.Require().NoError(err)
	defer readConnectionClient.Close()

	// First event
	event1 := readConnectionClient.Recv()
	s.Require().NoError(err)
	s.Require().NotNil(event1.EventAppeared, "EventAppeared should not be nil")
	firstReadEvent := event1.EventAppeared.Event
	s.Require().NotNil(firstReadEvent)

	// Second event
	event2 := readConnectionClient.Recv()
	s.Require().NoError(err)
	s.Require().NotNil(event2.EventAppeared, "EventAppeared should not be nil")
	secondReadEvent := event2.EventAppeared.Event
	s.Require().NotNil(secondReadEvent)

	// Ack the first event
	err = readConnectionClient.Ack(firstReadEvent)
	s.Require().NoError(err)

	// Third event
	event3 := readConnectionClient.Recv()
	s.Require().NoError(err)
	s.Require().NotNil(event3.EventAppeared, "EventAppeared should not be nil")
	thirdReadEvent := event3.EventAppeared.Event
	s.Require().NotNil(thirdReadEvent)

	// Ack the third event
	err = readConnectionClient.Ack(thirdReadEvent)
	s.Require().NoError(err)
}
