package nats_messaging

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/google/uuid"

	"github.com/nats-io/nats.go"
)

type NatsMessaging struct {
	nc *nats.Conn
	jc nats.JetStreamContext
}

func NewNatsMessaging() (*NatsMessaging, error) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal("failed to connect to NATS server")
		return nil, err
	}

	if !nc.IsConnected() {
		log.Fatal("failed to connect to NATS server")
		return nil, err
	}

	jc, err := nc.JetStream()
	if err != nil {
		log.Fatal("failed to create nats jet stream")
		return nil, err
	}

	log.Println("👉 successfully connected to the NATS server")
	log.Println("👉 successfully create NATS jet stream")

	return &NatsMessaging{nc: nc, jc: jc}, nil
}

func (n *NatsMessaging) Close() {
	n.nc.Close()
}

func (n *NatsMessaging) AddStream(streamName string, subjects ...string) error {
	if len(subjects) == 0 {
		subjects = []string{streamName + ".*"}
	}

	_, err := n.jc.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: subjects,
	})

	return err
}

type ProductEventType string

const (
	ProductCreatedEvent ProductEventType = "product.created"
	ProductUpdatedEvent ProductEventType = "product.updated"
	ProductDeletedEvent ProductEventType = "product.deleted"
)

type ProductPayload struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Description string    `json:"description"`
	Category  string    `json:"category"`
	Price     float64   `json:"price"`
	Quantity  uint      `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (n *NatsMessaging) Publish(eventType ProductEventType, data ProductPayload) error {
	payload, err := json.Marshal(data)
	if err != nil {
		log.Printf("failed to marshal product data: %v", err)
		return err
	}

	err = n.nc.Publish(string(eventType), payload)
	if err != nil {
		log.Printf("failed to publish %s event: %v", eventType, err)
		return err
	}

	return nil
}

func (n *NatsMessaging) Subscribe(evenType ProductEventType, elasticClient *elasticsearch.Client) {
	if evenType == ProductCreatedEvent {
		if _, err := n.nc.Subscribe(string(evenType), func(m *nats.Msg) {
			var data ProductPayload
			err := json.Unmarshal(m.Data, &data)
			if err != nil {
				log.Printf("failed to unmarshal product data: %v", err)
				return
			}

			// Prepare the request
			req := esapi.IndexRequest{
				Index:      "products",
				DocumentID: data.ID.String(),
				Body:       bytes.NewReader(m.Data),
				Refresh:    "true", // Refresh the index after inserting the document
			}

			// Send the request
			res, err := req.Do(context.Background(), elasticClient)
			if err != nil {
				log.Printf("Error inserting product: %s", err)
			}
			defer res.Body.Close()

			if res.IsError() {
				log.Printf("Error inserting book: %s", res.String())
			}

			log.Printf("New product created to elastic search: %v", data.ID)
		}); err != nil {
			log.Printf("failed to subscribe to product.created event: %v", err)
		} else {
			log.Println("👉 Subscribed to product.created event")
		}
	}
}
