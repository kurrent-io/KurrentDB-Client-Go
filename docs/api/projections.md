---
order: 6
title: Projections
---

# Projection management

The client provides a way to manage projections in EventStoreDB. 

For a detailed explanation of projections, see the [server documentation](@server/features/projections/README.md).

## Create a client

```go
conf, err := esdb.ParseConnectionString(connectionString)

if err != nil {
  panic(err)
}

client, err := esdb.NewProjectionClient(conf)
```

## Create a projection

Creates a projection that runs until the last event in the store, and then continues processing new events as they are appended to the store. The query parameter contains the JavaScript you want created as a projection.
Projections have explicit names, and you can enable or disable them via this name.

```go
script := `
fromAll()
.when({
	$init:function(){
		return {
			count: 0
		}
	},
	myEventUpdatedType: function(state, event){
		state.count += 1;
	}
})
.transformBy(function(state){
	state.count = 10;
})
.outputState()
`

name := fmt.Sprintf("countEvent_Create_%s", uuid.New())
err := client.Create(context.Background(), name, script, esdb.CreateProjectionOptions{})

if err != nil {
    panic(err)
}
```

Trying to create projections with the same name will result in an error:

```go
err := client.Create(context.Background(), name, script, esdb.CreateProjectionOptions{})

if esdbErr, ok := esdb.FromError(err); !ok {
    if esdbErr.IsErrorCode(esdb.ErrorCodeUnknown) && strings.Contains(esdbErr.Err().Error(), "Conflict") {
        log.Printf("projection %s already exists", name)
        return
    }
}
```

## Restart the subsystem

It is possible to restart the entire projection subsystem using the projections management client API. The user must be in the `$ops` or `$admin` group to perform this operation.

```go
err := client.RestartSubsystem(context.Background(), esdb.GenericProjectionOptions{})

if err != nil {
    panic(err)
}
```

## Enable a projection

This Enables an existing projection by name. Once enabled, the projection will
start to process events even after restarting the server or the projection
subsystem. You must have access to a projection to enable it, see the [ACL documentation](@server/security/user-authorization.md).

```go
err := client.Enable(context.Background(), "$by_category", esdb.GenericProjectionOptions{})

if err != nil {
    panic(err)
}
```

You can only enable an existing projection. When you try to enable a non-existing projection, you'll get an error:

 ```go
err := client.Enable(context.Background(), "projection that doesn't exist", esdb.GenericProjectionOptions{})

if esdbError, ok := esdb.FromError(err); !ok {
    if esdbError.IsErrorCode(esdb.ErrorCodeResourceNotFound) {
        log.Printf("projection not found")
        return
    }
}
```

## Disable a projection

Disables a projection, this will save the projection checkpoint.
Once disabled, the projection will not process events even after restarting the server or the projection subsystem.
You must have access to a projection to disable it, see the [ACL documentation](@server/security/user-authorization.md).

```go
err := client.Disable(context.Background(), "$by_category", esdb.GenericProjectionOptions{})

if err != nil {
    panic(err)
}
```

You can only disable an existing projection. When you try to disable a non-existing projection, you'll get an error:

```go
err := client.Disable(context.Background(), "projection that doesn't exist", esdb.GenericProjectionOptions{})

if esdbError, ok := esdb.FromError(err); !ok {
    if esdbError.IsErrorCode(esdb.ErrorCodeResourceNotFound) {
        log.Printf("projection not found")
        return
    }
}
```

## Delete a projection

```go
err := client.Delete(context.Background(), "$by_category", esdb.DeleteProjectionOptions{})

if err != nil {
    panic(err)
}
```

## Abort a projection

Aborts a projection, this will not save the projection's checkpoint.

```go
err := client.Abort(context.Background(), "$by_category", esdb.GenericProjectionOptions{})

if err != nil {
    panic(err)
}
```

You can only abort an existing projection. When you try to abort a non-existing projection, you'll get an error:

```go
err := client.Abort(context.Background(), "projection that doesn't exist", esdb.GenericProjectionOptions{})

if esdbError, ok := esdb.FromError(err); !ok {
    if esdbError.IsErrorCode(esdb.ErrorCodeResourceNotFound) {
        log.Printf("projection not found")
        return
    }
}
```

## Reset a projection

Resets a projection, which causes deleting the projection checkpoint. This will force the projection to start afresh and re-emit events. Streams that are written to from the projection will also be soft-deleted.

```go
err := client.Reset(context.Background(), "$by_category", esdb.ResetProjectionOptions{})

if err != nil {
    panic(err)
}
```

Resetting a projection that does not exist will result in an error.

```go
err := client.Reset(context.Background(), "projection that doesn't exist", esdb.ResetProjectionOptions{})

if esdbError, ok := esdb.FromError(err); !ok {
    if esdbError.IsErrorCode(esdb.ErrorCodeResourceNotFound) {
        log.Printf("projection not found")
        return
    }
}
```

## Update a projection

Updates a projection with a given name. The query parameter contains the new JavaScript. Updating system projections using this operation is not supported at the moment.

```go
err := client.Create(context.Background(), name, script, esdb.CreateProjectionOptions{})

if err != nil {
    panic(err)
}

err = client.Update(context.Background(), name, newScript, esdb.UpdateProjectionOptions{})

if err != nil {
    panic(err)
}
```

You can only update an existing projection. When you try to update a non-existing projection, you'll get an error:

```go
err := client.Update(context.Background(), "projection that doesn't exist", script, esdb.UpdateProjectionOptions{})

if esdbError, ok := esdb.FromError(err); !ok {
    if esdbError.IsErrorCode(esdb.ErrorCodeResourceNotFound) {
        log.Printf("projection not found")
        return
    }
}
```

## List all projections

Returns a list of all projections, user defined & system projections.
See the [projection details](#projection-details) section for an explanation of the returned values.

```go
projections, err := client.ListAll(context.Background(), esdb.GenericProjectionOptions{})

if err != nil {
    panic(err)
}

for i := range projections {
    projection := projections[i]

    log.Printf(
        "%s, %s, %s, %s, %f",
        projection.Name,
        projection.Status,
        projection.CheckpointStatus,
        projection.Mode,
        projection.Progress,
    )
}
```

## List continuous projections

Returns a list of all continuous projections.
See the [projection details](#projection-details) section for an explanation of the returned values.

```go
projections, err := client.ListContinuous(context.Background(), esdb.GenericProjectionOptions{})

if err != nil {
    panic(err)
}

for i := range projections {
    projection := projections[i]

    log.Printf(
        "%s, %s, %s, %s, %f",
        projection.Name,
        projection.Status,
        projection.CheckpointStatus,
        projection.Mode,
        projection.Progress,
    )
}
```

## Get status

Gets the status of a named projection.
See the [projection details](#projection-details) section for an explanation of the returned values.

```go
projection, err := client.GetStatus(context.Background(), "$by_category", esdb.GenericProjectionOptions{})

if err != nil {
    panic(err)
}

log.Printf(
    "%s, %s, %s, %s, %f",
    projection.Name,
    projection.Status,
    projection.CheckpointStatus,
    projection.Mode,
    projection.Progress,
)
```

## Get state

Retrieves the state of a projection.

```go
type Foobar struct {
    Count int64
}

value, err := client.GetState(context.Background(), projectionName, esdb.GetStateProjectionOptions{})

if err != nil {
    panic(err)
}

jsonContent, err := value.MarshalJSON()

if err != nil {
    panic(err)
}

var foobar Foobar

if err = json.Unmarshal(jsonContent, &foobar); err != nil {
    panic(err)
}

log.Printf("count %d", foobar.Count)
```

## Get result

Retrieves the result of the named projection and partition.

```go
type Baz struct {
    Result int64
}

value, err := client.GetResult(context.Background(), projectionName, esdb.GetResultProjectionOptions{})

if err != nil {
    panic(err)
}

jsonContent, err := value.MarshalJSON()

if err != nil {
    panic(err)
}

var baz Baz

if err = json.Unmarshal(jsonContent, &baz); err != nil {
    panic(err)
}

log.Printf("result %d", baz.Result)
```

## Projection Details

The `ListAll`, `ListContinuous`, and `GetStatus` methods return detailed statistics and information
about projections. Below is an explanation of the fields included in the
projection details:

| Field                                | Description                                                                                                                                                                                           |
| ------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Name`, `EffectiveName`              | The name of the projection                                                                                                                                                                            |
| `Status`                             | A human readable string of the current statuses of the projection (see below)                                                                                                                         |
| `StateReason`                        | A human readable string explaining the reason of the current projection state                                                                                                                         |
| `CheckpointStatus`                   | A human readable string explaining the current operation performed on the checkpoint : `requested`, `writing`                                                                                         |
| `Mode`                               | `Continuous`, `OneTime` , `Transient`                                                                                                                                                                 |
| `CoreProcessingTime`                 | The total time, in ms, the projection took to handle events since the last restart                                                                                                                    |
| `Progress`                           | The progress, in %, indicates how far this projection has processed event, in case of a restart this could be -1% or some number. It will be updated as soon as a new event is appended and processed |
| `WritesInProgress`                   | The number of write requests to emitted streams currently in progress, these writes can be batches of events                                                                                          |
| `ReadsInProgress`                    | The number of read requests currently in progress                                                                                                                                                     |
| `PartitionsCached`                   | The number of cached projection partitions                                                                                                                                                            |
| `Position`                           | The Position of the last processed event                                                                                                                                                              |
| `LastCheckpoint`                     | The Position of the last checkpoint of this projection                                                                                                                                                |
| `EventsProcessedAfterRestart`        | The number of events processed since the last restart of this projection                                                                                                                              |
| `BufferedEvents`                     | The number of events in the projection read buffer                                                                                                                                                    |
| `WritePendingEventsBeforeCheckpoint` | The number of events waiting to be appended to emitted streams before the pending checkpoint can be written                                                                                           |
| `WritePendingEventsAfterCheckpoint`  | The number of events to be appended to emitted streams since the last checkpoint                                                                                                                      |
| `Version`                            | This is used internally, the version is increased when the projection is edited or reset                                                                                                              |
| `Epoch`                              | This is used internally, the epoch is increased when the projection is reset                                                                                                                          |

The `status` string is a combination of the following values.

The first 3 are the most common one, as the other one are transient values while
the projection is initialised or stopped

| Value              | Description                                                                                                             |
|--------------------|-------------------------------------------------------------------------------------------------------------------------|
| Running            | The projection is running and processing events                                                                         |
| Stopped            | The projection is stopped and is no longer processing new events                                                        |
| Faulted            | An error occurred in the projection, `StateReason` will give the fault details, the projection is not processing events |
| Initial            | This is the initial state, before the projection is fully initialised                                                   |
| Suspended          | The projection is suspended and will not process events, this happens while stopping the projection                     |
| LoadStateRequested | The state of the projection is being retrieved, this happens while the projection is starting                           |
| StateLoaded        | The state of the projection is loaded, this happens while the projection is starting                                    |
| Subscribed         | The projection has successfully subscribed to its readers, this happens while the projection is starting                |
| FaultedStopping    | This happens before the projection is stopped due to an error in the projection                                         |
| Stopping           | The projection is being stopped                                                                                         |
| CompletingPhase    | This happens while the projection is stopping                                                                           |
| PhaseCompleted     | This happens while the projection is stopping                                                                           |
