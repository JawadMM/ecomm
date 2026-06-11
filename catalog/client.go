package catalog

import (
	"context"

	"github.com/JawadMM/ecomm/catalog/pb"
	"google.golang.org/grpc"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.CatalogServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	c := pb.NewCatalogServiceClient(conn)
	return &Client{conn: conn, service: c}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) PostProduct(ctx context.Context, name, description string, price float64) (*Product, error) {
	res, err := c.service.PostProduct(ctx, &pb.PostProductRequest{Name: name, Description: description, Price: price})
	if err != nil {
		return nil, err
	}
	return &Product{
		ID:          res.Product.Id,
		Name:        res.Product.Name,
		Description: res.Product.Description,
		Price:       res.Product.Price,
	}, nil
}

func (c *Client) GetProduct(ctx context.Context, id string) (*Product, error) {
	res, err := c.service.GetProduct(
		ctx,
		&pb.GetProductRequest{Id: id},
	)
	if err != nil {
		return nil, err
	}
	return &Product{
		ID:          res.Product.Id,
		Name:        res.Product.Name,
		Description: res.Product.Description,
		Price:       res.Product.Price,
	}, nil
}

func (c *Client) GetProducts(ctx context.Context, query string, ids []string, skip, take uint64) ([]*Product, error) {
	res, err := c.service.GetProducts(
		ctx,
		&pb.GetProductsRequest{
			Query: query,
			Ids:   ids,
			Skip:  skip,
			Take:  take,
		},
	)
	if err != nil {
		return nil, err
	}
	
	products := make([]*Product, len(res.Products))
	for i, p := range res.Products {
		products[i] = &Product{
			ID:          p.Id,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		}
	}
	return products, nil
}
