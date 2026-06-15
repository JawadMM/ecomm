package catalog

import (
	"context"
	"log"

	"github.com/JawadMM/ecomm/events"
	"github.com/segmentio/ksuid"
)

type Service interface {
	PostProduct(ctx context.Context, name, description string, price float64, stock uint32) (*Product, error)
	GetProduct(ctx context.Context, id string) (*Product, error)
	GetProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error)
	GetProductsByIds(ctx context.Context, ids []string) ([]Product, error)
	SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error)
	DecreaseStock(ctx context.Context, id string, quantity uint32) (*Product, error)
	IncreaseStock(ctx context.Context, id string, quantity uint32) (*Product, error)
}

type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       uint32  `json:"stock"`
}

type catalogService struct {
	repository Respository
	publisher  *events.Publisher
}

func NewService(repository Respository, pub *events.Publisher) *catalogService {
	return &catalogService{repository: repository, publisher: pub}
}

func (s *catalogService) PostProduct(ctx context.Context, name, description string, price float64, stock uint32) (*Product, error) {
	product := &Product{
		ID:          ksuid.New().String(),
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
	}
	product, err := s.repository.PutProduct(ctx, product)
	if err != nil {
		return nil, err
	}
	if s.publisher != nil {
		s.publisher.Publish(events.SubjectProductUpdated, events.ProductUpdatedEvent{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
		})
		log.Printf("[NATS pub] product.updated — id=%s name=%s price=%.2f", product.ID, product.Name, product.Price)
	}
	return product, nil
}

func (s *catalogService) GetProduct(ctx context.Context, id string) (*Product, error) {
	return s.repository.GetProduct(ctx, id)
}

func (s *catalogService) GetProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error) {
	if take > 100 || (take == 0 && skip == 0) {
		take = 100
	}
	return s.repository.ListProducts(ctx, skip, take)
}

func (s *catalogService) GetProductsByIds(ctx context.Context, ids []string) ([]Product, error) {
	return s.repository.ListProductsWithIds(ctx, ids)
}

func (s *catalogService) SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error) {
	if take > 100 || (take == 0 && skip == 0) {
		take = 100
	}
	return s.repository.SearchProducts(ctx, query, skip, take)
}

func (s *catalogService) DecreaseStock(ctx context.Context, id string, quantity uint32) (*Product, error) {
	return s.repository.DecreaseStock(ctx, id, quantity)
}

func (s *catalogService) IncreaseStock(ctx context.Context, id string, quantity uint32) (*Product, error) {
	return s.repository.IncreaseStock(ctx, id, quantity)
}
