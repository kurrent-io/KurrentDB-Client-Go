package kurrentdb_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/require"
)

func TestProjection(t *testing.T) {
	t.Run("cluster", func(t *testing.T) {
		secureClusterClientFixture := NewSecureClusterClientFixture(t)
		fixture := NewProjectFixture(t, secureClusterClientFixture)
		defer fixture.Close(t)

		runProjectionTests(t, fixture)
	})

	t.Run("singleNode", func(t *testing.T) {
		secureSingleNodeClientFixture := NewSecureSingleNodeClientFixture(t)
		fixture := NewProjectFixture(t, secureSingleNodeClientFixture)
		defer fixture.Close(t)

		runProjectionTests(t, fixture)
	})

}

func runProjectionTests(t *testing.T, fixture *ProjectFixture) {
	client := kurrentdb.NewProjectionClientFromExistingClient(fixture.Client())

	t.Run("createProjection", func(t *testing.T) {
		script, err := os.ReadFile("../resources/test/projection.js")
		require.NoError(t, err)
		name := fixture.NewProjectionName()

		err = client.Create(context.Background(), name, string(script), kurrentdb.CreateProjectionOptions{})
		require.NoError(t, err)

		fixture.WaitUntilProjectionStatusIs(t, 5*time.Minute, name, "Running")
	})

	t.Run("deleteProjection", func(t *testing.T) {
		script, err := os.ReadFile("../resources/test/projection.js")
		require.NoError(t, err)
		name := fixture.NewProjectionName()

		err = client.Create(context.Background(), name, string(script), kurrentdb.CreateProjectionOptions{})
		require.NoError(t, err)

		fixture.WaitUntilProjectionStatusIs(t, 5*time.Minute, name, "Running")

		err = client.Disable(context.Background(), name, kurrentdb.GenericProjectionOptions{})
		require.NoError(t, err)

		fixture.WaitUntilProjectionStatusIs(t, 5*time.Minute, name, "Stopped")

		done := make(chan bool)
		go func() {
			for {
				err = client.Delete(context.Background(), name, kurrentdb.DeleteProjectionOptions{})
				if esdbErr, ok := kurrentdb.FromError(err); !ok {
					if !esdbErr.IsErrorCode(kurrentdb.ErrorCodeUnknown) {
						t.Errorf("Delete error: %v", esdbErr)
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
			t.Fatal("Delete timeout")
		}
	})

	t.Run("updateProjection", func(t *testing.T) {
		script, err := os.ReadFile("../resources/test/projection.js")
		require.NoError(t, err)
		name := fixture.NewProjectionName()

		err = client.Create(context.Background(), name, string(script), kurrentdb.CreateProjectionOptions{})
		for i := 0; i < 100 && err != nil; i++ {
			if esdbErr, ok := kurrentdb.FromError(err); ok && esdbErr.IsErrorCode(kurrentdb.ErrorCodeUnknown) {
				time.Sleep(1 * time.Second)
				err = client.Create(context.Background(), name, string(script), kurrentdb.CreateProjectionOptions{})
			}
		}
		require.NoError(t, err)

		fixture.WaitUntilProjectionStatusIs(t, 5*time.Minute, name, "Running")

		updatedScript, err := os.ReadFile("../resources/test/projection-updated.js")
		require.NoError(t, err)

		err = client.Update(context.Background(), name, string(updatedScript), kurrentdb.UpdateProjectionOptions{})
		require.NoError(t, err)

		status, err := client.GetStatus(context.Background(), name, kurrentdb.GenericProjectionOptions{})
		require.NoError(t, err)
		require.Equal(t, int64(1), status.Version)
	})

	t.Run("enableProjection", func(t *testing.T) {
		script, err := os.ReadFile("../resources/test/projection.js")
		require.NoError(t, err)
		name := fixture.NewProjectionName()

		err = client.Create(context.Background(), name, string(script), kurrentdb.CreateProjectionOptions{})
		require.NoError(t, err)

		fixture.WaitUntilProjectionStatusIs(t, 5*time.Minute, name, "Running")

		err = client.Enable(context.Background(), name, kurrentdb.GenericProjectionOptions{})
		require.NoError(t, err)
	})

	t.Run("disableProjection", func(t *testing.T) {
		script, err := os.ReadFile("../resources/test/projection.js")
		require.NoError(t, err)
		name := fixture.NewProjectionName()

		err = client.Create(context.Background(), name, string(script), kurrentdb.CreateProjectionOptions{})
		require.NoError(t, err)

		fixture.WaitUntilProjectionStatusIs(t, 5*time.Minute, name, "Running")

		err = client.Abort(context.Background(), name, kurrentdb.GenericProjectionOptions{})
		require.NoError(t, err)

		fixture.WaitUntilProjectionStatusIs(t, 5*time.Minute, name, "Stopped")
	})

	t.Run("resetProjection", func(t *testing.T) {
		script, err := os.ReadFile("../resources/test/projection.js")
		require.NoError(t, err)
		name := fixture.NewProjectionName()

		err = client.Create(context.Background(), name, string(script), kurrentdb.CreateProjectionOptions{})
		require.NoError(t, err)

		fixture.WaitUntilProjectionStatusIs(t, 5*time.Minute, name, "Running")

		err = client.Reset(context.Background(), name, kurrentdb.ResetProjectionOptions{})
		require.NoError(t, err)
	})

	t.Run("getStateProjection", func(t *testing.T) {
		streamName := fixture.NewStreamId()
		projName := fixture.NewProjectionName()
		events := make([]kurrentdb.EventData, 10)
		for i := range events {
			events[i] = fixture.CreateTestEvent()
		}

		_, err := client.Client().AppendToStream(context.Background(), streamName, kurrentdb.AppendToStreamOptions{}, events...)
		require.NoError(t, err)

		script, err := os.ReadFile("../resources/test/projection.js")
		require.NoError(t, err)
		require.NoError(t, client.Create(context.Background(), projName, string(script), kurrentdb.CreateProjectionOptions{}))

		fixture.WaitUntilProjectionStatusIs(t, 5*time.Minute, projName, "Running")
		require.NoError(t, client.Enable(context.Background(), projName, kurrentdb.GenericProjectionOptions{}))

		fixture.WaitUntilProjectionStateReady(t, 5*time.Minute, projName)
	})

	t.Run("getResultProjection", func(t *testing.T) {
		streamName := fixture.NewStreamId()
		projName := fixture.NewProjectionName()
		events := make([]kurrentdb.EventData, 10)
		for i := range events {
			events[i] = fixture.CreateTestEvent()
		}

		_, err := client.Client().AppendToStream(context.Background(), streamName, kurrentdb.AppendToStreamOptions{}, events...)
		require.NoError(t, err)

		script, err := os.ReadFile("../resources/test/projection.js")
		require.NoError(t, err)
		require.NoError(t, client.Create(context.Background(), projName, string(script), kurrentdb.CreateProjectionOptions{}))

		fixture.WaitUntilProjectionStatusIs(t, 5*time.Minute, projName, "Running")
		require.NoError(t, client.Enable(context.Background(), projName, kurrentdb.GenericProjectionOptions{}))

		fixture.WaitUntilProjectionResultReady(t, 5*time.Minute, projName)
	})
}
