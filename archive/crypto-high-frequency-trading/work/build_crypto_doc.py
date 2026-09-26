from docx import Document
from docx.shared import Inches, Pt, RGBColor
from docx.enum.text import WD_ALIGN_PARAGRAPH
from docx.enum.table import WD_TABLE_ALIGNMENT, WD_CELL_VERTICAL_ALIGNMENT
from docx.oxml import OxmlElement
from docx.oxml.ns import qn
import os


OUT = "/Users/yitwah/Documents/Codex/2026-06-07/video-youtube-video-open-document-recently/outputs/crypto_learning_resources_2026.docx"
OUT_WITH_DIAGRAM = "/Users/yitwah/Documents/Codex/2026-06-07/video-youtube-video-open-document-recently/outputs/crypto_learning_resources_2026_with_chain_diagram.docx"
CHAIN_DIAGRAM = "/Users/yitwah/Library/Application Support/CleanShot/media/media_P3liaXmKed/CleanShot 2026-06-07 at 14.25.07@2x.png"


def rgb(hex_color):
    return RGBColor.from_string(hex_color)


def set_cell_shading(cell, fill):
    tc_pr = cell._tc.get_or_add_tcPr()
    shd = OxmlElement("w:shd")
    shd.set(qn("w:fill"), fill)
    tc_pr.append(shd)


def set_cell_text(cell, text, bold=False, size=8.6, color="000000"):
    cell.text = ""
    p = cell.paragraphs[0]
    p.paragraph_format.space_after = Pt(0)
    p.paragraph_format.line_spacing = 1.06
    r = p.add_run(str(text))
    r.bold = bold
    r.font.name = "Arial"
    r._element.rPr.rFonts.set(qn("w:eastAsia"), "Arial")
    r.font.size = Pt(size)
    r.font.color.rgb = rgb(color)
    cell.vertical_alignment = WD_CELL_VERTICAL_ALIGNMENT.CENTER


def set_cell_width(cell, width_in):
    cell.width = Inches(width_in)
    tc_pr = cell._tc.get_or_add_tcPr()
    tc_w = tc_pr.first_child_found_in("w:tcW")
    if tc_w is None:
        tc_w = OxmlElement("w:tcW")
        tc_pr.append(tc_w)
    tc_w.set(qn("w:w"), str(int(width_in * 1440)))
    tc_w.set(qn("w:type"), "dxa")


def set_table_widths(table, widths):
    for row in table.rows:
        for i, width in enumerate(widths):
            set_cell_width(row.cells[i], width)


def add_table(doc, headers, rows, widths, font_size=8.2):
    table = doc.add_table(rows=1, cols=len(headers))
    table.style = "Table Grid"
    table.alignment = WD_TABLE_ALIGNMENT.CENTER
    table.autofit = False
    for i, header in enumerate(headers):
        set_cell_text(table.rows[0].cells[i], header, bold=True, size=8.6, color="0B2545")
        set_cell_shading(table.rows[0].cells[i], "E8EEF5")
    for row in rows:
        cells = table.add_row().cells
        for i, text in enumerate(row):
            set_cell_text(cells[i], text, size=font_size)
    set_table_widths(table, widths)
    doc.add_paragraph("")
    return table


def add_bullet(doc, text):
    p = doc.add_paragraph(style="List Bullet")
    p.paragraph_format.space_after = Pt(3)
    p.add_run(text)


def add_number(doc, text):
    p = doc.add_paragraph(style="List Number")
    p.paragraph_format.space_after = Pt(3)
    p.add_run(text)


def add_labeled_para(doc, label, body):
    p = doc.add_paragraph()
    p.paragraph_format.space_after = Pt(5)
    r = p.add_run(label + ": ")
    r.bold = True
    p.add_run(body)


def build_doc():
    doc = Document()
    section = doc.sections[0]
    section.top_margin = Inches(0.75)
    section.bottom_margin = Inches(0.75)
    section.left_margin = Inches(0.75)
    section.right_margin = Inches(0.75)

    styles = doc.styles
    normal = styles["Normal"]
    normal.font.name = "Arial"
    normal._element.rPr.rFonts.set(qn("w:eastAsia"), "Arial")
    normal.font.size = Pt(10)
    normal.paragraph_format.space_after = Pt(6)
    normal.paragraph_format.line_spacing = 1.12

    style_tokens = {
        "Title": (24, "0B2545", False, 0, 8),
        "Heading 1": (16, "0B2545", True, 14, 6),
        "Heading 2": (13, "1F4D78", True, 10, 5),
        "Heading 3": (11, "1F4D78", True, 8, 4),
    }
    for style_name, (size, color, bold, before, after) in style_tokens.items():
        style = styles[style_name]
        style.font.name = "Arial"
        style._element.rPr.rFonts.set(qn("w:eastAsia"), "Arial")
        style.font.size = Pt(size)
        style.font.color.rgb = rgb(color)
        style.font.bold = bold
        style.paragraph_format.space_before = Pt(before)
        style.paragraph_format.space_after = Pt(after)

    p = doc.add_paragraph(style="Title")
    p.alignment = WD_ALIGN_PARAGRAPH.LEFT
    p.add_run("Crypto 交易入门资料清单（2026 年 6 月版）")
    p = doc.add_paragraph()
    p.paragraph_format.space_after = Pt(10)
    r = p.add_run(
        "面向小白：先建立交易常识，再学习风险控制、套利逻辑和自动化工具。"
        "筛选时间：视频优先近三个月；书籍优先近一年；公开报告和文档优先近三到六个月。"
    )
    r.font.size = Pt(10)
    r.font.color.rgb = rgb("555555")

    callout = doc.add_table(rows=1, cols=1)
    callout.alignment = WD_TABLE_ALIGNMENT.CENTER
    cell = callout.cell(0, 0)
    set_cell_shading(cell, "FFF4D6")
    set_cell_text(
        cell,
        "重要提醒：这是一份学习资料清单，不是投资建议。Crypto 交易有高波动、滑点、爆仓、"
        "交易所风险、智能合约风险和诈骗风险。新手不要上来就用杠杆、合约、网格重仓或所谓"
        "“AI 套利稳赚 bot”。",
        bold=True,
        size=9,
        color="5C3B00",
    )
    doc.add_paragraph("")

    doc.add_heading("1. 推荐学习顺序", level=1)
    for item in [
        "先看交易基础：现货、订单类型、K 线、成交量、盘口、手续费、滑点、资金费率。",
        "再学风险管理：仓位、止损、最大回撤、交易日志、不要把单次亏损变成账户级事故。",
        "第三步学市场结构：CEX/DEX、稳定币、永续合约、流动性、做市、资金费率套利。",
        "第四步再碰自动化：先回测和纸交易，再小资金实盘；API 权限只开交易，不开提现。",
        "最后才研究复杂策略：跨所套利、现货-永续套利、网格/做市、链上 MEV/DEX 套利。",
    ]:
        add_number(doc, item)

    doc.add_heading("2. 资料分层总表", level=1)
    add_table(
        doc,
        ["等级", "类别", "代表资料", "适合用途", "建议"],
        [
            ("A", "官方/研究报告", "CoinGecko、Binance Research、Coinbase Institutional、Kaiko", "理解市场环境、交易量、CEX/DEX、稳定币、衍生品", "优先读"),
            ("A", "官方学习中心", "Kraken Learn、Coinbase Learn、Binance Academy、CoinMarketCap Support", "安全开户、买卖、基础订单和风险提示", "优先读"),
            ("A", "开源工具官方文档", "Freqtrade、Hummingbot、CCXT、Jesse", "自动化交易、回测、做市、交易所 API", "动手时读"),
            ("B", "长视频候选", "Simplilearn、Coin Bureau、MoneyZG、Dapp University 等", "体系化入门、bot/套利概念", "核对发布日期"),
            ("B", "近一年书籍/论文", "2025-2026 出版/发表的 crypto trading、AI trading、algorithmic trading 资料", "结构化学习和深入研究", "选读"),
            ("C", "营销型短视频/bot 推广", "被动收入、提现展示、稳赚、私密套利脚本", "诈骗和过拟合风险很高", "不建议"),
        ],
        [0.45, 1.25, 2.1, 2.55, 0.65],
        8.0,
    )

    doc.add_heading("3. 最近公开报告与市场资料", level=1)
    add_table(
        doc,
        ["编号", "资料", "时间", "为什么有用", "读法"],
        [
            ("R1", "CoinGecko 2026 Q1 Crypto Industry Report", "2026-04-16", "55 页报告，覆盖总市值、BTC/ETH、DeFi、CEX/DEX、稳定币、交易量。报告列出 Q1 总市值下跌 20.4%，CEX 现货交易量下降 39.1%。", "入门后必读"),
            ("R2", "Binance Research Monthly Market Insights - May 2026", "2026-05", "月度市场洞察，适合跟踪市场叙事、链上/交易所数据、行业趋势。", "每月读"),
            ("R3", "Coinbase Crypto Market Positioning - May 2026", "2026-05-12", "机构视角：4 月风险仓位恢复但交易量和 altcoin 风险偏好仍弱。", "中级读"),
            ("R4", "Kaiko Research: Crypto in 2026, What Breaks, What Scales, What Consolidates", "2026-01", "市场结构视角，关注监管、稳定币流动性、链上衍生品和机构化。", "中级读"),
            ("R5", "Kraken: Introduction to cryptocurrency trading", "2026-03-09 更新", "交易基础说明，适合第一次理解交易所、订单和 API 入口。", "新手先读"),
            ("R6", "CoinMarketCap: How can I buy coins/tokens?", "2026-05-07 更新", "解释 CMC 自身不是交易所/钱包，强调 DEX 交易风险、滑点和第三方路由风险。", "安全必读"),
            ("R7", "CoinMarketCap: Ranking / Market Pair / Cryptoasset", "2026-01-21 更新", "解释市场对、流动性评分、虚假交易量问题；适合学会不要只看 reported volume。", "安全必读"),
            ("R8", "Kraken Learn Center", "持续更新", "Crypto 101、钱包、安全、交易工具、AI trading bots、期货策略等栏目。", "长期资料库"),
        ],
        [0.45, 2.05, 0.9, 3.35, 0.9],
        7.8,
    )

    doc.add_heading("4. 近三个月 YouTube 视频候选清单", level=1)
    p = doc.add_paragraph()
    p.add_run(
        "说明：YouTube 对自动化元数据抓取触发了反机器人限制；我能确认标题、频道、链接和部分播放量，"
        "但不是所有条目都能自动确认精确发布日期。以下只列标题明确指向 2026、内容方向清晰、可作为候选的视频；"
        "打开观看前请再确认发布日期是否在 2026-03-07 之后。"
    ).bold = True
    add_table(
        doc,
        ["编号", "视频", "频道", "长度", "看它学什么", "权重"],
        [
            ("V1", "Cryptocurrency Full Course 2026 | Cryptocurrency Tutorial For Beginners | Blockchain", "Simplilearn", "约 6.3 小时", "区块链/crypto 基础长课，适合第一轮建立词汇表和技术背景。", "候选 A"),
            ("V2", "Crypto Trading Guide: Step-by-Step For Complete Beginners", "Coin Bureau", "约 20 分钟", "偏交易入门：交易前准备、平台、基本订单、风险意识。", "补充"),
            ("V3", "Beginners Guide to Trading Crypto & Memecoins in 2026 (Free Course)", "Alex Choi Crypto", "约 37 分钟", "贴近 2026 散户交易语境，包含 memecoin 风险。", "候选 B"),
            ("V4", "How to create a profitable crypto arbitrage bot in 2026", "Dapp University", "约 12 分钟", "理解“套利 bot”从代码/流程角度到底在做什么。", "勿直接实盘"),
            ("V5", "Best Crypto Trading Bots 2026? (My Trading Bot Results)", "MoneyZG", "约 21 分钟", "对比交易 bot 类型和实际结果，适合形成工具判断。", "候选 B"),
            ("V6", "How I Built a Self-Healing Trading Bot That Fixes Its Own Losses (OpenClaw Tutorial)", "Sharbel A.", "约 19 分钟", "工程视角的自动化交易 bot 案例，适合程序员看系统结构。", "中级"),
            ("V7", "How To Actually Build a Trading Bot With Claude Code (Fully Automated)", "AI Pathways", "约 34 分钟", "偏自动化搭建流程；适合看 AI 辅助开发交易 bot 的边界。", "中级"),
            ("V8", "Crypto Trading Bot Masterclass (2026)", "Crypto Tips", "约 13 分钟", "bot 概览候选。若有“稳赚/被动收入”话术，作为反面样本看。", "低权重"),
            ("V9", "ULTIMATE Cryptocurrency Trading Course (From Beginner To PRO)", "The Trading Geek", "约 2.1 小时", "结构完整，但需确认是否近三个月；可作为基础补课。", "备选"),
            ("V10", "Crypto Trading for Beginners 2026: Best Crypto Trading Strategies?", "Top Crypto", "约 4 分钟", "短入门视频，只适合快速扫概念。", "低权重"),
        ],
        [0.42, 2.2, 0.95, 0.65, 2.5, 0.75],
        7.4,
    )

    doc.add_heading("5. 最近一年左右的书籍、论文与深度资料", level=1)
    add_table(
        doc,
        ["编号", "资料", "作者/来源", "时间", "价值与限制"],
        [
            ("B1", "Crypto Trading Decoded: A Beginner’s Guide to Profitable Strategies", "Alex Sterling / KDP", "2025-09-11", "入门定位。适合做术语和策略框架补充，但独立出版、评论少，不应作为唯一教材。"),
            ("B2", "Crypto Algorithmic Trading Mastery", "Khushabu Gupta / Google Play", "2025-10", "62 页，聚焦 Python、bot、回测和 AI；适合作为自动化入门小册。注意篇幅短，需配合官方工具文档。"),
            ("B3", "Crypto Algorithmic Trading: 16 Proven Strategies...", "Kaito Amatsuki", "2025-03-26", "覆盖 Python bot、动量、套利、机器学习。略超一年，但与自动化方向高度相关，可备选。"),
            ("B4", "Hands-On AI Trading with Python, QuantConnect, and AWS", "Jiri Pik, Ernest P. Chan 等 / Wiley", "2025-02-18", "不是纯 crypto，但作者和出版社更强，适合学 AI/量化交易流程、回测、云部署。"),
            ("B5", "Algorithmic crypto trading using information-driven bars...", "Financial Innovation / Springer", "2025-12-15", "开放论文，研究 BTC/ETH 的信息驱动采样、triple barrier 标签和深度学习。适合进阶理解“回测不要自欺欺人”。"),
            ("B6", "Bitcoin Price Prediction: Peer-Reviewed Evidence and Social Media Discourse", "arXiv", "2026-05-20", "综述指出短中期预测很难稳定战胜朴素基线，强调 walk-forward、多市场状态、交易成本。适合防止迷信 AI 预测。"),
            ("B7", "The Red Queen’s Trap: Limits of Deep Evolution in High-Frequency Trading", "arXiv", "2025-12", "用 crypto 高频环境展示训练收益和实盘表现可能严重背离。适合自动化交易前的风险教育。"),
            ("B8", "Freedom of Money: The Inside Story of Binance and Changpeng Zhao", "Changpeng Zhao", "2026-04-08", "行业人物/交易所历史，不是交易教材；适合理解 Binance、监管、交易所商业逻辑。"),
        ],
        [0.42, 2.45, 1.45, 0.8, 2.78],
        7.5,
    )

    doc.add_heading("6. 自动化交易与套利工具栈", level=1)
    add_table(
        doc,
        ["编号", "工具", "是什么", "适合用途", "关键注意"],
        [
            ("T1", "Freqtrade", "开源 Python crypto trading bot；适合策略回测、参数优化、纸交易、现货/期货自动化。", "新手做系统交易实验的首选之一；先 Docker/本地跑回测。", "官方文档建议用预构建 Docker 快速开始；stable 分支通常约每月发布。"),
            ("T2", "Hummingbot", "开源 Python 框架，偏做市、跨交易所、套利、连接 CEX/DEX。", "想学 market making、spot-perp/cross-exchange 逻辑时用。", "官方文档强调模块化、连接器、Strategy V2、Condor/Client quickstart。"),
            ("T3", "CCXT", "交易所 API 统一封装库，支持大量 CEX；适合自己写数据抓取、下单、账户查询。", "程序员必学；但它只是 API 工具，不帮你管理策略风险。", "先学 rate limit、错误重试、订单状态和 API 权限。"),
            ("T4", "Jesse", "Python 交易策略框架，含项目模板、回测、dashboard。", "适合学习策略结构、回测和多交易对逻辑。", "官方文档包含 backtest、strategies、routes、import candles 等入口。"),
            ("T5", "TradingView + webhook", "图表、告警、策略信号，再接 webhook 到交易执行端。", "适合非程序员先练信号和纪律。", "需自己处理延迟、重复信号、断网、API 错误。"),
            ("T6", "Exchange API sandbox / testnet", "Binance/OKX/Bybit 等常有 testnet 或 demo trading。", "任何 bot 实盘前必须先 testnet 或 paper trading。", "API key 禁止开提现权限。"),
        ],
        [0.42, 0.9, 2.35, 2.15, 2.08],
        7.6,
    )

    doc.add_heading("7. 不同链、跨链和手续费关系图", level=1)
    p = doc.add_paragraph()
    p.add_run(
        "这张图解释了中心化交易所、你的钱包、不同公链、跨链桥和手续费之间的关系。"
        "重点是：每条链都是独立账本；Bitget App 内部买卖更像中心化账本；"
        "从 Bitget 提现到钱包并在某条链上签名交易，才是链上操作。"
        "链选错、地址错、memo/tag 漏填都可能造成资产损失。"
    )
    pic_p = doc.add_paragraph()
    pic_p.alignment = WD_ALIGN_PARAGRAPH.CENTER
    run = pic_p.add_run()
    run.add_picture(CHAIN_DIAGRAM, width=Inches(6.75))
    cap = doc.add_paragraph()
    cap.alignment = WD_ALIGN_PARAGRAPH.CENTER
    cap.paragraph_format.space_after = Pt(10)
    r = cap.add_run("图：不同链、跨链、CEX 提现网络与手续费的关系")
    r.italic = True
    r.font.size = Pt(9)
    r.font.color.rgb = rgb("555555")

    doc.add_page_break()
    doc.add_heading("8. 新手应该理解的套利方式", level=1)
    p = doc.add_paragraph()
    p.add_run(
        "先把一个概念讲清楚：套利不是“无风险赚钱”，而是利用不同市场、不同时间、"
        "不同订单簿或不同合约之间的价格差。真正的利润要扣掉手续费、滑点、资金占用、"
        "失败交易、延迟和极端行情风险。新手学习套利，重点不是马上实盘，而是学会拆成本。"
    )
    add_table(
        doc,
        ["类型", "看起来赚什么", "主要成本", "新手建议"],
        [
            ("跨交易所价差", "A 所便宜买，B 所贵卖", "提现时间、手续费、盘口深度、限额、KYC", "先用表格模拟，不急着搬币"),
            ("现货-永续资金费率", "持有现货、做空永续，收 funding", "保证金、强平、基差、资金费率反转", "先学 funding 机制和保证金"),
            ("DEX/AMM 套利", "链上池子价格偏离外部价格", "gas、MEV、滑点、失败交易、合约风险", "小白不要主网实战"),
            ("网格交易", "震荡区间反复低买高卖", "单边行情、库存、手续费、弱币长期下跌", "只当区间策略，不当利息"),
            ("做市/挂单 spread", "同时挂买卖单赚价差", "逆向选择、撤单延迟、库存偏移", "先模拟盘口，不直接实盘"),
        ],
        [1.35, 1.65, 2.35, 2.25],
        7.7,
    )

    doc.add_heading("8.1 跨交易所价差套利", level=2)
    add_labeled_para(
        doc,
        "基本原理",
        "同一个币在不同中心化交易所的价格可能短暂不同。例如某币在 A 交易所卖 100 USDT，"
        "在 B 交易所买盘愿意以 101 USDT 买入，表面价差是 1%。真正套利需要在便宜市场买入，"
        "在贵市场卖出，或者两边提前都放好资金，同时买卖。"
    )
    add_labeled_para(
        doc,
        "为什么普通散户经常赚不到",
        "价格差通常不是免费午餐。你看到价差时，可能已经被专业做市商、搬砖机器人或交易所内部资金"
        "吃掉；等你充值、提现、等待区块确认时，价差可能消失。小币的盘口很浅，你一买就把价格推高，"
        "一卖又把价格砸低。"
    )
    add_table(
        doc,
        ["检查项", "为什么重要"],
        [
            ("两边盘口深度", "不能只看最新价，要看你真实下单金额能不能在目标价格附近成交。"),
            ("充值/提现状态", "某些交易所会临时关闭充值或提现；价差很大时尤其要怀疑。"),
            ("链和手续费", "提现网络不同会影响到账时间和成本，选错链还可能丢币。"),
            ("KYC/限额", "账户等级、地区限制和风控冻结会让套利链条断掉。"),
            ("时间风险", "等待确认期间价格可能反转，套利变成裸露方向仓位。"),
        ],
        [1.65, 5.6],
        8.0,
    )
    add_labeled_para(
        doc,
        "新手怎么学",
        "先不用真钱搬砖。用 CoinMarketCap 或交易所页面记录同一交易对在多个交易所的价格、盘口深度、"
        "提现费和到账时间，做一张“扣费后是否还有利润”的表。能连续几周算清楚成本，再考虑极小资金测试。"
    )

    doc.add_heading("8.2 现货-永续资金费率套利", level=2)
    add_labeled_para(
        doc,
        "基本原理",
        "永续合约没有到期日，交易所用资金费率让永续价格贴近现货价格。若多头需要向空头支付资金费，"
        "你可以理论上买入现货，同时做空同等数量永续，方向风险大致抵消，并尝试收取 funding。"
    )
    add_labeled_para(
        doc,
        "一个简化例子",
        "假设你买入 1 BTC 现货，同时做空 1 BTC 永续。如果 BTC 上涨，现货盈利、空头亏损；如果 BTC 下跌，"
        "现货亏损、空头盈利。理论上价格方向被对冲，收益主要来自资金费率。但现实中还有手续费、保证金、"
        "基差变化和强平风险。"
    )
    add_table(
        doc,
        ["风险", "解释"],
        [
            ("资金费率反转", "今天收 funding，不代表明天继续收；市场情绪改变后可能变成你付钱。"),
            ("基差变化", "现货和永续价格不可能完全同步，平仓时基差变化会影响结果。"),
            ("保证金/强平", "合约腿需要保证金；极端波动下即使整体对冲，也可能因保证金不足被强平。"),
            ("交易所风险", "现货和合约常在同一或不同交易所，均有宕机、风控、穿仓、提款限制风险。"),
            ("收益被成本吃掉", "maker/taker 费、资金划转、借贷成本和滑点会侵蚀看似稳定的收益。"),
        ],
        [1.55, 5.7],
        8.0,
    )
    add_labeled_para(
        doc,
        "新手怎么学",
        "先只观察交易所的 funding rate 历史，不开仓。把 funding、现货价格、永续价格、手续费和保证金率"
        "放进表格，模拟如果持有 7 天、30 天会发生什么。能解释强平价和基差后，再考虑模拟盘。"
    )

    doc.add_heading("8.3 DEX/AMM 套利", level=2)
    add_labeled_para(
        doc,
        "基本原理",
        "AMM 池子用智能合约报价，例如 Uniswap、Raydium 这类池子。若链上池子里的价格和外部市场价格偏离，"
        "套利者会在便宜处买、贵处卖，把价格推回去。"
    )
    add_labeled_para(
        doc,
        "为什么小白不适合直接做",
        "链上套利竞争非常激烈，专业搜索者会监控 mempool、用私有 RPC 或 bundle 抢顺序。你看到机会后手动提交交易，"
        "通常已经太慢。更糟的是，失败交易也要付 gas，滑点设置太高会被夹子攻击，授权错合约可能直接丢资产。"
    )
    add_table(
        doc,
        ["术语", "新手版解释"],
        [
            ("Gas", "链上交易手续费；失败也可能照付。"),
            ("Slippage", "你预期价格和实际成交价格的差；池子越浅，滑点越大。"),
            ("MEV", "验证者、构建者或机器人通过交易排序提取价值，比如夹子攻击。"),
            ("私有 mempool/RPC", "把交易发送到较私密的通道，减少被抢跑或夹击的概率，但不是绝对安全。"),
            ("授权", "钱包允许合约花你的 token；无限授权给恶意合约可能导致资产被转走。"),
        ],
        [1.55, 5.7],
        8.0,
    )
    add_labeled_para(
        doc,
        "新手怎么学",
        "先在测试网或小额钱包观察 AMM 池子的 swap。重点学习池子深度、滑点、gas、授权和交易失败原因。"
        "不要导入主钱包私钥，不要给陌生合约无限授权，不要相信“链上套利脚本一键赚钱”。"
    )

    doc.add_heading("8.4 网格交易", level=2)
    add_labeled_para(
        doc,
        "基本原理",
        "网格不是严格意义的无风险套利，而是一种区间策略。你预设价格区间和网格间距，价格下跌时买入，"
        "上涨时卖出，在震荡行情中反复赚小价差。"
    )
    add_labeled_para(
        doc,
        "为什么收益展示容易误导",
        "很多交易所会展示“网格收益”，但你还要看总资产净值。如果币价单边下跌，网格会不断买入，"
        "最后变成持有一堆下跌的币；如果币价单边上涨，网格可能过早卖飞。网格收益不是银行利息。"
    )
    add_table(
        doc,
        ["关键参数", "含义"],
        [
            ("价格区间", "区间设错时，价格跑出去后策略效果会明显变差。"),
            ("网格数量/间距", "越密成交越多，但手续费也越多；越疏机会少但成本低。"),
            ("投入资金", "决定每格买卖数量；资金太小可能被最小下单量和手续费吃掉。"),
            ("标的选择", "强波动弱币看起来网格收益高，但长期下跌风险也大。"),
            ("退出条件", "必须预先设定止损、止盈或停止条件，否则容易长期套住。"),
        ],
        [1.55, 5.7],
        8.0,
    )
    add_labeled_para(
        doc,
        "新手怎么学",
        "先用历史 K 线手工画区间：如果价格突破上沿或跌破下沿，你会怎么办？再用交易所模拟网格或小额现货网格，"
        "每天记录总资产净值，而不是只看网格成交利润。"
    )

    doc.add_heading("8.5 做市/挂单赚 spread", level=2)
    add_labeled_para(
        doc,
        "基本原理",
        "做市商在买一和卖一附近同时挂买单和卖单，希望低价买入、高价卖出，赚 bid-ask spread。"
        "做市的核心不是预测方向，而是管理库存、价差、撤单速度和成交质量。"
    )
    add_labeled_para(
        doc,
        "最大难点：逆向选择",
        "你的挂单被成交，往往不是因为别人送你钱，而是因为市场正在朝不利方向移动。比如你挂买单成交，"
        "可能是价格继续下跌前别人把货卖给你；你挂卖单成交，可能是价格继续上涨前你把货卖早了。"
    )
    add_table(
        doc,
        ["风险", "解释"],
        [
            ("库存偏移", "单边行情中只成交一侧订单，账户变成持有过多币或过多现金。"),
            ("撤单延迟", "行情快速变化时，旧报价来不及撤掉，容易被更快的交易者吃掉。"),
            ("手续费等级", "没有低手续费或 maker rebate，spread 很容易不够覆盖成本。"),
            ("盘口竞争", "专业做市商有更低延迟、更好风控和更大资金。"),
            ("交易所/API 风险", "断网、API 限速、订单状态未知时，挂单策略可能失控。"),
        ],
        [1.55, 5.7],
        8.0,
    )
    add_labeled_para(
        doc,
        "新手怎么学",
        "先在交易所盘口上观察 spread、深度和成交方向。再用 Hummingbot 这类工具在 paper trading 或极小资金环境"
        "学习 inventory、order refresh、spread、stop-loss。不要一开始就在流动性差的小币上做市。"
    )

    add_labeled_para(
        doc,
        "本节底线",
        "任何套利都要先问四个问题：第一，价差是否足够覆盖全部成本；第二，成交和转账是否来得及；"
        "第三，最坏情况下会亏多少；第四，系统出错时能不能立刻停下来。回答不清楚，就只学习不实盘。"
    )

    doc.add_heading("9. 90 天学习路线", level=1)
    add_table(
        doc,
        ["阶段", "目标", "资料/动作", "纪律"],
        [
            ("第 1-2 周", "只学基础和安全", "Kraken/Coinbase/Binance/CoinMarketCap 基础资料；建立词汇表。", "不要交易或只用极小金额现货。"),
            ("第 3-4 周", "学订单和风险", "限价/市价/止损、手续费、滑点、仓位、交易日志。", "每笔交易先写计划再下单。"),
            ("第 5-6 周", "看市场报告", "CoinGecko Q1、Binance May、Coinbase Market Positioning、Kaiko。", "理解现在是强趋势、震荡还是高风险宏观期。"),
            ("第 7-8 周", "纸交易/模拟盘", "TradingView 记录信号；Freqtrade 或 Jesse 跑历史回测。", "不许把回测收益当实盘收益。"),
            ("第 9-10 周", "小资金现货策略", "只测试 1-2 个策略：DCA、低频趋势、简单网格。", "设置每日/每周最大亏损。"),
            ("第 11-12 周", "自动化和套利实验", "Hummingbot/Freqtrade testnet；研究资金费率和跨所价差。", "API 禁止提现；先跑日志和风控。"),
        ],
        [0.9, 1.2, 3.25, 2.15],
        8.0,
    )

    doc.add_heading("10. 看到这些就先远离", level=1)
    for item in [
        "承诺“risk-free profit”“guaranteed passive income”“每天固定收益”的套利 bot。",
        "要求导入私钥/助记词、开提现权限 API key、连接陌生钱包签无限授权。",
        "视频主要展示提现截图、豪车、收益表，但没有解释数据、手续费、滑点、失败交易。",
        "让你下载闭源脚本、浏览器插件、Telegram bot 或“私密节点”。",
        "用 100x 杠杆、马丁格尔、亏损加仓来包装成“AI 自动修复亏损”。",
        "只展示短期回测，不包含手续费、滑点、资金费率、交易失败和不同市场状态。",
    ]:
        add_bullet(doc, item)

    doc.add_heading("11. 最小工具清单", level=1)
    for item in [
        "数据：CoinGecko、CoinMarketCap、TradingView、交易所公开行情。",
        "学习：Kraken Learn、Coinbase Learn、Binance Academy、CoinGecko/Coinbase/Binance/Kaiko 报告。",
        "记录：Notion/Excel/Google Sheets 交易日志，字段至少包括入场理由、止损、仓位、费用、结果、复盘。",
        "自动化：先 Freqtrade/Jesse 回测，再 Hummingbot 做市/套利，再 CCXT 自写执行层。",
        "安全：硬件钱包或独立热钱包；交易所 2FA；API key 最小权限；不要把全部资金放在一个平台。",
    ]:
        add_bullet(doc, item)

    doc.add_heading("12. 链接索引", level=1)
    links = [
        ("R1", "https://www.coingecko.com/research/publications/2026-q1-crypto-industry-report"),
        ("R1-PDF", "https://assets.coingecko.com/reports/2026/CoinGecko-2026-Q1-Crypto-Industry-Report-VN.pdf"),
        ("R2", "https://public.bnbstatic.com/static/files/research/monthly-market-insights-2026-05.pdf"),
        ("R3", "https://www.coinbase.com/institutional/research-insights/research/trading-insights/crypto-market-positioning-may-2026"),
        ("R4", "https://research.kaiko.com/insights/crypto-in-2026-what-breaks-what-scales-what-consolidates"),
        ("R5", "https://support.kraken.com/hc/en-us/articles/360000674406-Introduction-to-cryptocurrency-trading"),
        ("R6", "https://support.coinmarketcap.com/hc/en-us/articles/360013803852-How-can-I-buy-coins-tokens"),
        ("R7", "https://support.coinmarketcap.com/hc/en-us/articles/360043836851-Cryptoasset-Rank"),
        ("R8", "https://www.kraken.com/learn/"),
        ("T1", "https://docs.freqtrade.io/en/stable/installation/"),
        ("T2", "https://hummingbot.org/docs/"),
        ("T3", "https://docs.ccxt.com/en/latest/manual.html"),
        ("T4", "https://docs.jesse.trade/docs/getting-started/"),
        ("B2", "https://play.google.com/store/books/details/Khushabu_Gupta_Crypto_Algorithmic_Trading_Mastery?id=dGOLEQAAQBAJ"),
        ("B5", "https://link.springer.com/article/10.1186/s40854-025-00866-w"),
        ("B6", "https://arxiv.org/abs/2606.00071"),
        ("B7", "https://arxiv.org/abs/2512.15732"),
        ("B8", "https://freedomofmoney.net/"),
        ("V1", "https://www.youtube.com/watch?v=bRLIeRRV6X4"),
        ("V2", "https://www.youtube.com/watch?v=14HIIUjOLGY"),
        ("V3", "https://www.youtube.com/watch?v=TwnuvydK20U"),
        ("V4", "https://www.youtube.com/watch?v=-PWyM6adiIE"),
        ("V5", "https://www.youtube.com/watch?v=pliyHuME5J0"),
        ("V6", "https://www.youtube.com/watch?v=btG5YpvPkwE"),
        ("V7", "https://www.youtube.com/watch?v=y_bsjZThP0o"),
        ("V8", "https://www.youtube.com/watch?v=Er4KvQtZpxw"),
        ("V9", "https://www.youtube.com/watch?v=UYnCQEHx7ZU"),
        ("V10", "https://www.youtube.com/watch?v=_sLJleSaKMk"),
    ]
    for key, url in links:
        p = doc.add_paragraph(style="List Bullet")
        p.paragraph_format.space_after = Pt(2)
        r = p.add_run(f"{key}: ")
        r.bold = True
        p.add_run(url)

    for section in doc.sections:
        footer = section.footer.paragraphs[0]
        footer.alignment = WD_ALIGN_PARAGRAPH.RIGHT
        run = footer.add_run("Crypto learning resources 2026 | generated 2026-06-07")
        run.font.size = Pt(8)
        run.font.color.rgb = rgb("666666")

    for table in doc.tables:
        table.autofit = False
        for row in table.rows:
            for cell in row.cells:
                for p in cell.paragraphs:
                    for run in p.runs:
                        run.font.name = "Arial"
                        run._element.rPr.rFonts.set(qn("w:eastAsia"), "Arial")

    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    doc.save(OUT)
    doc.save(OUT_WITH_DIAGRAM)


if __name__ == "__main__":
    build_doc()
    print(OUT_WITH_DIAGRAM)
