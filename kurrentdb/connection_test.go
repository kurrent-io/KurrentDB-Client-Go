package kurrentdb

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDiscoveryFailureReleasesConcurrentCallers(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	config, err := ParseConnectionString("kurrentdb://" + listener.Addr().String() + "?tls=false&maxDiscoverAttempts=1&gossipTimeout=1&discoveryInterval=100")
	require.NoError(t, err)

	for _, timeout := range []time.Duration{0, 250 * time.Millisecond} {
		client, err := NewClient(config)
		require.NoError(t, err)

		errs := make(chan error, 2)
		for i := 0; i < 2; i++ {
			go func() {
				ctx := context.Background()
				if timeout > 0 {
					var cancel context.CancelFunc
					ctx, cancel = context.WithTimeout(ctx, timeout)
					defer cancel()
				}
				_, err := client.ReadStream(ctx, "repro", ReadStreamOptions{}, 1)
				errs <- err
			}()
		}

		for i := 0; i < 2; i++ {
			select {
			case err := <-errs:
				require.Error(t, err)
				if timeout == 0 {
					require.ErrorContains(t, err, "maximum discovery attempt count reached")
				}
			case <-time.After(5 * time.Second):
				t.Fatalf("caller still blocked (timeout %v)", timeout)
			}
		}

		require.NoError(t, client.Close())
	}
}

func TestCloseReleasesCallerDuringDiscovery(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	config, err := ParseConnectionString("kurrentdb://" + listener.Addr().String() + "?tls=false&maxDiscoverAttempts=10&gossipTimeout=5")
	require.NoError(t, err)

	client, err := NewClient(config)
	require.NoError(t, err)

	errs := make(chan error, 1)
	go func() {
		_, err := client.ReadStream(context.Background(), "repro", ReadStreamOptions{}, 1)
		errs <- err
	}()

	time.Sleep(100 * time.Millisecond)
	require.NoError(t, client.Close())

	select {
	case err := <-errs:
		kurrentErr, _ := FromError(err)
		require.Equal(t, ErrorCodeConnectionClosed, kurrentErr.Code())
	case <-time.After(time.Second):
		t.Fatal("caller still blocked after Close")
	}
}
