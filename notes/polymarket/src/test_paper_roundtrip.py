from __future__ import annotations

import json
from pathlib import Path

from bitget_client import BitgetPublicClient
from fee_model import DEFAULT_POLYMARKET_CRYPTO_TAKER_FEE_RATE
from market_discovery import MarketSnapshot, utc_now_iso
from paper_broker import PaperBroker
from strategies import Signal


def main() -> int:
    out_dir = Path("notes/polymarket/outputs")
    out_dir.mkdir(parents=True, exist_ok=True)
    stamp = utc_now_iso().replace(":", "").replace("-", "").replace("Z", "Z")

    bitget = BitgetPublicClient(timeout=10.0)
    ticker, ticker_result = bitget.parsed_ticker("BTCUSDT")
    candles, candle_result = bitget.parsed_candles("BTCUSDT", "1min", 20)

    market = _mock_market()
    broker = PaperBroker(starting_cash=1000.0, taker_fee_rate=DEFAULT_POLYMARKET_CRYPTO_TAKER_FEE_RATE)

    buy_signal = Signal(
        market_slug=market.market_slug,
        question=market.question,
        action="BUY_YES",
        target_outcome="YES",
        size_usdc=10.0,
        max_price=0.515,
        fair_probability=0.8,
        reason="roundtrip test buy: cross mock ask",
    )
    buy_order, buy_fill = broker.submit(market, buy_signal)

    # Simulate a later market move where YES bid improves.
    market.best_bid = 0.56
    market.best_ask = 0.58
    market.spread = 0.02
    market.clob_buy_price = 0.56
    market.clob_sell_price = 0.58

    sell_signal = Signal(
        market_slug=market.market_slug,
        question=market.question,
        action="SELL_YES",
        target_outcome="YES",
        size_usdc=10.0,
        max_price=None,
        fair_probability=0.55,
        reason="roundtrip test sell: hit mock bid after price move",
    )
    sell_order, sell_fill = broker.submit(market, sell_signal)

    result = {
        "timestamp": utc_now_iso(),
        "mode": "paper-roundtrip-only",
        "api_reads": {
            "bitget_ticker": ticker_result.to_dict(),
            "bitget_candles": candle_result.to_dict(),
            "ticker": None if ticker is None else ticker.to_dict(),
            "candle_count": len(candles),
        },
        "api_used": {
            "bitget_public_ticker": "GET https://api.bitget.com/api/v2/spot/market/tickers?symbol=BTCUSDT",
            "bitget_public_candles": "GET https://api.bitget.com/api/v2/spot/market/candles?symbol=BTCUSDT&granularity=1min&limit=20",
            "polymarket_real_order": "not used",
            "bitget_real_order": "not used",
        },
        "market": market.to_dict(),
        "orders": [buy_order.to_dict(), sell_order.to_dict()],
        "fills": [
            None if buy_fill is None else buy_fill.to_dict(),
            None if sell_fill is None else sell_fill.to_dict(),
        ],
        "broker_summary": broker.summary(),
        "safety": {
            "uses_bitget_private_api": False,
            "uses_polymarket_private_key": False,
            "uses_real_order_endpoint": False,
            "real_trading_enabled": False,
        },
    }

    json_path = out_dir / f"paper-roundtrip-test-{stamp}.json"
    report_path = out_dir / f"paper-roundtrip-test-{stamp}.md"
    json_path.write_text(json.dumps(result, ensure_ascii=False, indent=2), encoding="utf-8")
    report_path.write_text(_render_report(result, json_path), encoding="utf-8")

    print(f"json={json_path}")
    print(f"report={report_path}")
    print(f"fills={broker.summary()['fills']} cash={broker.summary()['cash']} positions={broker.summary()['positions']}")
    return 0 if ticker_result.ok and broker.summary()["fills"] == 2 else 2


def _mock_market() -> MarketSnapshot:
    return MarketSnapshot(
        asset="BTCUSDT",
        event_title="MOCK Bitcoin Up or Down 5M",
        event_slug="mock-bitcoin-updown-5m",
        event_closed=False,
        question="MOCK: Bitcoin Up or Down - paper buy/sell roundtrip",
        market_slug="mock-btc-updown-roundtrip-paper-only",
        end_date=None,
        accepting_orders=True,
        enable_order_book=True,
        best_bid=0.485,
        best_ask=0.515,
        spread=0.03,
        min_order_size=1.0,
        min_tick_size=0.01,
        volume_24h=0.0,
        liquidity=0.0,
        yes_token_id="mock-yes-token",
        no_token_id="mock-no-token",
        market_type="up_down_5m",
        threshold=None,
        source_query="paper-roundtrip-test",
        clob_buy_price=0.485,
        clob_sell_price=0.515,
        clob_status="mock",
        raw={"mock": True},
    )


def _render_report(result: dict, json_path: Path) -> str:
    ticker = result["api_reads"]["ticker"]
    lines = [
        "# Paper buy/sell roundtrip test",
        "",
        f"运行时间：{result['timestamp']}",
        "",
        "## API reads",
        "",
        f"- Bitget ticker: {'成功' if result['api_reads']['bitget_ticker']['ok'] else '失败'}",
        f"- Bitget candles: {'成功' if result['api_reads']['bitget_candles']['ok'] else '失败'}",
    ]
    if ticker:
        lines.append(f"- BTCUSDT last: `{ticker['last_price']}`, bid: `{ticker['bid_price']}`, ask: `{ticker['ask_price']}`")
    lines.extend(
        [
            f"- candle count: `{result['api_reads']['candle_count']}`",
            "",
            "## Simulated orders",
            "",
        ]
    )
    for order, fill in zip(result["orders"], result["fills"], strict=False):
        lines.extend(
            [
                f"- `{order['action']}` status `{order['status']}`",
                f"  - fill: `{fill}`",
            ]
        )
    summary = result["broker_summary"]
    lines.extend(
        [
            "",
            "## Broker summary",
            "",
            f"- starting cash: `{summary['starting_cash']}`",
            f"- ending cash: `{summary['cash']}`",
            f"- fills: `{summary['fills']}`",
            f"- open positions: `{summary['positions']}`",
            "",
            "## APIs used",
            "",
            "- Bitget public ticker endpoint.",
            "- Bitget public spot candle endpoint.",
            "- No Bitget private trading endpoint.",
            "- No Polymarket real order endpoint.",
            "",
            "## Output",
            "",
            f"- JSON: `{json_path}`",
            "",
        ]
    )
    return "\n".join(lines)


if __name__ == "__main__":
    raise SystemExit(main())
