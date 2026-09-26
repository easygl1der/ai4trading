"""
生成 IV/事件表:
- korea_holidays.parquet: 韩国 KRX 全天休市日期（窗口内），作为 IV 候选
- earnings.parquet: MU/SNDK 财报日期与前后 pre/post 标记
- events_daily.parquet: 每日事件哑变量表 (date x event)
窗口: 2026-04-11 到 2026-07-01
"""
import os
import pandas as pd

OUT = "/home/user/workspace/data/meta"
os.makedirs(OUT, exist_ok=True)

# ============ 韩国 KRX 假日 (窗口内) ============
# 数据源: KRX Holidays 2026 (多源交叉验证)
korea_holidays = [
    # (date, holiday_name_en, korean_name, type)
    ("2026-05-01", "Labor Day",           "근로자의 날",  "full_close"),
    ("2026-05-05", "Children's Day",      "어린이날",     "full_close"),
    ("2026-05-25", "Buddha's Birthday (Obs)", "부처님오신날", "full_close"),
    ("2026-06-03", "Election Day",        "지방선거일",   "full_close"),
    # 注: 6月6日 Memorial Day 是周六，KRX 本就不开; Aug 17 Liberation Day 超出窗口
]
kh = pd.DataFrame(korea_holidays, columns=["date", "name_en", "name_ko", "type"])
kh["date"] = pd.to_datetime(kh["date"])
kh.to_parquet(f"{OUT}/korea_holidays.parquet", index=False)
print(f"[OK] Korea holidays: {len(kh)} rows -> {OUT}/korea_holidays.parquet")
print(kh)

# ============ MU / SNDK 财报日期 (窗口内) ============
# MU: 2026-06-24 Wed, 2:30pm MT = 16:30 ET = 收盘后 (after-hours)
# SNDK: 2026-04-30 Thu, 时间未细查, 一般 SanDisk 是收盘后
earnings = pd.DataFrame([
    {"symbol": "MU",   "date": "2026-06-24", "session": "after_close", "fiscal_period": "FY26 Q3", "note": "2:30pm MT"},
    {"symbol": "SNDK", "date": "2026-04-30", "session": "after_close", "fiscal_period": "FY26 Q3", "note": "typical after-hours"},
])
earnings["date"] = pd.to_datetime(earnings["date"])
earnings.to_parquet(f"{OUT}/earnings.parquet", index=False)
print(f"\n[OK] Earnings: {len(earnings)} rows -> {OUT}/earnings.parquet")
print(earnings)

# ============ 每日事件哑变量表 ============
# 索引: 2026-04-11 到 2026-07-01 全日历
dates = pd.date_range("2026-04-11", "2026-07-01", freq="D")
ev = pd.DataFrame({"date": dates})
ev["kr_holiday"] = ev["date"].isin(kh["date"]).astype(int)
# 周末 KRX 本就不开 (韩国是 Sat=5, Sun=6)
ev["kr_weekend"] = (ev["date"].dt.dayofweek >= 5).astype(int)
# 韩国"完全不开盘" = weekend | holiday
ev["kr_closed"] = ((ev["kr_holiday"] == 1) | (ev["kr_weekend"] == 1)).astype(int)
# 美国 (纽交所) 周末
ev["us_weekend"] = (ev["date"].dt.dayofweek >= 5).astype(int)
# 美国窗口内节假日: 2026-05-25 Memorial Day (US), 2026-06-19 Juneteenth (Fri), 2026-07-03 (Fri obs of July 4)
us_holidays_us = pd.to_datetime(["2026-05-25", "2026-06-19"])  # 窗口内
ev["us_holiday"] = ev["date"].isin(us_holidays_us).astype(int)
# 财报
ev["mu_earnings"]  = ev["date"].isin(earnings[earnings.symbol=="MU"]["date"]).astype(int)
ev["sndk_earnings"] = ev["date"].isin(earnings[earnings.symbol=="SNDK"]["date"]).astype(int)
# IV 候选: 韩国休市但美国开盘的日子 (最强 IV)
ev["kr_closed_us_open"] = ((ev["kr_closed"] == 1) & (ev["us_weekend"] == 0) & (ev["us_holiday"] == 0)).astype(int)
# 韩国开盘美国休市 (反向 IV)
ev["us_closed_kr_open"] = (((ev["us_weekend"] == 1) | (ev["us_holiday"] == 1)) & (ev["kr_closed"] == 0)).astype(int)

ev.to_parquet(f"{OUT}/events_daily.parquet", index=False)
print(f"\n[OK] Events daily: {len(ev)} rows -> {OUT}/events_daily.parquet")
print(f"KR closed & US open (IV=1 days): {ev['kr_closed_us_open'].sum()}")
print(f"US closed & KR open: {ev['us_closed_kr_open'].sum()}")
print(f"MU earnings days: {ev['mu_earnings'].sum()}")
print(f"SNDK earnings days: {ev['sndk_earnings'].sum()}")

# 打印 IV=1 的具体日子
print("\n韩国休市但美国开盘的日子 (IV 候选):")
print(ev[ev["kr_closed_us_open"] == 1][["date"]].to_string(index=False))
