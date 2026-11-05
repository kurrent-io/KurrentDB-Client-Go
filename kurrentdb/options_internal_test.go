package kurrentdb

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestConfigureGrpcCallSendsRequiresLeader(t *testing.T) {
	tests := []struct {
		name       string
		preference NodePreference
		options    options
		force      bool
		want       bool
	}{
		{"read option, follower preference", NodePreferenceFollower, &ReadStreamOptions{RequiresLeader: true}, true, true},
		{"leader preference", NodePreferenceLeader, &SubscribeToAllOptions{}, true, true},
		{"neither", NodePreferenceFollower, &ReadStreamOptions{}, true, false},
		{"not forwarded", NodePreferenceLeader, &ReadStreamOptions{RequiresLeader: true}, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := &Configuration{NodePreference: tt.preference}

			_, ctx, cancel := configureGrpcCall_(context.Background(), conf, tt.options, nil, nil, tt.force)
			defer cancel()

			// Only outgoing metadata is sent to the server.
			md, _ := metadata.FromOutgoingContext(ctx)
			values := md.Get("requires-leader")
			got := len(values) == 1 && values[0] == "true"

			if got != tt.want {
				t.Fatalf("requires-leader sent = %v, want %v (outgoing metadata: %v)", got, tt.want, md)
			}
		})
	}
}
