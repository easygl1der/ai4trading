# Polymarket API Notes

## Public Endpoints

No authentication required for data endpoints.

### Get Event by Slug

```
GET https://gamma-api.polymarket.com/events/slug/{slug}
```

**Example:** `https://gamma-api.polymarket.com/events/slug/presidential-election-winner-2024`

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Event ID |
| `slug` | string | URL-friendly identifier |
| `title` | string | Event title |
| `active` | bool | Whether event is active |
| `closed` | bool | Whether event is closed |
| `volume` | float | Total trading volume |
| `negRisk` | bool | Uses Negative Risk pricing |
| `markets` | array | Array of market objects |

**Market Object Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Market ID |
| `question` | string | Prediction question |
| `conditionId` | string | On-chain condition identifier |
| `clobTokenIds` | array | [Yes_token_id, No_token_id] |
| `outcomePrices` | array | Current prices, e.g. ["1", "0"] = resolved |
| `volume` | float | Market trading volume |
| `orderMinSize` | float | Minimum order size |
| `orderPriceMinTickSize` | float | Price tick size |
| `negRisk` | bool | Negative risk market |

## Authentication (Trading)

Requires `py_clob_client` (Python >= 3.9.10):

```python
from py_clob_client.client import ClobClient

client = ClobClient(
    "https://clob.polymarket.com",
    key="<private-key>",
    chain_id=137
)
client.set_api_creds(client.create_or_derive_api_creds())
```

## Event Identifiers

### Event ID

Numeric integer uniquely identifying each event.

```bash
# Get event by ID
curl https://gamma-api.polymarket.com/events/35723

# Get tags for an event
curl https://gamma-api.polymarket.com/events/35723/tags
```

### Event Slug

URL-friendly text identifier extracted from event's web address.

```
URL: https://polymarket.com/event/fed-decision-in-october
                                    ↑
Slug: fed-decision-in-october
```

```bash
# Get event by slug (query parameter)
curl "https://gamma-api.polymarket.com/events?slug=fed-decision-in-october"

# Or use path endpoint
curl "https://gamma-api.polymarket.com/events/slug/fed-decision-in-october"
```

### Event Tags

Category labels for filtering events.

```bash
# Filter events by tag ID
curl "https://gamma-api.polymarket.com/events?tag_id=100381&active=true"
```

Common tags: `politics`, `crypto`, `sports`, `elections`, `economics`

## Markets API Endpoints

All GET endpoints are public (no auth required).

### 1. List markets (keyset pagination)
Cursor-based pagination for large datasets. Use `next_cursor` for pagination.

```bash
curl "https://gamma-api.polymarket.com/markets/keyset?limit=20&ascending=true"
curl "https://gamma-api.polymarket.com/markets/keyset?limit=50&closed=false&volume_num_min=10000"
curl "https://gamma-api.polymarket.com/markets/keyset?limit=20&after_cursor=eyJpZCI6MTIzNDU..."
```

**Use when:** Iterating through many markets, syncing large datasets.

### 2. List markets (offset pagination)
Standard listing with `offset` and `limit`.

```bash
curl "https://gamma-api.polymarket.com/markets?limit=100&offset=0&closed=false"
```

**Use when:** Simple pagination with filters like minimum volume or event ID.

### 3. Get market by ID
Fetch single market by numeric ID.

```bash
curl "https://gamma-api.polymarket.com/markets/21742"
```

**Use when:** You have a market ID and need full details.

### 4. Get market by slug
Fetch market by URL slug.

```bash
curl "https://gamma-api.polymarket.com/markets/will-bitcoin-hit-100k-by-2024"
```

**Use when:** Working with market URLs or descriptive names.

### 5. Get market tags by ID
Get categorization tags for a market.

```bash
curl "https://gamma-api.polymarket.com/markets/21742/tags"
```

**Use when:** Categorizing or filtering markets by topic.

### 6. Get market by token
Fetch market by CLOB token ID (ERC-1155 on Polygon).

```bash
curl "https://gamma-api.polymarket.com/markets?clob_token_ids=52114319..."
```

**Use when:** Working with on-chain data or blockchain explorers.

### 7. Get top holders
See largest positions in a market.

```bash
curl "https://gamma-api.polymarket.com/markets/top-holders?market_id=21742"
```

**Use when:** Analyzing whale activity, market sentiment.

### 8. Get open interest
Total outstanding positions in a market.

```bash
curl "https://gamma-api.polymarket.com/markets/21742/open-interest"
```

**Use when:** Assessing liquidity and market participation.

### 9. Get live volume for event
Real-time trading volume for all markets in an event.

```bash
curl "https://gamma-api.polymarket.com/events/100381/live-volume"
```

**Use when:** Monitoring trending events, measuring event popularity.

## Notes

- Python 3.9.6 is too old for `py_clob_client` (requires >= 3.9.10)
- For Python 3.9.6, use REST API directly with `requests` library
- Public data endpoints work without authentication
