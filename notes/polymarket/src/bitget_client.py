from __future__ import annotations

import json
import urllib.parse
import urllib.request
from dataclasses import asdict, dataclass
from typing import Any


@dataclass(frozen=True)
class BitgetResult:
    ok: bool
    url: str
    data: Any | None = None
    error: str | None = None

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


@dataclass(frozen=True)
class BitgetTicker:
    symbol: str
    last_price: float
    bid_price: float | None
    ask_price: float | None
    timestamp_ms: int | None
    raw: dict[str, Any]

    def to_dict(self) -> dict[str, Any]:
        data = asdict(self)
        data.pop("raw", None)
        return data


@dataclass(frozen=True)
class BitgetCandle:
    timestamp_ms: int
    open: float
    high: float
    low: float
    close: float
    base_volume: float
    quote_volume: float

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


class BitgetPublicClient:
    """Read-only Bitget public market-data client."""

    base_url = "https://api.bitget.com"

    def __init__(self, timeout: float = 10.0) -> None:
        self.timeout = timeout
        self.headers = {
            "User-Agent": "ai4trading-bitget-paper/0.1 (+read-only)",
            "Accept": "application/json",
        }

    def spot_ticker(self, symbol: str) -> BitgetResult:
        return self._get_json("/api/v2/spot/market/tickers", {"symbol": symbol})

    def spot_candles(self, symbol: str, granularity: str = "1min", limit: int = 100) -> BitgetResult:
        return self._get_json(
            "/api/v2/spot/market/candles",
            {"symbol": symbol, "granularity": granularity, "limit": str(limit)},
        )

    def parsed_ticker(self, symbol: str) -> tuple[BitgetTicker | None, BitgetResult]:
        result = self.spot_ticker(symbol)
        if not result.ok:
            return None, result
        rows = result.data if isinstance(result.data, list) else []
        if not rows:
            return None, BitgetResult(ok=False, url=result.url, error="empty ticker data")
        row = rows[0]
        try:
            ticker = BitgetTicker(
                symbol=symbol,
                last_price=float(row["lastPr"]),
                bid_price=_optional_float(row.get("bidPr")),
                ask_price=_optional_float(row.get("askPr")),
                timestamp_ms=_optional_int(row.get("ts")),
                raw=row,
            )
        except (KeyError, TypeError, ValueError) as exc:
            return None, BitgetResult(ok=False, url=result.url, data=result.data, error=f"bad ticker shape: {exc}")
        return ticker, result

    def parsed_candles(
        self,
        symbol: str,
        granularity: str = "1min",
        limit: int = 100,
    ) -> tuple[list[BitgetCandle], BitgetResult]:
        result = self.spot_candles(symbol, granularity, limit)
        if not result.ok:
            return [], result
        candles: list[BitgetCandle] = []
        for row in result.data or []:
            try:
                candles.append(
                    BitgetCandle(
                        timestamp_ms=int(row[0]),
                        open=float(row[1]),
                        high=float(row[2]),
                        low=float(row[3]),
                        close=float(row[4]),
                        base_volume=float(row[5]),
                        quote_volume=float(row[6]),
                    )
                )
            except (IndexError, TypeError, ValueError):
                continue
        candles.sort(key=lambda candle: candle.timestamp_ms)
        return candles, result

    def _get_json(self, path: str, params: dict[str, Any]) -> BitgetResult:
        url = self.base_url + path + "?" + urllib.parse.urlencode(params)
        request = urllib.request.Request(url, headers=self.headers)
        try:
            with urllib.request.urlopen(request, timeout=self.timeout) as response:
                payload = response.read().decode("utf-8", errors="replace")
            data = json.loads(payload)
            if data.get("code") != "00000":
                return BitgetResult(ok=False, url=url, data=data, error=f"bitget code {data.get('code')}: {data.get('msg')}")
            return BitgetResult(ok=True, url=url, data=data.get("data"))
        except Exception as exc:  # noqa: BLE001
            return BitgetResult(ok=False, url=url, error=str(exc))


def _optional_float(value: Any) -> float | None:
    if value in (None, ""):
        return None
    return float(value)


def _optional_int(value: Any) -> int | None:
    if value in (None, ""):
        return None
    return int(value)
