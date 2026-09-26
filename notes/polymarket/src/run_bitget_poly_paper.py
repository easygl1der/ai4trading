from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any

from bitget_client import BitgetCandle, BitgetPublicClient, BitgetTicker
from fee_model import DEFAULT_POLYMARKET_CRYPTO_TAKER_FEE_RATE
from market_discovery import MarketSnapshot, discover_crypto_markets, enrich_with_clob_prices, utc_now_iso
from paper_broker import PaperBroker
from polymarket_client import PolymarketClient
from strategies import BitgetFiveMinuteUpDownStrategy


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Run a Bitget-driven, Polymarket-read-only paper model for crypto up/down markets."
    )
    parser.add_argument("--output-dir", default="notes/polymarket/outputs")
    parser.add_argument("--symbols", nargs="+", default=["BTCUSDT", "ETHUSDT"])
    parser.add_argument("--max-markets", type=int, default=8)
    parser.add_argument("--bankroll", type=float, default=1000.0)
    parser.add_argument("--polymarket-timeout", type=float, default=5.0)
    parser.add_argument("--polymarket-retries", type=int, default=0)
    parser.add_argument(
        "--target",
        choices=["updown", "all"],
        default="updown",
        help="Use updown to evaluate only crypto up/down candidates; use all to scan every discovered crypto market.",
    )
    parser.add_argument("--mock-if-no-polymarket", action="store_true")
    args = parser.parse_args()

    out_dir = Path(args.output_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    stamp = utc_now_iso().replace(":", "").replace("-", "").replace("Z", "Z")

    bitget = BitgetPublicClient()
    bitget_snapshots: dict[str, dict[str, Any]] = {}
    tickers: dict[str, BitgetTicker] = {}
    candles_by_symbol: dict[str, list[BitgetCandle]] = {}

    for symbol in args.symbols:
        ticker, ticker_result = bitget.parsed_ticker(symbol)
        candles, candle_result = bitget.parsed_candles(symbol, "1min", 20)
        if ticker is not None:
            tickers[symbol] = ticker
        if candles:
            candles_by_symbol[symbol] = candles
        bitget_snapshots[symbol] = {
            "ticker_result": ticker_result.to_dict(),
            "candle_result": candle_result.to_dict(),
            "ticker": None if ticker is None else ticker.to_dict(),
            "candles": [candle.to_dict() for candle in candles],
        }

    poly = PolymarketClient(timeout=args.polymarket_timeout, retries=args.polymarket_retries)
    health = poly.health()
    markets, discovery_checks = discover_crypto_markets(poly)
    clob_checks = enrich_with_clob_prices(poly, markets, max_markets=args.max_markets)

    discovered_market_count = len(markets)
    if args.target == "updown":
        markets = [market for market in markets if market.market_type in {"up_down_5m", "up_down"}]

    used_mock_market = False
    if not markets and args.mock_if_no_polymarket:
        markets = [_mock_updown_market("BTCUSDT", tickers.get("BTCUSDT"))]
        used_mock_market = True

    strategy = BitgetFiveMinuteUpDownStrategy(
        bankroll=args.bankroll,
        taker_fee_rate=DEFAULT_POLYMARKET_CRYPTO_TAKER_FEE_RATE,
    )
    broker = PaperBroker(
        starting_cash=args.bankroll,
        taker_fee_rate=DEFAULT_POLYMARKET_CRYPTO_TAKER_FEE_RATE,
    )

    decisions = []
    for market in markets[: args.max_markets]:
        ticker = tickers.get(market.asset)
        candles = candles_by_symbol.get(market.asset, [])
        signal = strategy.signal(market, ticker, candles)
        order, fill = broker.submit(market, signal)
        decisions.append(
            {
                "market": market.to_dict(),
                "bitget_ticker": None if ticker is None else ticker.to_dict(),
                "bitget_candle_count": len(candles),
                "signal": signal.to_dict(),
                "order": order.to_dict(),
                "fill": None if fill is None else fill.to_dict(),
            }
        )

    scan = {
        "timestamp": utc_now_iso(),
        "mode": "bitget-polymarket-paper-only-read-only",
        "used_mock_market": used_mock_market,
        "bitget": bitget_snapshots,
        "polymarket": {
            "clob_health": {"ok": health.ok, "url": health.url, "data": health.data, "error": health.error},
            "discovery_checks": discovery_checks,
            "clob_checks": clob_checks,
            "discovered_market_count": discovered_market_count,
            "target": args.target,
            "candidate_count": len(markets),
            "candidates": [market.to_dict() for market in markets],
        },
        "fee_model": {
            "polymarket_crypto_taker_fee_rate": DEFAULT_POLYMARKET_CRYPTO_TAKER_FEE_RATE,
            "formula": "shares * fee_rate * price * (1 - price)",
        },
        "decisions": decisions,
        "broker_summary": broker.summary(),
        "safety": {
            "uses_bitget_private_api": False,
            "uses_polymarket_private_key": False,
            "uses_real_order_endpoint": False,
            "real_trading_enabled": False,
        },
    }

    scan_path = out_dir / f"bitget-poly-paper-scan-{stamp}.json"
    trades_path = out_dir / f"bitget-poly-paper-trades-{stamp}.jsonl"
    report_path = out_dir / f"bitget-poly-paper-report-{stamp}.md"
    scan_path.write_text(json.dumps(scan, ensure_ascii=False, indent=2), encoding="utf-8")

    with trades_path.open("w", encoding="utf-8") as handle:
        for decision in decisions:
            handle.write(json.dumps(decision, ensure_ascii=False) + "\n")

    report_path.write_text(_render_report(scan, scan_path, trades_path), encoding="utf-8")

    print(f"scan={scan_path}")
    print(f"trades={trades_path}")
    print(f"report={report_path}")
    print(
        "bitget_ok="
        f"{any(snapshot['ticker_result']['ok'] for snapshot in bitget_snapshots.values())} "
        f"polymarket_markets={len(markets)} fills={broker.summary()['fills']}"
    )
    return 0 if tickers and decisions else 2


def _mock_updown_market(symbol: str, ticker: BitgetTicker | None) -> MarketSnapshot:
    midpoint = 0.5
    if ticker is not None and ticker.bid_price is not None and ticker.ask_price is not None:
        micro_move = (ticker.ask_price - ticker.bid_price) / max(ticker.last_price, 1.0)
        midpoint = min(max(0.5 + micro_move, 0.05), 0.95)
    best_bid = round(max(midpoint - 0.015, 0.01), 3)
    best_ask = round(min(midpoint + 0.015, 0.99), 3)
    return MarketSnapshot(
        asset=symbol,
        event_title="MOCK Bitcoin Up or Down 5M",
        event_slug="mock-bitcoin-updown-5m",
        event_closed=False,
        question="MOCK: Bitcoin Up or Down - five-minute paper model connectivity fallback",
        market_slug="mock-btc-updown-5m-paper-only",
        end_date=None,
        accepting_orders=True,
        enable_order_book=True,
        best_bid=best_bid,
        best_ask=best_ask,
        spread=round(best_ask - best_bid, 6),
        min_order_size=1.0,
        min_tick_size=0.01,
        volume_24h=0.0,
        liquidity=0.0,
        yes_token_id="mock-yes-token",
        no_token_id="mock-no-token",
        market_type="up_down_5m",
        threshold=None,
        source_query="mock-if-no-polymarket",
        clob_buy_price=best_bid,
        clob_sell_price=best_ask,
        clob_status="mock",
        raw={"mock": True},
    )


def _render_report(scan: dict[str, Any], scan_path: Path, trades_path: Path) -> str:
    lines = [
        "# Bitget + Polymarket paper model report",
        "",
        f"运行时间：{scan['timestamp']}",
        "",
        "## 安全边界",
        "",
        "- 本次没有使用 Bitget private API。",
        "- 本次没有接入 Polymarket 私钥或钱包。",
        "- 本次没有调用任何真实下单 endpoint。",
        "- 所有订单都是本地 paper simulation。",
        "",
        "## Bitget 行情",
        "",
    ]

    for symbol, snapshot in scan["bitget"].items():
        ticker = snapshot["ticker"]
        lines.append(f"- {symbol}: {'成功' if snapshot['ticker_result']['ok'] else '失败'}")
        if ticker:
            lines.append(
                f"  - last `{ticker['last_price']}`, bid `{ticker['bid_price']}`, ask `{ticker['ask_price']}`, candles `{len(snapshot['candles'])}`"
            )
        elif snapshot["ticker_result"]["error"]:
            lines.append(f"  - error: {snapshot['ticker_result']['error']}")

    poly = scan["polymarket"]
    lines.extend(
        [
            "",
            "## Polymarket 只读检查",
            "",
            f"- CLOB health：{'成功' if poly['clob_health']['ok'] else '失败'}",
            f"- 市场发现查询数：{len(poly['discovery_checks'])}",
            f"- 原始候选市场数量：{poly['discovered_market_count']}",
            f"- 目标筛选：`{poly['target']}`",
            f"- 候选市场数量：{poly['candidate_count']}",
            f"- 使用 mock candidate：{scan['used_mock_market']}",
            "",
            "## Fee model",
            "",
            f"- crypto taker fee rate: `{scan['fee_model']['polymarket_crypto_taker_fee_rate']}`",
            "- formula: `shares * fee_rate * price * (1 - price)`",
            "",
            "## 决策",
            "",
        ]
    )

    for idx, decision in enumerate(scan["decisions"], start=1):
        signal = decision["signal"]
        fill = decision["fill"]
        lines.extend(
            [
                f"{idx}. `{signal['action']}` - {decision['market']['question']}",
                f"   - fair_probability: `{signal['fair_probability']}`",
                f"   - reason: {signal['reason']}",
                f"   - order status: `{decision['order']['status']}`",
                f"   - fill: `{fill}`",
                "",
            ]
        )

    summary = scan["broker_summary"]
    lines.extend(
        [
            "## Paper broker",
            "",
            f"- starting cash: `{summary['starting_cash']}`",
            f"- cash: `{summary['cash']}`",
            f"- fills: `{summary['fills']}`",
            f"- open positions: `{summary['open_position_count']}`",
            "",
            "## 输出文件",
            "",
            f"- scan JSON: `{scan_path}`",
            f"- trade JSONL: `{trades_path}`",
            "",
            "## 结论",
            "",
            "这个模型已经可以用 Bitget 的真实短周期行情生成 Polymarket up/down paper decision。"
            "如果 Polymarket 直连接口失败，报告会明确标出，并只用 mock candidate 验证模型和日志链路。",
            "",
        ]
    )
    return "\n".join(lines)


if __name__ == "__main__":
    raise SystemExit(main())
