# 01 完整因果推断数学框架

本文档给出本项目所有因果参数的**形式化定义**、**识别假设**、**估计器**。

---

## 一、符号约定

| 符号 | 含义 | 频率 | 单位 |
|---|---|---|---|
| \(h\) | 小时索引 | 1 to 1848 | 1h |
| \(t\) | 交易日索引 | 1 to 55 | 1d |
| \(s\) | 时段类别 | {OPEN, MID, CLOSE} | 分类 |
| \(D_h\) | 处理: 韩国半导体加权收益率 | 1h | 对数收益率 |
| \(Y_h\) | 结果: Bitget MUUSDT 下一小时收益率 | 1h | 对数收益率 |
| \(Z_h\) | 工具变量三维向量 | 1h | 混合 |
| \(M_h\) | 中介: EWY 或 KORU 盘前收益率 | 1h | 对数收益率 |
| \(\mathbf{X}_h\) | 控制变量向量 | 1h/1d | 混合 |
| \(\eta_h\) | 韩国本地冲击残差 | 1h | 对数收益率 |
| \(U\) | 未观察 confounder | — | — |

**处理变量精确定义**：
\[
D_h = \Delta \log P^{KR\_semi}_h, \quad P^{KR\_semi}_h = w_{SS} \cdot P^{005930}_h + w_{SK} \cdot P^{000660}_h
\]
其中 \(w_{SS} = 0.615, w_{SK} = 0.385\)（窗口起始日市值权重）。

**结果变量**：
\[
Y_h = \log P^{MUUSDT}_{h+1} - \log P^{MUUSDT}_h
\]

**工具变量构造（关键）**：
\[
r^{KR}_h = \alpha_s + \beta_s \cdot r^{SOX}_{t-1} + \eta_h, \quad s \in \{\text{OPEN, MID, CLOSE}\}
\]
\[
\mathbf{Z}_h = (\eta_h \cdot \mathbb{1}[s=\text{OPEN}],\ \eta_h \cdot \mathbb{1}[s=\text{MID}],\ \eta_h \cdot \mathbb{1}[s=\text{CLOSE}])^\top
\]

**控制变量集**：
\[
\mathbf{X}_h = (r^{VIX}_{t-1},\ r^{JPY}_h,\ r^{SOX}_{t-1},\ Y_{h-1})^\top
\]

---

## 二、三个因果目标

### 2.1 ATE — 平均因果效应（Backdoor 识别）

**Potential outcome**（潜在结果）: \(Y_h(d)\) 表示"如果强制 \(D_h = d\)"下的 MU 收益率。

**因果参数**：
\[
\tau^{ATE} = \mathbb{E}[Y_h(d+\Delta) - Y_h(d)]
\]

**识别假设**（Pearl 2009；Causalinference-cn Ch. 3）:
\[
Y_h(d) \perp\!\!\!\perp D_h \mid \mathbf{X}_h \quad \text{(conditional ignorability)}
\]
\[
0 < P(D_h = d \mid \mathbf{X}_h) < 1 \quad \text{(overlap / positivity)}
\]

**识别公式**：
\[
\tau^{ATE} = \mathbb{E}_\mathbf{X}\left[\mathbb{E}[Y \mid D = d+\Delta, \mathbf{X}] - \mathbb{E}[Y \mid D = d, \mathbf{X}]\right]
\]

**估计器 — 因连续 D 用 partialling-out（Frisch-Waugh-Lovell）**:
\[
\tilde{Y}_h = Y_h - \hat{\mathbb{E}}[Y_h \mid \mathbf{X}_h], \quad \tilde{D}_h = D_h - \hat{\mathbb{E}}[D_h \mid \mathbf{X}_h]
\]
\[
\hat{\tau}^{ATE} = \frac{\sum_h \tilde{Y}_h \tilde{D}_h}{\sum_h \tilde{D}_h^2}
\]

---

### 2.2 LATE / CACE — 依从者平均因果效应（IV 识别）

**Compliance types**（Angrist, Imbens & Rubin 1996；Causalinference-cn Ch. 6）:

| Type | \(D(z=0)\) | \(D(z=1)\) |
|---|---|---|
| Complier | 0 | 1 |
| Always-taker | 1 | 1 |
| Never-taker | 0 | 0 |
| Defier（假设排除） | 1 | 0 |

**LATE / CACE 定义**：
\[
\tau^{LATE} = \mathbb{E}[Y_h(1) - Y_h(0) \mid \text{Complier}]
\]

**IV 识别假设（4条 + 我们的连续 D 推广）**：

**(A1) Relevance**（相关性）:
\[
\text{Cov}(Z_h, D_h) \neq 0
\]
可检验，通过 Kleibergen-Paap F 统计量。

**(A2) Exclusion restriction**（排他性）:
\[
Y_h(d, z) = Y_h(d), \quad \forall z
\]
不可直接检验（是识别假设），但过度识别情形下可用 Hansen J-test 检验模型内一致性。

**(A3) Independence**（独立性 / 无 unmeasured Z-Y confounder）:
\[
Z_h \perp\!\!\!\perp \{Y_h(d), D_h(z)\} \mid \mathbf{X}_h
\]

**(A4) Monotonicity**（单调性）:
\[
D_h(z_1) \geq D_h(z_2) \quad \text{a.s.}, \quad \forall z_1 > z_2
\]
连续 D 版本: first-stage 系数符号在所有子样本一致。

**Wald 估计量（一维 IV）**：
\[
\hat{\tau}^{LATE} = \frac{\text{Cov}(Y_h, Z_h)}{\text{Cov}(D_h, Z_h)}
\]

**2SLS 估计器（多维 Z，一维 D）**：
\[
\text{First stage: } D_h = \pi_0 + \boldsymbol{\pi}^\top \mathbf{Z}_h + \boldsymbol{\gamma}^\top \mathbf{X}_h + \nu_h
\]
\[
\text{Second stage: } Y_h = \tau^{LATE} \cdot \hat{D}_h + \boldsymbol{\delta}^\top \mathbf{X}_h + \varepsilon_h
\]

**弱 IV 诊断 — Kleibergen-Paap rk Wald F**:
\[
F^{KP} = \frac{(n-k)}{L} \cdot \hat{\boldsymbol{\pi}}^\top \hat{\mathbf{V}}_{\boldsymbol{\pi}}^{-1} \hat{\boldsymbol{\pi}}
\]
经验阈值：
- \(F^{KP} \geq 10\): Stock-Yogo 传统弱 IV 边界
- \(F^{KP} \geq 23.1\): Lee et al. (2022) 新标准，保证 t-test size distortion < 5%

**弱-IV 稳健推断 — Anderson-Rubin (AR) CI**:
\[
AR(\tau_0) = n \cdot \frac{[\mathbf{Z}^\top(\mathbf{Y} - \mathbf{D}\tau_0)]^\top (\mathbf{Z}^\top\mathbf{Z})^{-1} [\mathbf{Z}^\top(\mathbf{Y} - \mathbf{D}\tau_0)]}{\hat{\sigma}^2(\tau_0)}
\]
95% CI 通过反演 \(\{\tau_0 : AR(\tau_0) \leq \chi^2_{L, 0.95}\}\) 得到。**即使 F 很小，AR CI 仍保 size**。

**过度识别检验 — Hansen J**（df = L − dim(D) = 3 − 1 = 2）:
\[
J = n \cdot \bar{g}(\hat{\tau})^\top \hat{\mathbf{S}}^{-1} \bar{g}(\hat{\tau}) \sim \chi^2_{L-1}
\]

---

### 2.3 Mediation — 自然直接/间接效应（Pearl 2001; VanderWeele 2015）

**反事实定义**：\(Y_h(d, M_h(d'))\) = "把 D 设为 d，同时 M 设为它在 D=d' 时该有的值"。

**分解**：
\[
\tau^{TE} = \mathbb{E}[Y_h(1) - Y_h(0)] = \underbrace{\tau^{NDE}}_{\text{Natural Direct Effect}} + \underbrace{\tau^{NIE}}_{\text{Natural Indirect Effect}}
\]

\[
\begin{aligned}
\tau^{NDE} &= \mathbb{E}[Y_h(1, M_h(0)) - Y_h(0, M_h(0))] \\
\tau^{NIE} &= \mathbb{E}[Y_h(1, M_h(1)) - Y_h(1, M_h(0))]
\end{aligned}
\]

**Sequential ignorability 假设**（Imai, Keele & Yamamoto 2010）:
\[
\{Y_h(d, m), M_h(d')\} \perp\!\!\!\perp D_h \mid \mathbf{X}_h
\]
\[
Y_h(d, m) \perp\!\!\!\perp M_h \mid D_h, \mathbf{X}_h
\]

**线性模型下的 Baron-Kenny 因果化版本**：
\[
\begin{aligned}
M_h &= a_0 + a \cdot D_h + \boldsymbol{\gamma}_M^\top \mathbf{X}_h + u_h \\
Y_h &= b_0 + c' \cdot D_h + b \cdot M_h + \boldsymbol{\gamma}_Y^\top \mathbf{X}_h + v_h
\end{aligned}
\]
\[
\hat{\tau}^{NDE} = c', \quad \hat{\tau}^{NIE} = a \cdot b, \quad \text{Prop.mediated} = \frac{ab}{ab + c'}
\]

**Sobel SE 与 Bootstrap CI**：
\[
SE_{Sobel}(ab) = \sqrt{b^2 \cdot SE_a^2 + a^2 \cdot SE_b^2}
\]
Bootstrap CI 更稳健（不依赖 ab 的正态近似）。

---

## 三、Sensitivity 框架

### 3.1 E-value（VanderWeele & Ding 2017）

**问题**：需要一个 unmeasured confounder \(U\) 与 D 和 Y 的关联度多强，才能"解释掉"当前估计的 τ？

对连续 outcome:
\[
\text{RR} \approx \exp(0.91 \cdot \hat{\tau} / \sigma_Y)
\]
\[
\text{E-value} = \text{RR} + \sqrt{\text{RR}(\text{RR}-1)}
\]

**解读**：E-value = 5 意味着必须存在一个 unmeasured U，同时与 D 关联 RR ≥ 5、与 Y 关联 RR ≥ 5，才能推翻估计的显著性。E-value 越大，结果越 robust。

### 3.2 Rosenbaum Γ-sensitivity（Rosenbaum 2002）

设 \(P(D=1 \mid \mathbf{X}, U=1) / P(D=1 \mid \mathbf{X}, U=0) \leq \Gamma\)。对每个 Γ 计算 worst-case p-value。**Break point** = 使 p > 0.05 的最小 Γ。

### 3.3 Manski Bounds（Manski 1990）

不做任何 identification 假设，只用 Y 的支撑范围：
\[
\tau^{Manski\_upper} = \mathbb{E}[Y \mid D=1] \cdot P(D=1) + Y_{\max} \cdot P(D=0) - \left(\mathbb{E}[Y \mid D=0] \cdot P(D=0) + Y_{\min} \cdot P(D=1)\right)
\]
\[
\tau^{Manski\_lower} = \mathbb{E}[Y \mid D=1] \cdot P(D=1) + Y_{\min} \cdot P(D=0) - \left(\mathbb{E}[Y \mid D=0] \cdot P(D=0) + Y_{\max} \cdot P(D=1)\right)
\]

### 3.4 Placebo test

用 \(D_{h+1}\)（**未来**的 D）代替 \(D_h\) 预测 \(Y_h\)。如果因果方向真是 D → Y，placebo τ 应该 ≈ 0。若 placebo τ 大且显著，说明**反向因果或共同趋势**。

---

## 四、条件独立测试（Fisher-Z）

对 partial correlation \(r_{XY|\mathbf{Z}}\)（把 X, Y 分别对 Z 回归后残差的相关性）:
\[
z = \frac{1}{2}\log\frac{1+r}{1-r} \cdot \sqrt{n - |\mathbf{Z}| - 3} \sim \mathcal{N}(0, 1)
\]
用于 PC 算法（Peter Spirtes-Clark）的每一步条件独立判定。

---

## 五、Newey-West HAC 稳健标准误

对小时级金融数据的自相关 + 异方差：
\[
\hat{\mathbf{V}}^{NW} = \hat{\mathbf{V}}^{OLS} + \sum_{\ell=1}^{L} \omega_\ell \left(\hat{\mathbf{S}}_\ell + \hat{\mathbf{S}}_\ell^\top\right)
\]
\[
\omega_\ell = 1 - \frac{\ell}{L+1}, \quad L = 6 \text{ (本项目)}
\]
