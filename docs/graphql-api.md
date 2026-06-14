# GraphQL API

The GraphQL gateway is the single entry point for all client interactions. It translates GraphQL operations into gRPC calls to the Account, Catalog, and Order services.

## HTTP Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/graphql` | Execute a GraphQL query or mutation |
| `GET` | `/playground` | Interactive GraphQL Playground IDE |

**Port:** `8080` (externally published)

**Content-Type:** `application/json`

---

## Schema

```graphql
scalar Time

# ── Types ──────────────────────────────────────────────────────────

type Account {
    id: String!
    name: String!
    orders: [Order!]!
}

type Product {
    id: String!
    name: String!
    description: String!
    price: Float!
}

type Order {
    id: String!
    createdAt: Time!
    totalPrice: Float!
    products: [OrderedProduct!]!
}

type OrderedProduct {
    id: String!
    name: String!
    description: String!
    price: Float!
    quantity: Int!
}

# ── Inputs ─────────────────────────────────────────────────────────

input PaginationInput {
    skip: Int   # default 0
    take: Int   # default / max 100
}

input AccountInput {
    name: String!
}

input ProductInput {
    name: String!
    description: String!
    price: Float!
}

input OrderProductInput {
    id: String!
    quantity: Int!
}

input OrderInput {
    accountId: String!
    products: [OrderProductInput!]!
}

# ── Queries ────────────────────────────────────────────────────────

type Query {
    accounts(pagination: PaginationInput, id: String): [Account!]!
    products(pagination: PaginationInput, query: String, id: String): [Product!]!
}

# ── Mutations ──────────────────────────────────────────────────────

type Mutation {
    createAccount(input: AccountInput!): Account!
    createProduct(input: ProductInput!): Product!
    createOrder(input: OrderInput!): Order!
}
```

---

## Queries

### `accounts`

Returns a list of accounts. Optionally filter by a single ID or paginate.

**Arguments**

| Argument | Type | Description |
|----------|------|-------------|
| `pagination` | `PaginationInput` | `skip` / `take` for offset pagination |
| `id` | `String` | If provided, returns only the matching account |

**Resolver behavior**
- When `id` is set → calls `Account.GetAccount(id)` on the Account Service.
- Otherwise → calls `Account.ListAccounts(skip, take)` on the Account Service.

The `orders` field on each `Account` is resolved lazily: when requested, the gateway calls `Order.GetOrdersForAccount(accountId)`.

**Example — list all accounts**
```graphql
query {
  accounts {
    id
    name
  }
}
```

```json
{
  "data": {
    "accounts": [
      { "id": "2abc...xyz", "name": "Jane Doe" }
    ]
  }
}
```

**Example — fetch one account with orders**
```graphql
query {
  accounts(id: "2abc...xyz") {
    id
    name
    orders {
      id
      createdAt
      totalPrice
      products {
        id
        name
        description
        price
        quantity
      }
    }
  }
}
```

```json
{
  "data": {
    "accounts": [
      {
        "id": "2abc...xyz",
        "name": "Jane Doe",
        "orders": [
          {
            "id": "2def...uvw",
            "createdAt": "2024-03-01T12:00:00Z",
            "totalPrice": 59.98,
            "products": [
              {
                "id": "2ghi...rst",
                "name": "Widget",
                "description": "A useful widget",
                "price": 29.99,
                "quantity": 2
              }
            ]
          }
        ]
      }
    ]
  }
}
```

**Example — paginated list**
```graphql
query {
  accounts(pagination: { skip: 0, take: 10 }) {
    id
    name
  }
}
```

---

### `products`

Returns a list of products. Supports full-text search, single-ID fetch, and paginated listing.

**Arguments**

| Argument | Type | Description |
|----------|------|-------------|
| `pagination` | `PaginationInput` | `skip` / `take` for offset pagination |
| `query` | `String` | Full-text search across `name` and `description` |
| `id` | `String` | If provided, returns only the matching product |

**Resolver behavior**
- When `id` is set → fetches that single product.
- When `query` is set → delegates to Elasticsearch multi-match search on `name` and `description`.
- Otherwise → lists all products with pagination.

**Example — list all products**
```graphql
query {
  products {
    id
    name
    description
    price
  }
}
```

**Example — search**
```graphql
query {
  products(query: "widget") {
    id
    name
    description
    price
  }
}
```

```json
{
  "data": {
    "products": [
      {
        "id": "2ghi...rst",
        "name": "Widget",
        "description": "A useful widget",
        "price": 29.99
      }
    ]
  }
}
```

**Example — paginated**
```graphql
query {
  products(pagination: { skip: 0, take: 20 }) {
    id
    name
    price
  }
}
```

---

## Mutations

### `createAccount`

Creates a new user account.

**Input**

| Field | Type | Constraints |
|-------|------|-------------|
| `name` | `String!` | Max 24 characters |

**Example**
```graphql
mutation {
  createAccount(input: { name: "Jane Doe" }) {
    id
    name
  }
}
```

```json
{
  "data": {
    "createAccount": {
      "id": "2abc...xyz",
      "name": "Jane Doe"
    }
  }
}
```

---

### `createProduct`

Adds a new product to the catalog.

**Input**

| Field | Type | Description |
|-------|------|-------------|
| `name` | `String!` | Product name |
| `description` | `String!` | Product description (searchable) |
| `price` | `Float!` | Unit price |

**Example**
```graphql
mutation {
  createProduct(input: {
    name: "Widget"
    description: "A useful widget for everyday tasks"
    price: 29.99
  }) {
    id
    name
    description
    price
  }
}
```

```json
{
  "data": {
    "createProduct": {
      "id": "2ghi...rst",
      "name": "Widget",
      "description": "A useful widget for everyday tasks",
      "price": 29.99
    }
  }
}
```

---

### `createOrder`

Places a new order for an account.

**Input**

| Field | Type | Description |
|-------|------|-------------|
| `accountId` | `String!` | ID of the account placing the order |
| `products` | `[OrderProductInput!]!` | List of products and quantities |

`OrderProductInput`:

| Field | Type | Description |
|-------|------|-------------|
| `id` | `String!` | Product ID |
| `quantity` | `Int!` | Number of units |

The gateway forwards the product list directly to the Order Service. The Order Service calculates `totalPrice` as `Σ (product.price × quantity)`.

**Example**
```graphql
mutation {
  createOrder(input: {
    accountId: "2abc...xyz"
    products: [
      { id: "2ghi...rst", quantity: 2 }
    ]
  }) {
    id
    createdAt
    totalPrice
    products {
      id
      name
      description
      price
      quantity
    }
  }
}
```

```json
{
  "data": {
    "createOrder": {
      "id": "2def...uvw",
      "createdAt": "2024-03-01T12:00:00Z",
      "totalPrice": 59.98,
      "products": [
        {
          "id": "2ghi...rst",
          "name": "Widget",
          "description": "A useful widget for everyday tasks",
          "price": 29.99,
          "quantity": 2
        }
      ]
    }
  }
}
```

---

## Pagination

Both `accounts` and `products` accept a `PaginationInput`:

| Field | Default | Max | Description |
|-------|---------|-----|-------------|
| `skip` | `0` | — | Number of records to skip |
| `take` | `100` | `100` | Number of records to return |

If `take` exceeds 100 or is not provided, the backend clamps it to 100.

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `ACCOUNT_SERVICE_URL` | gRPC address of Account Service (e.g. `account:8080`) |
| `CATALOG_SERVICE_URL` | gRPC address of Catalog Service (e.g. `catalog:8080`) |
| `ORDER_SERVICE_URL` | gRPC address of Order Service (e.g. `order:8080`) |
