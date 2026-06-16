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

func productToPb(p *Product) *pb.Product {
	return &pb.Product{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
	}
}

func (s *grpcServer) PostProduct(ctx context.Context, req *pb.PostProductRequest) (*pb.PostProductResponse, error) {
	product, err := s.service.PostProduct(ctx, req.Name, req.Description, req.Price, req.Stock)
	if err != nil {
		return nil, err
	}
	return &pb.PostProductResponse{Product: productToPb(product)}, nil
}

func (s *grpcServer) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	product, err := s.service.GetProduct(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetProductResponse{Product: productToPb(product)}, nil
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
		pbProducts[i] = productToPb(&product)
	}
	return &pb.GetProductsResponse{Products: pbProducts}, nil
}

func (s *grpcServer) DecreaseStock(ctx context.Context, req *pb.DecreaseStockRequest) (*pb.DecreaseStockResponse, error) {
	product, err := s.service.DecreaseStock(ctx, req.ProductId, req.Quantity)
	if err != nil {
		return nil, err
	}
	return &pb.DecreaseStockResponse{Product: productToPb(product)}, nil
}

func (s *grpcServer) IncreaseStock(ctx context.Context, req *pb.IncreaseStockRequest) (*pb.IncreaseStockResponse, error) {
	product, err := s.service.IncreaseStock(ctx, req.ProductId, req.Quantity)
	if err != nil {
		return nil, err
	}
	return &pb.IncreaseStockResponse{Product: productToPb(product)}, nil
}
