from __future__ import annotations

from dataclasses import asdict, dataclass
from typing import Any


DEFAULT_POLYMARKET_CRYPTO_TAKER_FEE_RATE = 0.07


@dataclass(frozen=True)
class FeeQuote:
    shares: float
    price: float
    fee_rate: float
    fee_usdc: float
    fee_bps_of_notional: float

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


def polymarket_fee_usdc(shares: float, price: float, fee_rate: float) -> float:
    return max(shares, 0.0) * max(fee_rate, 0.0) * price * (1.0 - price)


def quote_polymarket_fee(shares: float, price: float, fee_rate: float) -> FeeQuote:
    fee = polymarket_fee_usdc(shares, price, fee_rate)
    notional = max(shares * price, 0.0)
    bps = 0.0 if notional <= 0 else fee / notional * 10000.0
    return FeeQuote(
        shares=round(shares, 8),
        price=round(price, 6),
        fee_rate=fee_rate,
        fee_usdc=round(fee, 6),
        fee_bps_of_notional=round(bps, 2),
    )
