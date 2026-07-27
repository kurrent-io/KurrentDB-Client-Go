package kurrentdb

import (
	"context"
	"errors"
	"fmt"
	"github.com/kurrent-io/KurrentDB-Client-Go/protos/kurrentdb/protocols/v1/projections"
	"github.com/kurrent-io/KurrentDB-Client-Go/protos/kurrentdb/protocols/v1/shared"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

func projectionMetadata(md map[string]interface{}) (map[string]*structpb.Value, error) {
	if len(md) == 0 {
		return nil, nil
	}

	out := make(map[string]*structpb.Value, len(md))
	for key, value := range md {
		v, err := structpb.NewValue(value)
		if err != nil {
			return nil, fmt.Errorf("invalid projection metadata value for %q: %w", key, err)
		}
		out[key] = v
	}

	return out, nil
}

type ProjectionClient struct {
	inner *Client
}

func NewProjectionClient(configuration *Configuration) (*ProjectionClient, error) {
	client, err := NewClient(configuration)

	if err != nil {
		return nil, err
	}

	return &ProjectionClient{
		inner: client,
	}, nil
}

func NewProjectionClientFromExistingClient(client *Client) *ProjectionClient {
	return &ProjectionClient{
		inner: client,
	}
}

func (client *ProjectionClient) Client() *Client {
	return client.inner
}

func (client *ProjectionClient) Close() error {
	return client.inner.Close()
}

func (client *ProjectionClient) Create(
	context context.Context,
	name string,
	query string,
	opts CreateProjectionOptions,
) error {
	opts.setDefaults()

	if opts.EngineVersion == ProjectionEngineVersionV2 && opts.TrackEmittedStreams {
		return errors.New("trackEmittedStreams is not supported when engineVersion is V2")
	}

	metadataProps, err := projectionMetadata(opts.Metadata)
	if err != nil {
		return err
	}

	handle, err := client.inner.grpcClient.getConnectionHandle()
	if err != nil {
		return err
	}

	projClient := projections.NewProjectionsClient(handle.Connection())
	var headers, trailers metadata.MD
	callOptions := []grpc.CallOption{grpc.Header(&headers), grpc.Trailer(&trailers)}
	callOptions, ctx, cancel := configureGrpcCall(context, client.inner.config, &opts, callOptions, client.inner.grpcClient.perRPCCredentials)
	defer cancel()

	_, err = projClient.Create(ctx, &projections.CreateReq{
		Options: &projections.CreateReq_Options{
			Query:         query,
			EngineVersion: int32(opts.EngineVersion),
			Properties:    metadataProps,
			Mode: &projections.CreateReq_Options_Continuous_{
				Continuous: &projections.CreateReq_Options_Continuous{
					Name:                name,
					EmitEnabled:         opts.Emit,
					TrackEmittedStreams: opts.TrackEmittedStreams,
				},
			},
		},
	}, callOptions...)

	return err
}

func (client *ProjectionClient) Update(
	context context.Context,
	name string,
	query string,
	opts UpdateProjectionOptions,
) error {
	opts.setDefaults()

	metadataProps, err := projectionMetadata(opts.Metadata)
	if err != nil {
		return err
	}

	handle, err := client.inner.grpcClient.getConnectionHandle()
	if err != nil {
		return err
	}

	projClient := projections.NewProjectionsClient(handle.Connection())
	var headers, trailers metadata.MD
	callOptions := []grpc.CallOption{grpc.Header(&headers), grpc.Trailer(&trailers)}
	callOptions, ctx, cancel := configureGrpcCall(context, client.inner.config, &opts, callOptions, client.inner.grpcClient.perRPCCredentials)
	defer cancel()

	options := &projections.UpdateReq_Options{
		Name:       name,
		Query:      query,
		Properties: metadataProps,
	}

	if opts.Emit == nil {
		options.EmitOption = &projections.UpdateReq_Options_NoEmitOptions{}
	} else {
		options.EmitOption = &projections.UpdateReq_Options_EmitEnabled{
			EmitEnabled: *opts.Emit,
		}
	}

	_, err = projClient.Update(ctx, &projections.UpdateReq{
		Options: options,
	}, callOptions...)

	return err
}

func (client *ProjectionClient) Delete(
	context context.Context,
	name string,
	opts DeleteProjectionOptions,
) error {
	opts.setDefaults()
	handle, err := client.inner.grpcClient.getConnectionHandle()
	if err != nil {
		return err
	}

	projClient := projections.NewProjectionsClient(handle.Connection())
	var headers, trailers metadata.MD
	callOptions := []grpc.CallOption{grpc.Header(&headers), grpc.Trailer(&trailers)}
	callOptions, ctx, cancel := configureGrpcCall(context, client.inner.config, &opts, callOptions, client.inner.grpcClient.perRPCCredentials)
	defer cancel()

	_, err = projClient.Delete(ctx, &projections.DeleteReq{
		Options: &projections.DeleteReq_Options{
			Name:                   name,
			DeleteEmittedStreams:   opts.DeleteEmittedStreams,
			DeleteStateStream:      opts.DeleteStateStream,
			DeleteCheckpointStream: opts.DeleteCheckpointStream,
		},
	}, callOptions...)

	return err
}

func (client *ProjectionClient) Enable(
	context context.Context,
	name string,
	opts GenericProjectionOptions,
) error {
	opts.setDefaults()
	handle, err := client.inner.grpcClient.getConnectionHandle()
	if err != nil {
		return err
	}

	projClient := projections.NewProjectionsClient(handle.Connection())
	var headers, trailers metadata.MD
	callOptions := []grpc.CallOption{grpc.Header(&headers), grpc.Trailer(&trailers)}
	callOptions, ctx, cancel := configureGrpcCall(context, client.inner.config, &opts, callOptions, client.inner.grpcClient.perRPCCredentials)
	defer cancel()

	_, err = projClient.Enable(ctx, &projections.EnableReq{
		Options: &projections.EnableReq_Options{
			Name: name,
		},
	}, callOptions...)

	return err
}

func (client *ProjectionClient) Disable(
	context context.Context,
	name string,
	opts GenericProjectionOptions,
) error {
	return client.disable(context, name, true, opts)
}

func (client *ProjectionClient) Abort(
	context context.Context,
	name string,
	opts GenericProjectionOptions,
) error {
	return client.disable(context, name, false, opts)
}

func (client *ProjectionClient) disable(
	context context.Context,
	name string,
	writeCheckpoint bool,
	opts GenericProjectionOptions,
) error {
	opts.setDefaults()
	handle, err := client.inner.grpcClient.getConnectionHandle()
	if err != nil {
		return err
	}

	projClient := projections.NewProjectionsClient(handle.Connection())
	var headers, trailers metadata.MD
	callOptions := []grpc.CallOption{grpc.Header(&headers), grpc.Trailer(&trailers)}
	callOptions, ctx, cancel := configureGrpcCall(context, client.inner.config, &opts, callOptions, client.inner.grpcClient.perRPCCredentials)
	defer cancel()

	_, err = projClient.Disable(ctx, &projections.DisableReq{
		Options: &projections.DisableReq_Options{
			Name:            name,
			WriteCheckpoint: writeCheckpoint,
		},
	}, callOptions...)

	return err
}

func (client *ProjectionClient) Reset(
	context context.Context,
	name string,
	opts ResetProjectionOptions,
) error {
	opts.setDefaults()
	handle, err := client.inner.grpcClient.getConnectionHandle()
	if err != nil {
		return err
	}

	projClient := projections.NewProjectionsClient(handle.Connection())
	var headers, trailers metadata.MD
	callOptions := []grpc.CallOption{grpc.Header(&headers), grpc.Trailer(&trailers)}
	callOptions, ctx, cancel := configureGrpcCall(context, client.inner.config, &opts, callOptions, client.inner.grpcClient.perRPCCredentials)
	defer cancel()

	_, err = projClient.Reset(ctx, &projections.ResetReq{
		Options: &projections.ResetReq_Options{
			Name:            name,
			WriteCheckpoint: opts.WriteCheckpoint,
		},
	}, callOptions...)

	return err
}

func (client *ProjectionClient) GetResult(
	context context.Context,
	name string,
	opts GetResultProjectionOptions,
) (*structpb.Value, error) {
	opts.setDefaults()
	handle, err := client.inner.grpcClient.getConnectionHandle()
	if err != nil {
		return nil, err
	}

	projClient := projections.NewProjectionsClient(handle.Connection())
	var headers, trailers metadata.MD
	callOptions := []grpc.CallOption{grpc.Header(&headers), grpc.Trailer(&trailers)}
	callOptions, ctx, cancel := configureGrpcCall(context, client.inner.config, &opts, callOptions, client.inner.grpcClient.perRPCCredentials)
	defer cancel()

	resp, err := projClient.Result(ctx, &projections.ResultReq{
		Options: &projections.ResultReq_Options{
			Name:      name,
			Partition: opts.Partition,
		},
	}, callOptions...)

	if err != nil {
		return nil, err
	}

	return resp.Result, nil
}

func (client *ProjectionClient) GetState(
	context context.Context,
	name string,
	opts GetStateProjectionOptions,
) (*structpb.Value, error) {
	opts.setDefaults()
	handle, err := client.inner.grpcClient.getConnectionHandle()
	if err != nil {
		return nil, err
	}

	projClient := projections.NewProjectionsClient(handle.Connection())
	var headers, trailers metadata.MD
	callOptions := []grpc.CallOption{grpc.Header(&headers), grpc.Trailer(&trailers)}
	callOptions, ctx, cancel := configureGrpcCall(context, client.inner.config, &opts, callOptions, client.inner.grpcClient.perRPCCredentials)
	defer cancel()

	resp, err := projClient.State(ctx, &projections.StateReq{
		Options: &projections.StateReq_Options{
			Name:      name,
			Partition: opts.Partition,
		},
	}, callOptions...)

	if err != nil {
		return nil, err
	}

	return resp.State, nil
}

func (client *ProjectionClient) RestartSubsystem(
	context context.Context,
	opts GenericProjectionOptions,
) error {
	opts.setDefaults()
	handle, err := client.inner.grpcClient.getConnectionHandle()
	if err != nil {
		return err
	}

	projClient := projections.NewProjectionsClient(handle.Connection())
	var headers, trailers metadata.MD
	callOptions := []grpc.CallOption{grpc.Header(&headers), grpc.Trailer(&trailers)}
	callOptions, ctx, cancel := configureGrpcCall(context, client.inner.config, &opts, callOptions, client.inner.grpcClient.perRPCCredentials)
	defer cancel()

	_, err = projClient.RestartSubsystem(ctx, &shared.Empty{}, callOptions...)

	return err
}

type ProjectionStatus struct {
	CoreProcessingTime                 int64
	Version                            int64
	Epoch                              int64
	EffectiveName                      string
	WritesInProgress                   int32
	ReadsInProgress                    int32
	PartitionsCached                   int32
	Status                             string
	StateReason                        string
	Name                               string
	Mode                               string
	Position                           string
	Progress                           float32
	LastCheckpoint                     string
	EventsProcessedAfterRestart        int64
	CheckpointStatus                   string
	BufferedEvents                     int64
	WritePendingEventsBeforeCheckpoint int32
	WritePendingEventsAfterCheckpoint  int32
}

type projectionKind interface {
	setProjectionMode(opts *projections.StatisticsReq_Options)
}

type named struct {
	name string
}

func (name named) setProjectionMode(opts *projections.StatisticsReq_Options) {
	opts.Mode = &projections.StatisticsReq_Options_Name{
		Name: name.name,
	}
}

type projectionSelect int

const (
	projectionSelectContinuous projectionSelect = iota
	projectionSelectAll
)

func (projSelect projectionSelect) setProjectionMode(opts *projections.StatisticsReq_Options) {
	switch projSelect {
	case projectionSelectContinuous:
		opts.Mode = &projections.StatisticsReq_Options_Continuous{}
	case projectionSelectAll:
		opts.Mode = &projections.StatisticsReq_Options_All{}
	}
}

func (client *ProjectionClient) GetStatus(
	context context.Context,
	name string,
	opts GenericProjectionOptions,
) (*ProjectionStatus, error) {
	projs, err := client.listInternal(context, named{name: name}, opts)

	if err != nil {
		return nil, err
	}

	if len(projs) == 0 {
		return nil, &Error{code: ErrorCodeResourceNotFound, err: fmt.Errorf("projection '%s' is not found", name)}
	}

	return &projs[0], nil
}

func (client *ProjectionClient) ListContinuous(
	context context.Context,
	opts GenericProjectionOptions,
) ([]ProjectionStatus, error) {
	return client.listInternal(context, projectionSelectContinuous, opts)
}

func (client *ProjectionClient) ListAll(
	context context.Context,
	opts GenericProjectionOptions,
) ([]ProjectionStatus, error) {
	return client.listInternal(context, projectionSelectAll, opts)
}

func (client *ProjectionClient) listInternal(
	context context.Context,
	kind projectionKind,
	opts GenericProjectionOptions,
) ([]ProjectionStatus, error) {
	opts.setDefaults()
	handle, err := client.inner.grpcClient.getConnectionHandle()
	if err != nil {
		return nil, err
	}

	projClient := projections.NewProjectionsClient(handle.Connection())
	var headers, trailers metadata.MD
	callOptions := []grpc.CallOption{grpc.Header(&headers), grpc.Trailer(&trailers)}
	callOptions, ctx, cancel := configureGrpcCall(context, client.inner.config, &opts, callOptions, client.inner.grpcClient.perRPCCredentials)
	defer cancel()

	options := projections.StatisticsReq_Options{}
	kind.setProjectionMode(&options)
	stream, err := projClient.Statistics(ctx, &projections.StatisticsReq{
		Options: &options,
	}, callOptions...)
	if err != nil {
		return nil, client.inner.grpcClient.handleError(handle, trailers, err)
	}

	var projs []ProjectionStatus

	for {
		item, err := stream.Recv()

		if err != nil {
			if !errors.Is(err, io.EOF) {
				err = client.inner.grpcClient.handleError(handle, trailers, err)
				return nil, err
			}

			return projs, nil
		}

		details := item.GetDetails()
		if details == nil {
			continue
		}

		proj := ProjectionStatus{
			CoreProcessingTime:                 details.CoreProcessingTime,
			Version:                            details.Version,
			Epoch:                              details.Epoch,
			EffectiveName:                      details.EffectiveName,
			WritesInProgress:                   details.WritesInProgress,
			ReadsInProgress:                    details.ReadsInProgress,
			PartitionsCached:                   details.PartitionsCached,
			Status:                             details.Status,
			StateReason:                        details.StateReason,
			Name:                               details.Name,
			Mode:                               details.Mode,
			Position:                           details.Position,
			Progress:                           details.Progress,
			LastCheckpoint:                     details.LastCheckpoint,
			EventsProcessedAfterRestart:        details.EventsProcessedAfterRestart,
			CheckpointStatus:                   details.CheckpointStatus,
			BufferedEvents:                     details.BufferedEvents,
			WritePendingEventsBeforeCheckpoint: details.WritePendingEventsBeforeCheckpoint,
			WritePendingEventsAfterCheckpoint:  details.WritePendingEventsAfterCheckpoint,
		}

		projs = append(projs, proj)
	}
}
