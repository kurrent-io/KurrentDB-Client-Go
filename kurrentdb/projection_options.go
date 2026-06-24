package kurrentdb

import "time"

// ProjectionEngineVersion selects which projection engine the server should
// use when creating a projection. The engine version is pinned at create time
// and cannot be changed later.
type ProjectionEngineVersion int32

const (
	// ProjectionEngineVersionV1 selects the original projection engine. This is
	// the default when EngineVersion is left unset.
	ProjectionEngineVersionV1 ProjectionEngineVersion = 1
	// ProjectionEngineVersionV2 selects the next-generation projection engine
	// that processes partitions in parallel. V2 is opt-in and does not support
	// TrackEmittedStreams, bi-state projections, or live outputState result
	// streams. See the KurrentDB documentation for the full list of limitations
	// before choosing V2.
	ProjectionEngineVersionV2 ProjectionEngineVersion = 2
)

type CreateProjectionOptions struct {
	// Asks for authenticated request.
	Authenticated *Credentials
	// A length of time to use for gRPC deadlines.
	Deadline *time.Duration
	// Requires the request to be performed by the leader of the cluster.
	RequiresLeader bool
	// If the server should track all the stream the projection ever emit an event into.
	TrackEmittedStreams bool
	// If the projection should be able to write events.
	Emit bool
	// EngineVersion selects the projection engine version. Defaults to V1
	// when unset.
	EngineVersion ProjectionEngineVersion
	// Metadata is caller-supplied metadata stamped onto the projection
	// definition event. It does not affect the projection engine and is
	// intended for deployment tooling to annotate projections (for example
	// with a deployment identifier or the name of the managing tool). Values
	// must be JSON-compatible: string, bool, number, nil, []byte, or nested
	// []interface{} and map[string]interface{} of those. Ignored by servers
	// that don't support it.
	Metadata map[string]interface{}
}

func (o *CreateProjectionOptions) kind() operationKind {
	return regularOperation
}

func (o *CreateProjectionOptions) credentials() *Credentials {
	return o.Authenticated
}

func (o *CreateProjectionOptions) deadline() *time.Duration {
	return o.Deadline
}

func (o *CreateProjectionOptions) requiresLeader() bool {
	return o.RequiresLeader
}

func (o *CreateProjectionOptions) setDefaults() {
}

type UpdateProjectionOptions struct {
	// Asks for authenticated request.
	Authenticated *Credentials
	// A length of time to use for gRPC deadlines.
	Deadline *time.Duration
	// Requires the request to be performed by the leader of the cluster.
	RequiresLeader bool
	// If the projection should be able to write events.
	Emit *bool
	// Metadata is caller-supplied metadata stamped onto the projection
	// definition event. It does not affect the projection engine and is
	// intended for deployment tooling to annotate projections (for example
	// with a deployment identifier or the name of the managing tool). Values
	// must be JSON-compatible: string, bool, number, nil, []byte, or nested
	// []interface{} and map[string]interface{} of those. Ignored by servers
	// that don't support it.
	Metadata map[string]interface{}
}

func (o *UpdateProjectionOptions) kind() operationKind {
	return regularOperation
}

func (o *UpdateProjectionOptions) credentials() *Credentials {
	return o.Authenticated
}

func (o *UpdateProjectionOptions) deadline() *time.Duration {
	return o.Deadline
}

func (o *UpdateProjectionOptions) requiresLeader() bool {
	return o.RequiresLeader
}

func (o *UpdateProjectionOptions) setDefaults() {
}

type DeleteProjectionOptions struct {
	// Asks for authenticated request.
	Authenticated *Credentials
	// A length of time to use for gRPC deadlines.
	Deadline *time.Duration
	// Requires the request to be performed by the leader of the cluster.
	RequiresLeader bool
	// If the server should delete all the streams the projection created.
	DeleteEmittedStreams bool
	// If the server should delete the stream storing the projection's state.
	DeleteStateStream bool
	// if the server should delete the stream where all the projection's checkpoints are stored.
	DeleteCheckpointStream bool
}

func (o *DeleteProjectionOptions) kind() operationKind {
	return regularOperation
}

func (o *DeleteProjectionOptions) credentials() *Credentials {
	return o.Authenticated
}

func (o *DeleteProjectionOptions) deadline() *time.Duration {
	return o.Deadline
}

func (o *DeleteProjectionOptions) requiresLeader() bool {
	return o.RequiresLeader
}

func (o *DeleteProjectionOptions) setDefaults() {
}

type GetStateProjectionOptions struct {
	// Asks for authenticated request.
	Authenticated *Credentials
	// A length of time to use for gRPC deadlines.
	Deadline *time.Duration
	// Requires the request to be performed by the leader of the cluster.
	RequiresLeader bool
	// Partition of the projection we are getting the state from. Leave empty
	// if the projection doesn't have any partition.
	Partition string
}

func (o *GetStateProjectionOptions) kind() operationKind {
	return regularOperation
}

func (o *GetStateProjectionOptions) credentials() *Credentials {
	return o.Authenticated
}

func (o *GetStateProjectionOptions) deadline() *time.Duration {
	return o.Deadline
}

func (o *GetStateProjectionOptions) requiresLeader() bool {
	return o.RequiresLeader
}

func (o *GetStateProjectionOptions) setDefaults() {
}

type GetResultProjectionOptions struct {
	// Asks for authenticated request.
	Authenticated *Credentials
	// A length of time to use for gRPC deadlines.
	Deadline *time.Duration
	// Requires the request to be performed by the leader of the cluster.
	RequiresLeader bool
	// Partition of the projection we are getting the result from. Leave empty
	// if the projection doesn't have any partition.
	Partition string
}

func (o *GetResultProjectionOptions) kind() operationKind {
	return regularOperation
}

func (o *GetResultProjectionOptions) credentials() *Credentials {
	return o.Authenticated
}

func (o *GetResultProjectionOptions) deadline() *time.Duration {
	return o.Deadline
}

func (o *GetResultProjectionOptions) requiresLeader() bool {
	return o.RequiresLeader
}

func (o *GetResultProjectionOptions) setDefaults() {
}

type ResetProjectionOptions struct {
	// Asks for authenticated request.
	Authenticated *Credentials
	// A length of time to use for gRPC deadlines.
	Deadline *time.Duration
	// Requires the request to be performed by the leader of the cluster.
	RequiresLeader bool
	// If the server writes a checkpoint upon resetting.
	WriteCheckpoint bool
}

func (o *ResetProjectionOptions) kind() operationKind {
	return regularOperation
}

func (o *ResetProjectionOptions) credentials() *Credentials {
	return o.Authenticated
}

func (o *ResetProjectionOptions) deadline() *time.Duration {
	return o.Deadline
}

func (o *ResetProjectionOptions) requiresLeader() bool {
	return o.RequiresLeader
}

func (o *ResetProjectionOptions) setDefaults() {
}

type GenericProjectionOptions struct {
	// Asks for authenticated request.
	Authenticated *Credentials
	// A length of time to use for gRPC deadlines.
	Deadline *time.Duration
	// Requires the request to be performed by the leader of the cluster.
	RequiresLeader bool
}

func (o *GenericProjectionOptions) kind() operationKind {
	return regularOperation
}

func (o *GenericProjectionOptions) credentials() *Credentials {
	return o.Authenticated
}

func (o *GenericProjectionOptions) deadline() *time.Duration {
	return o.Deadline
}

func (o *GenericProjectionOptions) requiresLeader() bool {
	return o.RequiresLeader
}

func (o *GenericProjectionOptions) setDefaults() {
}
