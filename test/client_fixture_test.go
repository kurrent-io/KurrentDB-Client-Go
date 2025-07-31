package test

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type KurrentDBVersion struct {
	Maj   int
	Min   int
	Patch int
}

type VersionPredicateFn = func(KurrentDBVersion) bool

type ClientFixture struct {
	mu     sync.Mutex
	client *kurrentdb.Client
	config *kurrentdb.Configuration
}

func NewInsecureClientFixture(t *testing.T) *ClientFixture {
	t.Helper()

	config, err := kurrentdb.ParseConnectionString(fmt.Sprintf("kurrentdb://admin:changeit@localhost:2114?tls=false"))
	require.NoError(t, err, "Failed to parse connection string")

	client, err := kurrentdb.NewClient(config)
	require.NoError(t, err, "Failed to create KurrentDB client")

	return &ClientFixture{
		client: client,
		config: config,
	}
}

func NewSecureSingleNodeClientFixture(t *testing.T) *ClientFixture {
	t.Helper()
	tlsCaFile := "../certs/ca/ca.crt"

	config, err := kurrentdb.ParseConnectionString(fmt.Sprintf("kurrentdb://admin:changeit@localhost:2115?tls=true&tlscafile=%s", tlsCaFile))
	require.NoError(t, err, "Failed to parse connection string")

	client, err := kurrentdb.NewClient(config)
	require.NoError(t, err, "Failed to create KurrentDB client")

	return &ClientFixture{
		client: client,
		config: config,
	}
}

func NewSecureClusterClientFixture(t *testing.T) *ClientFixture {
	t.Helper()
	tlsCaFile := "../certs/ca/ca.crt"

	config, err := kurrentdb.ParseConnectionString(fmt.Sprintf("kurrentdb://admin:changeit@localhost:2111,localhost:2112,localhost:2113?tls=true&tlscafile=%s", tlsCaFile))
	require.NoError(t, err, "Failed to parse connection string")

	client, err := kurrentdb.NewClient(config)
	require.NoError(t, err, "Failed to create KurrentDB client")

	return &ClientFixture{
		client: client,
		config: config,
	}
}

func (f *ClientFixture) Client() *kurrentdb.Client {
	return f.client
}

func (f *ClientFixture) ProjectionClient() *kurrentdb.ProjectionClient {
	return kurrentdb.NewProjectionClientFromExistingClient(f.client)
}

type TestEventOptions struct {
	EventType   string
	ContentType kurrentdb.ContentType
	Data        []byte
	Metadata    []byte
}

func (f *ClientFixture) CreateTestEvent(options ...TestEventOptions) kurrentdb.EventData {
	eventType := "TestEvent"
	contentType := kurrentdb.ContentTypeBinary
	data := []byte{0xb, 0xe, 0xe, 0xf}
	var metadata []byte

	if len(options) > 0 {
		opt := options[0]
		if opt.EventType != "" {
			eventType = opt.EventType
		}
		if opt.ContentType != 0 {
			contentType = opt.ContentType
		}
		if opt.Data != nil {
			data = opt.Data
		}
		if opt.Metadata != nil {
			metadata = opt.Metadata
		}
	}

	event := kurrentdb.EventData{
		EventType:   eventType,
		ContentType: contentType,
		EventID:     uuid.New(),
		Data:        data,
		Metadata:    metadata,
	}

	return event
}

func (f *ClientFixture) CreateTestEvents(streamId string, count uint32) []kurrentdb.EventData {
	events := make([]kurrentdb.EventData, count)
	var i uint32 = 0
	for ; i < count; i++ {
		events[i] = f.CreateTestEvent(TestEventOptions{
			EventType: fmt.Sprintf("TestEvent-%d", i+1),
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()

	opts := kurrentdb.AppendToStreamOptions{
		StreamState: kurrentdb.Any{},
	}

	_, err := f.client.AppendToStream(ctx, streamId, opts, events...)
	if err != nil {
		panic(err)
	}

	return events
}

func (f *ClientFixture) CollectEvents(stream *kurrentdb.ReadStream) ([]*kurrentdb.ResolvedEvent, error) {
	var events []*kurrentdb.ResolvedEvent

	for {
		event, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		events = append(events, event)
	}

	return events, nil
}

func (f *ClientFixture) NewStreamId() string {
	return uuid.New().String()
}

func (f *ClientFixture) NewGroupId() string {
	return uuid.New().String()
}

func (f *ClientFixture) NewProjectionName() string {
	return uuid.New().String()
}

func (f *ClientFixture) WaitWithTimeout(wg *sync.WaitGroup, duration time.Duration) bool {
	channel := make(chan struct{})
	go func() {
		defer close(channel)
		wg.Wait()
	}()
	select {
	case <-channel:
		return false
	case <-time.After(duration):
		return true
	}
}

func (f *ClientFixture) Close(t *testing.T) {
	if f.client != nil {
		if err := f.client.Close(); err != nil {
			t.Error("Error closing client:", err)
		}
	}
}

func (f *ClientFixture) IsKurrentDbVersion20() bool {
	return isKurrentDbVersion(func(version KurrentDBVersion) bool {
		return version.Maj < 21
	})
}

func isKurrentDbVersion(predicate VersionPredicateFn) bool {
	value, exists := os.LookupEnv("KURRENTDB_DOCKER_TAG")
	if !exists || value == "ci" {
		return false
	}

	parts := strings.Split(value, "-")
	versionNumbers := strings.Split(parts[0], ".")

	version := KurrentDBVersion{
		Maj:   mustConvertToInt(versionNumbers[0]),
		Min:   mustConvertToInt(versionNumbers[1]),
		Patch: mustConvertToInt(versionNumbers[2]),
	}

	return predicate(version)
}

func mustConvertToInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return val
}
