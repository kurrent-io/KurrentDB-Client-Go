package kurrentdb

import (
	"testing"

	streamErrors "github.com/kurrent-io/KurrentDB-Client-Go/protos/kurrentdb/protocols/v2/streams/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestExtractRichErrorDetails_StreamRevisionConflict(t *testing.T) {
	// Create a StreamRevisionConflictErrorDetails message
	details := &streamErrors.StreamRevisionConflictErrorDetails{
		Stream:           "test-stream",
		ExpectedRevision: 5,
		ActualRevision:   10,
	}

	// Convert to Any
	any, err := anypb.New(details)
	if err != nil {
		t.Fatalf("Failed to create Any from details: %v", err)
	}

	// Create a gRPC status with the details
	st := status.New(codes.FailedPrecondition, "Stream revision conflict")
	st, err = st.WithDetails(any)
	if err != nil {
		t.Fatalf("Failed to add details to status: %v", err)
	}

	// Extract rich error details
	originalErr := st.Err()
	richErr := getDetail(originalErr)

	// Verify the result
	if richErr == originalErr {
		t.Error("Expected rich error details to be extracted, but got original error")
	}

	wrongVersionErr, ok := richErr.(*WrongExpectedVersionException)
	if !ok {
		t.Errorf("Expected WrongExpectedVersionException, got %T", richErr)
	}

	if wrongVersionErr.Stream != "test-stream" {
		t.Errorf("Expected stream 'test-stream', got '%s'", wrongVersionErr.Stream)
	}

	expectedRev := StreamRevision{Value: 5}
	if wrongVersionErr.ExpectedRevision != expectedRev {
		t.Errorf("Expected revision %v, got %v", expectedRev, wrongVersionErr.ExpectedRevision)
	}

	actualRev := StreamRevision{Value: 10}
	if wrongVersionErr.ActualRevision != actualRev {
		t.Errorf("Expected revision %v, got %v", actualRev, wrongVersionErr.ActualRevision)
	}
}

func TestExtractRichErrorDetails_NoRichDetails(t *testing.T) {
	// Create a regular gRPC error without rich details
	st := status.New(codes.NotFound, "Stream not found")
	originalErr := st.Err()

	// Extract rich error details
	richErr := getDetail(originalErr)

	// Verify the result - should return the original error since no rich details
	if richErr != originalErr {
		t.Error("Expected original error to be returned when no rich details are present")
	}
}

func TestConvertInt64ToStreamState(t *testing.T) {
	tests := []struct {
		input    int64
		expected StreamState
	}{
		{-1, NoStream{}},
		{-2, Any{}},
		{-4, StreamExists{}},
		{0, StreamRevision{Value: 0}},
		{42, StreamRevision{Value: 42}},
	}

	for _, test := range tests {
		result := convertInt64ToStreamState(test.input)
		if result != test.expected {
			t.Errorf("For input %d, expected %v, got %v", test.input, test.expected, result)
		}
	}
}
