# FinRobot 项目分析文档

分析日期：2026-06-06

本地路径：`/Users/yitwah/Projects/FinRobot`

GitHub：https://github.com/AI4Finance-Foundation/FinRobot

当前 checkout：

- 分支：`master`
- HEAD：`6a8161f 2026-05-11 gcloud deploy setup`
- 子模块：`FinNLP` 指向 `587f04f473507ddea6453e43796797fce17155ce`，当前未初始化为完整内容

## 一句话结论

FinRobot 不是单一 trading bot，而是两个东西叠在一个仓库里：

1. `finrobot/`：基于 Microsoft AutoGen 的通用金融 agent 框架，提供角色库、工具注册、RAG、数据源工具和多 agent 工作流。
2. `finrobot_equity/`：较新的股票研究报告生成产品，围绕 Financial Modeling Prep、OpenAI、图表、HTML/PDF 渲染和 FastAPI Web UI，形成端到端 equity research pipeline。

如果你要“研究它怎么工作”，优先看 `finrobot_equity/`，因为它是目前更完整、可运行、产品化的部分；如果你要学习“金融 agent 框架怎么搭”，看 `finrobot/agents/workflow.py` 和 `finrobot/toolkits.py`。

## 顶层结构

```text
FinRobot/
├── finrobot/                 # 通用 AutoGen 金融 agent 框架
│   ├── agents/               # agent 配置库和 workflow
│   ├── data_source/          # Finnhub/FMP/SEC/YFinance/Reddit/filings 等数据工具
│   ├── functional/           # 报告分析、图表、RAG、代码工具、量化工具
│   ├── toolkits.py           # 把 Python 函数注册成 AutoGen tools
│   └── utils.py              # 配置、日期、key 注册等工具
├── finrobot_equity/          # 股票研究报告生成系统
│   ├── core/
│   │   ├── config/           # config.ini.example
│   │   ├── src/              # CLI 入口和核心模块
│   │   ├── output/           # 示例报告输出
│   │   └── tests/            # 模块/报告生成测试
│   └── web_app/              # FastAPI Web UI
├── tutorials_beginner/       # Notebook 教程
├── tutorials_advanced/       # 进阶 Notebook
├── experiments/              # 组合投资、多因子、portfolio optimization 实验
├── deploy.sh                 # Web app 部署脚本
├── run_web_app.py            # FastAPI 启动入口
├── requirements.txt          # 老 finrobot 框架依赖
└── requirements-equity.txt   # equity research 系统依赖
```

## 运行方式

### Web UI

项目 README 推荐：

```bash
cd /Users/yitwah/Projects/FinRobot
cp finrobot_equity/core/config/config.ini.example finrobot_equity/core/config/config.ini
# 编辑 config.ini，填入 fmp_api_key 和 openai_api_key
chmod +x deploy.sh
./deploy.sh start
```

默认访问：

```text
http://127.0.0.1:8001
```

`deploy.sh` 会创建 `venv/`，安装 `requirements-equity.txt`，再调用 `run_web_app.py` 启动 `finrobot_equity.web_app.main:app`。

### CLI 方式

第一步：生成财务分析数据、预测、新闻、文本段落。

```bash
cd /Users/yitwah/Projects/FinRobot/finrobot_equity/core/src
python generate_financial_analysis.py \
  --company-ticker NVDA \
  --company-name "NVIDIA Corporation" \
  --config-file ../config/config.ini \
  --peer-tickers AMD INTC \
  --generate-text-sections
```

第二步：用第一步产出的 CSV/TXT/JSON 渲染 HTML report。

```bash
python create_equity_report.py \
  --company-ticker NVDA \
  --company-name "NVIDIA Corporation" \
  --analysis-csv output/NVDA/analysis/financial_metrics_and_forecasts.csv \
  --ratios-csv output/NVDA/analysis/ratios_raw_data.csv \
  --tagline-file output/NVDA/analysis/tagline.txt \
  --company-overview-file output/NVDA/analysis/company_overview.txt \
  --investment-overview-file output/NVDA/analysis/investment_overview.txt \
  --valuation-overview-file output/NVDA/analysis/valuation_overview.txt \
  --risks-file output/NVDA/analysis/risks.txt \
  --competitor-analysis-file output/NVDA/analysis/competitor_analysis.txt \
  --major-takeaways-file output/NVDA/analysis/major_takeaways.txt \
  --config-file ../config/config.ini
```

可选第三步：生成 PDF。

```bash
python generate_pdf_report.py \
  --company-ticker NVDA \
  --company-name "NVIDIA Corporation" \
  --analysis-dir output/NVDA/analysis \
  --output-dir output/NVDA/report \
  --config-file ../config/config.ini
```

## API Key 和外部服务

配置文件：`finrobot_equity/core/config/config.ini.example`

需要：

- `fmp_api_key`：Financial Modeling Prep，负责财报、财务指标、公司 profile、股价、新闻、peer data。
- `openai_api_key`：OpenAI 或兼容 OpenAI API 的 provider，负责文本段落生成。

可选：

- `openai_base_url`：用于代理或兼容服务。
- `openai_model`：默认模板里写的是 `gpt-4.1`。
- `adanos_api_key` / `adanos_base_url`：retail sentiment insights。

没有这些 key 时，完整 pipeline 不能真实跑完。代码里有部分 fallback 文本，但财务数据主流程依赖 FMP。

## `finrobot/` 通用 Agent 框架

### 核心思想

`finrobot/` 基于 AutoGen：

- 用 `AssistantAgent` 表示金融角色。
- 用 `UserProxyAgent` 执行工具/代码。
- 用 `register_function` 把数据源函数暴露给 agent。
- 用 GroupChat/MultiAssistant 实现多角色协作。

核心文件：

- `finrobot/agents/agent_library.py`
- `finrobot/agents/workflow.py`
- `finrobot/toolkits.py`
- `finrobot/data_source/*.py`
- `finrobot/functional/*.py`

### Agent 配置库

`agent_library.py` 定义了一组 agent profile，例如：

- `Market_Analyst`
- `Expert_Investor`
- `Financial_Analyst`
- `Data_Analyst`
- `Programmer`

每个 agent 可以绑定 toolkits。比如 `Market_Analyst` 绑定：

- `FinnHubUtils.get_company_profile`
- `FinnHubUtils.get_company_news`
- `FinnHubUtils.get_basic_financials`
- `YFinanceUtils.get_stock_data`

`Expert_Investor` 绑定：

- `FMPUtils.get_sec_report`
- `IPythonUtils.display_image`
- `TextUtils.check_text_length`
- `ReportLabUtils.build_annual_report`
- `ReportAnalysisUtils`
- `ReportChartUtils`

### Workflow 类

`finrobot/agents/workflow.py` 里主要类：

- `FinRobot`：继承 AutoGen `AssistantAgent`，从 agent config/library 生成 agent，并注册工具。
- `SingleAssistant`：单个 assistant + user proxy 的最小对话流程。
- `SingleAssistantRAG`：在 `SingleAssistant` 基础上注册 RAG 工具。
- `SingleAssistantShadow`：给主 assistant 增加 shadow assistant，用于嵌套反思/指令处理。
- `MultiAssistant` / `MultiAssistantWithLeader`：多 agent group chat 流程。

工作机制：

1. 用户传入 agent 名称或配置。
2. `FinRobot` 从 library 读取 profile 和 toolkits。
3. `UserProxyAgent` 作为执行者。
4. `register_toolkits` 把 Python 函数注册为 AutoGen 可调用工具。
5. agent 通过对话决定何时调用工具、何时生成结论。

### Toolkit 注册

`finrobot/toolkits.py` 是理解框架的关键。

它做了三件事：

1. `stringify_output`：把工具函数返回值转成字符串，DataFrame 会转成 table string。
2. `register_toolkits`：把函数或类方法注册给 AutoGen。
3. `register_code_writing`：注册文件查看/修改/创建等代码工具。

这说明 `finrobot/` 的 agent 本质上不是“金融模型”，而是“LLM + 已注册金融数据/分析函数 + AutoGen 对话编排”。

## `finrobot_equity/` 股票研究报告系统

这是当前仓库里最完整的产品形态。

### 两阶段 Pipeline

```text
generate_financial_analysis.py
  -> output/[TICKER]/analysis/*.csv|*.json|*.txt
  -> create_equity_report.py
      -> output/[TICKER]/report/*.html
      -> generate_pdf_report.py
          -> output/[TICKER]/report/*.pdf
```

### 第一步：生成分析数据

入口：`finrobot_equity/core/src/generate_financial_analysis.py`

主要做：

1. 读取 `config.ini`。
2. 用 FMP 拉取 income statement、balance sheet、cash flow、ratios、key metrics。
3. 从 API 数据抽取历史指标。
4. 根据默认/命令行假设生成 2025E-2027E 预测。
5. 如果传入 peers，生成 peer EBITDA 和 EV/EBITDA comparison。
6. 拉取 company news。
7. 可选增强：
   - sensitivity analysis
   - catalyst analysis
   - enhanced news
   - retail sentiment
8. 如果开启 `--generate-text-sections`，调用 OpenAI 生成：
   - tagline
   - company_overview
   - investment_overview
   - valuation_overview
   - risks
   - competitor_analysis
   - major_takeaways
   - news_summary
9. 保存 raw financial statements 和 `analysis_summary.json`。

关键模块：

- `market_data_api.py`：FMP/yfinance 数据抓取。
- `financial_data_processor.py`：指标抽取和预测。
- `text_generator_agents.py`：OpenAI 文本段落生成。
- `sensitivity_analyzer.py`：收入/利润率敏感性分析。
- `catalyst_analyzer.py`：从新闻识别催化剂。
- `news_integrator.py`：增强新闻处理。
- `retail_sentiment_client.py`：Adanos retail sentiment。

### 第二步：生成 HTML 报告

入口：`finrobot_equity/core/src/create_equity_report.py`

主要做：

1. 读取第一步输出的 analysis CSV、ratios CSV、peer CSV、文本文件。
2. 自动拉取当前 market metrics，例如 share price、target price、rating、market cap、sector。
3. 校验文本内容，发现 CSV-like text 时会尝试重新生成。
4. 生成图表：
   - revenue/EBITDA
   - EV/EBITDA peer comparison
   - EPS × PE
   - margin trend
   - stock price
   - technical indicators
   - cash flow
   - enhanced chart set
5. 可选 valuation analysis。
6. 把数据塞进 HTML template，生成 professional HTML report。

关键模块：

- `html_renderer.py`
- `html_template_professional.py`
- `chart_generator.py`
- `enhanced_chart_generator.py`
- `valuation_engine.py`
- `report_structure.py`
- `report_data_loader.py`

### 第三步：PDF

入口：`finrobot_equity/core/src/generate_pdf_report.py`

PDF 模块有两套：

- `pdf_generator.py`
- `professional_pdf_report.py`

它们基于 ReportLab/WeasyPrint 生成专业格式报告。

## Web App 如何工作

入口：

- `run_web_app.py`
- `finrobot_equity/web_app/main.py`

技术栈：

- FastAPI
- Jinja2 templates
- SQLAlchemy
- SQLite/local DB 层
- bcrypt password hashing
- 可选 GitHub OAuth

主要页面/API：

- `/`：登录后主页。
- `/login`：登录页。
- `/api/auth/*`：登录、注册、登出、改密码、GitHub OAuth。
- `/api/run`：提交分析任务。
- `/api/status/{task_id}`：查询后台任务状态和日志。
- `/output/...`：静态暴露生成的报告文件。

Web app 的核心不是重新实现分析，而是：

1. 接收用户请求。
2. 创建 `task_id`。
3. 后台执行 `generate_financial_analysis.py`。
4. 执行 `create_equity_report.py`。
5. 可选执行 `generate_pdf_report.py`。
6. 将日志写入 `finrobot_equity/logs/task_{task_id}.log`。
7. 将输出文件位置回传给前端。

也就是说，Web UI 是 CLI pipeline 的 wrapper。

## 数据流图

```text
User / Web UI / CLI
        |
        v
config.ini
        |
        v
FMP API + yfinance + optional Adanos
        |
        v
market_data_api.py
        |
        v
financial_data_processor.py
        |
        v
output/[TICKER]/analysis/
  - financial_metrics_and_forecasts.csv
  - ratios_raw_data.csv
  - income_statement_raw_data.csv
  - company_news.json
  - sensitivity_analysis.json
  - catalyst_analysis.json
  - *.txt text sections
        |
        v
create_equity_report.py
        |
        v
chart_generator.py + html_template_professional.py
        |
        v
output/[TICKER]/report/
  - Professional HTML report
  - optional PDF report
```

## 主要优点

1. 端到端完整：从数据抓取到 HTML/PDF 报告都有。
2. 模块边界清楚：数据、预测、文本、图表、HTML、PDF、Web UI 基本分开。
3. 有示例输出：`finrobot_equity/core/output/` 下已有 5 个 HTML 示例报告。
4. 可用 CLI，也有 Web UI。
5. 对 OpenAI 兼容 API 友好：支持 `openai_base_url` 和 `openai_model`。
6. 报告形态明确：更像 equity research report，而不是抽象聊天 bot。

## 主要问题和风险

### 1. 它不是交易系统

FinRobot 可以生成研究报告和投资分析，但没有可靠的 broker execution、order management、position tracking、risk engine。不能把它直接当自动交易系统。

### 2. 数据源强依赖 FMP

核心财务数据、peer comparison、market metrics、news 大多来自 Financial Modeling Prep。没有 FMP key 或额度不足，pipeline 会失败或退化。

### 3. LLM 文本不是强校验结论

`text_generator_agents.py` 直接把 DataFrame 转成 markdown 放进 prompt，然后让 OpenAI 生成段落。它有 fallback，但没有事实一致性审计、引用检查、数字一致性校验。

### 4. 预测假设较粗

默认预测参数是固定增长率和 margin improvement：

- `revenue_growth_2025 = 0.05`
- `revenue_growth_2026 = 0.06`
- `revenue_growth_2027 = 0.04`
- `margin_improvement = 0.01`

这适合 demo，不适合直接出严肃投资结论。

### 5. 代码风格混合

仓库里同时存在老的 AutoGen framework、新的 equity product、中文注释、Notebook、实验脚本、报告模板、Web app。学习时要分清层次，否则容易迷路。

### 6. 许可证信息不完全一致

`setup.py` 写 MIT，`finrobot_equity/README.md` 写 Apache 2.0，根目录有 `LICENSE`。二次开发或商业使用前应确认仓库当前 license 文件和项目方声明。

## 二次开发建议

### 如果目标是做一个金融研究 agent

优先扩展：

- `finrobot_equity/core/src/modules/market_data_api.py`
- `finrobot_equity/core/src/modules/text_generator_agents.py`
- `finrobot_equity/core/src/modules/html_template_professional.py`
- `finrobot_equity/core/src/create_equity_report.py`

建议加：

- SEC filings / earnings call RAG。
- 数据引用和来源表。
- 文本生成后的数字一致性检查。
- analyst assumptions UI。
- 输出 JSON schema，方便被其他 agent 消费。

### 如果目标是做交易 agent

不要直接在这个仓库里硬接交易 API。更合理是：

1. FinRobot 只负责研究报告和交易建议。
2. 单独建 `risk-engine`，纯规则控制仓位、ticker、订单金额。
3. 单独接 Alpaca/IBKR/Coinbase/交易所。
4. 所有真实订单必须经过人工确认或 hard policy gate。

### 如果目标是 MCP 化

可以拆成几个 MCP tools：

- `generate_equity_analysis(ticker, peers, assumptions)`：调用 `generate_financial_analysis.py`。
- `create_equity_report(ticker, analysis_dir)`：调用 `create_equity_report.py`。
- `get_report_artifact(ticker)`：返回 HTML/PDF 文件。
- `summarize_report(ticker)`：读取生成报告并输出结构化摘要。

不要把 API keys 直接暴露给 LLM。MCP server 应该封装 key，并限制 ticker、输出目录和运行时间。

## 快速入口索引

| 想做什么 | 看哪个文件 |
|---|---|
| 理解通用 agent 框架 | `finrobot/agents/workflow.py` |
| 看 agent 角色库 | `finrobot/agents/agent_library.py` |
| 看工具如何注册给 agent | `finrobot/toolkits.py` |
| 看数据源封装 | `finrobot/data_source/*.py` |
| 看 equity 分析主流程 | `finrobot_equity/core/src/generate_financial_analysis.py` |
| 看 HTML 报告生成 | `finrobot_equity/core/src/create_equity_report.py` |
| 看 FMP/yfinance API | `finrobot_equity/core/src/modules/market_data_api.py` |
| 看财务指标和预测 | `finrobot_equity/core/src/modules/financial_data_processor.py` |
| 看 LLM 文本生成 | `finrobot_equity/core/src/modules/text_generator_agents.py` |
| 看 Web API | `finrobot_equity/web_app/main.py` |
| 看部署脚本 | `deploy.sh` |
| 看配置模板 | `finrobot_equity/core/config/config.ini.example` |

## 我做过的验证

已完成：

- Clone 到 `/Users/yitwah/Projects/FinRobot`。
- 检查当前 git HEAD。
- 枚举项目结构。
- 阅读 README、equity README、核心 workflow、toolkits、analysis/report/web app 入口。
- 用 Python AST parse 验证关键文件语法可解析：
  - `finrobot/agents/workflow.py`
  - `finrobot/toolkits.py`
  - `finrobot_equity/core/src/generate_financial_analysis.py`
  - `finrobot_equity/core/src/create_equity_report.py`
  - `finrobot_equity/web_app/main.py`
- 确认 `finrobot_equity/core/output/` 有 5 个示例 HTML equity reports。

未运行完整 pipeline：

- 原因：需要 FMP/OpenAI/可选 Adanos API key，会消耗外部 API。
- 如果要实际跑，先填 `finrobot_equity/core/config/config.ini`，然后用 CLI 或 `./deploy.sh start`。

## 总体评价

FinRobot 对“金融 agent”学习很有价值，但要用对位置：

- 作为 equity research report generator：值得深入改造。
- 作为 AutoGen 金融 agent 示例：值得参考工具注册和多 agent workflow。
- 作为自动真钱交易 bot：不够，需要单独风控、执行、审计、回测和权限控制。

