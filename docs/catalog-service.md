# Catalog Service

Manages the product catalog and provides full-text product search. Exposes a gRPC API consumed by the GraphQL gateway and by the Order Service for stock management.

## Data Model

```go
type Product struct {
    ID          string  // KSUID, 27 chars
    Name        string
    Description string
    Price       float64
    Stock       uint32  // available units
}
```

## Database

**Technology:** Elasticsearch 7

**Connection:** configured via `DATABASE_URL` environment variable
(default compose value: `http://catalog_db:9200`)

**Index:** `product`

**Document format**
```json
{
  "_id": "<ksuid>",
  "name": "string",
  "description": "string",
  "price": 0.0,
  "stock": 0
}
```

The document `_id` is the product's KSUID. Elasticsearch handles indexing automatically; no schema migration is required.

---

## gRPC API

**Port:** `8080` (internal to Docker network; not exposed on the host)

**Proto service**

```protobuf
service CatalogService {
  rpc PostProduct   (PostProductRequest)   returns (PostProductResponse);
  rpc GetProduct    (GetProductRequest)    returns (GetProductResponse);
  rpc GetProducts   (GetProductsRequest)   returns (GetProductsResponse);
  rpc DecreaseStock (DecreaseStockRequest) returns (DecreaseStockResponse);
  rpc IncreaseStock (IncreaseStockRequest) returns (IncreaseStockResponse);
}
```

---

### `PostProduct`

Creates a new product in the catalog.

**Request**

| Field | Type | Description |
|-------|------|-------------|
| `name` | `string` | Product name |
| `description` | `string` | Product description (full-text indexed) |
| `price` | `double` | Unit price |
| `stock` | `uint32` | Initial available units |

**Response**

| Field | Type |
|-------|------|
| `product` | `Product` |

`Product` message:

| Field | Type |
|-------|------|
| `id` | `string` (KSUID) |
| `name` | `string` |
| `description` | `string` |
| `price` | `double` |
| `stock` | `uint32` |

---

### `GetProduct`

Fetches a single product by ID.

**Request**

| Field | Type |
|-------|------|
| `id` | `string` (KSUID) |

**Response**

| Field | Type |
|-------|------|
| `product` | `Product` |

---

### `GetProducts`

A multipurpose endpoint that handles three distinct access patterns based on which fields are set in the request.

**Request**

| Field | Type | Description |
|-------|------|-------------|
| `skip` | `uint64` | Offset for pagination |
| `take` | `uint64` | Page size (default/max `100`) |
| `query` | `string` | Full-text search string |
| `ids` | `repeated string` | Batch fetch by specific IDs |

**Routing logic (evaluated in order)**

| Condition | Backend call | Elasticsearch query |
|-----------|-------------|---------------------|
| `query` is non-empty | `SearchProducts(query, skip, take)` | Multi-match on `name` + `description` |
| `ids` is non-empty | `GetProductsByIds(ids)` | `mget` by document IDs |
| neither | `GetProducts(skip, take)` | Match-all, offset pagination |

**Response**

| Field | Type |
|-------|------|
| `products` | `repeated Product` |

---

### `DecreaseStock`

Atomically checks and decrements stock for a product. Called by the Order Service before persisting a new order.

**Request**

| Field | Type | Description |
|-------|------|-------------|
| `product_id` | `string` | KSUID of the product |
| `quantity` | `uint32` | Units to deduct |

**Response**

| Field | Type |
|-------|------|
| `product` | `Product` (updated stock) |

**Error:** Returns an error if `stock < quantity` (insufficient stock). The update is performed atomically via an Elasticsearch Painless script — no separate read is needed before decrementing.

---

### `IncreaseStock`

Unconditionally increments stock for a product. Used by the Order Service as a compensation step when a partial stock-decrease needs to be rolled back.

**Request**

| Field | Type | Description |
|-------|------|-------------|
| `product_id` | `string` | KSUID of the product |
| `quantity` | `uint32` | Units to restore |

**Response**

| Field | Type |
|-------|------|
| `product` | `Product` (updated stock) |

---

## Business Rules

- Pagination is capped at 100 results per request; if `take` is `0` or exceeds `100`, it is clamped to `100`.
- Full-text search uses Elasticsearch's `multi_match` query across both `name` and `description` fields.
- Batch-by-IDs uses Elasticsearch's `mget` API for efficient parallel document retrieval.
- IDs are generated server-side using KSUID; callers do not supply IDs.
- `DecreaseStock` uses an Elasticsearch scripted update (Painless), making the check-and-decrement atomic. A `noop` result (stock was already insufficient) is surfaced as an error to the caller.

---

## Events

### Published

| Subject | Trigger | Payload |
|---------|---------|---------|
| `product.updated` | After `PostProduct` succeeds | `{ id, name, description, price }` |

### Subscribed

_(none)_

---

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DATABASE_URL` | Elasticsearch base URL | `http://catalog_db:9200` |
| `NATS_URL` | NATS broker URL | `nats://nats:4222` |

---

## Startup Behavior

1. Reads `DATABASE_URL` and `NATS_URL` from environment.
2. Connects to Elasticsearch, retrying every 2 seconds until successful.
3. Connects to NATS for event publishing (optional; service starts even if NATS is unavailable).
4. Starts the gRPC server on `:8080`.
