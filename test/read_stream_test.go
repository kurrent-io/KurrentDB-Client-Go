package test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestReadStreamSuite(t *testing.T) {
	suite.Run(t, new(ReadStreamTestSuite))
}

type ReadStreamTestSuite struct {
	suite.Suite
	fixture *ClientFixture
}

func (s *ReadStreamTestSuite) SetupSuite() {
	s.fixture = NewSecureSingleNodeClientFixture(s.T())
}

func (s *ReadStreamTestSuite) TearDownSuite() {
	s.fixture.Close(s.T())
}

func (s *ReadStreamTestSuite) TestReadStreamEventsForwardFromZero() {
	fixture := s.fixture

	client := fixture.Client()

	// Arrange
	streamId := fixture.NewStreamId()
	testEvents := fixture.CreateTestEvents(streamId, 10)
	opts := kurrentdb.ReadStreamOptions{
		Direction:      kurrentdb.Forwards,
		ResolveLinkTos: true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Act
	stream, err := client.ReadStream(ctx, streamId, opts, uint64(len(testEvents)))
	assert.NoError(s.T(), err, "Unexpected failure")
	defer stream.Close()

	// Assert
	events, err := fixture.CollectEvents(stream)
	assert.NoError(s.T(), err, "Unexpected failure when collecting events")
	assert.Equal(s.T(), 10, len(events), "Expected the correct number of messages to be returned")
	assert.Equal(s.T(), testEvents[0].EventID, events[0].OriginalEvent().EventID)
}

func (s *ReadStreamTestSuite) TestReadStreamEventsBackwardFromEnd() {
	fixture := s.fixture

	client := fixture.Client()

	// Arrange
	streamId := fixture.NewStreamId()
	testEvents := fixture.CreateTestEvents(streamId, 10)
	opts := kurrentdb.ReadStreamOptions{
		Direction:      kurrentdb.Backwards,
		From:           kurrentdb.End{},
		ResolveLinkTos: true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Act
	stream, err := client.ReadStream(ctx, streamId, opts, uint64(len(testEvents)))
	assert.NoError(s.T(), err, "Unexpected failure")
	defer stream.Close()

	// Assert
	events, err := fixture.CollectEvents(stream)
	assert.NoError(s.T(), err, "Unexpected failure when collecting events")
	assert.Equal(s.T(), 10, len(events), "Expected the correct number of messages to be returned")
	assert.Equal(s.T(), testEvents[len(testEvents)-1].EventID, events[0].OriginalEvent().EventID,
		"Last event appended should be the first event read")
}

func (s *ReadStreamTestSuite) TestReadStreamReturnsEOFAfterCompletion() {
	fixture := s.fixture

	client := fixture.Client()

	// Arrange
	streamId := fixture.NewStreamId()
	fixture.CreateTestEvents(streamId, 10)

	// Act
	stream, err := client.ReadStream(context.Background(), streamId, kurrentdb.ReadStreamOptions{}, 1_024)
	require.NoError(s.T(), err)
	defer stream.Close()

	_, err = fixture.CollectEvents(stream)
	require.NoError(s.T(), err)

	// Assert
	_, err = stream.Recv()
	require.Error(s.T(), err)
	require.True(s.T(), err == io.EOF)
}

func (s *ReadStreamTestSuite) TestReadStreamNotFound() {
	fixture := s.fixture

	client := fixture.Client()

	// Arrange
	streamId := fixture.NewStreamId()

	// Act
	stream, err := client.ReadStream(context.Background(), streamId, kurrentdb.ReadStreamOptions{}, 1)
	require.NoError(s.T(), err)
	defer stream.Close()

	// Assert
	evt, err := stream.Recv()
	require.Nil(s.T(), evt)

	kurrentDbError, ok := kurrentdb.FromError(err)
	require.False(s.T(), ok)
	require.Equal(s.T(), kurrentDbError.Code(), kurrentdb.ErrorCodeResourceNotFound)
}

func (s *ReadStreamTestSuite) TestReadStreamWithMaxAge() {
	fixture := s.fixture

	client := fixture.Client()

	// Arrange
	streamId := fixture.NewStreamId()
	fixture.CreateTestEvents(streamId, 10)

	metadata := kurrentdb.StreamMetadata{}
	metadata.SetMaxAge(time.Second)

	_, err := client.SetStreamMetadata(context.Background(), streamId, kurrentdb.AppendToStreamOptions{}, metadata)
	assert.NoError(s.T(), err)

	// Wait for events to expire
	time.Sleep(2 * time.Second)

	// Act
	stream, err := client.ReadStream(context.Background(), streamId, kurrentdb.ReadStreamOptions{}, 10)
	require.NoError(s.T(), err)
	defer stream.Close()

	// Assert
	evt, err := stream.Recv()
	require.Nil(s.T(), evt)
	require.Error(s.T(), err)
	require.True(s.T(), errors.Is(err, io.EOF))
}

func (s *ReadStreamTestSuite) TestReadStreamWithCredentialsOverride() {
	fixture := s.fixture

	client := fixture.Client()

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
	assert.NoError(s.T(), err)

	streamId = fixture.NewStreamId()
	opts.Authenticated.Password = "invalid"
	_, err = client.AppendToStream(context.Background(), streamId, opts, fixture.CreateTestEvent())
	assert.Error(s.T(), err)
}
