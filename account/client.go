package account

import (
	"context"

	"github.com/JawadMM/ecomm/account/pb"
	"google.golang.org/grpc"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.AccountServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	c := pb.NewAccountServiceClient(conn)
	return &Client{
		conn:    conn,
		service: c,
	}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) PostAccount(ctx context.Context, name string) (*Account, error) {
	resp, err := c.service.PostAccount(ctx, &pb.PostAccountRequest{Name: name})
	if err != nil {
		return nil, err
	}
	return &Account{
		ID:   resp.Account.Id,
		Name: resp.Account.Name,
	}, nil
}

func (c *Client) GetAccount(ctx context.Context, id string) (*Account, error) {
	resp, err := c.service.GetAccount(ctx, &pb.GetAccountRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return &Account{
		ID:   resp.Account.Id,
		Name: resp.Account.Name,
	}, nil
}

func (c *Client) ListAccounts(ctx context.Context, skip, take uint64) ([]Account, error) {
	resp, err := c.service.ListAccounts(ctx, &pb.ListAccountsRequest{Skip: skip, Take: take})
	if err != nil {
		return nil, err
	}
	accounts := make([]Account, len(resp.Accounts))
	for i, a := range resp.Accounts {
		accounts[i] = Account{
			ID:   a.Id,
			Name: a.Name,
		}
	}
	return accounts, nil
}
