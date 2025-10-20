package samples

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"slices"

	"github.com/google/uuid"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

type TestEvent struct {
	Id            string
	ImportantData string
}

func AppendToStream(db *kurrentdb.Client) {
	// region append-to-stream
	data := TestEvent{
		Id:            "1",
		ImportantData: "some value",
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	options := kurrentdb.AppendToStreamOptions{
		StreamState: kurrentdb.NoStream{},
	}

	result, err := db.AppendToStream(context.Background(), "some-stream", options, kurrentdb.EventData{
		ContentType: kurrentdb.ContentTypeJson,
		EventType:   "some-event",
		Data:        bytes,
	})
	// endregion append-to-stream

	log.Printf("Result: %v", result)
}

func AppendWithSameId(db *kurrentdb.Client) {
	// region append-duplicate-event
	data := TestEvent{
		Id:            "1",
		ImportantData: "some value",
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	id := uuid.New()
	event := kurrentdb.EventData{
		ContentType: kurrentdb.ContentTypeJson,
		EventType:   "some-event",
		EventID:     id,
		Data:        bytes,
	}

	_, err = db.AppendToStream(context.Background(), "some-stream", kurrentdb.AppendToStreamOptions{}, event)

	if err != nil {
		panic(err)
	}

	// attempt to append the same event again
	_, err = db.AppendToStream(context.Background(), "some-stream", kurrentdb.AppendToStreamOptions{}, event)

	if err != nil {
		panic(err)
	}

	// endregion append-duplicate-event
}

func AppendWithNoStream(db *kurrentdb.Client) {
	// region append-with-no-stream
	data := TestEvent{
		Id:            "1",
		ImportantData: "some value",
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	options := kurrentdb.AppendToStreamOptions{
		StreamState: kurrentdb.NoStream{},
	}

	_, err = db.AppendToStream(context.Background(), "same-event-stream", options, kurrentdb.EventData{
		ContentType: kurrentdb.ContentTypeJson,
		EventType:   "some-event",
		Data:        bytes,
	})

	if err != nil {
		panic(err)
	}

	bytes, err = json.Marshal(TestEvent{
		Id:            "2",
		ImportantData: "some other value",
	})
	if err != nil {
		panic(err)
	}

	// attempt to append the same event again
	_, err = db.AppendToStream(context.Background(), "same-event-stream", options, kurrentdb.EventData{
		ContentType: kurrentdb.ContentTypeJson,
		EventType:   "some-event",
		Data:        bytes,
	})
	// endregion append-with-no-stream
}

func AppendWithConcurrencyCheck(db *kurrentdb.Client) {
	// region append-with-concurrency-check
	ropts := kurrentdb.ReadStreamOptions{
		Direction: kurrentdb.Backwards,
		From:      kurrentdb.End{},
	}

	stream, err := db.ReadStream(context.Background(), "concurrency-stream", ropts, 1)

	if err != nil {
		panic(err)
	}

	defer stream.Close()

	lastEvent, err := stream.Recv()

	if err != nil {
		panic(err)
	}

	data := TestEvent{
		Id:            "1",
		ImportantData: "clientOne",
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	aopts := kurrentdb.AppendToStreamOptions{
		StreamState: lastEvent.OriginalStreamRevision(),
	}

	_, err = db.AppendToStream(context.Background(), "concurrency-stream", aopts, kurrentdb.EventData{
		ContentType: kurrentdb.ContentTypeJson,
		EventType:   "some-event",
		Data:        bytes,
	})

	data = TestEvent{
		Id:            "1",
		ImportantData: "clientTwo",
	}
	bytes, err = json.Marshal(data)
	if err != nil {
		panic(err)
	}

	_, err = db.AppendToStream(context.Background(), "concurrency-stream", aopts, kurrentdb.EventData{
		ContentType: kurrentdb.ContentTypeJson,
		EventType:   "some-event",
		Data:        bytes,
	})
	// endregion append-with-concurrency-check
}

func AppendToStreamOverridingUserCredentials(db *kurrentdb.Client) {
	data := TestEvent{
		Id:            "1",
		ImportantData: "some value",
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	event := kurrentdb.EventData{
		ContentType: kurrentdb.ContentTypeJson,
		EventType:   "some-event",
		Data:        bytes,
	}

	// region overriding-user-credentials
	credentials := &kurrentdb.Credentials{Login: "admin", Password: "changeit"}

	result, err := db.AppendToStream(context.Background(), "some-stream", kurrentdb.AppendToStreamOptions{Authenticated: credentials}, event)
	// endregion overriding-user-credentials

	log.Printf("Result: %v", result)
}

func AppendToMultipleStreams(db *kurrentdb.Client) {
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

	_, err := db.MultiStreamAppend(context.Background(), slices.Values(requests))
	if err != nil {
		var streamRevisionConflictErr *kurrentdb.StreamRevisionConflictError
		if errors.As(err, &streamRevisionConflictErr) {
			log.Printf("Stream revision conflict on stream %s: expected %v but was %v", streamRevisionConflictErr.Stream, streamRevisionConflictErr.ExpectedRevision, streamRevisionConflictErr.ActualRevision)
		}
	}
}
