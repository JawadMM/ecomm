# Order Service

Handles order creation and retrieval. Stores orders with embedded product snapshots so order history is preserved even if catalog prices change. Exposes a gRPC API consumed by the GraphQL gateway.

## Data Models

```go
type Order struct {
    ID         string           // KSUID, 27 chars
    CreatedAt  time.Time        // UTC
    TotalPrice float64          // Σ (product.price × quantity)
    AccountID  string           // KSUID of the owning account
    Products   []OrderedProduct
}

type OrderedProduct struct {
    ID          string
    Name        string
    Description string
    Price       float64  // price at time of order
    Quantity    uint32
}
```

## Database

**Technology:** MongoDB 7

**Connection:** configured via `DATABASE_URL` environment variable
(default compose value: `mongodb://order_db:27017/orders_db`)

**Database:** `orders_db`
**Collection:** `orders`

**Document format**
```json
{
  "_id": "<ksuid>",
  "created_at": "ISODate",
  "account_id": "<ksuid>",
  "total_price": 59.98,
  "products": [
    {
      "id": "<ksuid>",
      "name": "Widget",
      "description": "A useful widget",
      "price": 29.99,
      "quantity": 2
    }
  ]
}
```

Products are stored as embedded documents (snapshot), not references — this ensures order history is accurate even after product prices change.

---

## gRPC API

**Port:** `8080` (internal to Docker network; not exposed on the host)

**Proto service**

```protobuf
service OrderService {
  rpc PostOrder              (PostOrderRequest)              returns (PostOrderResponse);
  rpc GetOrdersForAccount    (GetOrdersForAccountRequest)    returns (GetOrdersForAccountResponse);
}
```

---

### `PostOrder`

Places a new order for an account.

**Request**

| Field | Type | Description |
|-------|------|-------------|
| `account_id` | `string` | KSUID of the account placing the order |
| `products` | `repeated OrderProduct` | Products and quantities |

`OrderProduct` message:

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Product KSUID |
| `name` | `string` | Product name (snapshot) |
| `description` | `string` | Product description (snapshot) |
| `price` | `double` | Unit price at order time (snapshot) |
| `quantity` | `uint32` | Number of units ordered |

**Response**

| Field | Type |
|-------|------|
| `order` | `Order` |

`Order` message:

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | KSUID of the new order |
| `created_at` | `string` | RFC3339 UTC timestamp |
| `total_price` | `double` | Calculated total |
| `account_id` | `string` | Owning account ID |
| `products` | `repeated OrderProduct` | Snapshot of ordered products |

---

### `GetOrdersForAccount`

Returns all orders for a given account, ordered by insertion order.

**Request**

| Field | Type |
|-------|------|
| `account_id` | `string` (KSUID) |

**Response**

| Field | Type |
|-------|------|
| `orders` | `repeated Order` |

Each `Order` contains the full product snapshot stored at order creation time.

---

## Business Rules

- **ID generation:** The service generates a new KSUID for each order; callers do not supply IDs.
- **Timestamp:** `created_at` is set server-side to the current UTC time at the moment of insertion.
- **Total price:** Calculated as `Σ (product.price × product.quantity)` over all products in the request.
- **Product snapshot:** Full product details (name, description, price) are copied into the order document at creation time, decoupling order history from future catalog changes.
- **No pagination** on `GetOrdersForAccount` — all orders for an account are returned in a single response.

---

## Events

### Published

| Subject | Trigger | Payload |
|---------|---------|---------|
| `order.created` | After `PostOrder` succeeds | `{ id, account_id, total_price, created_at }` |

### Subscribed

_(none)_

---

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DATABASE_URL` | MongoDB connection string | `mongodb://order_db:27017/orders_db` |
| `NATS_URL` | NATS broker URL | `nats://nats:4222` |

---

## Startup Behavior

1. Reads `DATABASE_URL` and `NATS_URL` from environment.
2. Connects to MongoDB, retrying every 2 seconds until successful.
3. Connects to NATS for event publishing (optional; service starts even if NATS is unavailable).
4. Starts the gRPC server on `:8080`.
