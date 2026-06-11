package graphql

import "context"

type accountResolver struct{
	server *GraphQLServer
}

func (r *accountResolver) Orders(ctx context.Context, obj *Account) ([]Order, error) {
	orders, err := r.server.orderClient.GetOrdersByAccountID(ctx, obj.ID)
	if err != nil {
		return nil, err
	}
	return orders, nil
}	