package kurrentdb

// Any means the write should not conflict with anything and should always succeed.
type Any struct{}

// StreamExists means the stream should exist.
type StreamExists struct{}

// NoStream means the stream being written to should not yet exist.
type NoStream struct{}

// StreamState the use of expected revision can be a bit tricky especially when discussing guaranties given by
// KurrentDB server. The KurrentDB server will assure idempotency for all requests using any value in
// StreamState except Any. When using Any, the KurrentDB server will do its best to assure idempotency but
// will not guarantee it.
type StreamState interface {
	toRawInt64() int64
}

func (r Any) toRawInt64() int64 {
	return -2
}

func (r StreamExists) toRawInt64() int64 {
	return -4
}

func (r NoStream) toRawInt64() int64 {
	return -1
}

func (r StreamRevision) toRawInt64() int64 {
	return int64(r.Value)
}
