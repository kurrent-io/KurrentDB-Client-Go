package test_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestAppendEventsSuite(t *testing.T) {
	suite.Run(t, new(AppendTestSuite))
}

type AppendTestSuite struct {
	suite.Suite
	fixture *ClientFixture
}

func (s *AppendTestSuite) SetupTest() {
	s.fixture = NewSecureSingleNodeClientFixture(s.T())
}

func (s *AppendTestSuite) TestAppendToStreamSingleEventNoStream() {
	client := s.fixture.Client()

	// Arrange
	testEvent := s.fixture.CreateTestEvent()
	testEvent.EventID = uuid.New()
	streamId := s.fixture.NewStreamId()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	opts := kurrentdb.AppendToStreamOptions{StreamState: kurrentdb.NoStream{}}

	// Act
	_, err := client.AppendToStream(ctx, streamId, opts, testEvent)
	assert.NoError(s.T(), err, "Unexpected failure when appending to stream")

	stream, err := client.ReadStream(ctx, streamId, kurrentdb.ReadStreamOptions{}, 1)
	assert.NoError(s.T(), err, "Unexpected failure when reading stream")
	defer stream.Close()

	events, err := s.fixture.CollectEvents(stream)
	assert.NoError(s.T(), err, "Unexpected failure when collecting events")

	// Assert
	assert.Len(s.T(), events, 1, "Expected a single event")
	assert.Equal(s.T(), testEvent.EventID, events[0].OriginalEvent().EventID)
	assert.Equal(s.T(), testEvent.EventType, events[0].OriginalEvent().EventType)
	assert.Equal(s.T(), streamId, events[0].OriginalEvent().StreamID)
	assert.Equal(s.T(), testEvent.Data, events[0].OriginalEvent().Data)
	assert.Equal(s.T(), testEvent.Metadata, events[0].OriginalEvent().UserMetadata)
}

func (s *AppendTestSuite) TestAppendToStreamFailsOnNonExistentStream() {
	client := s.fixture.Client()

	// Arrange
	streamId := s.fixture.NewStreamId()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	opts := kurrentdb.AppendToStreamOptions{StreamState: kurrentdb.StreamExists{}}

	// Act
	_, err := client.AppendToStream(ctx, streamId, opts, s.fixture.CreateTestEvent())
	kurrentDbError, ok := kurrentdb.FromError(err)

	// Assert
	assert.False(s.T(), ok)
	assert.Equal(s.T(), kurrentdb.ErrorCodeWrongExpectedVersion, kurrentDbError.Code())
}

func (s *AppendTestSuite) TestMetadataOperation() {
	client := s.fixture.Client()

	// Arrange
	streamId := s.fixture.NewStreamId()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	opts := kurrentdb.AppendToStreamOptions{StreamState: kurrentdb.Any{}}

	_, err := client.AppendToStream(ctx, streamId, opts, s.fixture.CreateTestEvent())
	assert.NoError(s.T(), err, "Error writing event")

	acl := kurrentdb.Acl{}
	acl.AddReadRoles("admin")
	meta := kurrentdb.StreamMetadata{}
	meta.SetMaxAge(2 * time.Second)
	meta.SetAcl(acl)

	result, err := client.SetStreamMetadata(ctx, streamId, opts, meta)
	assert.NoError(s.T(), err, "Error writing stream metadata")
	assert.NotNil(s.T(), result, "Metadata write result should not be nil")

	metaActual, err := client.GetStreamMetadata(ctx, streamId, kurrentdb.ReadStreamOptions{})
	assert.NoError(s.T(), err, "Error reading stream metadata")
	assert.Equal(s.T(), meta, *metaActual, "Metadata should match")
}
