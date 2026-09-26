from __future__ import annotations

import argparse
import json
import urllib.parse
import urllib.request
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path
from typing import Any

from market_discovery import utc_now_iso


COINS = {
    "BTC": "btc",
    "BNB": "bnb",
    "ETH": "eth",
    "SOL": "sol",
    "XRP": "xrp",
    "DOGE": "doge",
    "HYPE": "hype",
}


def main() -> int:
    parser = argparse.ArgumentParser(description="Scan Polymarket 5-minute crypto up/down markets by epoch.")
    parser.add_argument("--epoch", type=int, required=True, help="Window start epoch, for example 1781151000.")
    parser.add_argument("--output-dir", default="notes/polymarket/outputs")
    parser.add_argument("--coins", nargs="+", default=list(COINS.keys()))
    parser.add_argument("--timeout", type=float, default=8.0)
    parser.add_argument("--workers", type=int, default=8)
    args = parser.parse_args()

    out_dir = Path(args.output_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    stamp = utc_now_iso().replace(":", "").replace("-", "").replace("Z", "Z")

    selected = {coin.upper(): COINS[coin.upper()] for coin in args.coins if coin.upper() in COINS}
    with ThreadPoolExecutor(max_workers=args.workers) as pool:
        futures = [
            pool.submit(scan_coin, coin, prefix, args.epoch, args.timeout)
            for coin, prefix in selected.items()
        ]
        rows = [future.result() for future in as_completed(futures)]

    rows.sort(key=lambda row: list(selected).index(row["coin"]))
    result = {
        "timestamp": utc_now_iso(),
        "epoch": args.epoch,
        "api_used": {
            "gamma_event_by_slug": "GET https://gamma-api.polymarket.com/events/slug/{slug}",
            "clob_price": "GET https://clob.polymarket.com/price?token_id={token_id}&side={BUY|SELL}",
        },
        "safety": {
            "read_only": True,
            "uses_polymarket_private_key": False,
            "uses_real_order_endpoint": False,
            "places_real_orders": False,
        },
        "markets": rows,
    }

    json_path = out_dir / f"updown-5m-scan-{args.epoch}-{stamp}.json"
    report_path = out_dir / f"updown-5m-scan-{args.epoch}-{stamp}.md"
    json_path.write_text(json.dumps(result, ensure_ascii=False, indent=2), encoding="utf-8")
    report_path.write_text(render_report(result, json_path), encoding="utf-8")

    ok_count = sum(1 for row in rows if row["ok"])
    tradable_count = sum(1 for row in rows if row.get("acceptingOrders"))
    print(f"json={json_path}")
    print(f"report={report_path}")
    print(f"ok={ok_count}/{len(rows)} accepting_orders={tradable_count}/{len(rows)}")
    return 0 if ok_count else 2


def scan_coin(coin: str, prefix: str, epoch: int, timeout: float) -> dict[str, Any]:
    slug = f"{prefix}-updown-5m-{epoch}"
    event_url = f"https://gamma-api.polymarket.com/events/slug/{slug}"
    row: dict[str, Any] = {
        "coin": coin,
        "slug": slug,
        "url": f"https://polymarket.com/event/{slug}",
        "gamma_url": event_url,
    }
    try:
        event = get_json(event_url, timeout=timeout)
        market = (event.get("markets") or [{}])[0]
        token_ids = parse_json_list(market.get("clobTokenIds"))
        row.update(
            {
                "ok": True,
                "title": event.get("title"),
                "active": event.get("active"),
                "closed": event.get("closed"),
                "restricted": event.get("restricted"),
                "acceptingOrders": market.get("acceptingOrders"),
                "enableOrderBook": market.get("enableOrderBook"),
                "eventStartTime": market.get("eventStartTime"),
                "endDate": market.get("endDate"),
                "resolutionSource": market.get("resolutionSource"),
                "outcomes": parse_json_list(market.get("outcomes")),
                "bestBid": market.get("bestBid"),
                "bestAsk": market.get("bestAsk"),
                "spread": market.get("spread"),
                "liquidity": market.get("liquidity"),
                "volume24hr": market.get("volume24hr"),
                "orderMinSize": market.get("orderMinSize"),
                "tick": market.get("orderPriceMinTickSize"),
                "feeSchedule": market.get("feeSchedule"),
                "conditionId": market.get("conditionId"),
                "tokenIds": {
                    "UP": token_ids[0] if len(token_ids) > 0 else None,
                    "DOWN": token_ids[1] if len(token_ids) > 1 else None,
                },
            }
        )
        if len(token_ids) >= 2:
            row["clobPrices"] = {
                "UP": {
                    "BUY": clob_price(token_ids[0], "BUY", timeout),
                    "SELL": clob_price(token_ids[0], "SELL", timeout),
                },
                "DOWN": {
                    "BUY": clob_price(token_ids[1], "BUY", timeout),
                    "SELL": clob_price(token_ids[1], "SELL", timeout),
                },
            }
    except Exception as exc:  # noqa: BLE001
        row.update({"ok": False, "error": str(exc)})
    return row


def get_json(url: str, timeout: float) -> Any:
    request = urllib.request.Request(
        url,
        headers={
            "User-Agent": "ai4trading-polymarket-updown-scan/0.1 (+read-only)",
            "Accept": "application/json",
        },
    )
    with urllib.request.urlopen(request, timeout=timeout) as response:
        return json.loads(response.read().decode("utf-8", errors="replace"))


def clob_price(token_id: str, side: str, timeout: float) -> dict[str, Any]:
    url = "https://clob.polymarket.com/price?" + urllib.parse.urlencode(
        {"token_id": token_id, "side": side}
    )
    try:
        data = get_json(url, timeout=timeout)
        return {"ok": True, "price": data.get("price"), "url": url}
    except Exception as exc:  # noqa: BLE001
        return {"ok": False, "error": str(exc), "url": url}


def parse_json_list(value: Any) -> list[Any]:
    if isinstance(value, list):
        return value
    if isinstance(value, str):
        try:
            parsed = json.loads(value)
            return parsed if isinstance(parsed, list) else []
        except json.JSONDecodeError:
            return []
    return []


def render_report(result: dict[str, Any], json_path: Path) -> str:
    lines = [
        "# Polymarket 5m Up/Down Market Scan",
        "",
        f"运行时间：{result['timestamp']}",
        f"epoch：`{result['epoch']}`",
        "",
        "## API",
        "",
        "- Gamma event by slug：读取 event、market、token id、fee、bid/ask。",
        "- CLOB price：按 token id 读取 UP/DOWN 的 BUY/SELL 价格。",
        "- 未使用真实下单 API，未接入钱包或私钥。",
        "",
        "## Markets",
        "",
        "| Coin | OK | Accepting | Bid/Ask | UP CLOB | DOWN CLOB | Min | Fee | URL |",
        "| --- | --- | --- | --- | --- | --- | --- | --- | --- |",
    ]
    for row in result["markets"]:
        up = row.get("clobPrices", {}).get("UP", {})
        down = row.get("clobPrices", {}).get("DOWN", {})
        up_prices = format_clob(up)
        down_prices = format_clob(down)
        fee = row.get("feeSchedule") or {}
        fee_text = f"{fee.get('rate')}" if fee else ""
        lines.append(
            "| "
            + " | ".join(
                [
                    str(row["coin"]),
                    str(row.get("ok")),
                    str(row.get("acceptingOrders")),
                    f"{row.get('bestBid')}/{row.get('bestAsk')}",
                    up_prices,
                    down_prices,
                    str(row.get("orderMinSize")),
                    fee_text,
                    f"[link]({row['url']})",
                ]
            )
            + " |"
        )
    lines.extend(
        [
            "",
            "## JSON",
            "",
            f"- `{json_path}`",
            "",
        ]
    )
    return "\n".join(lines)


def format_clob(prices: dict[str, Any]) -> str:
    buy = prices.get("BUY", {})
    sell = prices.get("SELL", {})
    buy_text = buy.get("price") if buy.get("ok") else "ERR"
    sell_text = sell.get("price") if sell.get("ok") else "ERR"
    return f"B {buy_text} / S {sell_text}"


if __name__ == "__main__":
    raise SystemExit(main())
