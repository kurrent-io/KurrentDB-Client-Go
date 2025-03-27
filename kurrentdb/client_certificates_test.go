package kurrentdb_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/kurrent-io/KurrentDB-Client-Go/v1/kurrentdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ClientCertificatesSingleNodeTests(t *testing.T, emptyDBContainer *Container) {
	t.Run("AuthenticationTests", func(t *testing.T) {
		t.Run("ValidClientCertificatesTlsRelativePath", func(t *testing.T) {
			testValidClientCertificatesTlsWithRelativePath(t, emptyDBContainer.Endpoint)
		})
		t.Run("ValidClientCertificatesTlsWithAbsolutePath", func(t *testing.T) {
			testValidClientCertificatesTlsWithAbsolutePath(t, emptyDBContainer.Endpoint)
		})
		t.Run("InvalidUserCertificates", func(t *testing.T) {
			testInvalidUserCertificates(t, emptyDBContainer.Endpoint)
		})
		t.Run("MissingCertificateFile", func(t *testing.T) {
			testMissingCertificateFile(t, emptyDBContainer.Endpoint)
		})
	})
}

func ClientCertificatesClusterNodesTests(t *testing.T) {
	t.Run("AuthenticationTests", func(t *testing.T) {
		var endpoint = "localhost:2111,localhost:2112,localhost:2113"

		t.Run("ValidClientCertificatesTlsRelativePath", func(t *testing.T) {
			testValidClientCertificatesTlsWithRelativePath(t, endpoint)
		})
		t.Run("ValidClientCertificatesTlsWithAbsolutePath", func(t *testing.T) {
			testValidClientCertificatesTlsWithAbsolutePath(t, endpoint)
		})
		t.Run("InvalidUserCertificates", func(t *testing.T) {
			testInvalidUserCertificates(t, endpoint)
		})
		t.Run("MissingCertificateFile", func(t *testing.T) {
			testMissingCertificateFile(t, endpoint)
		})
	})
}

func testValidClientCertificatesTlsWithRelativePath(t *testing.T, endpoint string) {
	tlsCaFile := "../certs/ca/ca.crt"
	userCertFile := "../certs/user-admin/user-admin.crt"
	userKeyFile := "../certs/user-admin/user-admin.key"

	config, err := kurrentdb.ParseConnectionString(fmt.Sprintf("esdb://admin:changeit@%s?tls=true&tlscafile=%s&userCertFile=%s&userKeyFile=%s", endpoint, tlsCaFile, userCertFile, userKeyFile))
	if err != nil {
		t.Fatalf("Unexpected configuration error: %s", err.Error())
	}

	c, err := kurrentdb.NewClient(config)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
	defer c.Close()

	numberOfEventsToRead := 1
	numberOfEvents := uint64(numberOfEventsToRead)
	opts := kurrentdb.ReadAllOptions{
		From:           kurrentdb.Start{},
		Direction:      kurrentdb.Backwards,
		ResolveLinkTos: true,
	}

	stream, err := c.ReadAll(context.Background(), opts, numberOfEvents)
	require.NoError(t, err)
	defer stream.Close()
	evt, err := stream.Recv()
	require.Nil(t, evt)
	require.True(t, errors.Is(err, io.EOF))
}

func testValidClientCertificatesTlsWithAbsolutePath(t *testing.T, endpoint string) {
	userCertFile, err := filepath.Abs("../certs/user-admin/user-admin.crt")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	userKeyFile, err := filepath.Abs("../certs/user-admin/user-admin.key")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	tlsCaFile, err := filepath.Abs("../certs/ca/ca.crt")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	config, err := kurrentdb.ParseConnectionString(fmt.Sprintf("esdb://admin:changeit@%s?tls=true&tlscafile=%s&usercertfile=%s&userkeyfile=%s", endpoint, tlsCaFile, userCertFile, userKeyFile))
	if err != nil {
		t.Fatalf("Unexpected configuration error: %s", err.Error())
	}

	c, err := kurrentdb.NewClient(config)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
	defer c.Close()

	numberOfEventsToRead := 1
	numberOfEvents := uint64(numberOfEventsToRead)
	opts := kurrentdb.ReadAllOptions{
		From:           kurrentdb.Start{},
		Direction:      kurrentdb.Backwards,
		ResolveLinkTos: true,
	}

	stream, err := c.ReadAll(context.Background(), opts, numberOfEvents)
	require.NoError(t, err)
	defer stream.Close()
	evt, err := stream.Recv()
	require.Nil(t, evt)
	require.True(t, errors.Is(err, io.EOF))
}

func testInvalidUserCertificates(t *testing.T, endpoint string) {
	tlsCaFile := "../certs/ca/ca.crt"
	userCertFile := "../certs/user-invalid/user-invalid.crt"
	userKeyFile := "../certs/user-invalid/user-invalid.key"

	config, err := kurrentdb.ParseConnectionString(fmt.Sprintf("esdb://admin:changeit@%s?tls=true&tlscafile=%s&usercertfile=%s&userkeyfile=%s", endpoint, tlsCaFile, userCertFile, userKeyFile))
	if err != nil {
		t.Fatalf("Unexpected configuration error: %s", err.Error())
	}

	c, err := kurrentdb.NewClient(config)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
	defer c.Close()

	testEvent := createTestEvent()

	streamID := uuid.NewString()
	opts := kurrentdb.AppendToStreamOptions{
		StreamState: kurrentdb.Any{},
	}

	result, err := c.AppendToStream(context.Background(), streamID, opts, testEvent)
	require.Nil(t, result)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Unauthenticated")
}

func testMissingCertificateFile(t *testing.T, endpoint string) {
	tlsCaFile := "../certs/ca/ca.crt"
	userCertFile := "../certs/user-admin/user-admin.crt"

	config, err := kurrentdb.ParseConnectionString(fmt.Sprintf("esdb://admin:changeit@%s?tls=true&tlscafile=%s&usercertfile=%s", endpoint, tlsCaFile, userCertFile))

	_, err = kurrentdb.NewClient(config)
	esdbErr, ok := kurrentdb.FromError(err)
	require.False(t, ok)
	require.NotNil(t, esdbErr)
	assert.Contains(t, esdbErr.Error(), "both userCertFile and userKeyFile must be provided")
}
