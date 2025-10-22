package kurrentdb

import (
	"fmt"

	streamErrors "github.com/kurrent-io/KurrentDB-Client-Go/protos/kurrentdb/protocols/v2/streams/errors"
	"google.golang.org/grpc/status"
)

// ErrorCode KurrentDB error code.
type ErrorCode int

const (
	// ErrorCodeUnknown unclassified error.
	ErrorCodeUnknown ErrorCode = iota
	// ErrorCodeUnsupportedFeature a request not supported by the targeted KurrentDB node was sent.
	ErrorCodeUnsupportedFeature
	// ErrorCodeDeadlineExceeded a gRPC deadline exceeded error.
	ErrorCodeDeadlineExceeded
	// ErrorCodeUnauthenticated a request requires authentication and the authentication failed.
	ErrorCodeUnauthenticated
	// ErrorCodeResourceNotFound a remote resource was not found or because its access was denied.
	ErrorCodeResourceNotFound
	// ErrorCodeResourceAlreadyExists a creation request was made for a resource that already exists.
	ErrorCodeResourceAlreadyExists
	// ErrorCodeConnectionClosed when a connection is already closed.
	ErrorCodeConnectionClosed
	// ErrorCodeWrongExpectedVersion when an append request failed the optimistic concurrency on the server.
	ErrorCodeWrongExpectedVersion
	// ErrorCodeStreamRevisionConflict when an append request failed due to a stream revision conflict.
	ErrorCodeStreamRevisionConflict
	// ErrorCodeStreamTombstoned requested stream is tombstoned.
	ErrorCodeStreamTombstoned
	// ErrorCodeAppendRecordSizeExceeded when an append record exceeds the maximum allowed size.
	ErrorCodeAppendRecordSizeExceeded
	// ErrorCodeAppendTransactionSizeExceeded when an append transaction exceeds the maximum allowed size.
	ErrorCodeAppendTransactionSizeExceeded
	// ErrorCodeAccessDenied a request requires the right ACL.
	ErrorCodeAccessDenied
	// ErrorCodeStreamDeleted requested stream is deleted.
	ErrorCodeStreamDeleted
	// ErrorCodeParsing error when parsing data.
	ErrorCodeParsing
	// ErrorCodeInternalClient unexpected error from the client library, worthy of a GitHub issue.
	ErrorCodeInternalClient
	// ErrorCodeInternalServer unexpected error from the server, worthy of a GitHub issue.
	ErrorCodeInternalServer
	// ErrorCodeNotLeader when a request needing a leader node was executed on a follower node.
	ErrorCodeNotLeader
	// ErrorAborted when the server aborted the request.
	ErrorAborted
	// ErrorUnavailable when the KurrentDB node became unavailable.
	ErrorUnavailable
)

// Error main client error type.
type Error struct {
	code ErrorCode
	err  error
}

// Code returns an error code.
func (e *Error) Code() ErrorCode {
	return e.code
}

// Err returns underlying error.
func (e *Error) Err() error {
	return e.err
}

// IsErrorCode checks if the error code is the same as the given one.
func (e *Error) IsErrorCode(code ErrorCode) bool {
	return e.code == code
}

func (e *Error) Error() string {
	msg := ""

	switch e.code {
	case ErrorCodeUnsupportedFeature:
		msg = "[ErrorCodeUnsupportedFeature] request not supported by the targeted KurrentDB node"
	case ErrorCodeDeadlineExceeded:
		msg = "[ErrorCodeDeadlineExceeded] gRPC deadline exceeded error"
	case ErrorCodeUnauthenticated:
		msg = "[ErrorCodeUnauthenticated] request requires authentication and the authentication failed"
	case ErrorCodeResourceNotFound:
		msg = "[ErrorCodeResourceNotFound] a remote resource was not found or its access was denied"
	case ErrorCodeResourceAlreadyExists:
		msg = "[ErrorCodeResourceAlreadyExists] a creation request was made for a resource that already exists"
	case ErrorCodeConnectionClosed:
		msg = "[ErrorCodeConnectionClosed] the connection is already closed"
	case ErrorCodeWrongExpectedVersion:
		msg = "[ErrorCodeWrongExpectedVersion] an append request failed the optimistic concurrency on the server"
	case ErrorCodeAccessDenied:
		msg = "[ErrorCodeAccessDenied] the request requires the right ACL"
	case ErrorCodeStreamDeleted:
		msg = "[ErrorCodeStreamDeleted] requested stream is deleted"
	case ErrorCodeParsing:
		msg = "[ErrorCodeParsing] error when parsing data"
	case ErrorCodeInternalClient:
		msg = "[ErrorCodeInternalClient] unexpected error from the client library, worthy of a GitHub issue"
	case ErrorCodeInternalServer:
		msg = "[ErrorCodeInternalServer] unexpected error from the server, worthy of a GitHub issue"
	case ErrorCodeNotLeader:
		msg = "[ErrorCodeNotLeader] the request needing a leader node was executed on a follower node"
	case ErrorUnavailable:
		msg = "[ErrorUnavailable] the server is not ready to accept requests"

	default:
		msg = fmt.Sprintf("[ErrorCode %d] (sorry, this error code is not supported by the Error() method)", e.code)
	}

	if e.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Err())
	}

	return msg
}

func (e *Error) Unwrap() error {
	return e.Err()
}

func FromError(err error) (*Error, bool) {
	if err == nil {
		return nil, true
	}

	if esErr, ok := err.(*Error); ok {
		return esErr, false
	}

	return &Error{code: ErrorCodeUnknown, err: err}, false
}

func unsupportedFeatureError() error {
	return &Error{code: ErrorCodeUnsupportedFeature}
}

func unknownError() error {
	return &Error{code: ErrorCodeUnknown}
}

// getDetail attempts to extract rich error details from a gRPC status error
func getDetail(err error) *Error {
	s, ok := status.FromError(err)
	if !ok || s == nil {
		return nil
	}

	for _, d := range s.Details() {
		switch detail := d.(type) {
		case *streamErrors.StreamRevisionConflictErrorDetails:
			expectedState := convertInt64ToStreamState(detail.ExpectedRevision)
			actualState := convertInt64ToStreamState(detail.ActualRevision)
			return &Error{
				code: ErrorCodeStreamRevisionConflict,
				err: &StreamRevisionConflictError{
					Stream:           detail.Stream,
					ExpectedRevision: expectedState,
					ActualRevision:   actualState,
				},
			}
		case *streamErrors.StreamTombstonedErrorDetails:
			return &Error{
				code: ErrorCodeStreamTombstoned,
				err:  &StreamTombstoneError{Stream: detail.Stream},
			}
		case *streamErrors.AppendRecordSizeExceededErrorDetails:
			return &Error{
				code: ErrorCodeAppendRecordSizeExceeded,
				err: &AppendRecordSizeExceededError{
					Stream:   detail.Stream,
					RecordId: detail.RecordId,
					Size:     detail.Size,
					MaxSize:  detail.MaxSize,
				},
			}
		case *streamErrors.AppendTransactionSizeExceededErrorDetails:
			return &Error{
				code: ErrorCodeAppendTransactionSizeExceeded,
				err: &AppendTransactionSizeExceededError{
					Size:    detail.Size,
					MaxSize: detail.MaxSize,
				},
			}
		}
	}

	return nil
}

// convertInt64ToStreamState converts an int64 revision value to the appropriate StreamState implementation
func convertInt64ToStreamState(revision int64) StreamState {
	switch revision {
	case -1:
		return NoStream{}
	case -2:
		return Any{}
	case -4:
		return StreamExists{}
	default:
		if revision >= 0 {
			return StreamRevision{Value: uint64(revision)}
		}
		// For any other negative values, treat as StreamRevision with the raw value
		return StreamRevision{Value: uint64(revision)}
	}
}

// StreamRevisionConflictError represents a conflict in stream revision during append.
type StreamRevisionConflictError struct {
	Stream           string
	ExpectedRevision StreamState
	ActualRevision   StreamState
}

func (e *StreamRevisionConflictError) Error() string {
	return fmt.Sprintf("[ErrorCodeStreamRevisionConflict] stream revision conflict: stream=%s expected_revision=%v actual_revision=%v", e.Stream, e.ExpectedRevision, e.ActualRevision)
}

// StreamDeletedError represents an error when attempting to access a deleted stream.
type StreamDeletedError struct {
	Stream string
}

func (e *StreamDeletedError) Error() string {
	return fmt.Sprintf("[ErrorCodeStreamDeleted] stream deleted: stream=%s", e.Stream)
}

// StreamTombstoneError represents an error when attempting to access a deleted stream.
type StreamTombstoneError struct {
	Stream string
}

func (e *StreamTombstoneError) Error() string {
	return fmt.Sprintf("[ErrorCodeStreamTombstone] stream deleted: stream=%s", e.Stream)
}

// AppendRecordSizeExceededError represents an error when an append record exceeds the maximum allowed size.
type AppendRecordSizeExceededError struct {
	Stream   string
	RecordId string
	Size     int32
	MaxSize  int32
}

func (e *AppendRecordSizeExceededError) Error() string {
	exceededBy := e.Size - e.MaxSize
	return fmt.Sprintf("[ErrorCodeAppendRecordSizeExceeded] The size of record %s (%d bytes) exceeds the maximum allowed size of %d bytes by %d bytes", e.RecordId, e.Size, e.MaxSize, exceededBy)
}

// AppendTransactionSizeExceededError represents an error when an append transaction exceeds the maximum allowed size.
type AppendTransactionSizeExceededError struct {
	Size    int32
	MaxSize int32
}

func (e *AppendTransactionSizeExceededError) Error() string {
	exceededBy := e.Size - e.MaxSize
	return fmt.Sprintf("[ErrorCodeAppendTransactionSizeExceeded] The total size of the append transaction (%d bytes) exceeds the maximum allowed size of %d bytes by %d bytes", e.Size, e.MaxSize, exceededBy)
}
