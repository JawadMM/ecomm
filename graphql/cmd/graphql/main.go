package main

import (
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	gql "github.com/JawadMM/ecomm/graphql"
	"github.com/kelseyhightower/envconfig"
)

type AppConfig struct {
	AccountServiceURL string `envconfig:"ACCOUNT_SERVICE_URL"`
	CatalogServiceURL string `envconfig:"CATALOG_SERVICE_URL"`
	OrderServiceURL   string `envconfig:"ORDER_SERVICE_URL"`
}

func main() {
	var config AppConfig
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatalf("Failed to process environment variables: %v", err)
	}

	server, err := gql.NewGraphQLServer(config.AccountServiceURL, config.CatalogServiceURL, config.OrderServiceURL)
	if err != nil {
		log.Fatalf("Failed to create GraphQL server: %v", err)
	}

	http.Handle("/graphql", handler.NewDefaultServer(server.ToExecutableSchema()))
	http.Handle("/playground", playground.Handler("GraphQL Playground", "/graphql"))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
