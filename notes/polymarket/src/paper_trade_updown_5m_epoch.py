from __future__ import annotations

import argparse
import json
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path
from typing import Any

from bitget_client import BitgetPublicClient
from fee_model import DEFAULT_POLYMARKET_CRYPTO_TAKER_FEE_RATE
from market_discovery import MarketSnapshot, utc_now_iso
from paper_broker import PaperBroker
from scan_updown_5m_events import COINS, scan_coin
from strategies import BitgetFiveMinuteUpDownStrategy


BITGET_SYMBOLS = {
    "BTC": "BTCUSDT",
    "BNB": "BNBUSDT",
    "ETH": "ETHUSDT",
    "SOL": "SOLUSDT",
    "XRP": "XRPUSDT",
    "DOGE": "DOGEUSDT",
    "HYPE": "HYPEUSDT",
}


def main() -> int:
    parser = argparse.ArgumentParser(description="Paper trade real Polymarket 5m up/down markets for one epoch.")
    parser.add_argument("--epoch", type=int, required=True)
    parser.add_argument("--coins", nargs="+", default=["BTC", "ETH", "SOL", "XRP", "DOGE", "BNB", "HYPE"])
    parser.add_argument("--output-dir", default="notes/polymarket/outputs")
    parser.add_argument("--bankroll", type=float, default=1000.0)
    parser.add_argument("--timeout", type=float, default=8.0)
    parser.add_argument("--max-markets", type=int, default=7)
    parser.add_argument("--workers", type=int, default=7)
    args = parser.parse_args()

    out_dir = Path(args.output_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    stamp = utc_now_iso().replace(":", "").replace("-", "").replace("Z", "Z")

    selected_coins = [coin.upper() for coin in args.coins if coin.upper() in COINS][: args.max_markets]
    market_rows = scan_markets_concurrently(selected_coins, args.epoch, args.timeout, args.workers)

    bitget = BitgetPublicClient(timeout=args.timeout)
    symbols = sorted({BITGET_SYMBOLS[coin] for coin in selected_coins if coin in BITGET_SYMBOLS})
    bitget_data = fetch_bitget_concurrently(bitget, symbols, args.workers)
    strategy = BitgetFiveMinuteUpDownStrategy(
        bankroll=args.bankroll,
        taker_fee_rate=DEFAULT_POLYMARKET_CRYPTO_TAKER_FEE_RATE,
    )
    broker = PaperBroker(
        starting_cash=args.bankroll,
        taker_fee_rate=DEFAULT_POLYMARKET_CRYPTO_TAKER_FEE_RATE,
    )

    decisions = []
    for row in market_rows:
        if not row.get("ok") or not row.get("acceptingOrders"):
            decisions.append({"market_row": row, "skipped": True, "reason": "market unavailable or not accepting orders"})
            continue

        symbol = BITGET_SYMBOLS.get(row["coin"])
        snapshot = market_snapshot_from_scan_row(row, symbol)
        ticker_obj = bitget_data[symbol]["_ticker_obj"]
        candles_obj = bitget_data[symbol]["_candles_obj"]
        signal = strategy.signal(snapshot, ticker_obj, candles_obj)
        order, fill = broker.submit(snapshot, signal)
        decisions.append(
            {
                "coin": row["coin"],
                "symbol": symbol,
                "market": snapshot.to_dict(),
                "token_ids": row.get("tokenIds"),
                "market_url": row.get("url"),
                "bitget_ticker": None if ticker_obj is None else ticker_obj.to_dict(),
                "bitget_candle_count": len(candles_obj),
                "signal": signal.to_dict(),
                "order": order.to_dict(),
                "fill": None if fill is None else fill.to_dict(),
            }
        )

    for data in bitget_data.values():
        data.pop("_ticker_obj", None)
        data.pop("_candles_obj", None)

    result = {
        "timestamp": utc_now_iso(),
        "mode": "paper-trade-real-polymarket-5m-market-data",
        "epoch": args.epoch,
        "api_used": {
            "polymarket_gamma_event_by_slug": "GET https://gamma-api.polymarket.com/events/slug/{coin}-updown-5m-{epoch}",
            "polymarket_clob_price": "GET https://clob.polymarket.com/price?token_id={token_id}&side={BUY|SELL}",
            "bitget_public_ticker": "GET https://api.bitget.com/api/v2/spot/market/tickers?symbol={symbol}",
            "bitget_public_candles": "GET https://api.bitget.com/api/v2/spot/market/candles?symbol={symbol}&granularity=1min&limit=20",
            "real_order_api": "not used",
        },
        "safety": {
            "paper_only": True,
            "uses_bitget_private_api": False,
            "uses_polymarket_private_key": False,
            "uses_real_order_endpoint": False,
            "real_trading_enabled": False,
        },
        "markets": market_rows,
        "bitget": bitget_data,
        "decisions": decisions,
        "broker_summary": broker.summary(),
    }

    json_path = out_dir / f"updown-5m-paper-trade-{args.epoch}-{stamp}.json"
    trades_path = out_dir / f"updown-5m-paper-trades-{args.epoch}-{stamp}.jsonl"
    report_path = out_dir / f"updown-5m-paper-trade-{args.epoch}-{stamp}.md"
    json_path.write_text(json.dumps(result, ensure_ascii=False, indent=2), encoding="utf-8")
    with trades_path.open("w", encoding="utf-8") as handle:
        for decision in decisions:
            handle.write(json.dumps(decision, ensure_ascii=False) + "\n")
    report_path.write_text(render_report(result, json_path, trades_path), encoding="utf-8")

    print(f"json={json_path}")
    print(f"trades={trades_path}")
    print(f"report={report_path}")
    print(f"decisions={len(decisions)} fills={broker.summary()['fills']} cash={broker.summary()['cash']}")
    return 0 if decisions else 2


def scan_markets_concurrently(coins: list[str], epoch: int, timeout: float, workers: int) -> list[dict[str, Any]]:
    rows: list[dict[str, Any]] = []
    with ThreadPoolExecutor(max_workers=max(workers, 1)) as pool:
        futures = {
            pool.submit(scan_coin, coin, COINS[coin], epoch, timeout): coin
            for coin in coins
        }
        for future in as_completed(futures):
            rows.append(future.result())
    rows.sort(key=lambda row: coins.index(row["coin"]))
    return rows


def fetch_bitget_concurrently(
    bitget: BitgetPublicClient,
    symbols: list[str],
    workers: int,
) -> dict[str, dict[str, Any]]:
    def fetch(symbol: str) -> tuple[str, dict[str, Any]]:
        ticker, ticker_result = bitget.parsed_ticker(symbol)
        candles, candle_result = bitget.parsed_candles(symbol, "1min", 20)
        return symbol, {
            "ticker": None if ticker is None else ticker.to_dict(),
            "ticker_result": ticker_result.to_dict(),
            "candles": [candle.to_dict() for candle in candles],
            "candle_result": candle_result.to_dict(),
            "_ticker_obj": ticker,
            "_candles_obj": candles,
        }

    data: dict[str, dict[str, Any]] = {}
    with ThreadPoolExecutor(max_workers=max(workers, 1)) as pool:
        futures = [pool.submit(fetch, symbol) for symbol in symbols]
        for future in as_completed(futures):
            symbol, payload = future.result()
            data[symbol] = payload
    return data


def market_snapshot_from_scan_row(row: dict[str, Any], symbol: str) -> MarketSnapshot:
    token_ids = row.get("tokenIds") or {}
    clob = row.get("clobPrices") or {}
    up_buy = _price(clob.get("UP", {}).get("BUY", {}).get("price"))
    up_sell = _price(clob.get("UP", {}).get("SELL", {}).get("price"))
    best_bid = _price(row.get("bestBid"))
    best_ask = _price(row.get("bestAsk"))
    return MarketSnapshot(
        asset=symbol,
        event_title=row.get("title") or "",
        event_slug=row.get("slug") or "",
        event_closed=bool(row.get("closed")),
        question=row.get("title") or "",
        market_slug=row.get("slug") or "",
        end_date=row.get("endDate"),
        accepting_orders=bool(row.get("acceptingOrders")),
        enable_order_book=bool(row.get("enableOrderBook")),
        best_bid=best_bid,
        best_ask=best_ask,
        spread=_price(row.get("spread")),
        min_order_size=_price(row.get("orderMinSize")),
        min_tick_size=_price(row.get("tick")),
        volume_24h=_price(row.get("volume24hr")),
        liquidity=_price(row.get("liquidity")),
        yes_token_id=str(token_ids.get("UP") or ""),
        no_token_id=str(token_ids.get("DOWN") or ""),
        market_type="up_down_5m",
        threshold=None,
        source_query=f"epoch:{row.get('slug')}",
        clob_buy_price=up_buy or best_bid,
        clob_sell_price=up_sell or best_ask,
        clob_status="ok" if up_buy is not None or up_sell is not None else "missing",
    )


def _price(value: Any) -> float | None:
    if value in (None, ""):
        return None
    try:
        return float(value)
    except (TypeError, ValueError):
        return None


def render_report(result: dict[str, Any], json_path: Path, trades_path: Path) -> str:
    lines = [
        "# 5m Up/Down Paper Trade Run",
        "",
        f"运行时间：{result['timestamp']}",
        f"epoch：`{result['epoch']}`",
        "",
        "## 关键边界",
        "",
        "- Polymarket token id 只用于 Polymarket CLOB，不用于 Bitget。",
        "- Bitget 这里只作为行情源，不作为 Polymarket 下单通道。",
        "- 本次只做 paper trade，没有真实订单、钱包或私钥。",
        "",
        "## Decisions",
        "",
        "| Coin | Action | Price | Fill | Fee | Reason |",
        "| --- | --- | --- | --- | --- | --- |",
    ]
    for decision in result["decisions"]:
        if decision.get("skipped"):
            row = decision.get("market_row") or {}
            lines.append(
                "| "
                + " | ".join(
                    [
                        str(row.get("coin")),
                        "SKIP",
                        "",
                        "None",
                        "",
                        f"{decision.get('reason')}: {row.get('error', '')}".replace("|", "/"),
                    ]
                )
                + " |"
            )
            continue
        signal = decision.get("signal") or {}
        fill = decision.get("fill")
        fill_text = "None" if fill is None else f"{fill['side']} {fill['shares']} @ {fill['price']}"
        fee_text = "" if fill is None else str(fill["fee_usdc"])
        lines.append(
            "| "
            + " | ".join(
                [
                    str(decision.get("coin")),
                    str(signal.get("action")),
                    str(signal.get("max_price")),
                    fill_text,
                    fee_text,
                    str(signal.get("reason", "")).replace("|", "/"),
                ]
            )
            + " |"
        )
    summary = result["broker_summary"]
    lines.extend(
        [
            "",
            "## Broker",
            "",
            f"- starting cash: `{summary['starting_cash']}`",
            f"- ending cash: `{summary['cash']}`",
            f"- fills: `{summary['fills']}`",
            f"- positions: `{summary['positions']}`",
            "",
            "## API",
            "",
            "- Polymarket Gamma event by slug",
            "- Polymarket CLOB price by token id",
            "- Bitget public ticker/candles",
            "- No real order API",
            "",
            "## Files",
            "",
            f"- JSON: `{json_path}`",
            f"- Trades: `{trades_path}`",
            "",
        ]
    )
    return "\n".join(lines)


if __name__ == "__main__":
    raise SystemExit(main())
