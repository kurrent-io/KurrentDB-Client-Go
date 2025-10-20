package test

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MultiAppendTestSuite struct {
	suite.Suite
	fixture *ClientFixture
}

type ExpectedMetadata struct {
	StringValue      string        `json:"stringValue"`
	BoolValue        bool          `json:"boolValue"`
	Float32Value     float32       `json:"float32Value"`
	Float64Value     float64       `json:"float64Value"`
	IntValue         int           `json:"intValue"`
	Int32Value       int32         `json:"int32Value"`
	Int64Value       int64         `json:"int64Value"`
	NullValue        interface{}   `json:"nullValue"`
	TimeValue        time.Time     `json:"timeValue"`
	DurationValue    time.Duration `json:"durationValue"`
	ByteValue        []byte        `json:"byteValue"`
	DefaultValue     interface{}   `json:"defaultValue"`
	SchemaDataFormat string        `json:"$schema.data-format"`
	SchemaName       string        `json:"$schema.name"`
}

func TestMultiAppendEventsSuite(t *testing.T) {
	suite.Run(t, new(MultiAppendTestSuite))
}

func (s *MultiAppendTestSuite) SetupTest() {
	s.fixture = NewInsecureClientFixture(s.T())
}

func (s *MultiAppendTestSuite) TestMultiStreamAppend() {
	client := s.fixture.Client()

	version, err := client.GetServerVersion()
	assert.NoError(s.T(), err)

	if version.Major < 25 {
		s.T().Skip("Multi-stream append is not supported in versions prior to 25.0")
	}

	// Arrange
	expectedMetadata, _ := json.Marshal(map[string]interface{}{
		"stringValue":   "string",
		"boolValue":     true,
		"intValue":      1,
		"int32Value":    int32(32),
		"int64Value":    int64(64),
		"float32Value":  float32(3.14),
		"float64Value":  float64(2.718),
		"nullValue":     nil,
		"timeValue":     time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		"durationValue": 5 * time.Minute,
		"byteValue":     []byte{0x01, 0x02, 0x03},
		"defaultValue":  struct{ Value string }{Value: "test"},
	})

	stream1 := s.fixture.NewStreamId()
	stream2 := s.fixture.NewStreamId()

	event1 := s.fixture.CreateTestEvent(TestEventOptions{EventType: "OrderCreated", Metadata: expectedMetadata})
	event2 := s.fixture.CreateTestEvent(TestEventOptions{EventType: "OrderShipped", Metadata: expectedMetadata})

	// Act
	requests := []kurrentdb.AppendStreamRequest{
		{
			StreamName:          stream1,
			Events:              slices.Values([]kurrentdb.EventData{event1}),
			ExpectedStreamState: kurrentdb.NoStream{},
		},
		{
			StreamName:          stream2,
			Events:              slices.Values([]kurrentdb.EventData{event2}),
			ExpectedStreamState: kurrentdb.NoStream{},
		},
	}

	result, err := client.MultiStreamAppend(context.Background(), slices.Values(requests))

	// Assert
	assert.NoError(s.T(), err)
	assert.Greater(s.T(), result.Position, int64(0))

	s.assertMetadata(stream1, event1.EventType)
	s.assertMetadata(stream2, event2.EventType)
}

func (s *MultiAppendTestSuite) TestMultiStreamAppendUnsupported() {
	client := s.fixture.Client()

	version, err := client.GetServerVersion()
	assert.NoError(s.T(), err)

	if version.Major >= 25 {
		s.T().Skip("Multi-stream append is supported in from 25.0 and upwards")
	}

	stream := s.fixture.NewStreamId()
	events := s.fixture.CreateTestEvents(stream, 1)

	requests := []kurrentdb.AppendStreamRequest{
		{
			StreamName:          stream,
			Events:              slices.Values(events),
			ExpectedStreamState: kurrentdb.Any{},
		},
	}

	result, err := client.MultiStreamAppend(context.Background(), slices.Values(requests))

	assert.Nil(s.T(), result)

	kdbError, _ := kurrentdb.FromError(err)
	assert.Equal(s.T(), kurrentdb.ErrorCodeUnsupportedFeature, kdbError.Code(), "Expected unsupported feature error code")
}

func (s *MultiAppendTestSuite) TestMultiStreamAppendWithBinaryMetadataThrowsError() {
	client := s.fixture.Client()

	version, err := client.GetServerVersion()
	assert.NoError(s.T(), err)

	if version.Major < 25 {
		s.T().Skip("Multi-stream append is not supported in versions prior to 25.0")
	}

	// Arrange
	stream := s.fixture.NewStreamId()
	binaryMetadata := []byte{0x01, 0x02, 0x03, 0x04}

	event := s.fixture.CreateTestEvent(TestEventOptions{
		EventType:   "OrderCreated",
		ContentType: kurrentdb.ContentTypeBinary,
		Metadata:    binaryMetadata,
	})

	requests := []kurrentdb.AppendStreamRequest{
		{
			StreamName:          stream,
			Events:              slices.Values([]kurrentdb.EventData{event}),
			ExpectedStreamState: kurrentdb.NoStream{},
		},
	}

	// Act
	result, err := client.MultiStreamAppend(context.Background(), slices.Values(requests))

	// Assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)

	dbError, ok := kurrentdb.FromError(err)
	assert.False(s.T(), ok)
	assert.Error(s.T(), dbError)
}

func (s *MultiAppendTestSuite) TestMultiStreamAppendStreamRevisionConflict() {
	client := s.fixture.Client()

	version, err := client.GetServerVersion()
	assert.NoError(s.T(), err)

	if version.Major < 25 {
		s.T().Skip("Multi-stream append is not supported in versions prior to 25.0")
	}

	// Arrange
	stream := s.fixture.NewStreamId()

	s.fixture.CreateTestEvents(stream, 10)

	readStream, err := client.ReadStream(context.Background(), stream, kurrentdb.ReadStreamOptions{
		Direction: kurrentdb.Backwards,
		From:      kurrentdb.End{},
	}, 1)
	assert.NoError(s.T(), err)
	defer readStream.Close()
	lastEvent, err := readStream.Recv()
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), lastEvent)

	// Act
	requests := []kurrentdb.AppendStreamRequest{
		{
			StreamName:          stream,
			Events:              slices.Values([]kurrentdb.EventData{s.fixture.CreateTestEvent()}),
			ExpectedStreamState: kurrentdb.NoStream{},
		},
	}

	_, err = client.MultiStreamAppend(context.Background(), slices.Values(requests))
	kurrentDbError, ok := kurrentdb.FromError(err)

	assert.False(s.T(), ok)
	assert.Equal(s.T(), kurrentdb.ErrorCodeStreamRevisionConflict, kurrentDbError.Code())

	var streamRevisionConflictErr *kurrentdb.StreamRevisionConflictError
	if !errors.As(err, &streamRevisionConflictErr) {
		s.T().Fatal("Expected StreamRevisionConflictError")
	}

	assert.Equal(s.T(), stream, streamRevisionConflictErr.Stream)
	assert.Equal(s.T(), kurrentdb.NoStream{}, streamRevisionConflictErr.ExpectedRevision)
	assert.Equal(s.T(), kurrentdb.StreamRevision{Value: lastEvent.OriginalEvent().EventNumber}, streamRevisionConflictErr.ActualRevision)
}

func (s *MultiAppendTestSuite) TestMultiStreamAppendStreamTombstoned() {
	client := s.fixture.Client()

	version, err := client.GetServerVersion()
	assert.NoError(s.T(), err)

	if version.Major < 25 {
		s.T().Skip("Multi-stream append is not supported in versions prior to 25.0")
	}

	// Arrange
	stream := s.fixture.NewStreamId()

	s.fixture.CreateTestEvents(stream, 10)

	_, err = client.TombstoneStream(context.Background(), stream, kurrentdb.TombstoneStreamOptions{})
	assert.NoError(s.T(), err)

	// Act
	requests := []kurrentdb.AppendStreamRequest{
		{
			StreamName:          stream,
			Events:              slices.Values([]kurrentdb.EventData{s.fixture.CreateTestEvent()}),
			ExpectedStreamState: kurrentdb.NoStream{},
		},
	}

	_, err = client.MultiStreamAppend(context.Background(), slices.Values(requests))
	kurrentDbError, ok := kurrentdb.FromError(err)

	assert.False(s.T(), ok)
	assert.Equal(s.T(), kurrentdb.ErrorCodeStreamTombstoned, kurrentDbError.Code())
}

// region helpers
func (s *MultiAppendTestSuite) assertMetadata(streamName string, eventType string) {
	client := s.fixture.Client()

	readStream, err := client.ReadStream(context.Background(), streamName, kurrentdb.ReadStreamOptions{}, 1)
	assert.NoError(s.T(), err)
	defer readStream.Close()

	events, err := s.fixture.CollectEvents(readStream)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), events, 1)

	var readMetadata ExpectedMetadata
	err = json.Unmarshal(events[0].OriginalEvent().UserMetadata, &readMetadata)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), "Bytes", readMetadata.SchemaDataFormat)
	assert.Equal(s.T(), eventType, readMetadata.SchemaName)
	assert.Equal(s.T(), "string", readMetadata.StringValue)
	assert.Equal(s.T(), true, readMetadata.BoolValue)
	assert.Equal(s.T(), 1, readMetadata.IntValue)
	assert.Equal(s.T(), int32(32), readMetadata.Int32Value)
	assert.Equal(s.T(), int64(64), readMetadata.Int64Value)
	assert.Equal(s.T(), float32(3.14), readMetadata.Float32Value)
	assert.Equal(s.T(), float64(2.718), readMetadata.Float64Value)
	assert.Nil(s.T(), readMetadata.NullValue)
	assert.Equal(s.T(), time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), readMetadata.TimeValue)
	assert.Equal(s.T(), 5*time.Minute, readMetadata.DurationValue)
	assert.Equal(s.T(), []byte{0x01, 0x02, 0x03}, readMetadata.ByteValue)
	assert.Equal(s.T(), "map[Value:test]", readMetadata.DefaultValue)
}

//endregion
