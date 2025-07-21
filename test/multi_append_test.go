package test

import (
	"context"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"slices"
	"testing"
)

type MultiAppendTestSuite struct {
	suite.Suite
	fixture *ClientFixture
}

func TestMultiAppendEventsSuite(t *testing.T) {
	suite.Run(t, new(MultiAppendTestSuite))
}

func (s *MultiAppendTestSuite) SetupTest() {
	s.fixture = NewInsecureClientFixture(s.T())
}

func (s *MultiAppendTestSuite) TestMultiStreamAppendHappyPath() {
	client := s.fixture.Client()

	version, err := client.GetServerVersion()

	assert.NoError(s.T(), err)

	if version.Major < 25 {
		s.T().Skip("Multi-stream append is not supported in versions prior to 25.0")
	}

	stream1 := s.fixture.NewStreamId()
	stream2 := s.fixture.NewStreamId()
	events1 := s.fixture.CreateTestEvents(stream1, 10)
	events2 := s.fixture.CreateTestEvents(stream2, 10)

	requests := make([]kurrentdb.AppendStreamRequest, 2)
	requests[0] = kurrentdb.AppendStreamRequest{
		StreamName:          stream1,
		Events:              slices.Values(events1),
		ExpectedStreamState: kurrentdb.Any{},
	}

	requests[1] = kurrentdb.AppendStreamRequest{
		StreamName:          stream2,
		Events:              slices.Values(events2),
		ExpectedStreamState: kurrentdb.Any{},
	}

	result, err := client.MultiStreamAppend(context.Background(), slices.Values(requests), kurrentdb.AppendToStreamOptions{})

	assert.NoError(s.T(), err, "Unexpected error during multi-stream append")
	assert.True(s.T(), result.IsSuccessful())
}

func (s *MultiAppendTestSuite) TestMultiStreamAppendUnsupported() {
	client := s.fixture.Client()

	version, err := client.GetServerVersion()

	assert.NoError(s.T(), err)

	if version.Major >= 25 {
		s.T().Skip("Multi-stream append is supported in from 25.0 and upwards")
	}

	stream1 := s.fixture.NewStreamId()
	stream2 := s.fixture.NewStreamId()
	events1 := s.fixture.CreateTestEvents(stream1, 10)
	events2 := s.fixture.CreateTestEvents(stream2, 10)

	requests := make([]kurrentdb.AppendStreamRequest, 2)
	requests[0] = kurrentdb.AppendStreamRequest{
		StreamName:          stream1,
		Events:              slices.Values(events1),
		ExpectedStreamState: kurrentdb.Any{},
	}

	requests[1] = kurrentdb.AppendStreamRequest{
		StreamName:          stream2,
		Events:              slices.Values(events2),
		ExpectedStreamState: kurrentdb.Any{},
	}

	result, err := client.MultiStreamAppend(context.Background(), slices.Values(requests), kurrentdb.AppendToStreamOptions{})

	assert.Nil(s.T(), result)

	kdbError, _ := kurrentdb.FromError(err)
	assert.Equal(s.T(), kurrentdb.ErrorCodeUnsupportedFeature, kdbError.Code(), "Expected unsupported feature error code")
}
