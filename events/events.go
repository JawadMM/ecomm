package events

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	SubjectOrderCreated   = "order.created"
	SubjectAccountCreated = "account.created"
	SubjectProductUpdated = "product.updated"
)

type OrderCreatedEvent struct {
	ID         string    `json:"id"`
	AccountID  string    `json:"account_id"`
	TotalPrice float64   `json:"total_price"`
	CreatedAt  time.Time `json:"created_at"`
}

type AccountCreatedEvent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProductUpdatedEvent struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type Publisher struct {
	nc *nats.Conn
}

func NewPublisher(url string) (*Publisher, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	return &Publisher{nc: nc}, nil
}

func (p *Publisher) Publish(subject string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return p.nc.Publish(subject, data)
}

func (p *Publisher) Close() {
	p.nc.Drain()
}

func Subscribe[T any](nc *nats.Conn, subject string, handler func(T)) (*nats.Subscription, error) {
	return nc.Subscribe(subject, func(msg *nats.Msg) {
		var event T
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("events: failed to unmarshal %s: %v", subject, err)
			return
		}
		handler(event)
	})
}

func NewConn(url string) (*nats.Conn, error) {
	return nats.Connect(url)
}
