package kurrentdb

import "time"

// AppendRecord represents a record to be appended to a specific stream in an AppendRecords operation.
// Each record specifies its own target stream, allowing interleaved writes across multiple streams.
type AppendRecord struct {
	// Stream is the name of the target stream for this record.
	Stream string
	// Record is the event data to append.
	Record EventData
}

// ConsistencyCheck represents a pre-commit condition that must hold true for the transaction to succeed.
// Checks are decoupled from writes: a check can reference any stream, whether or not the request writes to it.
type ConsistencyCheck interface {
	isConsistencyCheck()
}

// StreamStateCheck asserts a stream is at a specific revision or lifecycle state before commit.
type StreamStateCheck struct {
	// Stream is the name of the stream to check.
	Stream string
	// ExpectedState is the expected state of the stream (revision number or state constant).
	ExpectedState StreamState
}

func (StreamStateCheck) isConsistencyCheck() {}

// appendRecordsOptions is an internal options type used to satisfy the options interface for gRPC call configuration.
type appendRecordsOptions struct{}

func (o *appendRecordsOptions) kind() operationKind {
	return regularOperation
}

func (o *appendRecordsOptions) credentials() *Credentials {
	return nil
}

func (o *appendRecordsOptions) deadline() *time.Duration {
	return nil
}

func (o *appendRecordsOptions) requiresLeader() bool {
	return false
}
