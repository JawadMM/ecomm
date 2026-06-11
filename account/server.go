//go:generate protoc --go_out=pb --go_opt=paths=source_relative --go-grpc_out=pb --go-grpc_opt=paths=source_relative -I . account.proto

package account

import (
	"context"
	"fmt"
	"net"

	"github.com/JawadMM/ecomm/account/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	pb.UnimplementedAccountServiceServer
	service *AccountService
}

func ListenGRPC(service *AccountService, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	pb.RegisterAccountServiceServer(server, NewGRPCServer(service))
	reflection.Register(server)
	return server.Serve(lis)
}

func NewGRPCServer(service *AccountService) *grpcServer {
	return &grpcServer{service: service}
}

func (s *grpcServer) PostAccount(ctx context.Context, req *pb.PostAccountRequest) (*pb.PostAccountResponse, error) {
	account, err := s.service.PostAccount(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	return &pb.PostAccountResponse{Account: &pb.Account{Id: account.ID, Name: account.Name}}, nil
}

func (s *grpcServer) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	account, err := s.service.GetAccount(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetAccountResponse{Account: &pb.Account{Id: account.ID, Name: account.Name}}, nil
}

func (s *grpcServer) ListAccounts(ctx context.Context, req *pb.ListAccountsRequest) (*pb.ListAccountsResponse, error) {
	accounts, err := s.service.ListAccounts(ctx, req.Skip, req.Take)
	if err != nil {
		return nil, err
	}
	resp := &pb.ListAccountsResponse{}
	for _, account := range accounts {
		resp.Accounts = append(resp.Accounts, &pb.Account{Id: account.ID, Name: account.Name})
	}
	return resp, nil
}
