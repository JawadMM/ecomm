package catalog

import (
	"context"
	"fmt"
	"net"

	"github.com/JawadMM/ecomm/catalog/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	pb.UnimplementedCatalogServiceServer
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
	pb.RegisterCatalogServiceServer(server, NewGRPCServer(service))
	reflection.Register(server)
	return server.Serve(lis)
}

func (s *grpcServer) PostProduct(ctx context.Context, req *pb.PostProductRequest) (*pb.PostProductResponse, error) {
	product, err := s.service.PostProduct(ctx, req.Name, req.Description, req.Price)
	if err != nil {
		return nil, err
	}
	return &pb.PostProductResponse{Product: &pb.Product{Id: product.ID, Name: product.Name, Description: product.Description, Price: product.Price}}, nil
}

func (s *grpcServer) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	product, err := s.service.GetProduct(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetProductResponse{Product: &pb.Product{
		Id: product.ID, 
		Name: product.Name, 
		Description: product.Description, 
		Price: product.Price}}, nil
}

func (s *grpcServer) GetProducts(ctx context.Context, req *pb.GetProductsRequest) (*pb.GetProductsResponse, error) {
	var res []Product
	var err error

	if req.Query != "" {
		res, err = s.service.SearchProducts(ctx, req.Query, req.Skip, req.Take)
	} else if len(req.Ids) != 0 {
		res, err = s.service.GetProductsByIds(ctx, req.Ids)
	} else {
		res, err = s.service.GetProducts(ctx, req.Skip, req.Take)
	}
	if err != nil {
		return nil, err
	}

	pbProducts := make([]*pb.Product, len(res))
	for i, product := range res {
		pbProducts[i] = &pb.Product{
			Id: product.ID,
			Name: product.Name, 
			Description: product.Description, 
			Price: product.Price}
	}
	return &pb.GetProductsResponse{Products: pbProducts}, nil
}
