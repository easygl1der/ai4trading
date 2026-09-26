from __future__ import annotations

import json
import time
import urllib.error
import urllib.parse
import urllib.request
from dataclasses import dataclass
from typing import Any


class PolymarketApiError(RuntimeError):
    pass


@dataclass(frozen=True)
class ApiResult:
    ok: bool
    url: str
    data: Any | None = None
    error: str | None = None


class PolymarketClient:
    """Read-only Polymarket API client for paper trading experiments."""

    gamma_base = "https://gamma-api.polymarket.com"
    clob_base = "https://clob.polymarket.com"

    def __init__(self, timeout: float = 20.0, retries: int = 2) -> None:
        self.timeout = timeout
        self.retries = retries
        self.headers = {
            "User-Agent": "ai4trading-polymarket-paper/0.1 (+read-only)",
            "Accept": "application/json",
        }

    def public_search(self, query: str) -> ApiResult:
        return self._get_json(
            f"{self.gamma_base}/public-search",
            {"q": query},
        )

    def clob_price(self, token_id: str, side: str = "BUY") -> ApiResult:
        return self._get_json(
            f"{self.clob_base}/price",
            {"token_id": token_id, "side": side.upper()},
        )

    def clob_order_book(self, token_id: str) -> ApiResult:
        return self._get_json(
            f"{self.clob_base}/book",
            {"token_id": token_id},
        )

    def health(self) -> ApiResult:
        return self._get_text(f"{self.clob_base}/ok")

    def _get_json(self, base_url: str, params: dict[str, Any]) -> ApiResult:
        url = base_url + "?" + urllib.parse.urlencode(params)
        result = self._request(url)
        if not result.ok:
            return result
        try:
            return ApiResult(ok=True, url=url, data=json.loads(result.data))
        except json.JSONDecodeError as exc:
            return ApiResult(ok=False, url=url, error=f"invalid json: {exc}")

    def _get_text(self, url: str) -> ApiResult:
        return self._request(url)

    def _request(self, url: str) -> ApiResult:
        last_error: str | None = None
        for attempt in range(self.retries + 1):
            request = urllib.request.Request(url, headers=self.headers)
            try:
                with urllib.request.urlopen(request, timeout=self.timeout) as response:
                    body = response.read().decode("utf-8", errors="replace")
                    return ApiResult(ok=True, url=url, data=body)
            except urllib.error.HTTPError as exc:
                detail = exc.read().decode("utf-8", errors="replace")[:500]
                last_error = f"http {exc.code}: {detail}"
            except (urllib.error.URLError, TimeoutError, OSError) as exc:
                last_error = str(exc)

            if attempt < self.retries:
                time.sleep(0.5 * (attempt + 1))

        return ApiResult(ok=False, url=url, error=last_error or "unknown error")

