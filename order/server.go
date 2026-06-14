//go:generate protoc --go_out=pb --go_opt=paths=source_relative --go-grpc_out=pb --go-grpc_opt=paths=source_relative -I . order.proto

package order

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/JawadMM/ecomm/order/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	pb.UnimplementedOrderServiceServer
	service Service
}

func NewGRPCServer(service Service) *grpcServer {
	return &grpcServer{service: service}
}

func ListenGRPC(service Service, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	server := grpc.NewServer()
	pb.RegisterOrderServiceServer(server, NewGRPCServer(service))
	reflection.Register(server)
	return server.Serve(lis)
}

func (s *grpcServer) PostOrder(ctx context.Context, req *pb.PostOrderRequest) (*pb.PostOrderResponse, error) {
	products := make([]OrderedProduct, len(req.Products))
	for i, p := range req.Products {
		products[i] = OrderedProduct{
			ID:          p.Id,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    p.Quantity,
		}
	}
	o, err := s.service.PostOrder(ctx, req.AccountId, products)
	if err != nil {
		return nil, err
	}
	pbProducts := make([]*pb.OrderProduct, len(o.Products))
	for i, p := range o.Products {
		pbProducts[i] = &pb.OrderProduct{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    p.Quantity,
		}
	}
	return &pb.PostOrderResponse{
		Order: &pb.Order{
			Id:         o.ID,
			CreatedAt:  o.CreatedAt.Format(time.RFC3339),
			TotalPrice: o.TotalPrice,
			AccountId:  o.AccountID,
			Products:   pbProducts,
		},
	}, nil
}

func (s *grpcServer) GetOrdersForAccount(ctx context.Context, req *pb.GetOrdersForAccountRequest) (*pb.GetOrdersForAccountResponse, error) {
	orders, err := s.service.GetOrdersForAccount(ctx, req.AccountId)
	if err != nil {
		return nil, err
	}
	pbOrders := make([]*pb.Order, len(orders))
	for i, o := range orders {
		pbProducts := make([]*pb.OrderProduct, len(o.Products))
		for j, p := range o.Products {
			pbProducts[j] = &pb.OrderProduct{
				Id:       p.ID,
				Quantity: p.Quantity,
			}
		}
		pbOrders[i] = &pb.Order{
			Id:         o.ID,
			CreatedAt:  o.CreatedAt.Format(time.RFC3339),
			TotalPrice: o.TotalPrice,
			AccountId:  o.AccountID,
			Products:   pbProducts,
		}
	}
	return &pb.GetOrdersForAccountResponse{Orders: pbOrders}, nil
}
