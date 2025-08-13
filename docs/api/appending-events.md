---
order: 2
---

# Appending events

When you start working with KurrentDB, your application streams are empty. The first meaningful operation is to add one or more events to the database using this API.

::: tip
Check the [Getting Started](getting-started.md) guide to learn how to configure and use the client SDK.
:::

## Append your first event

The simplest way to append an event to KurrentDB is to create an `EventData` object and call `appendToStream` method.

```go{20-24}
type OrderPlaced struct {
  OrderId string `json:"orderId"`
  Amount  float64 `json:"amount"`
}

data := OrderPlaced{
  OrderId: "ORD-123",
  Amount:  49.99,
}

bytes, err := json.Marshal(data)
if err != nil {
  panic(err)
}

options := kurrentdb.AppendToStreamOptions{
  StreamState: kurrentdb.NoStream{},
}

result, err := db.AppendToStream(context.Background(), "orders-123", options, kurrentdb.EventData{
  ContentType: kurrentdb.ContentTypeJson,
  EventType:   "OrderPlaced",
  Data:        bytes,
})
```

`AppendToStream` takes a collection or a single object that can be serialized in JSON or binary format, which allows you to save more than one event in a single batch.
 
Outside the example above, other options exist for dealing with different scenarios. 

::: tip
If you are new to Event Sourcing, please study the [Handling concurrency](#handling-concurrency) section below.
:::

## Working with EventData

Events appended to KurrentDB must be wrapped in an `EventData` object. This allows you to specify the event's content, the type of event, and whether it's in JSON format. In its simplest form, you need three arguments: **eventId**, **eventType**, and **eventData**.

### EventID

This takes the format of a `UUID` and is used to uniquely identify the event you are trying to append. If two events with the same `UUID` are appended to the same stream in quick succession, KurrentDB will only append one of the events to the stream. 

For example, the following code will only append a single event:

```go{7-12}
_, err = db.AppendToStream(context.Background(), "orders-456", kurrentdb.AppendToStreamOptions{}, event)

if err != nil {
    panic(err)
}

// attempt to append the same event again
_, err = db.AppendToStream(context.Background(), "orders-456", kurrentdb.AppendToStreamOptions{}, event)

if err != nil {
    panic(err)
}
```

### EventType

Each event should be supplied with an event type. This unique string is used to identify the type of event you are saving. 

It is common to see the explicit event code type name used as the type as it makes serialising and de-serialising of the event easy. However, we recommend against this as it couples the storage to the type and will make it more difficult if you need to version the event at a later date.

### Data

Representation of your event data. It is recommended that you store your events as JSON objects. This allows you to take advantage of all of KurrentDB's functionality, such as projections. That said, you can save events using whatever format suits your workflow. Eventually, the data will be stored as encoded bytes.

### Metadata

Storing additional information alongside your event that is part of the event itself is standard practice. This can be correlation IDs, timestamps, access information, etc. KurrentDB allows you to store a separate byte array containing this information to keep it separate.

### ContentType

The content type indicates whether the event is stored as JSON or binary format. You can choose between `kurrentdb.ContentTypeJson` and `kurrentdb.ContentTypeBinary` when creating your `EventData` object. 

## Handling concurrency

When appending events to a stream, you can supply a *stream state*. Your client uses this to inform KurrentDB of the state or version you expect the stream to be in when appending an event. If the stream isn't in that state, an exception will be thrown. 

For example, if you try to append the same record twice, expecting both times that the stream doesn't exist, you will get an exception on the second:

```go{12,15-19,34-39}
data := OrderPlaced{
    OrderId: "ORD-001",
    Amount:  29.99,
}

bytes, err := json.Marshal(data)
if err != nil {
    panic(err)
}

options := kurrentdb.AppendToStreamOptions{
    StreamState: kurrentdb.NoStream{},
}

_, err = db.AppendToStream(context.Background(), "order-123", options, kurrentdb.EventData{
    ContentType: kurrentdb.ContentTypeJson,
    EventType:   "OrderPlaced",
    Data:        bytes,
})

if err != nil {
    panic(err)
}

bytes, err = json.Marshal(OrderPlaced{
    OrderId: "ORD-002",
    Amount:  45.50,
})

if err != nil {
    panic(err)
}

// attempt to append the same event again
_, err = db.AppendToStream(context.Background(), "order-123", options, kurrentdb.EventData{
    ContentType: kurrentdb.ContentTypeJson,
    EventType:   "OrderPlaced",
    Data:        bytes,
})
```

There are several available expected revision options: 
- `kurrentdb.Any` - No concurrency check
- `kurrentdb.NoStream{}` - Stream should not exist
- `kurrentdb.StreamExists{}` - Stream should exist
- `kurrentdb.StreamRevision{}` - Stream should be at specific revision

This check can be used to implement optimistic concurrency. When retrieving a
stream from KurrentDB, note the current version number. When you save it back,
you can determine if somebody else has modified the record in the meantime.

```go{6,32,35-39,50-54}
ropts := kurrentdb.ReadStreamOptions{
    Direction: kurrentdb.Backwards,
    From:      kurrentdb.End{},
}

stream, err := db.ReadStream(context.Background(), "orders-123", ropts, 1)

if err != nil {
    panic(err)
}

defer stream.Close()

lastEvent, err := stream.Recv()

if err != nil {
    panic(err)
}

data := OrderPlaced{
    OrderId: "ORD-123",
    Amount:  29.99,
}

bytes, err := json.Marshal(data)

if err != nil {
    panic(err)
}

aopts := kurrentdb.AppendToStreamOptions{
    StreamState: lastEvent.OriginalStreamRevision(),
}

_, err = db.AppendToStream(context.Background(), "orders-123", aopts, kurrentdb.EventData{
    ContentType: kurrentdb.ContentTypeJson,
    EventType:   "OrderPlaced",
    Data:        bytes,
})

data = OrderPlaced{
    OrderId: "ORD-123",
    Amount:  39.99,
}
bytes, err = json.Marshal(data)
if err != nil {
    panic(err)
}

_, err = db.AppendToStream(context.Background(), "orders-123", aopts, kurrentdb.EventData{
    ContentType: kurrentdb.ContentTypeJson,
    EventType:   "OrderPlaced",
    Data:        bytes,
})
```

## User credentials

You can provide user credentials to append the data as follows. This will override the default credentials set on the connection.

```go{5-8}
result, err := db.AppendToStream(
  context.Background(),
  "orders",
  kurrentdb.AppendToStreamOptions{
    Authenticated: &kurrentdb.Credentials{
      Login: "admin",
      Password: "changeit"
    }
  },
  event
)
```

## Append to multiple streams

::: note
This feature is only available in KurrentDB 25.1 and later.
:::

You can append events to multiple streams in a single atomic operation. Either all streams are updated, or the entire operation fails.

The `MultiStreamAppend` method accepts a collection of `AppendStreamRequest` objects and returns a `MultiAppendWriteResult`. Each `AppendStreamRequest` contains:

- **StreamName** - The name of the stream
- **ExpectedStreamState** - The expected state of the stream for optimistic concurrency control
- **Events** - A collection of `EventData` objects to append

The operation returns either:
- `AppendStreamSuccess` - Successful append results for all streams
- `AppendStreamFailure` - Specific exceptions for any failed operations

::: warning
Event metadata in `EventData` must be valid JSON objects. This requirement will
be removed in a future major release.
:::

Here's a basic example of appending events to multiple streams:

```go
type OrderCreated struct {
	OrderId string  `json:"orderId"`
	Amount  float64 `json:"amount"`
}

type PaymentProcessed struct {
	PaymentId string  `json:"paymentId"`
	Amount    float64 `json:"amount"`
	Method    string  `json:"method"`
}

metadata := map[string]interface{}{
	"source": "web-store",
}
metadataBytes, _ := json.Marshal(metadata)

orderData, _ := json.Marshal(OrderCreated{OrderId: "12345", Amount: 99.99})
paymentData, _ := json.Marshal(PaymentProcessed{PaymentId: "PAY-789", Amount: 99.99, Method: "credit_card"})

requests := []kurrentdb.AppendStreamRequest{
	{
		StreamName: "order-stream-1",
		Events: slices.Values([]kurrentdb.EventData{{
			EventID:     uuid.New(),
			EventType:   "OrderCreated",
			ContentType: kurrentdb.ContentTypeJson,
			Data:        orderData,
			Metadata:    metadataBytes,
		}}),
		ExpectedStreamState: kurrentdb.Any{},
	},
	{
		StreamName: "payment-stream-1",
		Events: slices.Values([]kurrentdb.EventData{{
			EventID:     uuid.New(),
			EventType:   "PaymentProcessed",
			ContentType: kurrentdb.ContentTypeJson,
			Data:        paymentData,
			Metadata:    metadataBytes,
		}}),
		ExpectedStreamState: kurrentdb.Any{},
	},
}

result, err := client.MultiStreamAppend(context.Background(), slices.Values(requests))
if err != nil {
	panic(err)
}

if result.IsSuccessful() {
	for _, success := range result.Successes {
		fmt.Printf("Stream '%s' updated at position %d\n", success.StreamName, success.Position)
	}
}

```

If the operation doesn't succeed, you can handle the failures as follows:

```go
if result.HasFailed() {
  for _, failure := range result.Failures {
    fmt.Printf("Error in stream '%s': ", failure.StreamName)
    switch failure.ErrorCase {
    case kurrentdb.AppendStreamErrorCaseStreamRevisionConflict:
      fmt.Printf("version conflict (current: %d)\n", *failure.StreamRevision)
    case kurrentdb.AppendStreamErrorCaseAccessDenied:
      fmt.Printf("access denied - %s\n", failure.Reason)
    case kurrentdb.AppendStreamErrorCaseStreamDeleted:
      fmt.Printf("stream deleted\n")
    case kurrentdb.AppendStreamErrorCaseTransactionMaxSizeExceeded:
      fmt.Printf("transaction too large (max: %d bytes)\n", *failure.TransactionMaxSize)
    default:
      fmt.Printf("unexpected error\n")
    }
  }
}
```
