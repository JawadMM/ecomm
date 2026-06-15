package order

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/JawadMM/ecomm/events"
	"github.com/segmentio/ksuid"
)

type StockClient interface {
	DecreaseStock(ctx context.Context, productID string, quantity uint32) error
	IncreaseStock(ctx context.Context, productID string, quantity uint32) error
}

type Service interface {
	PostOrder(ctx context.Context, accountID string, products []OrderedProduct) (*Order, error)
	GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error)
}

type Order struct {
	ID         string           `bson:"_id"`
	CreatedAt  time.Time        `bson:"created_at"`
	TotalPrice float64          `bson:"total_price"`
	AccountID  string           `bson:"account_id"`
	Products   []OrderedProduct `bson:"products"`
}

type OrderedProduct struct {
	ID          string  `bson:"id"`
	Name        string  `bson:"name"`
	Description string  `bson:"description"`
	Price       float64 `bson:"price"`
	Quantity    uint32  `bson:"quantity"`
}

type orderService struct {
	repository  Repository
	publisher   *events.Publisher
	stockClient StockClient
}

func NewService(r Repository, pub *events.Publisher, sc StockClient) Service {
	return &orderService{repository: r, publisher: pub, stockClient: sc}
}

func (s orderService) PostOrder(ctx context.Context, accountID string, products []OrderedProduct) (*Order, error) {
	since := time.Now().UTC().Add(-time.Minute)
	hasRecent, err := s.repository.HasRecentOrder(ctx, accountID, since)
	if err != nil {
		return nil, err
	}
	if hasRecent {
		return nil, fmt.Errorf("duplicate order: an order was already placed within the last minute for this account")
	}

	if s.stockClient != nil {
		decreased := make([]OrderedProduct, 0, len(products))
		for _, p := range products {
			if err := s.stockClient.DecreaseStock(ctx, p.ID, p.Quantity); err != nil {
				for _, dp := range decreased {
					if restoreErr := s.stockClient.IncreaseStock(ctx, dp.ID, dp.Quantity); restoreErr != nil {
						log.Printf("[stock] failed to restore stock for product %s: %v", dp.ID, restoreErr)
					}
				}
				return nil, err
			}
			decreased = append(decreased, p)
		}
	}

	o := &Order{
		ID:        ksuid.New().String(),
		CreatedAt: time.Now().UTC(),
		AccountID: accountID,
		Products:  products,
	}
	for _, p := range products {
		o.TotalPrice += p.Price * float64(p.Quantity)
	}
	if err := s.repository.PutOrder(ctx, *o); err != nil {
		return nil, err
	}
	if s.publisher != nil {
		s.publisher.Publish(events.SubjectOrderCreated, events.OrderCreatedEvent{
			ID:         o.ID,
			AccountID:  o.AccountID,
			TotalPrice: o.TotalPrice,
			CreatedAt:  o.CreatedAt,
		})
		log.Printf("[NATS pub] order.created — id=%s account=%s total=%.2f", o.ID, o.AccountID, o.TotalPrice)
	}
	return o, nil
}

func (s orderService) GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error) {
	return s.repository.GetOrdersForAccount(ctx, accountID)
}
