from __future__ import annotations

from dataclasses import asdict, dataclass
from typing import Any

from fee_model import DEFAULT_POLYMARKET_CRYPTO_TAKER_FEE_RATE, quote_polymarket_fee
from market_discovery import MarketSnapshot, utc_now_iso
from strategies import Signal


@dataclass
class SimulatedOrder:
    timestamp: str
    market_slug: str
    question: str
    action: str
    outcome: str | None
    size_usdc: float
    limit_price: float | None
    status: str
    reason: str

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


@dataclass
class Fill:
    timestamp: str
    market_slug: str
    outcome: str
    side: str
    shares: float
    price: float
    notional: float
    fee_usdc: float
    total_cost: float

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


@dataclass
class Position:
    market_slug: str
    outcome: str
    shares: float = 0.0
    cost_basis: float = 0.0

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


class PaperBroker:
    def __init__(
        self,
        starting_cash: float = 1000.0,
        taker_fee_rate: float = DEFAULT_POLYMARKET_CRYPTO_TAKER_FEE_RATE,
    ) -> None:
        self.cash = starting_cash
        self.starting_cash = starting_cash
        self.taker_fee_rate = taker_fee_rate
        self.positions: dict[tuple[str, str], Position] = {}
        self.orders: list[SimulatedOrder] = []
        self.fills: list[Fill] = []

    def submit(self, market: MarketSnapshot, signal: Signal) -> tuple[SimulatedOrder, Fill | None]:
        if signal.action == "HOLD":
            order = SimulatedOrder(
                timestamp=utc_now_iso(),
                market_slug=market.market_slug,
                question=market.question,
                action=signal.action,
                outcome=signal.target_outcome,
                size_usdc=signal.size_usdc,
                limit_price=signal.max_price,
                status="rejected",
                reason=signal.reason,
            )
            self.orders.append(order)
            return order, None

        if signal.action not in {"BUY_YES", "BUY_NO", "SELL_YES", "SELL_NO"}:
            return self._reject(market, signal, f"unsupported action {signal.action}")
        if signal.size_usdc <= 0:
            return self._reject(market, signal, "non-positive size")

        if signal.action.startswith("SELL_"):
            return self._submit_sell(market, signal)

        outcome = "YES" if signal.action == "BUY_YES" else "NO"
        price = self._buy_execution_price(market, outcome)
        if price is None or price <= 0:
            return self._reject(market, signal, "missing executable paper price")
        if signal.max_price is not None and price > signal.max_price:
            return self._reject(market, signal, f"paper execution price {price:.4f} exceeds limit {signal.max_price:.4f}")

        shares = round(signal.size_usdc / price, 6)
        fee_quote = quote_polymarket_fee(shares, price, self.taker_fee_rate)
        total_cost = round(signal.size_usdc + fee_quote.fee_usdc, 6)
        if total_cost > self.cash:
            return self._reject(market, signal, "insufficient paper cash including fee")
        if market.min_order_size is not None and signal.size_usdc < market.min_order_size:
            return self._reject(market, signal, "below Polymarket minimum order size")

        self.cash = round(self.cash - total_cost, 6)
        key = (market.market_slug, outcome)
        position = self.positions.get(key) or Position(market_slug=market.market_slug, outcome=outcome)
        position.shares = round(position.shares + shares, 6)
        position.cost_basis = round(position.cost_basis + total_cost, 6)
        self.positions[key] = position

        timestamp = utc_now_iso()
        order = SimulatedOrder(
            timestamp=timestamp,
            market_slug=market.market_slug,
            question=market.question,
            action=signal.action,
            outcome=outcome,
            size_usdc=signal.size_usdc,
            limit_price=signal.max_price,
            status="filled",
            reason=signal.reason,
        )
        fill = Fill(
            timestamp=timestamp,
            market_slug=market.market_slug,
            outcome=outcome,
            side="BUY",
            shares=shares,
            price=round(price, 6),
            notional=signal.size_usdc,
            fee_usdc=fee_quote.fee_usdc,
            total_cost=total_cost,
        )
        self.orders.append(order)
        self.fills.append(fill)
        return order, fill

    def summary(self) -> dict[str, Any]:
        return {
            "starting_cash": self.starting_cash,
            "cash": self.cash,
            "taker_fee_rate": self.taker_fee_rate,
            "open_position_count": len(self.positions),
            "orders": len(self.orders),
            "fills": len(self.fills),
            "positions": [position.to_dict() for position in self.positions.values()],
        }

    def _reject(self, market: MarketSnapshot, signal: Signal, reason: str) -> tuple[SimulatedOrder, None]:
        order = SimulatedOrder(
            timestamp=utc_now_iso(),
            market_slug=market.market_slug,
            question=market.question,
            action=signal.action,
            outcome=signal.target_outcome,
            size_usdc=signal.size_usdc,
            limit_price=signal.max_price,
            status="rejected",
            reason=f"{reason}; signal reason: {signal.reason}",
        )
        self.orders.append(order)
        return order, None

    def _submit_sell(self, market: MarketSnapshot, signal: Signal) -> tuple[SimulatedOrder, Fill | None]:
        outcome = "YES" if signal.action == "SELL_YES" else "NO"
        price = self._sell_execution_price(market, outcome)
        if price is None or price <= 0:
            return self._reject(market, signal, "missing executable paper sell price")

        key = (market.market_slug, outcome)
        position = self.positions.get(key)
        if position is None or position.shares <= 0:
            return self._reject(market, signal, f"no open {outcome} position to sell")

        requested_shares = round(signal.size_usdc / price, 6)
        shares = min(position.shares, requested_shares)
        if shares <= 0:
            return self._reject(market, signal, "non-positive sell shares")

        notional = round(shares * price, 6)
        fee_quote = quote_polymarket_fee(shares, price, self.taker_fee_rate)
        proceeds = round(notional - fee_quote.fee_usdc, 6)
        if proceeds < 0:
            return self._reject(market, signal, "sell fee exceeds notional")

        position.shares = round(position.shares - shares, 6)
        average_cost = 0.0 if position.shares + shares <= 0 else position.cost_basis / (position.shares + shares)
        position.cost_basis = round(max(position.cost_basis - average_cost * shares, 0.0), 6)
        if position.shares <= 0:
            self.positions.pop(key, None)
        else:
            self.positions[key] = position
        self.cash = round(self.cash + proceeds, 6)

        timestamp = utc_now_iso()
        order = SimulatedOrder(
            timestamp=timestamp,
            market_slug=market.market_slug,
            question=market.question,
            action=signal.action,
            outcome=outcome,
            size_usdc=signal.size_usdc,
            limit_price=signal.max_price,
            status="filled",
            reason=signal.reason,
        )
        fill = Fill(
            timestamp=timestamp,
            market_slug=market.market_slug,
            outcome=outcome,
            side="SELL",
            shares=shares,
            price=round(price, 6),
            notional=notional,
            fee_usdc=fee_quote.fee_usdc,
            total_cost=round(-proceeds, 6),
        )
        self.orders.append(order)
        self.fills.append(fill)
        return order, fill

    def _buy_execution_price(self, market: MarketSnapshot, outcome: str) -> float | None:
        if outcome == "YES":
            return market.clob_sell_price or market.best_ask
        if outcome == "NO" and market.best_bid is not None:
            return round(1.0 - market.best_bid, 6)
        return None

    def _sell_execution_price(self, market: MarketSnapshot, outcome: str) -> float | None:
        if outcome == "YES":
            return market.clob_buy_price or market.best_bid
        if outcome == "NO" and market.best_ask is not None:
            return round(1.0 - market.best_ask, 6)
        return None
