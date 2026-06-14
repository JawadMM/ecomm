package order

import (
	"context"
	"time"

	"github.com/JawadMM/ecomm/order/pb"
	"google.golang.org/grpc"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.OrderServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:    conn,
		service: pb.NewOrderServiceClient(conn),
	}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) PostOrder(ctx context.Context, accountID string, products []OrderedProduct) (*Order, error) {
	pbProducts := make([]*pb.OrderProduct, len(products))
	for i, p := range products {
		pbProducts[i] = &pb.OrderProduct{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    p.Quantity,
		}
	}
	resp, err := c.service.PostOrder(ctx, &pb.PostOrderRequest{
		AccountId: accountID,
		Products:  pbProducts,
	})
	if err != nil {
		return nil, err
	}
	createdAt, _ := time.Parse(time.RFC3339, resp.Order.CreatedAt)
	o := &Order{
		ID:         resp.Order.Id,
		CreatedAt:  createdAt,
		TotalPrice: resp.Order.TotalPrice,
		AccountID:  resp.Order.AccountId,
	}
	for _, p := range resp.Order.Products {
		o.Products = append(o.Products, OrderedProduct{
			ID:          p.Id,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    p.Quantity,
		})
	}
	return o, nil
}

func (c *Client) GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error) {
	resp, err := c.service.GetOrdersForAccount(ctx, &pb.GetOrdersForAccountRequest{
		AccountId: accountID,
	})
	if err != nil {
		return nil, err
	}
	orders := make([]Order, len(resp.Orders))
	for i, o := range resp.Orders {
		createdAt, _ := time.Parse(time.RFC3339, o.CreatedAt)
		orders[i] = Order{
			ID:         o.Id,
			CreatedAt:  createdAt,
			TotalPrice: o.TotalPrice,
			AccountID:  o.AccountId,
		}
		for _, p := range o.Products {
			orders[i].Products = append(orders[i].Products, OrderedProduct{
				ID:          p.Id,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
				Quantity:    p.Quantity,
			})
		}
	}
	return orders, nil
}
