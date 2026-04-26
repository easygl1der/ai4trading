# Polymarket Trading Guide

Get up and running with the Polymarket API — fetch market data and place your first order.

## 1. Install the SDK

```bash
pip install py-clob-client-v2
```

## 2. Set Up Your Client

Derive API credentials and initialize the trading client. This example uses an EOA wallet (type `0`) — your wallet pays its own gas and acts as the funder:

```python
from py_clob_client.client import ClobClient
import os

host = "https://clob.polymarket.com"
chain = 137  # Polygon mainnet
private_key = os.getenv("PRIVATE_KEY")

# Derive API credentials
temp_client = ClobClient(host, key=private_key, chain=chain)
api_creds = temp_client.create_or_derive_api_creds()

# Initialize trading client
client = ClobClient(
    host,
    key=private_key,
    chain=chain,
    creds=api_creds,
    signature_type=0,  # EOA
    funder="YOUR_WALLET_ADDRESS"
)
```

> **Note:** If you have a Polymarket.com account, your funds are in a proxy wallet — use signature type `1` or `2` instead. See [Signature Types](/trading/overview#signature-types).

> **Warning:** Before trading, your funder address needs **pUSD** (for buying outcome tokens) and **POL** (for gas, if using EOA type `0`). Proxy wallet users (types `1` and `2`) can use Polymarket's gasless relayer instead.

## 3. Place an Order

Get a token ID from the [Markets API](/market-data/fetching-markets), then create and submit your order:

```python
from py_clob_client.clob_types import OrderArgs, OrderType
from py_clob_client.order_builder.constants import BUY

response = client.create_and_post_order(
    OrderArgs(
        token_id="YOUR_TOKEN_ID",
        price=0.50,
        size=10,
        side=BUY,
    ),
    options={
        "tick_size": "0.01",
        "neg_risk": False,  # Set to True for multi-outcome markets
    },
    order_type=OrderType.GTC
)

print("Order ID:", response["orderID"])
print("Status:", response["status"])
```

> **Tip:** Look up a market's `tickSize` and `negRisk` values using the SDK's `getTickSize()` and `getNegRisk()` methods, or from the market object returned by the API.

## 4. Check Your Orders

```python
# View all open orders
open_orders = client.get_orders()
print(f"You have {len(open_orders)} open orders")

# View your trade history
trades = client.get_trades()
print(f"You've made {len(trades)} trades")

# Cancel an order
client.cancel(order_id=response["orderID"])
```

## Troubleshooting

**L2 AUTH NOT AVAILABLE - Invalid Signature**
- Wrong private key, signature type, or funder address for the derived API credentials
- Check that `signatureType` matches your account type (`0`, `1`, `2`, or `3`)
- Ensure `funder` is correct for your wallet type
- Re-derive credentials with `createOrDeriveApiKey()` if unsure

**Order rejected - insufficient balance**
- Your funder address doesn't have enough tokens
- **BUY orders**: need pUSD in your funder address
- **SELL orders**: need outcome tokens in your funder address
- Ensure you have more pUSD than what's committed in open orders

**Order rejected - insufficient allowance**
- You need to approve the Exchange contract to spend your tokens
- Typically done through the Polymarket UI on your first trade

**What is my funder address**
- **EOA (type 0)**: Your wallet address directly
- **Proxy wallet (type 1 or 2)**: Go to [polymarket.com/settings](https://polymarket.com/settings) and look for the wallet address in the profile dropdown

**Blocked by Cloudflare or Geoblock**
- You're trying to place a trade from a restricted region
- See [Geographic Restrictions](/api-reference/geoblock) for details

## Next Steps

- [Create Orders](/trading/orders/create) — Order types, tick sizes, and error handling
- [Order Attribution](/trading/orders/attribution) — Attribute orders to your builder account for volume credit
