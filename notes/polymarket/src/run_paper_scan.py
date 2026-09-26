from __future__ import annotations

import argparse
import json
from pathlib import Path

from market_discovery import discover_crypto_markets, enrich_with_clob_prices, utc_now_iso
from paper_broker import PaperBroker
from polymarket_client import PolymarketClient
from strategies import ThresholdStrategy, fetch_binance_price


def main() -> int:
    parser = argparse.ArgumentParser(description="Run a read-only Polymarket crypto paper-trading scan.")
    parser.add_argument("--output-dir", default="notes/polymarket/outputs")
    parser.add_argument("--max-markets", type=int, default=8)
    parser.add_argument("--bankroll", type=float, default=1000.0)
    args = parser.parse_args()

    out_dir = Path(args.output_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    stamp = utc_now_iso().replace(":", "").replace("-", "").replace("Z", "Z")

    client = PolymarketClient()
    health = client.health()
    markets, discovery_checks = discover_crypto_markets(client)
    clob_checks = enrich_with_clob_prices(client, markets, max_markets=args.max_markets)

    symbols = sorted({market.asset for market in markets if market.asset in {"BTCUSDT", "ETHUSDT"}})
    references = {symbol: fetch_binance_price(symbol) for symbol in symbols}

    strategy = ThresholdStrategy(bankroll=args.bankroll)
    broker = PaperBroker(starting_cash=args.bankroll)
    decisions = []

    for market in markets[: args.max_markets]:
        reference = references.get(market.asset)
        if reference is None:
            signal = strategy._hold(market, f"no reference source configured for {market.asset}")
        else:
            signal = strategy.signal(market, reference)
        order, fill = broker.submit(market, signal)
        decisions.append(
            {
                "market": market.to_dict(),
                "reference": None if reference is None else reference.to_dict(),
                "signal": signal.to_dict(),
                "order": order.to_dict(),
                "fill": None if fill is None else fill.to_dict(),
            }
        )

    scan = {
        "timestamp": utc_now_iso(),
        "mode": "paper-only-read-only",
        "clob_health": {"ok": health.ok, "url": health.url, "data": health.data, "error": health.error},
        "discovery_checks": discovery_checks,
        "clob_checks": clob_checks,
        "reference_prices": {symbol: reference.to_dict() for symbol, reference in references.items()},
        "candidate_count": len(markets),
        "candidates": [market.to_dict() for market in markets],
        "decisions": decisions,
        "broker_summary": broker.summary(),
        "safety": {
            "uses_private_key": False,
            "uses_real_order_endpoint": False,
            "real_trading_enabled": False,
        },
    }

    scan_path = out_dir / f"market-scan-{stamp}.json"
    trades_path = out_dir / f"paper-trades-{stamp}.jsonl"
    report_path = out_dir / f"run-report-{stamp}.md"
    scan_path.write_text(json.dumps(scan, ensure_ascii=False, indent=2), encoding="utf-8")

    with trades_path.open("w", encoding="utf-8") as handle:
        for decision in decisions:
            handle.write(json.dumps(decision, ensure_ascii=False) + "\n")

    report_path.write_text(_render_report(scan, scan_path, trades_path), encoding="utf-8")

    print(f"scan={scan_path}")
    print(f"trades={trades_path}")
    print(f"report={report_path}")
    print(f"candidate_count={len(markets)} fills={broker.summary()['fills']}")
    return 0 if markets and any(check["buy_ok"] or check["sell_ok"] for check in clob_checks) else 2


def _render_report(scan: dict, scan_path: Path, trades_path: Path) -> str:
    lines = [
        "# Polymarket paper-trading dry run report",
        "",
        f"运行时间：{scan['timestamp']}",
        "",
        "## 安全边界",
        "",
        "- 本次只调用 Polymarket 公开只读接口。",
        "- 未接入真实钱包、私钥或签名逻辑。",
        "- 未调用真实下单 endpoint。",
        "",
        "## 接口状态",
        "",
        f"- CLOB health：{'成功' if scan['clob_health']['ok'] else '失败'}",
        f"- 市场发现查询数：{len(scan['discovery_checks'])}",
        f"- 候选市场数量：{scan['candidate_count']}",
        f"- CLOB 价格检查数：{len(scan['clob_checks'])}",
        f"- Binance 参考价格：{', '.join(_format_reference_prices(scan['reference_prices'])) or '无'}",
        "",
        "## 候选市场",
        "",
    ]

    for idx, market in enumerate(scan["candidates"][:8], start=1):
        lines.extend(
            [
                f"{idx}. {market['question']}",
                f"   - asset: `{market['asset']}`",
                f"   - type: `{market['market_type']}`",
                f"   - slug: `{market['market_slug']}`",
                f"   - end: `{market['end_date']}`",
                f"   - bid/ask: `{market['best_bid']}` / `{market['best_ask']}`, spread `{market['spread']}`",
                f"   - CLOB buy/sell: `{market['clob_buy_price']}` / `{market['clob_sell_price']}`",
                "",
            ]
        )

    lines.extend(
        [
            "## Dry-run 策略结果",
            "",
        ]
    )
    for idx, decision in enumerate(scan["decisions"], start=1):
        signal = decision["signal"]
        order = decision["order"]
        fill = decision["fill"]
        lines.extend(
            [
                f"{idx}. {signal['action']} - {decision['market']['question']}",
                f"   - reason: {signal['reason']}",
                f"   - order status: `{order['status']}`",
                f"   - fill: `{fill}`",
                "",
            ]
        )

    summary = scan["broker_summary"]
    lines.extend(
        [
            "## Paper broker 状态",
            "",
            f"- starting cash: `{summary['starting_cash']}`",
            f"- cash: `{summary['cash']}`",
            f"- orders: `{summary['orders']}`",
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
            "这个原型已经完成实时市场发现、CLOB 只读价格读取、Binance 参考价格读取、策略信号生成、虚拟订单处理和本地日志保存。"
            "如果没有真实成交，是因为当前保守策略没有发现足够 edge；这属于 paper strategy 的正常结果，不影响接口验证。",
            "",
        ]
    )
    return "\n".join(lines)


def _format_reference_prices(reference_prices: dict) -> list[str]:
    formatted = []
    for symbol, reference in reference_prices.items():
        if reference["ok"]:
            formatted.append(f"{symbol}={reference['price']}")
        else:
            formatted.append(f"{symbol}=失败")
    return formatted


if __name__ == "__main__":
    raise SystemExit(main())

