package graphql

import "context"

type mutationResolver struct{ 
	server *GraphQLServer
}

func (r *mutationResolver) CreateAccount(ctx context.Context, input AccountInput) (*Account, error) {
	account, err := r.server.accountClient.CreateAccount(ctx, input.Name)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (r *mutationResolver) CreateProduct(ctx context.Context, input ProductInput) (*Product, error) {
	product, err := r.server.catalogClient.CreateProduct(ctx, input.Name, input.Price)
	if err != nil {
		return nil, err
	}
	return product, nil
}	

func (r *mutationResolver) CreateOrder(ctx context.Context, input OrderInput) (*Order, error) {
	order, err := r.server.orderClient.CreateOrder(ctx, input.AccountID, input.Product, input.Price)
	if err != nil {
		return nil, err
	}
	return order, nil
}
