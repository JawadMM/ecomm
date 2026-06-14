# go-ecomm

A Go-based e-commerce platform built as a collection of microservices. Each service owns its own database and communicates via gRPC for synchronous calls and NATS for asynchronous events. A GraphQL gateway serves as the single entry point for clients.

## Architecture

```
                        ┌─────────────────────────┐
                        │    Client (HTTP/GraphQL)  │
                        └────────────┬────────────-┘
                                     │ HTTP :8080
                        ┌────────────▼─────────────┐
                        │      GraphQL Gateway      │
                        │    POST /graphql           │
                        │    GET  /playground        │
                        └──────┬──────┬──────┬──────┘
                               │      │      │  gRPC
                  ┌────────────┘      │      └────────────┐
                  │                   │                    │
       ┌──────────▼──────┐  ┌─────────▼──────┐  ┌────────▼────────┐
       │  Account Service │  │ Catalog Service │  │  Order Service  │
       │   gRPC :8080     │  │  gRPC :8080    │  │  gRPC :8080     │
       └──────────┬───────┘  └────────┬───────┘  └────────┬────────┘
                  │                   │                    │
           ┌──────▼──────┐   ┌────────▼──────┐   ┌────────▼────────┐
           │  PostgreSQL  │   │ Elasticsearch │   │    MongoDB      │
           └─────────────┘   └───────────────┘   └─────────────────┘
                  │                   │                    │
                  └───────────────────┴────────────────────┘
                                      │  publish events
                              ┌───────▼───────┐
                              │  NATS Broker  │
                              │   :4222       │
                              └───────────────┘
```

## Services

| Service | Description | Database | Docs |
|---------|-------------|----------|------|
| GraphQL Gateway | Client-facing API layer | — | [graphql-api.md](docs/graphql-api.md) |
| Account | User account management | PostgreSQL | [account-service.md](docs/account-service.md) |
| Catalog | Product catalog & search | Elasticsearch | [catalog-service.md](docs/catalog-service.md) |
| Order | Order processing | MongoDB | [order-service.md](docs/order-service.md) |

See [architecture.md](docs/architecture.md) for a full explanation of how the services communicate.

## Tech Stack

- **Language:** Go
- **API Gateway:** GraphQL ([gqlgen](https://github.com/99designs/gqlgen))
- **Inter-service RPC:** gRPC + Protocol Buffers
- **Async messaging:** NATS
- **Databases:** PostgreSQL · Elasticsearch 7 · MongoDB 7
- **Containerization:** Docker / Docker Compose
- **ID generation:** [KSUID](https://github.com/segmentio/ksuid) (27-character sortable IDs)

## Prerequisites

- Docker and Docker Compose

## Quick Start

```bash
git clone <repo-url>
cd go-ecomm
docker-compose up
```

Once all containers are healthy:

| Endpoint | URL |
|----------|-----|
| GraphQL API | http://localhost:8080/graphql |
| GraphQL Playground | http://localhost:8080/playground |

## Example Queries

```graphql
# Create an account
mutation {
  createAccount(input: { name: "Jane Doe" }) {
    id
    name
  }
}

# Create a product
mutation {
  createProduct(input: { name: "Widget", description: "A useful widget", price: 29.99 }) {
    id
    name
    price
  }
}

# Place an order
mutation {
  createOrder(input: { accountId: "<account-id>", products: [{ id: "<product-id>", quantity: 2 }] }) {
    id
    createdAt
    totalPrice
  }
}

# Query accounts with their orders
query {
  accounts {
    id
    name
    orders {
      id
      createdAt
      totalPrice
      products {
        name
        quantity
      }
    }
  }
}

# Search products
query {
  products(query: "widget") {
    id
    name
    description
    price
  }
}
```

## Documentation

- [Architecture & Communication](docs/architecture.md)
- [GraphQL API](docs/graphql-api.md)
- [Account Service](docs/account-service.md)
- [Catalog Service](docs/catalog-service.md)
- [Order Service](docs/order-service.md)
