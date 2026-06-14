package graphql

import "context"

type accountResolver struct {
	server *GraphQLServer
}

func (r *accountResolver) Orders(ctx context.Context, obj *Account) ([]*Order, error) {
	orders, err := r.server.orderClient.GetOrdersForAccount(ctx, obj.ID)
	if err != nil {
		return nil, err
	}
	result := make([]*Order, len(orders))
	for i, o := range orders {
		products := make([]OrderedProduct, len(o.Products))
		for j, p := range o.Products {
			products[j] = OrderedProduct{
				ID:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
				Quantity:    int(p.Quantity),
			}
		}
		result[i] = &Order{
			ID:         o.ID,
			CreatedAt:  o.CreatedAt,
			TotalPrice: o.TotalPrice,
			Products:   products,
		}
	}
	return result, nil
}
