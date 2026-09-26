# 金融与交易 Agent / MCP 选型调研

调研日期：2026-06-06

结论先说：现在已经有不少“金融 agent”项目，但真正可以放心接真钱自动交易的很少。比较成熟的落地方向不是让一个 LLM 直接下单，而是把系统拆成四层：

1. 数据与新闻层：OpenBB、Alpaca、Yahoo Finance、CoinGecko、Polygon、交易所 API。
2. 研究与决策层：TradingAgents、FinRobot、OpenAlice、自己写的多 agent 流程。
3. 执行层：Alpaca paper/live trading、Coinbase AgentKit/onchain wallet、交易所 SDK。
4. 风控与审计层：硬规则、限额、paper trading、人工确认、回测、日志、告警。

如果目标是“市场行情 + 新闻理解 + 策略讨论 + 生成交易建议”，现在已经可以搭得比较好。如果目标是“长期自动真钱交易”，建议先做 paper trading 和半自动执行，不建议第一版直接让 agent 拥有无限下单权限。

## 推荐优先级

| 优先级 | 方案 | 适合做什么 | 不适合做什么 |
|---|---|---|---|
| A | OpenBB + TradingAgents/FinRobot + Alpaca MCP | 股票/ETF 研究、报告、paper trading、半自动交易 | 低延迟高频交易 |
| A | Coinbase AgentKit + 自定义 crypto research agent | onchain 钱包、DeFi 操作、稳定币支付、链上 agent | 无风控的真钱自动操作 |
| B | OpenAlice | 一体化 trading agent 原型，覆盖股票、crypto、外汇、商品等 | 直接托管大资金 |
| B | FinRL / 自研回测层 | 强化学习、量化策略实验 | LLM 新闻理解和解释 |
| C | Yahoo Finance MCP 等轻量 MCP | 快速给 LLM 接行情、新闻、财报 | 机构级数据质量或执行 |
| A-企业 | Bloomberg / AlphaSense / FactSet / Kensho / Claude for Financial Services | 机构研究、文档/新闻/RAG、合规环境 | 开源可控、低成本个人交易 bot |

## 候选清单

### 1. TradingAgents

URL: https://github.com/TauricResearch/TradingAgents

定位：多 agent LLM 金融交易研究框架。它把交易流程拆成基本面分析、情绪分析、新闻分析、技术分析、多空研究员辩论、交易员、风险管理、投资组合经理等角色。

强项：

- 设计上贴近“投研团队”流程，不是一个 prompt 直接输出买卖。
- GitHub 活跃度很高。2026-06-06 查询时约 83k stars，最近 push 在 2026-06-01。
- README 明确有 2026 年多次版本更新，支持多 LLM provider、Docker、结构化输出、checkpoint、非美市场 ticker。
- 适合做“研究报告 + 交易建议 + 决策日志”的核心引擎。

接入：

- LLM provider：OpenAI、Anthropic、Google、xAI、DeepSeek、Qwen、GLM、MiniMax、OpenRouter、Ollama 等。
- 数据：README 提到 Yahoo Finance 覆盖的市场 ticker、Alpha Vantage 等。
- 执行：项目文档强调模拟交易/研究用途，不应直接等同于可真钱自动交易系统。

风险：

- 官方说明偏研究用途，并明确不是投资建议。
- LLM 输出有非确定性，回测结果和实盘差异可能很大。
- 要真钱交易必须另加风控、仓位、人工确认、broker 执行层。

适合度：研究/投研 agent 很适合；真钱交易只能作为决策辅助，不应裸接执行。

### 2. FinRobot

URL: https://github.com/AI4Finance-Foundation/FinRobot

定位：AI4Finance 的开源金融 AI agent 平台，偏“金融分析、研究报告、量化分析、风险评估”。它不是单一 trading bot，而是一个金融 agent 生态。

强项：

- 有白皮书和金融 agent 架构，结合 LLM、强化学习、量化分析。
- FinRobot Pro 方向偏 equity research assistant，可以自动抓财务数据、做预测、DCF、peer comparison、生成 HTML/PDF 报告。
- GitHub 活跃。2026-06-06 查询时约 7.2k stars，最近 push 在 2026-05-10。
- 数据源覆盖 FMP、SEC、Finnhub、yfinance 等。

接入：

- FMP API、OpenAI API、SEC、yfinance、Finnhub 等。
- 有 Web interface 和 CLI 示例。
- 有专业 equity report 生成流程。

风险：

- 更适合投研和报告，不是完整 execution system。
- 直接交易功能需要自己补 broker/exchange 连接、风控和审计。

适合度：做股票研究 agent 很适合；做自动交易需要和 Alpaca/IBKR/交易所执行层组合。

### 3. OpenAlice

URL: https://github.com/TraderAlice/OpenAlice

定位：开源 AI trading agent，项目描述是覆盖 equities、crypto、commodities、forex、macro，从 research 到 position entry、management、exit。

强项：

- 产品方向最接近“一个人版华尔街交易员”。
- GitHub 活跃。2026-06-06 查询时约 4.9k stars，最近 push 在 2026-06-05。
- 覆盖多资产类别，不只是股票或 crypto。

风险：

- 相比 TradingAgents/OpenBB/FinRobot，生态和长期验证还需要观察。
- 这类“一体化自动交易 agent”最容易出现过度授权、幻觉下单、风控薄弱问题。
- AGPL-3.0 许可证会影响商业闭源使用。

适合度：非常适合研究一体化交易 agent 产品形态；真钱交易前需要严格审计、paper trading 和权限隔离。

### 4. Alpaca MCP Server

URL: https://github.com/alpacahq/alpaca-mcp-server

定位：Alpaca 官方 MCP server，让 Claude/Cursor/VS Code 等 AI assistant 通过自然语言使用 Alpaca Trading API。

强项：

- 官方 broker/trading API MCP，不是第三方玩具。
- 支持股票、ETF、crypto、options、portfolio management、real-time market data。
- 支持 paper trading，适合先做安全验证。
- 2026-06-06 查询时约 800 stars，最近 push 在 2026-06-05。
- README 提到 v2 用 FastMCP 和 OpenAPI 重写，并支持通过 `ALPACA_TOOLSETS` 限制工具集。

风险：

- 这是执行和数据工具，不是完整投研脑。
- 一旦配置 live trading key，风险来自 LLM 调用工具的权限边界。
- 需要 server-side 风控：白名单 ticker、最大订单金额、最大日亏损、人工确认、只允许 paper trading 等。

适合度：股票/ETF/期权/crypto paper trading 执行层很适合；真钱交易必须强限权。

### 5. OpenBB / OpenBB Workspace MCP

URLs:

- https://github.com/OpenBB-finance/OpenBB
- https://github.com/OpenBB-finance/workspace-mcp

定位：OpenBB 是金融数据平台，README 称其服务对象包括 analysts、quants、AI agents。OpenBB Workspace MCP 可以让 agent 连接到当前 OpenBB Workspace 浏览器会话，读写 dashboard、widgets、apps、backends。

强项：

- 数据平台成熟度高。2026-06-06 查询时 OpenBB 约 68k stars，最近 push 在 2026-06-06。
- “connect once, consume everywhere”：Python、Workspace、Excel、MCP、REST API。
- 适合做数据接入和研究 dashboard，而不是重新造行情数据层。
- Workspace MCP 明确可让 agent inspect/update Workspace session、widgets、dashboards。

风险：

- OpenBB README 有明确免责声明：金融工具交易有高风险，数据不一定准确。
- Workspace MCP sidecar 明确只建议本地运行，因为连接的 MCP client 可以读写 active Workspace session。
- 不是交易执行层。

适合度：金融 agent 的数据/可视化/研究工作台层非常适合。

### 6. Coinbase AgentKit

URL: https://github.com/coinbase/agentkit

定位：Coinbase Developer Platform 的 agent wallet/onchain interaction toolkit。README 的核心口号是“Every agent deserves a wallet”。

强项：

- 官方 crypto/onchain agent 工具包。
- 支持 TypeScript 和 Python。
- 支持 LangChain、Vercel AI SDK、AutoGen、OpenAI Agents SDK、Pydantic AI、Strands Agents 等。
- README 显示 action providers 很多，TypeScript 50+ actions，Python 30+ actions，并有 Model Context Protocol extension。
- 2026-06-06 查询时约 1.2k stars，最近 push 在 2026-06-04。

接入：

- CDP wallet、Privy、viem 等 wallet providers。
- Base Sepolia 等测试网示例。
- 可创建 onchain-agent，全栈或 chatbot。

风险：

- 有钱包能力就等于有真实资产操作风险。
- 更偏 onchain action/payment/DeFi 操作，不是完整市场行情分析系统。
- 需要独立补价格、新闻、风险模型和交易策略验证。

适合度：crypto/onchain agent 的执行和钱包层很适合；不应作为独立 alpha engine。

### 7. elizaOS

URL: https://github.com/elizaOS/eliza

定位：开源 agentic operating system，crypto 圈常用来做社交/社区/onchain agent。

强项：

- 社区大，2026-06-06 查询时约 18.5k stars，最近 push 在 2026-06-06。
- 适合构建长期运行的 persona/社交 agent、链上交互 agent、Telegram/Discord/Twitter 类接口。
- 可以和 Coinbase AgentKit、钱包插件、行情工具组合。

风险：

- 它本身不是专业金融投研框架。
- 如果拿来做交易，需要自己实现严肃的数据质量、策略、回测和风控。

适合度：crypto agent shell/运行时可以考虑；交易决策核心不建议只靠它。

### 8. Yahoo Finance MCP Server

URL: https://github.com/AgentX-ai/yahoo-finance-server

定位：第三方 Yahoo Finance MCP server，提供股票信息、新闻、搜索、历史价格、期权链、earnings 等。

强项：

- 接入简单，适合快速让 LLM 查询 ticker/news/history/options。
- 2026-06-06 查询时约 47 stars，最近 push 在 2026-03-28。
- 支持 stdio 和 HTTP transport。

风险：

- 不是官方 Yahoo 产品，数据稳定性、授权、限流、延迟都要自己承担。
- README 建议配置 proxy 以提高可靠性，这说明生产稳定性不能默认假设。
- 不提供交易执行。

适合度：个人研究 demo 可以用；生产系统建议换成正式授权的数据源。

### 9. FinRL

URL: https://github.com/AI4Finance-Foundation/FinRL

定位：金融强化学习库，适合训练、回测和比较量化策略。

强项：

- 成熟度较高。2026-06-06 查询时约 15.4k stars，最近 push 在 2026-05-25。
- 适合做策略研究、RL agent、benchmark。
- 可以和 FinRobot 或自研 LLM agent 组合：LLM 负责解释和监控，FinRL/回测层负责策略验证。

风险：

- 强化学习回测容易过拟合。
- 不是新闻理解或自然语言 agent 平台。
- 实盘执行仍要另接 broker/exchange 和风控。

适合度：量化策略实验很适合；不是完整 agent 产品。

### 10. 企业级金融 AI 产品

#### Bloomberg

URL: https://www.bloomberg.com/professional/

定位：机构级金融数据、新闻、Terminal、AI/搜索/研究功能。Bloomberg 的优势是数据、新闻、历史沉淀、终端生态和机构合规，而不是开源可控。

适合：机构研究、交易员工作流、新闻和数据可信源。

不适合：个人低成本自建 agent、开源二次开发。

#### AlphaSense

URL: https://www.alpha-sense.com/

定位：企业市场情报和金融研究平台，重点在 earnings call、broker research、filings、news、expert call transcript 上做语义搜索和生成式总结。

适合：研究员、IR、PE/VC、企业战略团队。

不适合：自动交易执行。

#### FactSet Mercury / FactSet AI

URL: https://www.factset.com/

定位：机构金融数据和工作流产品，Mercury 方向是 conversational/generative AI for FactSet content。

适合：机构投研、组合分析、金融数据问答。

不适合：低成本个人 agent 或开源部署。

#### S&P Global / Kensho

URLs:

- https://www.spglobal.com/
- https://www.kensho.com/

定位：S&P Global 旗下 Kensho 做金融/商业数据的 AI、NLP、语音转录、文档处理等。更偏机构数据智能基础设施。

适合：机构文档处理、转录、实体识别、数据抽取、RAG。

不适合：个人自动交易 bot。

#### Claude for Financial Services / 通用企业 agent 平台

URL: https://www.anthropic.com/

定位：通用 LLM/agent 平台正在往金融服务场景集成，包括数据连接器、MCP、企业安全、文档分析和 agent workflow。适合做金融研究 copilots，但执行层和数据授权仍要自己处理。

适合：企业内部研究助手、审计友好型工作流、文档/表格/新闻分析。

不适合：单独作为交易系统。

## 现在比较靠谱的组合架构

### 股票/ETF/期权方向

推荐组合：

1. OpenBB 作为行情、财务、宏观、新闻、可视化层。
2. TradingAgents 做多 agent 投研和交易建议。
3. FinRobot 做 equity research report 和估值报告。
4. Alpaca MCP 做 paper trading 和受限执行。
5. 自研 Risk Gate 做最后一道硬限制。

Risk Gate 至少要包括：

- 默认 paper trading。
- live trading 必须单独开关。
- 每笔订单金额上限。
- 每日最大下单次数。
- 每日最大亏损/最大回撤停止。
- ticker 白名单/黑名单。
- 不允许 market order，或只允许小额 market order。
- 交易前输出结构化理由和引用数据。
- 大额订单人工确认。
- 完整日志：输入数据、agent 推理摘要、工具调用、订单、成交、PnL。

### Crypto / DeFi 方向

推荐组合：

1. Coinbase AgentKit 提供钱包和 onchain action。
2. CoinGecko/CoinMarketCap/交易所 API 提供行情。
3. Dune/Nansen/DefiLlama 等提供链上和协议数据。
4. 自研 news/social agent 处理 Twitter、Telegram、Discord、RSS、公告。
5. 自研 policy engine 限制合约、token、额度和链。

Crypto 的额外风险：

- 合约授权和 approve 风险。
- MEV、滑点、流动性、rug pull。
- 社交媒体噪声和操纵。
- CEX/DEX 价格差、资金费率、借贷清算。
- wallet key 管理。

第一版最好只支持测试网或极小额资金，并要求人工确认任何转账、swap、approve、bridge。

### 研究型金融 agent

如果不急着自动交易，最稳路线是：

1. OpenBB/授权数据源抓数据。
2. 新闻/RSS/SEC filings/earnings transcript 入库。
3. RAG 检索相关新闻、财报、电话会。
4. TradingAgents/FinRobot 生成观点。
5. 输出日报、watchlist、风险事件、交易计划。
6. 人手动交易或在 Alpaca paper account 验证。

这个方向最容易做出实际可用产品，因为它把 LLM 的价值放在总结、比较、解释、生成研究材料，而不是毫无保护地控制资金。

## MCP 在这里应该怎么用

MCP 很适合做“工具边界”，但不应该把所有能力暴露给 agent。

建议拆成多个 MCP server：

1. `market-data-mcp`：只读行情、K线、财务指标。
2. `news-rag-mcp`：只读新闻、公告、filings、transcripts。
3. `portfolio-mcp`：只读持仓、现金、风险暴露。
4. `paper-trading-mcp`：只允许 paper trading。
5. `execution-mcp`：真实下单，默认不启用，需要人工确认和硬风控。
6. `audit-mcp`：写日志、生成交易复盘、记录 agent decision。

不要把“读数据”和“真实下单”混在一个无约束工具包里。MCP 的优势是让 agent 发现工具，缺点也是让 agent 可以调用工具；金融系统里要把权限切得很细。

## 是否存在“专门金融 Scale 级别 agent”

如果这里的“Scale”指 Scale AI 这类企业 AI/data 公司：它们更偏数据标注、评测、企业 AI 基础设施和行业解决方案，不是主流意义上的个人可用交易 agent。金融领域真正成熟的“Scale 级别”通常是 Bloomberg、FactSet、AlphaSense、S&P Global/Kensho 这类机构产品；它们强在数据、授权、合规、搜索、研究工作流，而不是开源自动交易。

如果目标是自建一个“Scale 风格”的金融 agent 平台，应该关注：

- 高质量数据接入和授权。
- 可追溯的 RAG。
- agent evaluation。
- 工具权限隔离。
- 交易前后审计。
- 人机协作界面。
- 生产监控和异常停止。

## 选型建议

### 如果你想最快做出 demo

用：

- TradingAgents：生成股票交易研究建议。
- OpenBB：补数据和图表。
- Alpaca MCP：paper trading。

目标：让 agent 每天读 watchlist，生成多空观点、风险、建议仓位，并在 paper account 下模拟单。

### 如果你想做 crypto agent

用：

- Coinbase AgentKit：钱包/onchain action。
- 自己接 CoinGecko/交易所/DefiLlama/Nansen/Dune。
- 加 policy engine：只允许测试网或小额白名单 token。

目标：先做“分析 + 提案 + 人工确认执行”，不要先做完全自动。

### 如果你想做更严肃的研究平台

用：

- OpenBB + FinRobot + 自建 RAG。
- 把 SEC filings、earnings call、news、price history、macro data 入库。
- 先产出日报、个股研究报告、组合风险报告。

目标：做一个可靠研究助手，而不是赌博式 bot。

## 红线

这些不要做：

- 不要把 live trading API key 直接给一个没有限权的 LLM agent。
- 不要让 agent 根据单条新闻自动全仓交易。
- 不要相信没有 out-of-sample、paper trading、滑点/手续费模拟的回测。
- 不要把 Yahoo Finance 这类非授权/不稳定来源当生产级唯一数据源。
- 不要让 crypto agent 自动 approve 任意合约。
- 不要让 agent 自己修改风控规则。

## 我会怎么落地第一版

第一版可以做成 6 个服务：

1. Ingest：抓 OpenBB/Yahoo/Alpaca/交易所/新闻/RSS/SEC 数据。
2. Store：Postgres + pgvector，保存价格、新闻、filings、agent 输出。
3. Research Agent：TradingAgents 或 FinRobot + 自定义 prompts。
4. Risk Engine：纯代码规则，不让 LLM 决定硬限制。
5. Execution：Alpaca paper trading 或 Coinbase testnet。
6. Dashboard：展示观点、引用、订单、PnL、错误、人工确认。

MVP 验收标准：

- 能对 10 个 watchlist ticker 每天生成报告。
- 每条建议都能追溯到数据和新闻来源。
- paper trading 连续跑 30 天。
- 记录每次 agent 调用、输入、输出和工具调用。
- 每日自动生成复盘：命中率、盈亏、最大回撤、错误类型。
- live trading 开关默认关闭。

## 来源

- TradingAgents GitHub: https://github.com/TauricResearch/TradingAgents
- FinRobot GitHub: https://github.com/AI4Finance-Foundation/FinRobot
- OpenAlice GitHub: https://github.com/TraderAlice/OpenAlice
- Alpaca MCP Server GitHub: https://github.com/alpacahq/alpaca-mcp-server
- OpenBB GitHub: https://github.com/OpenBB-finance/OpenBB
- OpenBB Workspace MCP GitHub: https://github.com/OpenBB-finance/workspace-mcp
- Coinbase AgentKit GitHub: https://github.com/coinbase/agentkit
- elizaOS GitHub: https://github.com/elizaOS/eliza
- Yahoo Finance MCP Server GitHub: https://github.com/AgentX-ai/yahoo-finance-server
- FinRL GitHub: https://github.com/AI4Finance-Foundation/FinRL
- Bloomberg Professional: https://www.bloomberg.com/professional/
- AlphaSense: https://www.alpha-sense.com/
- FactSet: https://www.factset.com/
- S&P Global: https://www.spglobal.com/
- Kensho: https://www.kensho.com/
- Anthropic: https://www.anthropic.com/

