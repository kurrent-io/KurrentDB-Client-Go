package test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AppendRecordsTestSuite struct {
	suite.Suite
	fixture *ClientFixture
	client  *kurrentdb.Client
}

func TestAppendRecordsSuite(t *testing.T) {
	suite.Run(t, new(AppendRecordsTestSuite))
}

func (s *AppendRecordsTestSuite) SetupTest() {
	s.fixture = NewInsecureClientFixture(s.T())
	s.fixture.RequireMinServerVersion(s.T(), 26, 1, 0, "AppendRecords requires KurrentDB 26.1+")
	s.client = s.fixture.Client()
}

func (s *AppendRecordsTestSuite) recordFor(stream string) kurrentdb.AppendRecord {
	return kurrentdb.AppendRecord{
		Stream: stream,
		Record: s.fixture.CreateTestEvent(),
	}
}

func (s *AppendRecordsTestSuite) assertConsistencyViolation(err error, expectedViolationCount int) *kurrentdb.AppendConsistencyViolationError {
	s.T().Helper()

	kdbError, ok := kurrentdb.FromError(err)
	assert.False(s.T(), ok)
	assert.Equal(s.T(), kurrentdb.ErrorCodeAppendConsistencyViolation, kdbError.Code())

	var violationErr *kurrentdb.AppendConsistencyViolationError
	if !errors.As(err, &violationErr) {
		s.T().Fatal("Expected AppendConsistencyViolationError")
	}

	assert.Len(s.T(), violationErr.Violations, expectedViolationCount)
	return violationErr
}

// ==================== Base tests ====================

func (s *AppendRecordsTestSuite) TestAppendRecordsSingleStream() {
	stream := s.fixture.NewStreamId()
	events := []kurrentdb.AppendRecord{
		{Stream: stream, Record: s.fixture.CreateTestEvent(TestEventOptions{EventType: "Event1"})},
		{Stream: stream, Record: s.fixture.CreateTestEvent(TestEventOptions{EventType: "Event2"})},
		{Stream: stream, Record: s.fixture.CreateTestEvent(TestEventOptions{EventType: "Event3"})},
	}

	result, err := s.client.AppendRecords(context.Background(), events)

	assert.NoError(s.T(), err)
	assert.Greater(s.T(), result.Position, int64(0))
	assert.Len(s.T(), result.Responses, 1)
	assert.Equal(s.T(), stream, result.Responses[0].Stream)
}

func (s *AppendRecordsTestSuite) TestAppendRecordsMultipleStreams() {
	stream1 := s.fixture.NewStreamId()
	stream2 := s.fixture.NewStreamId()

	metadata, _ := json.Marshal(map[string]string{"Name": "Test User"})

	records := []kurrentdb.AppendRecord{
		{Stream: stream1, Record: s.fixture.CreateTestEvent(TestEventOptions{EventType: "Event1", Metadata: metadata})},
		{Stream: stream1, Record: s.fixture.CreateTestEvent(TestEventOptions{EventType: "Event2", Metadata: metadata})},
		{Stream: stream1, Record: s.fixture.CreateTestEvent(TestEventOptions{EventType: "Event3", Metadata: metadata})},
		{Stream: stream2, Record: s.fixture.CreateTestEvent(TestEventOptions{EventType: "Event4", Metadata: metadata})},
		{Stream: stream2, Record: s.fixture.CreateTestEvent(TestEventOptions{EventType: "Event5", Metadata: metadata})},
	}

	result, err := s.client.AppendRecords(context.Background(), records)

	assert.NoError(s.T(), err)
	assert.Greater(s.T(), result.Position, int64(0))
	assert.Len(s.T(), result.Responses, 2)
}

func (s *AppendRecordsTestSuite) TestInterleavedTracksRevisions() {
	stream1 := s.fixture.NewStreamId()
	stream2 := s.fixture.NewStreamId()

	records := []kurrentdb.AppendRecord{
		{Stream: stream1, Record: s.fixture.CreateTestEvent()},
		{Stream: stream2, Record: s.fixture.CreateTestEvent()},
		{Stream: stream1, Record: s.fixture.CreateTestEvent()},
		{Stream: stream2, Record: s.fixture.CreateTestEvent()},
		{Stream: stream1, Record: s.fixture.CreateTestEvent()},
	}

	result, err := s.client.AppendRecords(context.Background(), records)

	assert.NoError(s.T(), err)
	assert.Greater(s.T(), result.Position, int64(0))
	assert.Len(s.T(), result.Responses, 2)
}

func (s *AppendRecordsTestSuite) TestAppendRecordsEmptyRecordsReturnsError() {
	result, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{})

	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)
}

func (s *AppendRecordsTestSuite) TestAppendRecordsInvalidMetadataReturnsError() {
	stream := s.fixture.NewStreamId()
	records := []kurrentdb.AppendRecord{
		{
			Stream: stream,
			Record: kurrentdb.EventData{
				EventType: "TestEvent",
				Data:      []byte{0xb, 0xe, 0xe, 0xf},
				Metadata:  []byte("invalid"),
			},
		},
	}

	result, err := s.client.AppendRecords(context.Background(), records)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)
}

// ==================== WriteOnly / WhenExpectingAny ====================

func (s *AppendRecordsTestSuite) TestWriteAny_SucceedsWhenStreamHasRevision() {
	stream := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(stream, 3)

	result, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(stream),
	})

	assert.NoError(s.T(), err)
	assert.Len(s.T(), result.Responses, 1)
	assert.Equal(s.T(), stream, result.Responses[0].Stream)
}

func (s *AppendRecordsTestSuite) TestWriteAny_SucceedsWhenStreamNotFound() {
	stream := s.fixture.NewStreamId()

	result, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(stream),
	})

	assert.NoError(s.T(), err)
	assert.Len(s.T(), result.Responses, 1)
	assert.Equal(s.T(), stream, result.Responses[0].Stream)
}

func (s *AppendRecordsTestSuite) TestWriteAny_SucceedsWhenStreamIsDeleted() {
	stream := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(stream, 1)
	s.fixture.DeleteStream(stream)

	result, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(stream),
	})

	assert.NoError(s.T(), err)
	assert.Len(s.T(), result.Responses, 1)
	assert.Equal(s.T(), stream, result.Responses[0].Stream)
}

func (s *AppendRecordsTestSuite) TestWriteAny_FailsWhenStreamIsTombstoned() {
	stream := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(stream, 1)
	s.fixture.TombstoneStream(stream)

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(stream),
	})

	assert.Error(s.T(), err)
}

// ==================== WriteOnly / WhenExpectingNoStream ====================

func (s *AppendRecordsTestSuite) TestWriteNoStream_SucceedsWhenStreamNotFound() {
	stream := s.fixture.NewStreamId()

	result, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(stream),
	},
		kurrentdb.StreamStateCheck{Stream: stream, ExpectedState: kurrentdb.NoStream{}},
	)

	assert.NoError(s.T(), err)
	assert.Len(s.T(), result.Responses, 1)
	assert.Equal(s.T(), stream, result.Responses[0].Stream)
}

func (s *AppendRecordsTestSuite) TestWriteNoStream_SucceedsWhenStreamIsDeleted() {
	stream := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(stream, 1)
	s.fixture.DeleteStream(stream)

	result, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(stream),
	},
		kurrentdb.StreamStateCheck{Stream: stream, ExpectedState: kurrentdb.NoStream{}},
	)

	assert.NoError(s.T(), err)
	assert.Len(s.T(), result.Responses, 1)
}

func (s *AppendRecordsTestSuite) TestWriteNoStream_FailsWhenStreamHasRevision() {
	stream := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(stream, 3)

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(stream),
	},
		kurrentdb.StreamStateCheck{Stream: stream, ExpectedState: kurrentdb.NoStream{}},
	)

	v := s.assertConsistencyViolation(err, 1)
	assert.Equal(s.T(), stream, v.Violations[0].Stream)
}

func (s *AppendRecordsTestSuite) TestWriteNoStream_FailsWhenStreamIsTombstoned() {
	stream := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(stream, 1)
	s.fixture.TombstoneStream(stream)

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(stream),
	},
		kurrentdb.StreamStateCheck{Stream: stream, ExpectedState: kurrentdb.NoStream{}},
	)

	assert.Error(s.T(), err)
}

// ==================== WriteOnly / WhenExpectingRevision ====================

func (s *AppendRecordsTestSuite) TestWriteRevision_SucceedsWhenStreamHasRevision() {
	stream := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(stream, 11)

	result, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(stream),
	},
		kurrentdb.StreamStateCheck{Stream: stream, ExpectedState: kurrentdb.StreamRevision{Value: 10}},
	)

	assert.NoError(s.T(), err)
	assert.Len(s.T(), result.Responses, 1)
	assert.Equal(s.T(), stream, result.Responses[0].Stream)
}

func (s *AppendRecordsTestSuite) TestWriteRevision_FailsWhenStreamNotFound() {
	stream := s.fixture.NewStreamId()

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(stream),
	},
		kurrentdb.StreamStateCheck{Stream: stream, ExpectedState: kurrentdb.StreamRevision{Value: 10}},
	)

	v := s.assertConsistencyViolation(err, 1)
	assert.Equal(s.T(), stream, v.Violations[0].Stream)
}

func (s *AppendRecordsTestSuite) TestWriteRevision_FailsWhenStreamHasWrongRevision() {
	stream := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(stream, 5)

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(stream),
	},
		kurrentdb.StreamStateCheck{Stream: stream, ExpectedState: kurrentdb.StreamRevision{Value: 10}},
	)

	v := s.assertConsistencyViolation(err, 1)
	assert.Equal(s.T(), stream, v.Violations[0].Stream)
}

func (s *AppendRecordsTestSuite) TestWriteRevision_FailsWhenStreamIsDeleted() {
	stream := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(stream, 3)
	s.fixture.DeleteStream(stream)

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(stream),
	},
		kurrentdb.StreamStateCheck{Stream: stream, ExpectedState: kurrentdb.StreamRevision{Value: 10}},
	)

	assert.Error(s.T(), err)
}

func (s *AppendRecordsTestSuite) TestWriteRevision_FailsWhenStreamIsTombstoned() {
	stream := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(stream, 1)
	s.fixture.TombstoneStream(stream)

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(stream),
	},
		kurrentdb.StreamStateCheck{Stream: stream, ExpectedState: kurrentdb.StreamRevision{Value: 10}},
	)

	assert.Error(s.T(), err)
}

// ==================== WriteOnly / WhenExpectingExists ====================

func (s *AppendRecordsTestSuite) TestWriteExists_SucceedsWhenStreamHasRevision() {
	stream := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(stream, 3)

	result, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(stream),
	},
		kurrentdb.StreamStateCheck{Stream: stream, ExpectedState: kurrentdb.StreamExists{}},
	)

	assert.NoError(s.T(), err)
	assert.Len(s.T(), result.Responses, 1)
	assert.Equal(s.T(), stream, result.Responses[0].Stream)
}

func (s *AppendRecordsTestSuite) TestWriteExists_FailsWhenStreamNotFound() {
	stream := s.fixture.NewStreamId()

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(stream),
	},
		kurrentdb.StreamStateCheck{Stream: stream, ExpectedState: kurrentdb.StreamExists{}},
	)

	v := s.assertConsistencyViolation(err, 1)
	assert.Equal(s.T(), stream, v.Violations[0].Stream)
}

func (s *AppendRecordsTestSuite) TestWriteExists_FailsWhenStreamIsDeleted() {
	stream := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(stream, 1)
	s.fixture.DeleteStream(stream)

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(stream),
	},
		kurrentdb.StreamStateCheck{Stream: stream, ExpectedState: kurrentdb.StreamExists{}},
	)

	assert.Error(s.T(), err)
}

func (s *AppendRecordsTestSuite) TestWriteExists_FailsWhenStreamIsTombstoned() {
	stream := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(stream, 1)
	s.fixture.TombstoneStream(stream)

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(stream),
	},
		kurrentdb.StreamStateCheck{Stream: stream, ExpectedState: kurrentdb.StreamExists{}},
	)

	assert.Error(s.T(), err)
}

// ==================== CheckOnly / WhenExpectingNoStream ====================

func (s *AppendRecordsTestSuite) TestCheckNoStream_SucceedsWhenStreamNotFound() {
	checkStream := s.fixture.NewStreamId()
	writeStream := s.fixture.NewStreamId()

	result, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(writeStream),
	},
		kurrentdb.StreamStateCheck{Stream: checkStream, ExpectedState: kurrentdb.NoStream{}},
	)

	assert.NoError(s.T(), err)
	assert.Len(s.T(), result.Responses, 1)
	assert.Equal(s.T(), writeStream, result.Responses[0].Stream)
}

func (s *AppendRecordsTestSuite) TestCheckNoStream_SucceedsWhenStreamIsDeleted() {
	checkStream := s.fixture.NewStreamId()
	writeStream := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(checkStream, 1)
	s.fixture.DeleteStream(checkStream)

	result, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(writeStream),
	},
		kurrentdb.StreamStateCheck{Stream: checkStream, ExpectedState: kurrentdb.NoStream{}},
	)

	assert.NoError(s.T(), err)
	assert.Len(s.T(), result.Responses, 1)
	assert.Equal(s.T(), writeStream, result.Responses[0].Stream)
}

func (s *AppendRecordsTestSuite) TestCheckNoStream_SucceedsWhenStreamIsTombstoned() {
	checkStream := s.fixture.NewStreamId()
	writeStream := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(checkStream, 1)
	s.fixture.TombstoneStream(checkStream)

	result, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(writeStream),
	},
		kurrentdb.StreamStateCheck{Stream: checkStream, ExpectedState: kurrentdb.NoStream{}},
	)

	assert.NoError(s.T(), err)
	assert.Len(s.T(), result.Responses, 1)
	assert.Equal(s.T(), writeStream, result.Responses[0].Stream)
}

func (s *AppendRecordsTestSuite) TestCheckNoStream_FailsWhenStreamHasRevision() {
	checkStream := s.fixture.NewStreamId()
	writeStream := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(checkStream, 3)

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(writeStream),
	},
		kurrentdb.StreamStateCheck{Stream: checkStream, ExpectedState: kurrentdb.NoStream{}},
	)

	v := s.assertConsistencyViolation(err, 1)
	assert.Equal(s.T(), checkStream, v.Violations[0].Stream)
}

// ==================== CheckOnly / WhenExpectingRevision ====================

func (s *AppendRecordsTestSuite) TestCheckRevision_SucceedsWhenStreamHasRevision() {
	checkStream := s.fixture.NewStreamId()
	writeStream := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(checkStream, 11)

	result, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(writeStream),
	},
		kurrentdb.StreamStateCheck{Stream: checkStream, ExpectedState: kurrentdb.StreamRevision{Value: 10}},
	)

	assert.NoError(s.T(), err)
	assert.Len(s.T(), result.Responses, 1)
	assert.Equal(s.T(), writeStream, result.Responses[0].Stream)
}

func (s *AppendRecordsTestSuite) TestCheckRevision_FailsWhenStreamNotFound() {
	checkStream := s.fixture.NewStreamId()
	writeStream := s.fixture.NewStreamId()

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(writeStream),
	},
		kurrentdb.StreamStateCheck{Stream: checkStream, ExpectedState: kurrentdb.StreamRevision{Value: 10}},
	)

	v := s.assertConsistencyViolation(err, 1)
	assert.Equal(s.T(), checkStream, v.Violations[0].Stream)
}

func (s *AppendRecordsTestSuite) TestCheckRevision_FailsWhenStreamHasWrongRevision() {
	checkStream := s.fixture.NewStreamId()
	writeStream := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(checkStream, 5)

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(writeStream),
	},
		kurrentdb.StreamStateCheck{Stream: checkStream, ExpectedState: kurrentdb.StreamRevision{Value: 10}},
	)

	v := s.assertConsistencyViolation(err, 1)
	assert.Equal(s.T(), checkStream, v.Violations[0].Stream)
}

func (s *AppendRecordsTestSuite) TestCheckRevision_FailsWhenStreamIsDeleted() {
	checkStream := s.fixture.NewStreamId()
	writeStream := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(checkStream, 3)
	s.fixture.DeleteStream(checkStream)

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(writeStream),
	},
		kurrentdb.StreamStateCheck{Stream: checkStream, ExpectedState: kurrentdb.StreamRevision{Value: 10}},
	)

	v := s.assertConsistencyViolation(err, 1)
	assert.Equal(s.T(), checkStream, v.Violations[0].Stream)
}

func (s *AppendRecordsTestSuite) TestCheckRevision_FailsWhenStreamIsTombstoned() {
	checkStream := s.fixture.NewStreamId()
	writeStream := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(checkStream, 1)
	s.fixture.TombstoneStream(checkStream)

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(writeStream),
	},
		kurrentdb.StreamStateCheck{Stream: checkStream, ExpectedState: kurrentdb.StreamRevision{Value: 10}},
	)

	v := s.assertConsistencyViolation(err, 1)
	assert.Equal(s.T(), checkStream, v.Violations[0].Stream)
}

// ==================== CheckOnly / WhenExpectingExists ====================

func (s *AppendRecordsTestSuite) TestCheckExists_SucceedsWhenStreamHasRevision() {
	checkStream := s.fixture.NewStreamId()
	writeStream := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(checkStream, 3)

	result, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(writeStream),
	},
		kurrentdb.StreamStateCheck{Stream: checkStream, ExpectedState: kurrentdb.StreamExists{}},
	)

	assert.NoError(s.T(), err)
	assert.Len(s.T(), result.Responses, 1)
	assert.Equal(s.T(), writeStream, result.Responses[0].Stream)
}

func (s *AppendRecordsTestSuite) TestCheckExists_FailsWhenStreamNotFound() {
	checkStream := s.fixture.NewStreamId()
	writeStream := s.fixture.NewStreamId()

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(writeStream),
	},
		kurrentdb.StreamStateCheck{Stream: checkStream, ExpectedState: kurrentdb.StreamExists{}},
	)

	v := s.assertConsistencyViolation(err, 1)
	assert.Equal(s.T(), checkStream, v.Violations[0].Stream)
}

func (s *AppendRecordsTestSuite) TestCheckExists_FailsWhenStreamIsDeleted() {
	checkStream := s.fixture.NewStreamId()
	writeStream := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(checkStream, 1)
	s.fixture.DeleteStream(checkStream)

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(writeStream),
	},
		kurrentdb.StreamStateCheck{Stream: checkStream, ExpectedState: kurrentdb.StreamExists{}},
	)

	v := s.assertConsistencyViolation(err, 1)
	assert.Equal(s.T(), checkStream, v.Violations[0].Stream)
}

func (s *AppendRecordsTestSuite) TestCheckExists_FailsWhenStreamIsTombstoned() {
	checkStream := s.fixture.NewStreamId()
	writeStream := s.fixture.NewStreamId()
	s.fixture.CreateTestEvents(checkStream, 1)
	s.fixture.TombstoneStream(checkStream)

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(writeStream),
	},
		kurrentdb.StreamStateCheck{Stream: checkStream, ExpectedState: kurrentdb.StreamExists{}},
	)

	v := s.assertConsistencyViolation(err, 1)
	assert.Equal(s.T(), checkStream, v.Violations[0].Stream)
}

// ==================== CheckOnly / WhenMultipleChecks ====================

func (s *AppendRecordsTestSuite) TestMultipleChecks_SucceedsWhenAllChecksPass() {
	writeStream := s.fixture.NewStreamId()
	checkStreamA := s.fixture.NewStreamId()
	checkStreamB := s.fixture.NewStreamId()

	s.fixture.CreateTestEvents(checkStreamA, 3)
	s.fixture.CreateTestEvents(checkStreamB, 5)

	result, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(writeStream),
	},
		kurrentdb.StreamStateCheck{Stream: checkStreamA, ExpectedState: kurrentdb.StreamRevision{Value: 2}},
		kurrentdb.StreamStateCheck{Stream: checkStreamB, ExpectedState: kurrentdb.StreamRevision{Value: 4}},
	)

	assert.NoError(s.T(), err)
	assert.Len(s.T(), result.Responses, 1)
	assert.Equal(s.T(), writeStream, result.Responses[0].Stream)
}

func (s *AppendRecordsTestSuite) TestMultipleChecks_SucceedsWhenAllMixedCheckTypesPass() {
	writeStream := s.fixture.NewStreamId()
	revisionStream := s.fixture.NewStreamId()
	existsStream := s.fixture.NewStreamId()
	noStream := s.fixture.NewStreamId()

	s.fixture.CreateTestEvents(revisionStream, 3)
	s.fixture.CreateTestEvents(existsStream, 1)

	result, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(writeStream),
	},
		kurrentdb.StreamStateCheck{Stream: revisionStream, ExpectedState: kurrentdb.StreamRevision{Value: 2}},
		kurrentdb.StreamStateCheck{Stream: existsStream, ExpectedState: kurrentdb.StreamExists{}},
		kurrentdb.StreamStateCheck{Stream: noStream, ExpectedState: kurrentdb.NoStream{}},
	)

	assert.NoError(s.T(), err)
	assert.Len(s.T(), result.Responses, 1)
	assert.Equal(s.T(), writeStream, result.Responses[0].Stream)
}

func (s *AppendRecordsTestSuite) TestMultipleChecks_FailsWhenOneOfMultipleChecksFails() {
	writeStream := s.fixture.NewStreamId()
	checkStreamA := s.fixture.NewStreamId()
	checkStreamB := s.fixture.NewStreamId()

	s.fixture.CreateTestEvents(checkStreamA, 3)
	// checkStreamB not seeded — will fail the revision check

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(writeStream),
	},
		kurrentdb.StreamStateCheck{Stream: checkStreamA, ExpectedState: kurrentdb.StreamRevision{Value: 2}},
		kurrentdb.StreamStateCheck{Stream: checkStreamB, ExpectedState: kurrentdb.StreamRevision{Value: 5}},
	)

	v := s.assertConsistencyViolation(err, 1)
	assert.Equal(s.T(), int32(1), v.Violations[0].CheckIndex)
	assert.Equal(s.T(), checkStreamB, v.Violations[0].Stream)
}

func (s *AppendRecordsTestSuite) TestMultipleChecks_FailsWhenAllChecksFail() {
	writeStream := s.fixture.NewStreamId()
	checkStreamA := s.fixture.NewStreamId()
	checkStreamB := s.fixture.NewStreamId()

	// Neither stream is seeded — both will fail

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(writeStream),
	},
		kurrentdb.StreamStateCheck{Stream: checkStreamA, ExpectedState: kurrentdb.StreamRevision{Value: 3}},
		kurrentdb.StreamStateCheck{Stream: checkStreamB, ExpectedState: kurrentdb.StreamExists{}},
	)

	v := s.assertConsistencyViolation(err, 2)

	var violationA, violationB *kurrentdb.ConsistencyViolation
	for i := range v.Violations {
		if v.Violations[i].Stream == checkStreamA {
			violationA = &v.Violations[i]
		}
		if v.Violations[i].Stream == checkStreamB {
			violationB = &v.Violations[i]
		}
	}

	assert.NotNil(s.T(), violationA)
	assert.NotNil(s.T(), violationB)
	assert.Equal(s.T(), int32(0), violationA.CheckIndex)
	assert.Equal(s.T(), int32(1), violationB.CheckIndex)
}

func (s *AppendRecordsTestSuite) TestMultipleChecks_FailsWhenFirstCheckFailsAndSecondPasses() {
	writeStream := s.fixture.NewStreamId()
	checkStreamA := s.fixture.NewStreamId()
	checkStreamB := s.fixture.NewStreamId()

	// checkStreamA not seeded — will fail the revision check
	s.fixture.CreateTestEvents(checkStreamB, 6)

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(writeStream),
	},
		kurrentdb.StreamStateCheck{Stream: checkStreamA, ExpectedState: kurrentdb.StreamRevision{Value: 3}},
		kurrentdb.StreamStateCheck{Stream: checkStreamB, ExpectedState: kurrentdb.StreamRevision{Value: 5}},
	)

	v := s.assertConsistencyViolation(err, 1)
	assert.Equal(s.T(), int32(0), v.Violations[0].CheckIndex)
	assert.Equal(s.T(), checkStreamA, v.Violations[0].Stream)
}

func (s *AppendRecordsTestSuite) TestMultipleChecks_FailsWhenTwoOfThreeChecksFail() {
	writeStream := s.fixture.NewStreamId()
	checkStreamA := s.fixture.NewStreamId()
	checkStreamB := s.fixture.NewStreamId()
	checkStreamC := s.fixture.NewStreamId()

	s.fixture.CreateTestEvents(checkStreamA, 3)
	// checkStreamB not seeded — will fail
	s.fixture.CreateTestEvents(checkStreamC, 2)

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(writeStream),
	},
		kurrentdb.StreamStateCheck{Stream: checkStreamA, ExpectedState: kurrentdb.StreamRevision{Value: 2}},
		kurrentdb.StreamStateCheck{Stream: checkStreamB, ExpectedState: kurrentdb.StreamExists{}},
		kurrentdb.StreamStateCheck{Stream: checkStreamC, ExpectedState: kurrentdb.StreamRevision{Value: 10}},
	)

	v := s.assertConsistencyViolation(err, 2)

	var violationB, violationC *kurrentdb.ConsistencyViolation
	for i := range v.Violations {
		if v.Violations[i].Stream == checkStreamB {
			violationB = &v.Violations[i]
		}
		if v.Violations[i].Stream == checkStreamC {
			violationC = &v.Violations[i]
		}
	}

	assert.NotNil(s.T(), violationB)
	assert.NotNil(s.T(), violationC)
	assert.Equal(s.T(), int32(1), violationB.CheckIndex)
	assert.Equal(s.T(), int32(2), violationC.CheckIndex)
}

func (s *AppendRecordsTestSuite) TestMultipleChecks_FailsWithMixedViolationStates() {
	writeStream := s.fixture.NewStreamId()
	deletedStream := s.fixture.NewStreamId()
	tombstonedStream := s.fixture.NewStreamId()
	missingStream := s.fixture.NewStreamId()

	s.fixture.CreateTestEvents(deletedStream, 3)
	s.fixture.DeleteStream(deletedStream)

	s.fixture.CreateTestEvents(tombstonedStream, 1)
	s.fixture.TombstoneStream(tombstonedStream)

	// missingStream not seeded

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(writeStream),
	},
		kurrentdb.StreamStateCheck{Stream: deletedStream, ExpectedState: kurrentdb.StreamExists{}},
		kurrentdb.StreamStateCheck{Stream: tombstonedStream, ExpectedState: kurrentdb.StreamExists{}},
		kurrentdb.StreamStateCheck{Stream: missingStream, ExpectedState: kurrentdb.StreamRevision{Value: 10}},
	)

	v := s.assertConsistencyViolation(err, 3)

	var deletedV, tombstonedV, missingV *kurrentdb.ConsistencyViolation
	for i := range v.Violations {
		switch v.Violations[i].Stream {
		case deletedStream:
			deletedV = &v.Violations[i]
		case tombstonedStream:
			tombstonedV = &v.Violations[i]
		case missingStream:
			missingV = &v.Violations[i]
		}
	}

	assert.NotNil(s.T(), deletedV)
	assert.NotNil(s.T(), tombstonedV)
	assert.NotNil(s.T(), missingV)
	assert.Equal(s.T(), int32(0), deletedV.CheckIndex)
	assert.Equal(s.T(), int32(1), tombstonedV.CheckIndex)
	assert.Equal(s.T(), int32(2), missingV.CheckIndex)
}

func (s *AppendRecordsTestSuite) TestMultipleChecks_FailsWhenCheckOnWriteTargetAndSeparateCheckFails() {
	checkStream := s.fixture.NewStreamId()
	writeStream := s.fixture.NewStreamId()

	// checkStream not seeded — will fail
	s.fixture.CreateTestEvents(writeStream, 4)

	_, err := s.client.AppendRecords(context.Background(), []kurrentdb.AppendRecord{
		s.recordFor(writeStream),
	},
		kurrentdb.StreamStateCheck{Stream: checkStream, ExpectedState: kurrentdb.StreamRevision{Value: 5}},
		kurrentdb.StreamStateCheck{Stream: writeStream, ExpectedState: kurrentdb.StreamRevision{Value: 3}},
	)

	v := s.assertConsistencyViolation(err, 1)
	assert.Equal(s.T(), int32(0), v.Violations[0].CheckIndex)
	assert.Equal(s.T(), checkStream, v.Violations[0].Stream)
}
