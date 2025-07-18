package test

import (
	"context"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestDeleteSuite(t *testing.T) {
	suite.Run(t, new(DeleteTestSuite))
}

type DeleteTestSuite struct {
	suite.Suite
	fixture *ClientFixture
}

func (suite *DeleteTestSuite) SetupTest() {
	suite.fixture = NewSecureSingleNodeClientFixture(suite.T())
}

func (suite *DeleteTestSuite) TearDownTest() {
	suite.fixture.Close(suite.T())
}

func (suite *DeleteTestSuite) TestCanDeleteStream() {
	client := suite.fixture.Client()

	opts := kurrentdb.DeleteStreamOptions{
		StreamState: kurrentdb.Revision(0),
	}

	streamId := suite.fixture.NewStreamId()
	_, err := client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{}, suite.fixture.CreateTestEvent())
	assert.NoError(suite.T(), err)

	deleteResult, err := client.DeleteStream(context.Background(), streamId, opts)
	if err != nil {
		suite.T().Fatalf("Unexpected failure %+v", err)
	}

	assert.True(suite.T(), deleteResult.Position.Commit > 0)
	assert.True(suite.T(), deleteResult.Position.Prepare > 0)
}

func (suite *DeleteTestSuite) TestCanTombstoneStream() {
	client := suite.fixture.Client()

	streamId := suite.fixture.NewStreamId()
	_, err := client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{}, suite.fixture.CreateTestEvent())
	deleteResult, err := client.TombstoneStream(context.Background(), streamId, kurrentdb.TombstoneStreamOptions{
		StreamState: kurrentdb.Revision(0),
	})
	if err != nil {
		suite.T().Fatalf("Unexpected failure %+v", err)
	}

	assert.True(suite.T(), deleteResult.Position.Commit > 0)
	assert.True(suite.T(), deleteResult.Position.Prepare > 0)

	_, err = client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{}, suite.fixture.CreateTestEvent())
	require.Error(suite.T(), err)
}

func (suite *DeleteTestSuite) TestDetectStreamDeleted() {
	client := suite.fixture.Client()

	streamId := suite.fixture.NewStreamId()
	event := suite.fixture.CreateTestEvent()

	_, err := client.AppendToStream(context.Background(), streamId, kurrentdb.AppendToStreamOptions{}, event)
	require.NoError(suite.T(), err)

	_, err = client.TombstoneStream(context.Background(), streamId, kurrentdb.TombstoneStreamOptions{})
	require.NoError(suite.T(), err)

	stream, err := client.ReadStream(context.Background(), streamId, kurrentdb.ReadStreamOptions{}, 1)
	require.NoError(suite.T(), err)
	defer stream.Close()

	evt, err := stream.Recv()
	require.Nil(suite.T(), evt)

	kurrentDbError, ok := kurrentdb.FromError(err)
	require.False(suite.T(), ok)
	require.Equal(suite.T(), kurrentdb.ErrorCodeStreamDeleted, kurrentDbError.Code())
}
