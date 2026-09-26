# 09. Reproducibility Guide（复现指南）

> 本文档说明如何从零开始复现本项目所有结果。适合后续 AI 或研究者作为运行手册。

## 1. 环境要求

### 1.1 Python 版本
- Python **3.11+**（推荐 3.11.x）

### 1.2 依赖包
```
numpy>=1.24,<2.0
pandas>=2.0
scipy>=1.11
statsmodels>=0.14
linearmodels>=5.0
scikit-learn>=1.3
matplotlib>=3.7
pyarrow>=13.0
requests>=2.31
```

安装：
```bash
pip install -r project/requirements.txt
```

### 1.3 外部服务

| 服务 | 用途 | 凭证 |
|---|---|---|
| **Bitget REST API** | MUUSDT / SNDKUSDT 小时 K 线 | 无需（公共接口） |
| **Yahoo Finance MCP** | MU, SNDK, EWY, KORU, 韩股, SOX, VIX 数据 | 用 `yahoo_finance_4b64a6bdc06a4a80affb5af12e961bf0` connector |
| **FRED API** | DGS10, DFF, DGS3MO 等宏观数据 | Custom credential: `custom-cred:api.stlouisfed.org`（QueryParam api_key） |

### 1.4 时区设置
所有时间戳统一到 **UTC**。运行前建议：
```bash
export TZ=UTC
```

## 2. 目录结构

```
project/
├── README.md                    # 项目总览
├── docs/                        # 详细文档
│   ├── 01_causal_framework.md   # 因果框架 + Rubin PO
│   ├── 02_dag_mermaid.md        # 大 Mermaid DAG
│   ├── 03_data_pipeline.md      # 数据流程
│   ├── 04_iv_design.md          # 3D IV 设计
│   ├── 05_results_report.md     # 最终结果
│   ├── 06_pitfalls_and_gotchas.md   # 踩坑
│   ├── 07_user_preferences.md   # 用户偏好
│   ├── 08_alternatives_not_taken.md # 替代方案
│   └── 09_reproducibility.md    # 本文
├── data/                        # 数据
│   ├── bitget/                  # Bitget 1h K 线
│   ├── yahoo/                   # Yahoo 1h 现货 + ETF
│   ├── korea/                   # 韩股 1h (Samsung, SK Hynix)
│   ├── etf/                     # EWY, KORU 1h
│   ├── macro/                   # SOX, VIX, USDJPY, KRW
│   ├── fred/                    # FRED daily 宏观
│   ├── meta/                    # 韩国节假日、财报、事件
│   ├── analysis/                # η_hourly, master_clean, results.json
│   ├── manifest.json            # 数据源汇总
│   └── manifest.md              # 人类可读版本
├── scripts/                     # 运行脚本
│   ├── pull_bitget.py           # 拉 Bitget 数据
│   ├── process_yahoo.py         # 处理 Yahoo 数据
│   ├── pull_fred.py             # 拉 FRED
│   ├── build_events.py          # 构造事件表
│   ├── qc_manifest.py           # QC + manifest
│   ├── stage1_segmented_regression.py  # Stage 1 分段回归
│   ├── build_master_table.py    # 合并主表
│   └── stage234_analysis.py     # Stage 2/3/4 + Mediation + Granger
└── figures/                     # Mermaid 渲染图（待生成）
```

## 3. 完整复现流程

### Step 1: 数据采集

```bash
# 1.1 Bitget MUUSDT + SNDKUSDT 1h K 线
python scripts/pull_bitget.py

# 输出: data/bitget/muusdt_1h.parquet (1954 rows)
#      data/bitget/sndkusdt_1h.parquet (1851 rows)
```

```bash
# 1.2 Yahoo 数据（MU, SNDK, EWY, KORU, 005930.KS, 000660.KS, ^SOX, ^VIX, KRW=X, JPY=X）
# 通过 Yahoo Finance MCP，脚本内嵌 tool calls
python scripts/process_yahoo.py

# 输出:
#   data/yahoo/mu_1h.parquet (938 rows)
#   data/yahoo/sndk_1h.parquet (938 rows)
#   data/etf/ewy_1h.parquet (938 rows)
#   data/etf/koru_1h.parquet (938 rows)
#   data/korea/005930_ks_1h.parquet (378 rows)
#   data/korea/000660_ks_1h.parquet (378 rows)
#   data/macro/sox_1h.parquet
#   data/macro/vix_1h.parquet
#   data/macro/krwx_1h.parquet
#   data/macro/usdjpyx_1h.parquet
```

```bash
# 1.3 FRED 宏观 daily
python scripts/pull_fred.py

# 输出: data/fred/fred_daily_wide.parquet (65 days, TEDRATE 已废弃)
# 序列: DGS10, DFF, DGS3MO, DEXKOUS, VIXCLS
```

```bash
# 1.4 韩国节假日 + 财报 + 事件
python scripts/build_events.py

# 输出:
#   data/meta/korea_holidays.parquet
#   data/meta/earnings.parquet
#   data/meta/events_daily.parquet
```

### Step 2: 数据质检

```bash
python scripts/qc_manifest.py

# 输出:
#   data/manifest.json (机器可读)
#   data/manifest.md   (人类可读)
```

关键检查项：
- ✅ 时区统一到 UTC
- ✅ 无 NaN 或已 forward-fill
- ✅ 时间戳对齐（1h 分辨率）
- ✅ 76 个交易日窗口一致

### Step 3: Stage 1 —— η_h 构造

```bash
python scripts/stage1_segmented_regression.py
```

**输入**：
- `data/korea/005930_ks_1h.parquet` (Samsung)
- `data/korea/000660_ks_1h.parquet` (SK Hynix)
- 市值权重：Samsung 0.615，SK Hynix 0.385

**输出**：
- `data/analysis/eta_hourly.parquet` (n=306)

**关键公式**：
$$
\eta_h = 0.615 \cdot r^{Samsung}_h + 0.385 \cdot r^{SKHynix}_h
$$

在美国 session 用分段回归识别 β_OPEN, β_MID, β_CLOSE：
- β_OPEN = 0.485 (R²=22%)
- β_MID ≈ 0
- β_CLOSE = -0.078
- Wald χ² = 10.02, p = 0.0067

### Step 4: 主表构建

```bash
python scripts/build_master_table.py
```

**输入**：
- `data/analysis/eta_hourly.parquet`
- `data/bitget/muusdt_1h.parquet` （构造 Y_h）
- `data/etf/ewy_1h.parquet`, `data/etf/koru_1h.parquet` （中介 M_h）
- 所有宏观 covariates
- `data/meta/events_daily.parquet`

**输出**：
- `data/analysis/analysis_master_clean.parquet` (n=302)

**关键变量**：
| 变量 | 定义 |
|---|---|
| `Y_h` | Bitget MUUSDT log return, hour h → h+1 |
| `D_h` | Korea overnight signal (proxy) |
| `eta_h` | Korean semiconductor weighted return |
| `M_EWY_h` | EWY pre-market return |
| `M_KORU_h` | KORU pre-market return |
| `I_OPEN, I_MID, I_CLOSE` | US session dummies |
| `X_h` | Covariates: VIX, SOX, USDJPY, DGS10 |

### Step 5: 因果推断分析

```bash
python scripts/stage234_analysis.py
```

**输出**：
- `data/analysis/all_results.json`

**运行的分析**（按顺序）：

1. **Stage 2: ATE via backdoor**
   - IPW (Inverse Propensity Weighting)
   - DR-AIPW (Doubly Robust)
   - Partialling-out (Frisch-Waugh-Lovell)
   - **τ_ATE = +0.018, p = 0.61**

2. **Stage 3: CACE via 3D IV**
   - Instruments: η × [I_OPEN, I_MID, I_CLOSE]
   - 方案 A (I_OPEN only): τ = -0.314, KP-F = 4.57 ❌ 弱 IV
   - 方案 B (I_OPEN, I_CLOSE): τ = +0.010, KP-F = 113.88 ✅
   - 方案 C (三段联合): **τ = +0.018, KP-F = 951.97** ✅
   - AR 95% CI: [-0.049, +0.083]
   - Hansen J: NaN（linearmodels API 问题，需手动）

3. **Stage 3.5: Mediation NDE/NIE (EWY)**
   - a-path: +0.129, b-path: +0.164
   - NIE = +0.021 (bootstrap CI [+0.001, +0.049])
   - NDE = -0.006

4. **Stage 3.5: Mediation NDE/NIE (KORU)**
   - a-path: +0.450, b-path: +0.053
   - NIE = +0.024 (bootstrap CI [+0.002, +0.051])
   - NDE = -0.009

5. **Stage 4: Sensitivity**
   - E-value = 11.12
   - Manski bounds: [-0.047, +0.050]
   - Placebo test (D_{h+1} → Y_h): τ = +0.219, p < 0.0001 **⚠️ 反向因果**
   - Rosenbaum Γ: (未做，见 docs/08)

6. **Granger + PC**
   - Y → D: F = 93.80, p < 0.0001 ✅
   - D → Y: F = 1.09, p = 0.30 ❌
   - **结论：MU 领先 Korea，方向反转**

### Step 6: 输出与呈现

```bash
# 生成 Mermaid 渲染图（可选）
# 需要 mermaid-cli
npx @mermaid-js/mermaid-cli -i docs/02_dag_mermaid.md -o figures/dag.png
```

## 4. 关键命令一键跑

```bash
cd /home/user/workspace/project

# 数据 pipeline
python scripts/pull_bitget.py
python scripts/process_yahoo.py
python scripts/pull_fred.py
python scripts/build_events.py
python scripts/qc_manifest.py

# 分析 pipeline
python scripts/stage1_segmented_regression.py
python scripts/build_master_table.py
python scripts/stage234_analysis.py

# 查看结果
cat data/analysis/all_results.json | python -m json.tool
```

## 5. 复现时可能遇到的问题

### 5.1 Bitget API 数据回填

Bitget 会在几小时/几天内修正历史 K 线。如果重新拉取，早期时段的数据可能微调。**保存 parquet 快照**是保证复现的关键。

### 5.2 Yahoo `scale=3`

MU/SNDK 现价 > $100，Yahoo 返回 `priceHint=3`。**用对数收益率**免疫。

### 5.3 SSL 证书

在 Perplexity Sandbox 内运行 FRED API 需要：
```bash
curl --cacert /etc/ssl/certs/agent-proxy-ca-2.pem ...
```

### 5.4 时区

**必须**统一到 UTC，否则 session dummies (`I_OPEN`) 会错位。

### 5.5 依赖版本冲突

`linearmodels` 5.x 与旧代码不兼容。见 `docs/06_pitfalls_and_gotchas.md` 第 1 节。

## 6. 数据字典

参见 `data/manifest.md`。

## 7. 版本信息

- **数据快照**：2026-07-01
- **窗口**：2026-04-11 to 2026-07-01 (76 天)
- **文档创建**：2026-07-01
- **主要贡献者**：本项目由 Perplexity Computer 主导执行，用户（SYSU 数学系）作为 causal-inference 学习者与最终决策者

## 8. 联系与后续

- 项目 Space: `causal inference` (j4KXMsjIQfC9Lx5ok2B7XA)
- 用户教材：*Causalinference-cn*
- 反馈渠道：与用户对话 in Perplexity Space
- 建议扩展方向：见 `docs/08_alternatives_not_taken.md`

