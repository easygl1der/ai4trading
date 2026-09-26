# Mermaid Research Maps

This document keeps the valuable diagram ideas while removing the invalid old causal claims.

The diagrams below are research maps, not validated causal conclusions.

## 1. Valid High-Level Idea

```mermaid
flowchart LR
    subgraph Markets["Market clocks"]
        US["US equity market<br/>regular + pre/post"]
        KR["Korea equity market<br/>KOSPI/KOSDAQ hours"]
        BG["Bitget RWA perpetuals<br/>24/7 trading"]
    end

    subgraph Assets["Semiconductor assets"]
        MU["MU / MUUSDT"]
        SNDK["SNDK / SNDKUSDT"]
        SS["Samsung Electronics"]
        SK["SK Hynix"]
        ETF["EWY / KORU"]
    end

    subgraph Common["Common shocks"]
        SOX["SOX / semiconductor beta"]
        VIX["VIX / global risk"]
        FX["USD/KRW, USD/JPY"]
        NEWS["Earnings, AI/HBM, macro news"]
    end

    US --> MU
    BG --> MU
    KR --> SS
    KR --> SK
    US --> ETF

    MU -. "tracking / basis" .- BG
    SS --> ETF
    SK --> ETF

    SOX --> MU
    SOX --> SS
    SOX --> SK
    VIX --> MU
    VIX --> SS
    VIX --> SK
    FX --> SS
    FX --> SK
    NEWS --> MU
    NEWS --> SS
    NEWS --> SK

    style BG fill:#c8e6c9
    style KR fill:#bbdefb
    style US fill:#ffe0b2
    style Common fill:#f5f5f5
```

Value of this map:

- Bitget RWA perpetuals may provide a 24/7 price signal.
- Korea and US market clocks create measurable timing gaps.
- Semiconductor shocks must be separated from global risk and SOX beta.
- Basis/tracking quality must be validated before causal claims.

## 2. Predictive Lead-Lag Map

This is the currently defensible first-stage research object.

```mermaid
flowchart LR
    B0["Bitget MUUSDT<br/>overnight / 24h returns"]
    K0["Korea semi basket<br/>OPEN/MID/CLOSE returns"]
    U["Common shocks<br/>SOX, VIX, FX, news"]
    T["Market clock<br/>US closed / Korea open"]

    U --> B0
    U --> K0
    T --> B0
    T --> K0

    B0 -. "predictive lag test" .-> K0
    K0 -. "predictive lag test" .-> B0

    style B0 fill:#c8e6c9
    style K0 fill:#bbdefb
    style U fill:#ffcdd2
```

Allowed language:

- "Bitget predicts Korea at lag \(k\)."
- "Korea predicts Bitget at lag \(k\)."
- "The predictive relation is stronger in OPEN than MID."

Disallowed language unless a causal design is added:

- "Bitget causes Korea."
- "Korea causes MU."
- "The Granger direction establishes causality."

## 3. Candidate Reverse Causal Design

This is only a candidate. It needs an external shock before estimation.

```mermaid
flowchart LR
    Z["External US/MU shock<br/>earnings surprise, verified news,<br/>exchange-specific liquidity event"]
    D["D_t<br/>Bitget MUUSDT shock<br/>before Korea open"]
    Y["Y_t<br/>Korea semi OPEN return"]
    X["Controls<br/>SOX lag, VIX, FX,<br/>US post-market, events"]
    U["Unobserved semiconductor demand shock"]

    Z --> D
    D --> Y
    X --> D
    X --> Y
    U --> D
    U --> Y

    Z -. "must not directly affect Y<br/>except through D" .-> Y

    style Z fill:#fff9c4
    style D fill:#c8e6c9
    style Y fill:#bbdefb
    style U fill:#ffcdd2
```

Identification warning:

- If \(Z\) is just a transformed version of \(D\), it is invalid.
- If \(Z\) is broad semiconductor news, it likely affects Korean stocks directly and fails exclusion.
- If no valid \(Z\) exists, the study remains predictive-only.

## 4. Correct Mediation Timing

The old mediation was invalid because the mediator was measured after the outcome.

The valid time order must be:

\[
D_h \to M_{h+a} \to Y_{h+a+b}
\]

```mermaid
flowchart LR
    D["D_h<br/>earlier shock"]
    M["M_{h+a}<br/>mediator after D"]
    Y["Y_{h+a+b}<br/>outcome after M"]
    X["Controls before D"]

    X --> D
    X --> M
    X --> Y
    D --> M
    M --> Y
    D --> Y

    style D fill:#c8e6c9
    style M fill:#e1bee7
    style Y fill:#bbdefb
```

Valid candidate paths:

- Korea OPEN shock \(\to\) EWY/KORU pre-market \(\to\) US regular-session MU.
- Bitget overnight MU shock \(\to\) Korea OPEN \(\to\) EWY/KORU pre-market.

Invalid path:

- Korea \(h\) \(\to\) EWY/KORU \(h+8\) \(\to\) Bitget MU \(h+1\).
