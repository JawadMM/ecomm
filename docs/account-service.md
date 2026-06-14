# Account Service

Manages user accounts. Exposes a gRPC API consumed by the GraphQL gateway.

## Data Model

```go
type Account struct {
    ID   string // KSUID, 27 chars
    Name string // max 24 chars
}
```

## Database

**Technology:** PostgreSQL

**Connection:** configured via `DATABASE_URL` environment variable
(default compose value: `postgres://account:password@account_db/account?sslmode=disable`)

**Schema**

```sql
CREATE TABLE IF NOT EXISTS accounts (
    id   CHAR(27)     PRIMARY KEY,
    name VARCHAR(24)  NOT NULL
);
```

IDs are [KSUIDs](https://github.com/segmentio/ksuid) — 27-character, base-62 encoded, lexicographically sortable.

---

## gRPC API

**Port:** `8080` (internal to Docker network; not exposed on the host)

**Proto service**

```protobuf
service AccountService {
  rpc PostAccount    (PostAccountRequest)    returns (PostAccountResponse);
  rpc GetAccount     (GetAccountRequest)     returns (GetAccountResponse);
  rpc ListAccounts   (ListAccountsRequest)   returns (ListAccountsResponse);
}
```

---

### `PostAccount`

Creates a new account.

**Request**

| Field | Type | Description |
|-------|------|-------------|
| `name` | `string` | Account name (max 24 chars) |

**Response**

| Field | Type | Description |
|-------|------|-------------|
| `account` | `Account` | The newly created account |

`Account` message:

| Field | Type |
|-------|------|
| `id` | `string` (KSUID) |
| `name` | `string` |

---

### `GetAccount`

Fetches a single account by ID.

**Request**

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | KSUID of the account |

**Response**

| Field | Type |
|-------|------|
| `account` | `Account` |

Returns an error if no account with that ID exists.

---

### `ListAccounts`

Returns a paginated list of accounts ordered by ID.

**Request**

| Field | Type | Default | Max |
|-------|------|---------|-----|
| `skip` | `uint64` | `0` | — |
| `take` | `uint64` | `100` | `100` |

If `take` is `0` or greater than `100`, the service clamps it to `100`.

**Response**

| Field | Type |
|-------|------|
| `accounts` | `repeated Account` |

---

## Business Rules

- `name` is stored as-is (no normalization or uniqueness constraint).
- Pagination is capped at 100 results per request.
- IDs are generated server-side using KSUID; callers do not supply IDs.

---

## Events

### Published

| Subject | Trigger | Payload |
|---------|---------|---------|
| `account.created` | After `PostAccount` succeeds | `{ id, name }` |

### Subscribed

| Subject | Action |
|---------|--------|
| `order.created` | Logs the event (handler hook for future business logic) |

---

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://account:password@account_db/account?sslmode=disable` |
| `NATS_URL` | NATS broker URL | `nats://nats:4222` |

---

## Startup Behavior

1. Reads `DATABASE_URL` and `NATS_URL` from environment.
2. Connects to PostgreSQL, retrying every 2 seconds until successful.
3. Subscribes to `order.created` events on NATS.
4. Starts the gRPC server on `:8080`.
