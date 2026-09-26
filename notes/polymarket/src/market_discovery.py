from __future__ import annotations

import json
import re
from dataclasses import asdict, dataclass, field
from datetime import UTC, datetime
from typing import Any

from polymarket_client import PolymarketClient


ASSET_SYMBOLS = {
    "bitcoin": "BTCUSDT",
    "btc": "BTCUSDT",
    "ethereum": "ETHUSDT",
    "eth": "ETHUSDT",
}


@dataclass
class MarketSnapshot:
    asset: str
    event_title: str
    event_slug: str
    event_closed: bool
    question: str
    market_slug: str
    end_date: str | None
    accepting_orders: bool
    enable_order_book: bool
    best_bid: float | None
    best_ask: float | None
    spread: float | None
    min_order_size: float | None
    min_tick_size: float | None
    volume_24h: float | None
    liquidity: float | None
    yes_token_id: str
    no_token_id: str
    market_type: str
    threshold: float | None
    source_query: str
    clob_buy_price: float | None = None
    clob_sell_price: float | None = None
    clob_status: str = "not_checked"
    raw: dict[str, Any] = field(default_factory=dict, repr=False)

    def to_dict(self) -> dict[str, Any]:
        data = asdict(self)
        data.pop("raw", None)
        return data


def discover_crypto_markets(client: PolymarketClient) -> tuple[list[MarketSnapshot], list[dict[str, Any]]]:
    queries = [
        "bitcoin up down",
        "ethereum up down",
        "bitcoin above",
        "ethereum above",
        "btc updown 5m",
        "eth updown 5m",
    ]
    markets: dict[tuple[str, str], MarketSnapshot] = {}
    api_events: list[dict[str, Any]] = []

    for query in queries:
        result = client.public_search(query)
        api_events.append(
            {
                "query": query,
                "ok": result.ok,
                "url": result.url,
                "error": result.error,
                "event_count": len((result.data or {}).get("events", [])) if result.ok else 0,
            }
        )
        if not result.ok:
            continue
        for event in (result.data or {}).get("events", []):
            for snapshot in _snapshots_from_event(event, query):
                key = (snapshot.market_slug, snapshot.yes_token_id)
                current = markets.get(key)
                if current is None or _rank_tuple(snapshot) > _rank_tuple(current):
                    markets[key] = snapshot

    sorted_markets = sorted(markets.values(), key=_rank_tuple, reverse=True)
    return sorted_markets, api_events


def enrich_with_clob_prices(
    client: PolymarketClient,
    markets: list[MarketSnapshot],
    max_markets: int = 8,
) -> list[dict[str, Any]]:
    checks: list[dict[str, Any]] = []
    for market in markets[:max_markets]:
        buy = client.clob_price(market.yes_token_id, "BUY")
        sell = client.clob_price(market.yes_token_id, "SELL")
        market.clob_status = "ok" if buy.ok or sell.ok else "failed"
        if buy.ok:
            market.clob_buy_price = _safe_float((buy.data or {}).get("price"))
        if sell.ok:
            market.clob_sell_price = _safe_float((sell.data or {}).get("price"))
        checks.append(
            {
                "market_slug": market.market_slug,
                "question": market.question,
                "yes_token_id": market.yes_token_id,
                "buy_ok": buy.ok,
                "buy_url": buy.url,
                "buy_error": buy.error,
                "buy_price": market.clob_buy_price,
                "sell_ok": sell.ok,
                "sell_url": sell.url,
                "sell_error": sell.error,
                "sell_price": market.clob_sell_price,
            }
        )
    return checks


def _snapshots_from_event(event: dict[str, Any], query: str) -> list[MarketSnapshot]:
    event_title = event.get("title") or event.get("question") or ""
    event_slug = event.get("slug") or ""
    asset = _infer_asset(" ".join([event_title, event_slug, query]))
    snapshots: list[MarketSnapshot] = []

    for market in event.get("markets") or []:
        if not _is_candidate(market):
            continue
        token_ids = _parse_json_list(market.get("clobTokenIds"))
        if len(token_ids) < 2:
            continue
        question = market.get("question") or ""
        best_bid = _safe_float(market.get("bestBid"))
        best_ask = _safe_float(market.get("bestAsk"))
        spread = None if best_bid is None or best_ask is None else round(best_ask - best_bid, 6)
        snapshots.append(
            MarketSnapshot(
                asset=asset or _infer_asset(question) or "UNKNOWN",
                event_title=event_title,
                event_slug=event_slug,
                event_closed=bool(event.get("closed")),
                question=question,
                market_slug=market.get("slug") or "",
                end_date=market.get("endDate"),
                accepting_orders=bool(market.get("acceptingOrders")),
                enable_order_book=bool(market.get("enableOrderBook")),
                best_bid=best_bid,
                best_ask=best_ask,
                spread=spread,
                min_order_size=_safe_float(market.get("orderMinSize")),
                min_tick_size=_safe_float(market.get("orderPriceMinTickSize")),
                volume_24h=_safe_float(market.get("volume24hr")),
                liquidity=_safe_float(market.get("liquidity")),
                yes_token_id=str(token_ids[0]),
                no_token_id=str(token_ids[1]),
                market_type=_market_type(question, event_title, event_slug),
                threshold=_extract_threshold(question),
                source_query=query,
                raw={"event": event, "market": market},
            )
        )
    return snapshots


def _is_candidate(market: dict[str, Any]) -> bool:
    return bool(
        market.get("active")
        and not market.get("closed")
        and market.get("acceptingOrders")
        and market.get("enableOrderBook")
        and market.get("clobTokenIds")
    )


def _rank_tuple(market: MarketSnapshot) -> tuple[float, float, float, float]:
    spread_score = 1.0 - min(max(market.spread or 1.0, 0.0), 1.0)
    return (
        float(market.volume_24h or 0.0),
        float(market.liquidity or 0.0),
        spread_score,
        1.0 if market.market_type == "up_down_5m" else 0.0,
    )


def _infer_asset(text: str) -> str | None:
    lowered = text.lower()
    for key, symbol in ASSET_SYMBOLS.items():
        if key in lowered:
            return symbol
    return None


def _market_type(question: str, event_title: str, event_slug: str) -> str:
    text = " ".join([question, event_title, event_slug]).lower()
    if "up or down" in text or "updown" in text:
        if "5m" in text or "5 minute" in text or "5-minute" in text:
            return "up_down_5m"
        return "up_down"
    if "above" in text:
        return "threshold_above"
    return "other"


def _extract_threshold(question: str) -> float | None:
    if "above" not in question.lower():
        return None
    dollar_match = re.search(r"\$\s*([0-9][0-9,]*(?:\.[0-9]+)?)", question)
    if dollar_match:
        value = dollar_match.group(1)
    else:
        after_above = re.split(r"\babove\b", question, flags=re.IGNORECASE, maxsplit=1)[-1]
        matches = re.findall(r"([0-9][0-9,]*(?:\.[0-9]+)?)", after_above)
        if not matches:
            return None
        value = matches[0]
    try:
        return float(value.replace(",", ""))
    except ValueError:
        return None


def _parse_json_list(value: Any) -> list[Any]:
    if isinstance(value, list):
        return value
    if isinstance(value, str):
        try:
            parsed = json.loads(value)
            return parsed if isinstance(parsed, list) else []
        except json.JSONDecodeError:
            return []
    return []


def _safe_float(value: Any) -> float | None:
    if value in (None, ""):
        return None
    try:
        return float(value)
    except (TypeError, ValueError):
        return None


def utc_now_iso() -> str:
    return datetime.now(UTC).isoformat(timespec="seconds").replace("+00:00", "Z")
