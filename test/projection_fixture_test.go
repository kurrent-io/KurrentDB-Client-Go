package test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

const projectionsStreamPrefix = "$projections-"
const projectionUpdatedEventType = "$ProjectionUpdated"

type ProjectFixture struct {
	*ClientFixture
	projectionClient *kurrentdb.ProjectionClient
}

func NewProjectFixture(t *testing.T, clientFixture *ClientFixture) *ProjectFixture {
	t.Helper()

	projectionClient := kurrentdb.NewProjectionClientFromExistingClient(clientFixture.Client())

	return &ProjectFixture{
		projectionClient: projectionClient,
		ClientFixture:    clientFixture,
	}
}

func (f *ProjectFixture) WaitUntilProjectionStatusIs(t *testing.T, timeout time.Duration, name string, targetStatus string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Timeout waiting for status %s", targetStatus)
		default:
			status, err := f.projectionClient.GetStatus(context.Background(), name, kurrentdb.GenericProjectionOptions{})
			if err != nil {
				t.Logf("Status check error: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}
			if strings.Contains(status.Status, targetStatus) {
				return
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func (f *ProjectFixture) WaitUntilProjectionStateReady(t *testing.T, duration time.Duration, name string) {
	done := make(chan *state)
	go func() {
		for {
			result, err := f.projectionClient.GetState(context.Background(), name, kurrentdb.GetStateProjectionOptions{})

			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			var state state
			content, err := result.MarshalJSON()

			if err != nil {
				t.Errorf("error when unmarshalling protobuf types to json: %v", err)
			}

			err = json.Unmarshal(content, &state)
			if err != nil {
				t.Errorf("error when unmarshalling projection internal state type: %v", err)
			}

			done <- &state
			return
		}
	}()

	select {
	case <-done:
		return
	case <-time.After(duration):
		t.Errorf("unable to get projection '%s' internal state in a timely manner", name)
	}
}

func (f *ProjectFixture) WaitUntilProjectionResultReady(t *testing.T, duration time.Duration, name string) {
	done := make(chan *state)
	go func() {
		for {
			result, err := f.projectionClient.GetResult(context.Background(), name, kurrentdb.GetResultProjectionOptions{})

			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			var state state
			content, err := result.MarshalJSON()

			if err != nil {
				t.Errorf("error when unmarshalling protobuf types to json: %v", err)
			}

			err = json.Unmarshal(content, &state)
			if err != nil {
				t.Errorf("error when unmarshalling projection internal result type: %v", err)
			}

			done <- &state
			return
		}
	}()

	select {
	case <-done:
		return
	case <-time.After(duration):
		t.Errorf("unable to get projection '%s' internal state in a timely manner", name)
	}
}

// ProjectionDefinitionMetadata reads the caller-supplied metadata stamped on
// the most recent $ProjectionUpdated definition event for the named projection.
// The server serializes it as a protobuf Struct on the event's metadata.
func (f *ProjectFixture) ProjectionDefinitionMetadata(t *testing.T, name string) map[string]interface{} {
	t.Helper()

	stream, err := f.Client().ReadStream(context.Background(), projectionsStreamPrefix+name, kurrentdb.ReadStreamOptions{
		Direction: kurrentdb.Backwards,
		From:      kurrentdb.End{},
	}, 100)
	if err != nil {
		t.Fatalf("error reading projection definition stream: %v", err)
	}
	defer stream.Close()

	for {
		event, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			t.Fatalf("error receiving projection definition event: %v", err)
		}

		recorded := event.OriginalEvent()
		if recorded.EventType != projectionUpdatedEventType || len(recorded.UserMetadata) == 0 {
			continue
		}

		var s structpb.Struct
		if err := proto.Unmarshal(recorded.UserMetadata, &s); err != nil {
			t.Fatalf("error unmarshalling projection properties: %v", err)
		}

		return s.AsMap()
	}
}

func (f *ProjectFixture) Close(t *testing.T) {
	if f.projectionClient != nil {
		if err := f.projectionClient.Close(); err != nil {
			t.Error("Error closing client:", err)
		}
	}

	f.ClientFixture.Close(t)
}

type state struct {
	Foo struct {
		Baz struct {
			Count float64 `json:"count"`
		} `json:"baz"`
	} `json:"foo"`
}
