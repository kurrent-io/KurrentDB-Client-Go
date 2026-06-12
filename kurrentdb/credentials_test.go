package kurrentdb

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type fakeOptions struct{ creds *Credentials }

func (fakeOptions) kind() operationKind         { return regularOperation }
func (o fakeOptions) credentials() *Credentials { return o.creds }
func (fakeOptions) deadline() *time.Duration    { return nil }
func (fakeOptions) requiresLeader() bool        { return false }

func headerFrom(t *testing.T, creds credentials.PerRPCCredentials) string {
	t.Helper()
	if creds == nil {
		return ""
	}
	md, err := creds.GetRequestMetadata(context.Background())
	require.NoError(t, err)
	return md["Authorization"]
}

func TestCredentialsAuthorizationHeader(t *testing.T) {
	basic := "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:changeit"))

	for _, tc := range []struct {
		name  string
		creds *Credentials
		want  string
	}{
		{"nil", nil, ""},
		{"empty", &Credentials{}, ""},
		{"basic", &Credentials{Login: "admin", Password: "changeit"}, basic},
		{"bearer", &Credentials{BearerToken: "abc.def"}, "Bearer abc.def"},
		{"bearer wins over basic", &Credentials{Login: "admin", Password: "changeit", BearerToken: "abc.def"}, "Bearer abc.def"},
		{"login only", &Credentials{Login: "admin"}, "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:"))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.creds.authorizationHeader())
		})
	}
}

func TestStaticAuthPerRPCCredentials(t *testing.T) {
	t.Run("attaches the rendered header", func(t *testing.T) {
		creds := staticAuthPerRPCCredentials(&Credentials{BearerToken: "tok"})
		md, err := creds.GetRequestMetadata(context.Background())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"Authorization": "Bearer tok"}, md)
	})

	t.Run("no header when credentials are empty", func(t *testing.T) {
		creds := staticAuthPerRPCCredentials(&Credentials{})
		md, err := creds.GetRequestMetadata(context.Background())
		require.NoError(t, err)
		assert.Nil(t, md)
	})

	t.Run("requires transport security", func(t *testing.T) {
		assert.True(t, staticAuthPerRPCCredentials(&Credentials{}).RequireTransportSecurity())
	})
}

func TestProviderAuthPerRPCCredentials(t *testing.T) {
	t.Run("resolves the provider on every call", func(t *testing.T) {
		calls := 0
		creds := providerAuthPerRPCCredentials(func(context.Context) (*Credentials, error) {
			calls++
			return &Credentials{BearerToken: "tok"}, nil
		})

		for i := 0; i < 3; i++ {
			md, err := creds.GetRequestMetadata(context.Background())
			require.NoError(t, err)
			assert.Equal(t, map[string]string{"Authorization": "Bearer tok"}, md)
		}
		assert.Equal(t, 3, calls, "provider must run once per RPC so refreshed tokens are picked up")
	})

	t.Run("propagates the provider error", func(t *testing.T) {
		boom := errors.New("token fetch failed")
		creds := providerAuthPerRPCCredentials(func(context.Context) (*Credentials, error) {
			return nil, boom
		})
		md, err := creds.GetRequestMetadata(context.Background())
		assert.Nil(t, md)
		assert.ErrorIs(t, err, boom)
	})

	t.Run("no header when the provider returns nil credentials", func(t *testing.T) {
		creds := providerAuthPerRPCCredentials(func(context.Context) (*Credentials, error) {
			return nil, nil
		})
		md, err := creds.GetRequestMetadata(context.Background())
		require.NoError(t, err)
		assert.Nil(t, md)
	})
}

func TestClientPerRPCCredentials(t *testing.T) {
	provider := func(context.Context) (*Credentials, error) { return &Credentials{BearerToken: "tok"}, nil }
	basic := "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:changeit"))

	t.Run("nil over an insecure channel", func(t *testing.T) {
		assert.Nil(t, clientPerRPCCredentials(Configuration{DisableTLS: true, CredentialsProvider: provider}))
		assert.Nil(t, clientPerRPCCredentials(Configuration{DisableTLS: true, Username: "admin", Password: "changeit"}))
	})

	t.Run("provider takes precedence over static", func(t *testing.T) {
		creds := clientPerRPCCredentials(Configuration{CredentialsProvider: provider, Username: "admin", Password: "changeit"})
		assert.Equal(t, "Bearer tok", headerFrom(t, creds))
	})

	t.Run("static basic when only username/password", func(t *testing.T) {
		assert.Equal(t, basic, headerFrom(t, clientPerRPCCredentials(Configuration{Username: "admin", Password: "changeit"})))
	})

	t.Run("nil when nothing is set", func(t *testing.T) {
		assert.Nil(t, clientPerRPCCredentials(Configuration{}))
	})
}

func TestConfigureGrpcCallCredentialPrecedence(t *testing.T) {
	conf := &Configuration{}
	clientCreds := providerAuthPerRPCCredentials(func(context.Context) (*Credentials, error) {
		return &Credentials{BearerToken: "client"}, nil
	})

	extract := func(opts []grpc.CallOption) credentials.PerRPCCredentials {
		for _, o := range opts {
			if c, ok := o.(grpc.PerRPCCredsCallOption); ok {
				return c.Creds
			}
		}
		return nil
	}

	t.Run("per-call credentials override the client provider", func(t *testing.T) {
		opts, _, cancel := configureGrpcCall_(context.Background(), conf, fakeOptions{creds: &Credentials{BearerToken: "percall"}}, nil, clientCreds, true)
		defer cancel()
		assert.Equal(t, "Bearer percall", headerFrom(t, extract(opts)))
	})

	t.Run("client provider used when there are no per-call credentials", func(t *testing.T) {
		opts, _, cancel := configureGrpcCall_(context.Background(), conf, fakeOptions{}, nil, clientCreds, true)
		defer cancel()
		assert.Equal(t, "Bearer client", headerFrom(t, extract(opts)))
	})

	t.Run("per-call basic credentials render a Basic header", func(t *testing.T) {
		basic := "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:changeit"))
		opts, _, cancel := configureGrpcCall_(context.Background(), conf, fakeOptions{creds: &Credentials{Login: "admin", Password: "changeit"}}, nil, clientCreds, true)
		defer cancel()
		assert.Equal(t, basic, headerFrom(t, extract(opts)))
	})

	t.Run("per-call credentials are dropped over an insecure channel", func(t *testing.T) {
		insecure := &Configuration{DisableTLS: true}
		opts, _, cancel := configureGrpcCall_(context.Background(), insecure, fakeOptions{creds: &Credentials{BearerToken: "percall"}}, nil, nil, true)
		defer cancel()
		assert.Nil(t, extract(opts))
	})
}

func TestResolveRequestCredentials(t *testing.T) {
	ctx := context.Background()
	provider := func(context.Context) (*Credentials, error) { return &Credentials{BearerToken: "prov"}, nil }
	basic := "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:changeit"))

	t.Run("nil over an insecure channel", func(t *testing.T) {
		conf := &Configuration{DisableTLS: true, CredentialsProvider: provider, Username: "admin", Password: "changeit"}
		creds, err := conf.resolveRequestCredentials(ctx, &Credentials{BearerToken: "percall"})
		require.NoError(t, err)
		assert.Nil(t, creds)
	})

	t.Run("per-call takes precedence over provider and static", func(t *testing.T) {
		conf := &Configuration{CredentialsProvider: provider, Username: "admin", Password: "changeit"}
		creds, err := conf.resolveRequestCredentials(ctx, &Credentials{BearerToken: "percall"})
		require.NoError(t, err)
		assert.Equal(t, "Bearer percall", creds.authorizationHeader())
	})

	t.Run("provider takes precedence over static", func(t *testing.T) {
		conf := &Configuration{CredentialsProvider: provider, Username: "admin", Password: "changeit"}
		creds, err := conf.resolveRequestCredentials(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, "Bearer prov", creds.authorizationHeader())
	})

	t.Run("static basic when only username/password", func(t *testing.T) {
		conf := &Configuration{Username: "admin", Password: "changeit"}
		creds, err := conf.resolveRequestCredentials(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, basic, creds.authorizationHeader())
	})

	t.Run("username without password is not used (parity with the gRPC gate)", func(t *testing.T) {
		conf := &Configuration{Username: "admin"}
		creds, err := conf.resolveRequestCredentials(ctx, nil)
		require.NoError(t, err)
		assert.Nil(t, creds)
	})

	t.Run("provider error propagates", func(t *testing.T) {
		boom := errors.New("token fetch failed")
		conf := &Configuration{CredentialsProvider: func(context.Context) (*Credentials, error) { return nil, boom }}
		_, err := conf.resolveRequestCredentials(ctx, nil)
		assert.ErrorIs(t, err, boom)
	})

	t.Run("provider returning nil credentials yields nil", func(t *testing.T) {
		conf := &Configuration{CredentialsProvider: func(context.Context) (*Credentials, error) { return nil, nil }}
		creds, err := conf.resolveRequestCredentials(ctx, nil)
		require.NoError(t, err)
		assert.Nil(t, creds)
	})
}
