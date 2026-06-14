package graphql

import (
	"github.com/99designs/gqlgen/graphql"
	"github.com/JawadMM/ecomm/account"
	"github.com/JawadMM/ecomm/catalog"
	"github.com/JawadMM/ecomm/order"
)

type GraphQLServer struct {
	accountClient *account.Client
	catalogClient *catalog.Client
	orderClient   *order.Client
}

func NewGraphQLServer(accountUrl, catalogUrl, orderUrl string) (*GraphQLServer, error) {
	accountClient, err := account.NewClient(accountUrl)
	if err != nil {
		return nil, err
	}

	catalogClient, err := catalog.NewClient(catalogUrl)
	if err != nil {
		accountClient.Close()
		return nil, err
	}

	orderClient, err := order.NewClient(orderUrl)
	if err != nil {
		accountClient.Close()
		catalogClient.Close()
		return nil, err
	}

	return &GraphQLServer{
		accountClient: accountClient,
		catalogClient: catalogClient,
		orderClient:   orderClient,
	}, nil
}

func (s *GraphQLServer) Mutation() MutationResolver {
	return &mutationResolver{server: s}
}

func (s *GraphQLServer) Query() QueryResolver {
	return &queryResolver{server: s}
}

func (s *GraphQLServer) Account() AccountResolver {
	return &accountResolver{server: s}
}

func (s *GraphQLServer) ToExecutableSchema() graphql.ExecutableSchema {
	return NewExecutableSchema(Config{Resolvers: s})
}
