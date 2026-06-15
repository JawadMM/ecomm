package graphql

import "context"

type queryResolver struct {
	server *GraphQLServer
}

func (r *queryResolver) Accounts(ctx context.Context, pagination *PaginationInput, id *string) ([]*Account, error) {
	if id != nil {
		a, err := r.server.accountClient.GetAccount(ctx, *id)
		if err != nil {
			return nil, err
		}
		return []*Account{{ID: a.ID, Name: a.Name}}, nil
	}

	var skip, take uint64
	if pagination != nil {
		if pagination.Skip != nil {
			skip = uint64(*pagination.Skip)
		}
		if pagination.Take != nil {
			take = uint64(*pagination.Take)
		}
	}

	accounts, err := r.server.accountClient.ListAccounts(ctx, skip, take)
	if err != nil {
		return nil, err
	}
	result := make([]*Account, len(accounts))
	for i, a := range accounts {
		result[i] = &Account{ID: a.ID, Name: a.Name}
	}
	return result, nil
}

func (r *queryResolver) Products(ctx context.Context, pagination *PaginationInput, query *string, id *string) ([]*Product, error) {
	var skip, take uint64
	var q string
	var ids []string

	if pagination != nil {
		if pagination.Skip != nil {
			skip = uint64(*pagination.Skip)
		}
		if pagination.Take != nil {
			take = uint64(*pagination.Take)
		}
	}
	if query != nil {
		q = *query
	}
	if id != nil {
		ids = []string{*id}
	}

	products, err := r.server.catalogClient.GetProducts(ctx, q, ids, skip, take)
	if err != nil {
		return nil, err
	}
	result := make([]*Product, len(products))
	for i, p := range products {
		result[i] = &Product{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Stock:       int(p.Stock),
		}
	}
	return result, nil
}
