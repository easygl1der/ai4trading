from __future__ import annotations

import math
import urllib.parse
import urllib.request
from dataclasses import asdict, dataclass
from datetime import UTC, datetime
from typing import Any

from bitget_client import BitgetCandle, BitgetTicker
from market_discovery import MarketSnapshot


@dataclass
class ReferencePrice:
    symbol: str
    price: float | None
    source: str
    ok: bool
    error: str | None = None

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


@dataclass
class Signal:
    market_slug: str
    question: str
    action: str
    target_outcome: str | None
    size_usdc: float
    max_price: float | None
    fair_probability: float | None
    reason: str

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


class ThresholdStrategy:
    def __init__(
        self,
        bankroll: float = 1000.0,
        max_trade_fraction: float = 0.01,
        min_edge: float = 0.02,
        max_spread: float = 0.08,
    ) -> None:
        self.bankroll = bankroll
        self.max_trade_fraction = max_trade_fraction
        self.min_edge = min_edge
        self.max_spread = max_spread

    def signal(self, market: MarketSnapshot, reference: ReferencePrice) -> Signal:
        if market.market_type != "threshold_above" or market.threshold is None:
            return self._hold(market, "strategy only supports threshold_above markets")
        if not reference.ok or reference.price is None:
            return self._hold(market, f"missing reference price: {reference.error}")
        if market.best_ask is None:
            return self._hold(market, "missing best ask")
        if market.spread is not None and market.spread > self.max_spread:
            return self._hold(market, f"spread too wide: {market.spread:.4f}")

        fair = self._estimate_probability(reference.price, market.threshold, market.end_date)
        # CLOB /price side=SELL corresponds to the ask side a paper buyer crosses.
        yes_ask = market.clob_sell_price or market.best_ask
        no_ask = None if market.best_bid is None else 1.0 - market.best_bid
        trade_size = round(self.bankroll * self.max_trade_fraction, 2)

        if fair - yes_ask >= self.min_edge:
            return Signal(
                market_slug=market.market_slug,
                question=market.question,
                action="BUY_YES",
                target_outcome="YES",
                size_usdc=trade_size,
                max_price=round(yes_ask, 4),
                fair_probability=round(fair, 4),
                reason=(
                    f"estimated YES probability {fair:.3f} exceeds ask {yes_ask:.3f}; "
                    f"reference {reference.symbol}={reference.price:.2f}, threshold={market.threshold:.2f}"
                ),
            )

        if no_ask is not None and (1.0 - fair) - no_ask >= self.min_edge:
            return Signal(
                market_slug=market.market_slug,
                question=market.question,
                action="BUY_NO",
                target_outcome="NO",
                size_usdc=trade_size,
                max_price=round(no_ask, 4),
                fair_probability=round(fair, 4),
                reason=(
                    f"estimated NO probability {1.0 - fair:.3f} exceeds implied NO ask {no_ask:.3f}; "
                    f"reference {reference.symbol}={reference.price:.2f}, threshold={market.threshold:.2f}"
                ),
            )

        return self._hold(
            market,
            f"no edge: fair YES={fair:.3f}, yes ask={yes_ask:.3f}, no ask={no_ask}",
            fair,
        )

    def _estimate_probability(self, price: float, threshold: float, end_date: str | None) -> float:
        seconds = _seconds_to_expiry(end_date)
        days = max(seconds / 86400.0, 1.0 / 1440.0)
        annual_vol = 0.65
        sigma = annual_vol * math.sqrt(days / 365.0)
        if sigma <= 0:
            return 1.0 if price > threshold else 0.0
        z = math.log(price / threshold) / sigma
        return min(max(_normal_cdf(z), 0.001), 0.999)

    def _hold(self, market: MarketSnapshot, reason: str, fair: float | None = None) -> Signal:
        return Signal(
            market_slug=market.market_slug,
            question=market.question,
            action="HOLD",
            target_outcome=None,
            size_usdc=0.0,
            max_price=None,
            fair_probability=None if fair is None else round(fair, 4),
            reason=reason,
        )


class BitgetFiveMinuteUpDownStrategy:
    def __init__(
        self,
        bankroll: float = 1000.0,
        max_trade_fraction: float = 0.01,
        min_edge: float = 0.08,
        max_spread: float = 0.08,
        taker_fee_rate: float = 0.07,
    ) -> None:
        self.bankroll = bankroll
        self.max_trade_fraction = max_trade_fraction
        self.min_edge = min_edge
        self.max_spread = max_spread
        self.taker_fee_rate = taker_fee_rate

    def signal(
        self,
        market: MarketSnapshot,
        ticker: BitgetTicker | None,
        candles: list[BitgetCandle],
    ) -> Signal:
        if market.market_type not in {"up_down_5m", "up_down"}:
            return self._hold(market, "strategy only supports up/down markets")
        if ticker is None:
            return self._hold(market, "missing Bitget ticker")
        if len(candles) < 5:
            return self._hold(market, "need at least 5 one-minute Bitget candles")
        if market.spread is not None and market.spread > self.max_spread:
            return self._hold(market, f"spread too wide: {market.spread:.4f}")

        fair_up, model_reason = self._estimate_up_probability(ticker, candles)
        yes_ask = market.clob_sell_price or market.best_ask
        no_ask = None if market.best_bid is None else 1.0 - market.best_bid
        if yes_ask is None:
            return self._hold(market, "missing executable YES ask", fair_up)

        yes_fee_probability_cost = self._fee_probability_cost(yes_ask)
        yes_required_edge = self.min_edge + yes_fee_probability_cost
        trade_size = round(self.bankroll * self.max_trade_fraction, 2)

        if fair_up - yes_ask >= yes_required_edge:
            return Signal(
                market_slug=market.market_slug,
                question=market.question,
                action="BUY_YES",
                target_outcome="YES",
                size_usdc=trade_size,
                max_price=round(yes_ask, 4),
                fair_probability=round(fair_up, 4),
                reason=(
                    f"{model_reason}; fair_up={fair_up:.3f}, yes_ask={yes_ask:.3f}, "
                    f"required_edge={yes_required_edge:.3f}"
                ),
            )

        if no_ask is not None:
            no_fee_probability_cost = self._fee_probability_cost(no_ask)
            no_required_edge = self.min_edge + no_fee_probability_cost
            fair_down = 1.0 - fair_up
            if fair_down - no_ask >= no_required_edge:
                return Signal(
                    market_slug=market.market_slug,
                    question=market.question,
                    action="BUY_NO",
                    target_outcome="NO",
                    size_usdc=trade_size,
                    max_price=round(no_ask, 4),
                    fair_probability=round(fair_up, 4),
                    reason=(
                        f"{model_reason}; fair_down={fair_down:.3f}, no_ask={no_ask:.3f}, "
                        f"required_edge={no_required_edge:.3f}"
                    ),
                )

        return self._hold(
            market,
            f"no fee-adjusted edge: fair_up={fair_up:.3f}, yes_ask={yes_ask:.3f}, no_ask={no_ask}; {model_reason}",
            fair_up,
        )

    def _estimate_up_probability(
        self,
        ticker: BitgetTicker,
        candles: list[BitgetCandle],
    ) -> tuple[float, str]:
        latest = ticker.last_price
        closes = [candle.close for candle in candles[-6:]]
        start = closes[0]
        momentum = 0.0 if start <= 0 else (latest / start) - 1.0
        returns = []
        for prev, cur in zip(closes, closes[1:], strict=False):
            if prev > 0:
                returns.append(math.log(cur / prev))
        realized_sigma = _stddev(returns) if len(returns) >= 2 else 0.0
        scale = max(realized_sigma * math.sqrt(5.0), 0.0005)
        z = momentum / scale
        fair_up = min(max(_normal_cdf(z), 0.05), 0.95)
        return fair_up, f"Bitget 5m momentum={momentum:.5f}, realized_sigma_1m={realized_sigma:.5f}"

    def _fee_probability_cost(self, price: float) -> float:
        if price <= 0:
            return 1.0
        # Fee per share divided by price becomes an approximate probability-edge hurdle.
        return self.taker_fee_rate * price * (1.0 - price) / price

    def _hold(self, market: MarketSnapshot, reason: str, fair: float | None = None) -> Signal:
        return Signal(
            market_slug=market.market_slug,
            question=market.question,
            action="HOLD",
            target_outcome=None,
            size_usdc=0.0,
            max_price=None,
            fair_probability=None if fair is None else round(fair, 4),
            reason=reason,
        )


def fetch_binance_price(symbol: str, timeout: float = 10.0) -> ReferencePrice:
    url = "https://api.binance.com/api/v3/ticker/price?" + urllib.parse.urlencode({"symbol": symbol})
    request = urllib.request.Request(url, headers={"User-Agent": "ai4trading-paper/0.1"})
    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:
            payload = response.read().decode("utf-8")
        import json

        data = json.loads(payload)
        return ReferencePrice(symbol=symbol, price=float(data["price"]), source=url, ok=True)
    except Exception as exc:  # noqa: BLE001
        return ReferencePrice(symbol=symbol, price=None, source=url, ok=False, error=str(exc))


def _seconds_to_expiry(end_date: str | None) -> float:
    if not end_date:
        return 86400.0
    normalized = end_date.replace("Z", "+00:00")
    try:
        expiry = datetime.fromisoformat(normalized)
    except ValueError:
        return 86400.0
    now = datetime.now(UTC)
    if expiry.tzinfo is None:
        expiry = expiry.replace(tzinfo=UTC)
    return max((expiry - now).total_seconds(), 0.0)


def _normal_cdf(x: float) -> float:
    return 0.5 * (1.0 + math.erf(x / math.sqrt(2.0)))


def _stddev(values: list[float]) -> float:
    if len(values) < 2:
        return 0.0
    mean = sum(values) / len(values)
    variance = sum((value - mean) ** 2 for value in values) / (len(values) - 1)
    return math.sqrt(max(variance, 0.0))
