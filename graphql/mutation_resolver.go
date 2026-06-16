package graphql

import (
	"context"
	"fmt"

	"github.com/JawadMM/ecomm/catalog"
	"github.com/JawadMM/ecomm/order"
)

type mutationResolver struct {
	server *GraphQLServer
}

func (r *mutationResolver) CreateAccount(ctx context.Context, input AccountInput) (*Account, error) {
	a, err := r.server.accountClient.PostAccount(ctx, input.Name)
	if err != nil {
		return nil, err
	}
	return &Account{ID: a.ID, Name: a.Name}, nil
}

func (r *mutationResolver) CreateProduct(ctx context.Context, input ProductInput) (*Product, error) {
	p, err := r.server.catalogClient.PostProduct(ctx, input.Name, input.Description, input.Price, uint32(input.Stock))
	if err != nil {
		return nil, err
	}
	return &Product{ID: p.ID, Name: p.Name, Description: p.Description, Price: p.Price, Stock: int(p.Stock)}, nil
}

func (r *mutationResolver) CreateOrder(ctx context.Context, input OrderInput) (*Order, error) {
	// Collect product IDs from the input
	productIDs := make([]string, len(input.Products))
	for i, p := range input.Products {
		productIDs[i] = p.ID
	}

	// Fetch full product details from the Catalog Service
	catalogProducts, err := r.server.catalogClient.GetProducts(ctx, "", productIDs, 0, 0)
	if err != nil {
		return nil, err
	}
	productMap := make(map[string]*catalog.Product, len(catalogProducts))
	for _, p := range catalogProducts {
		productMap[p.ID] = p
	}

	// Build ordered products with name, description, and price populated
	products := make([]order.OrderedProduct, len(input.Products))
	for i, p := range input.Products {
		cp, ok := productMap[p.ID]
		if !ok {
			return nil, fmt.Errorf("product %s not found", p.ID)
		}
		products[i] = order.OrderedProduct{
			ID:          cp.ID,
			Name:        cp.Name,
			Description: cp.Description,
			Price:       cp.Price,
			Quantity:    uint32(p.Quantity),
		}
	}

	o, err := r.server.orderClient.PostOrder(ctx, input.AccountID, products)
	if err != nil {
		return nil, err
	}
	orderedProducts := make([]OrderedProduct, len(o.Products))
	for i, p := range o.Products {
		orderedProducts[i] = OrderedProduct{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    int(p.Quantity),
		}
	}
	return &Order{
		ID:         o.ID,
		CreatedAt:  o.CreatedAt,
		TotalPrice: o.TotalPrice,
		Products:   orderedProducts,
	}, nil
}
