package test

import (
	"context"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
	"time"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

func TestProjectionSuite(t *testing.T) {
	suite.Run(t, new(ProjectionSuite))
}

type ProjectionSuite struct {
	suite.Suite
	fixture *ProjectFixture
}

func (s *ProjectionSuite) SetupTest() {
	clientFixture := NewSecureSingleNodeClientFixture(s.T())
	s.fixture = NewProjectFixture(s.T(), clientFixture)
}

func (s *ProjectionSuite) TestCreateProjection() {
	fixture := s.fixture

	script, err := os.ReadFile("../resources/test/projection.js")
	s.NoError(err)
	name := fixture.NewProjectionName()

	err = s.fixture.ProjectionClient().Create(context.Background(), name, string(script), kurrentdb.CreateProjectionOptions{})
	s.NoError(err)

	fixture.WaitUntilProjectionStatusIs(s.T(), 5*time.Minute, name, "Running")
}

func (s *ProjectionSuite) TestDeleteProjection() {
	fixture := s.fixture
	client := s.fixture.ProjectionClient()
	script, err := os.ReadFile("../resources/test/projection.js")
	s.NoError(err)
	name := fixture.NewProjectionName()

	err = client.Create(context.Background(), name, string(script), kurrentdb.CreateProjectionOptions{})
	s.NoError(err)

	fixture.WaitUntilProjectionStatusIs(s.T(), 5*time.Minute, name, "Running")

	err = client.Disable(context.Background(), name, kurrentdb.GenericProjectionOptions{})
	s.NoError(err)

	fixture.WaitUntilProjectionStatusIs(s.T(), 5*time.Minute, name, "Stopped")

	done := make(chan bool)
	go func() {
		for {
			err = client.Delete(context.Background(), name, kurrentdb.DeleteProjectionOptions{})
			if esdbErr, ok := kurrentdb.FromError(err); !ok {
				if !esdbErr.IsErrorCode(kurrentdb.ErrorCodeUnknown) {
					s.T().Errorf("Delete error: %v", esdbErr)
				}
				time.Sleep(500 * time.Millisecond)
				continue
			}
			break
		}
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Minute):
		s.T().Fatal("Delete timeout")
	}
}

func (s *ProjectionSuite) TestUpdateProjection() {
	fixture := s.fixture
	client := s.fixture.ProjectionClient()

	script, err := os.ReadFile("../resources/test/projection.js")
	s.NoError(err)
	name := fixture.NewProjectionName()

	err = client.Create(context.Background(), name, string(script), kurrentdb.CreateProjectionOptions{})
	for i := 0; i < 100 && err != nil; i++ {
		if esdbErr, ok := kurrentdb.FromError(err); ok && esdbErr.IsErrorCode(kurrentdb.ErrorCodeUnknown) {
			time.Sleep(1 * time.Second)
			err = client.Create(context.Background(), name, string(script), kurrentdb.CreateProjectionOptions{})
		}
	}
	s.NoError(err)

	fixture.WaitUntilProjectionStatusIs(s.T(), 5*time.Minute, name, "Running")

	updatedScript, err := os.ReadFile("../resources/test/projection-updated.js")
	s.NoError(err)

	err = client.Update(context.Background(), name, string(updatedScript), kurrentdb.UpdateProjectionOptions{})
	s.NoError(err)

	status, err := client.GetStatus(context.Background(), name, kurrentdb.GenericProjectionOptions{})
	s.NoError(err)
	s.Equal(int64(1), status.Version)
}

func (s *ProjectionSuite) TestEnableProjection() {
	fixture := s.fixture
	client := s.fixture.ProjectionClient()

	script, err := os.ReadFile("../resources/test/projection.js")
	s.NoError(err)
	name := fixture.NewProjectionName()

	err = client.Create(context.Background(), name, string(script), kurrentdb.CreateProjectionOptions{})
	s.NoError(err)

	fixture.WaitUntilProjectionStatusIs(s.T(), 5*time.Minute, name, "Running")

	err = client.Enable(context.Background(), name, kurrentdb.GenericProjectionOptions{})
	s.NoError(err)
}

func (s *ProjectionSuite) TestDisableProjection() {
	fixture := s.fixture
	client := s.fixture.ProjectionClient()
	script, err := os.ReadFile("../resources/test/projection.js")
	s.NoError(err)
	name := fixture.NewProjectionName()

	err = client.Create(context.Background(), name, string(script), kurrentdb.CreateProjectionOptions{})
	s.NoError(err)

	fixture.WaitUntilProjectionStatusIs(s.T(), 5*time.Minute, name, "Running")

	err = client.Abort(context.Background(), name, kurrentdb.GenericProjectionOptions{})
	s.NoError(err)

	fixture.WaitUntilProjectionStatusIs(s.T(), 5*time.Minute, name, "Stopped")
}

func (s *ProjectionSuite) TestResetProjection() {
	fixture := s.fixture
	client := s.fixture.ProjectionClient()
	script, err := os.ReadFile("../resources/test/projection.js")
	s.NoError(err)
	name := fixture.NewProjectionName()

	err = client.Create(context.Background(), name, string(script), kurrentdb.CreateProjectionOptions{})
	s.NoError(err)

	fixture.WaitUntilProjectionStatusIs(s.T(), 5*time.Minute, name, "Running")

	err = client.Reset(context.Background(), name, kurrentdb.ResetProjectionOptions{})
	s.NoError(err)
}

func (s *ProjectionSuite) TestGetStateProjection() {
	fixture := s.fixture
	client := s.fixture.ProjectionClient()

	streamName := fixture.NewStreamId()
	projName := fixture.NewProjectionName()
	events := make([]kurrentdb.EventData, 10)
	for i := range events {
		events[i] = fixture.CreateTestEvent()
	}

	_, err := client.Client().AppendToStream(context.Background(), streamName, kurrentdb.AppendToStreamOptions{}, events...)
	s.NoError(err)

	script, err := os.ReadFile("../resources/test/projection.js")
	s.NoError(err)
	s.NoError(client.Create(context.Background(), projName, string(script), kurrentdb.CreateProjectionOptions{}))

	fixture.WaitUntilProjectionStatusIs(s.T(), 5*time.Minute, projName, "Running")
	s.NoError(client.Enable(context.Background(), projName, kurrentdb.GenericProjectionOptions{}))

	fixture.WaitUntilProjectionStateReady(s.T(), 5*time.Minute, projName)
}

func (s *ProjectionSuite) TestGetResultProjection() {
	fixture := s.fixture
	client := s.fixture.ProjectionClient()

	streamName := fixture.NewStreamId()
	projName := fixture.NewProjectionName()
	events := make([]kurrentdb.EventData, 10)
	for i := range events {
		events[i] = fixture.CreateTestEvent()
	}

	_, err := client.Client().AppendToStream(context.Background(), streamName, kurrentdb.AppendToStreamOptions{}, events...)
	s.NoError(err)

	script, err := os.ReadFile("../resources/test/projection.js")
	s.NoError(err)
	s.NoError(client.Create(context.Background(), projName, string(script), kurrentdb.CreateProjectionOptions{}))

	fixture.WaitUntilProjectionStatusIs(s.T(), 5*time.Minute, projName, "Running")
	s.NoError(client.Enable(context.Background(), projName, kurrentdb.GenericProjectionOptions{}))

	fixture.WaitUntilProjectionResultReady(s.T(), 5*time.Minute, projName)
}
