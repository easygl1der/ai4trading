from pathlib import Path

from docx import Document
from docx.enum.section import WD_SECTION
from docx.enum.table import WD_CELL_VERTICAL_ALIGNMENT, WD_TABLE_ALIGNMENT
from docx.enum.text import WD_ALIGN_PARAGRAPH
from docx.oxml import OxmlElement
from docx.oxml.ns import qn
from docx.shared import Inches, Pt, RGBColor


BASE = Path("/Users/yitwah/Documents/Codex/2026-06-07/video-youtube-video-open-document-recently")
OUT = BASE / "outputs/youtube_v6_openclaw_self_healing_trading_bot_notes_zh.docx"

VIDEO = {
    "title": "How I Built a Self-Healing Trading Bot That Fixes Its Own Losses (OpenClaw Tutorial)",
    "channel": "Sharbel A.",
    "url": "https://www.youtube.com/watch?v=btG5YpvPkwE",
    "published": "2026-03-19T15:21:57Z",
    "duration": "PT18M55S",
    "views": "82,315",
    "likes": "2,197",
    "comments": "171",
}

SOURCES = [
    ("YouTube 视频", VIDEO["url"]),
    ("Karpathy miniAutoResearch repo", "https://github.com/karpathy/miniAutoResearch"),
    ("OpenClaw", "https://openclaw.ai"),
    ("Claude Code", "https://claude.ai/code"),
]


def rgb(value: str) -> RGBColor:
    return RGBColor.from_string(value)


def set_run_font(run, size=10, color="111111", bold=False, italic=False):
    run.font.name = "Arial"
    run._element.rPr.rFonts.set(qn("w:ascii"), "Arial")
    run._element.rPr.rFonts.set(qn("w:hAnsi"), "Arial")
    run._element.rPr.rFonts.set(qn("w:eastAsia"), "PingFang SC")
    run.font.size = Pt(size)
    run.font.color.rgb = rgb(color)
    run.bold = bold
    run.italic = italic


def shade(cell, fill: str):
    tc_pr = cell._tc.get_or_add_tcPr()
    shd = OxmlElement("w:shd")
    shd.set(qn("w:fill"), fill)
    tc_pr.append(shd)


def set_cell_margins(cell, top=90, bottom=90, start=130, end=130):
    tc_pr = cell._tc.get_or_add_tcPr()
    tc_mar = tc_pr.first_child_found_in("w:tcMar")
    if tc_mar is None:
        tc_mar = OxmlElement("w:tcMar")
        tc_pr.append(tc_mar)
    for m, v in [("top", top), ("bottom", bottom), ("start", start), ("end", end)]:
        node = tc_mar.find(qn(f"w:{m}"))
        if node is None:
            node = OxmlElement(f"w:{m}")
            tc_mar.append(node)
        node.set(qn("w:w"), str(v))
        node.set(qn("w:type"), "dxa")


def set_cell_width(cell, inches: float):
    cell.width = Inches(inches)
    tc_pr = cell._tc.get_or_add_tcPr()
    tc_w = tc_pr.first_child_found_in("w:tcW")
    if tc_w is None:
        tc_w = OxmlElement("w:tcW")
        tc_pr.append(tc_w)
    tc_w.set(qn("w:w"), str(int(inches * 1440)))
    tc_w.set(qn("w:type"), "dxa")


def set_table_geometry(table, widths):
    table.autofit = False
    tbl = table._tbl
    tbl_pr = tbl.tblPr
    tbl_w = tbl_pr.first_child_found_in("w:tblW")
    if tbl_w is None:
        tbl_w = OxmlElement("w:tblW")
        tbl_pr.append(tbl_w)
    tbl_w.set(qn("w:w"), str(int(sum(widths) * 1440)))
    tbl_w.set(qn("w:type"), "dxa")
    tbl_ind = tbl_pr.first_child_found_in("w:tblInd")
    if tbl_ind is None:
        tbl_ind = OxmlElement("w:tblInd")
        tbl_pr.append(tbl_ind)
    tbl_ind.set(qn("w:w"), "120")
    tbl_ind.set(qn("w:type"), "dxa")

    grid = tbl.tblGrid
    if grid is None:
        grid = OxmlElement("w:tblGrid")
        tbl.append(grid)
    for child in list(grid):
        grid.remove(child)
    for width in widths:
        col = OxmlElement("w:gridCol")
        col.set(qn("w:w"), str(int(width * 1440)))
        grid.append(col)

    for row in table.rows:
        for i, width in enumerate(widths):
            set_cell_width(row.cells[i], width)
            set_cell_margins(row.cells[i])


def add_cell_text(cell, text, bold=False, size=8.6, color="111111", align=None, fill=None):
    if fill:
        shade(cell, fill)
    cell.text = ""
    p = cell.paragraphs[0]
    p.paragraph_format.space_after = Pt(0)
    p.paragraph_format.line_spacing = 1.12
    if align is not None:
        p.alignment = align
    r = p.add_run(str(text))
    set_run_font(r, size=size, color=color, bold=bold)
    cell.vertical_alignment = WD_CELL_VERTICAL_ALIGNMENT.CENTER


def add_table(doc, headers, rows, widths, size=8.3):
    table = doc.add_table(rows=1, cols=len(headers))
    table.style = "Table Grid"
    table.alignment = WD_TABLE_ALIGNMENT.CENTER
    for i, header in enumerate(headers):
        add_cell_text(table.rows[0].cells[i], header, bold=True, size=8.7, color="0B2545", fill="E8EEF5")
    for row in rows:
        cells = table.add_row().cells
        for i, value in enumerate(row):
            align = WD_ALIGN_PARAGRAPH.CENTER if i == 0 and len(str(value)) < 12 else WD_ALIGN_PARAGRAPH.LEFT
            add_cell_text(cells[i], value, size=size, align=align)
    set_table_geometry(table, widths)
    spacer(doc, 4)
    return table


def spacer(doc, points=6):
    p = doc.add_paragraph()
    p.paragraph_format.space_after = Pt(points)
    p.paragraph_format.space_before = Pt(0)


def para(doc, text, size=10.1, color="111111", bold_prefix=None):
    p = doc.add_paragraph()
    p.paragraph_format.space_before = Pt(0)
    p.paragraph_format.space_after = Pt(6)
    p.paragraph_format.line_spacing = 1.13
    if bold_prefix and text.startswith(bold_prefix):
        r1 = p.add_run(bold_prefix)
        set_run_font(r1, size=size, color=color, bold=True)
        text = text[len(bold_prefix):]
    r = p.add_run(text)
    set_run_font(r, size=size, color=color)


def bullet(doc, text, level=0):
    p = doc.add_paragraph(style="List Bullet")
    p.paragraph_format.left_indent = Inches(0.34 + level * 0.18)
    p.paragraph_format.first_line_indent = Inches(-0.16)
    p.paragraph_format.space_after = Pt(3.5)
    p.paragraph_format.line_spacing = 1.12
    r = p.add_run(text)
    set_run_font(r, size=9.6)


def new_numbering_id(doc, abstract_id="7"):
    numbering = doc.part.numbering_part.numbering_definitions._numbering
    existing = [
        int(num.get(qn("w:numId")))
        for num in numbering.findall(qn("w:num"))
        if num.get(qn("w:numId")) is not None
    ]
    num_id = max(existing, default=0) + 1
    num = OxmlElement("w:num")
    num.set(qn("w:numId"), str(num_id))
    abstract = OxmlElement("w:abstractNumId")
    abstract.set(qn("w:val"), str(abstract_id))
    num.append(abstract)
    lvl_override = OxmlElement("w:lvlOverride")
    lvl_override.set(qn("w:ilvl"), "0")
    start_override = OxmlElement("w:startOverride")
    start_override.set(qn("w:val"), "1")
    lvl_override.append(start_override)
    num.append(lvl_override)
    numbering.append(num)
    return num_id


def number(doc, text, num_id=None):
    p = doc.add_paragraph(style="List Number")
    if num_id is not None:
        p_pr = p._p.get_or_add_pPr()
        num_pr = p_pr.find(qn("w:numPr"))
        if num_pr is None:
            num_pr = OxmlElement("w:numPr")
            p_pr.append(num_pr)
        ilvl = num_pr.find(qn("w:ilvl"))
        if ilvl is None:
            ilvl = OxmlElement("w:ilvl")
            num_pr.append(ilvl)
        ilvl.set(qn("w:val"), "0")
        num_id_node = num_pr.find(qn("w:numId"))
        if num_id_node is None:
            num_id_node = OxmlElement("w:numId")
            num_pr.append(num_id_node)
        num_id_node.set(qn("w:val"), str(num_id))
    p.paragraph_format.left_indent = Inches(0.34)
    p.paragraph_format.first_line_indent = Inches(-0.16)
    p.paragraph_format.space_after = Pt(3.5)
    p.paragraph_format.line_spacing = 1.12
    r = p.add_run(text)
    set_run_font(r, size=9.6)


def numbered_list(doc, items):
    num_id = new_numbering_id(doc)
    for item in items:
        number(doc, item, num_id=num_id)


def add_callout(doc, title, text, fill="FFF4D6", color="5C3B00"):
    table = doc.add_table(rows=1, cols=1)
    table.alignment = WD_TABLE_ALIGNMENT.CENTER
    cell = table.cell(0, 0)
    shade(cell, fill)
    set_cell_margins(cell, top=120, bottom=120, start=150, end=150)
    cell.text = ""
    p = cell.paragraphs[0]
    p.paragraph_format.space_after = Pt(0)
    p.paragraph_format.line_spacing = 1.14
    r1 = p.add_run(title + "：")
    set_run_font(r1, size=9.3, color=color, bold=True)
    r2 = p.add_run(text)
    set_run_font(r2, size=9.3, color=color)
    set_table_geometry(table, [6.28])
    spacer(doc, 5)


def label_para(doc, label, text):
    p = doc.add_paragraph()
    p.paragraph_format.space_after = Pt(5)
    p.paragraph_format.line_spacing = 1.12
    r1 = p.add_run(label + "：")
    set_run_font(r1, size=9.9, color="0B2545", bold=True)
    r2 = p.add_run(text)
    set_run_font(r2, size=9.9)


def source_para(doc, label, url):
    p = doc.add_paragraph()
    p.paragraph_format.space_after = Pt(3)
    r = p.add_run(f"{label}: {url}")
    set_run_font(r, size=8.4, color="555555")


def paragraph_bottom_rule(paragraph, color="1F4D78", size="8"):
    p_pr = paragraph._p.get_or_add_pPr()
    p_bdr = p_pr.find(qn("w:pBdr"))
    if p_bdr is None:
        p_bdr = OxmlElement("w:pBdr")
        p_pr.append(p_bdr)
    bottom = OxmlElement("w:bottom")
    bottom.set(qn("w:val"), "single")
    bottom.set(qn("w:sz"), size)
    bottom.set(qn("w:space"), "4")
    bottom.set(qn("w:color"), color)
    p_bdr.append(bottom)


def setup_styles(doc):
    section = doc.sections[0]
    section.page_width = Inches(8.5)
    section.page_height = Inches(11)
    section.top_margin = Inches(0.72)
    section.bottom_margin = Inches(0.72)
    section.left_margin = Inches(0.72)
    section.right_margin = Inches(0.72)
    section.header_distance = Inches(0.38)
    section.footer_distance = Inches(0.38)

    styles = doc.styles
    normal = styles["Normal"]
    normal.font.name = "Arial"
    normal._element.rPr.rFonts.set(qn("w:eastAsia"), "PingFang SC")
    normal.font.size = Pt(10)
    normal.paragraph_format.space_after = Pt(6)
    normal.paragraph_format.line_spacing = 1.13

    for name, size, color, bold, before, after in [
        ("Title", 22, "0B2545", False, 0, 4),
        ("Subtitle", 10.5, "555555", False, 0, 10),
        ("Heading 1", 15.5, "0B2545", True, 13, 5),
        ("Heading 2", 12.4, "1F4D78", True, 9, 4),
        ("Heading 3", 10.8, "1F4D78", True, 7, 3),
    ]:
        style = styles[name]
        style.font.name = "Arial"
        style._element.rPr.rFonts.set(qn("w:eastAsia"), "PingFang SC")
        style.font.size = Pt(size)
        style.font.color.rgb = rgb(color)
        style.font.bold = bold
        style.paragraph_format.space_before = Pt(before)
        style.paragraph_format.space_after = Pt(after)
        style.paragraph_format.line_spacing = 1.08

    for list_style in ["List Bullet", "List Number"]:
        style = styles[list_style]
        style.font.name = "Arial"
        style._element.rPr.rFonts.set(qn("w:eastAsia"), "PingFang SC")
        style.font.size = Pt(9.6)
        style.paragraph_format.space_after = Pt(3.5)
        style.paragraph_format.line_spacing = 1.12


def setup_header_footer(doc):
    section = doc.sections[0]
    header = section.header.paragraphs[0]
    header.alignment = WD_ALIGN_PARAGRAPH.LEFT
    header.paragraph_format.space_after = Pt(0)
    r = header.add_run("YouTube V6 记录 | OpenClaw 自愈交易 bot 案例")
    set_run_font(r, size=8.3, color="777777")

    footer = section.footer.paragraphs[0]
    footer.alignment = WD_ALIGN_PARAGRAPH.RIGHT
    footer.paragraph_format.space_before = Pt(0)
    r2 = footer.add_run("中文学习记录 · 非投资建议")
    set_run_font(r2, size=8.2, color="777777")


def add_opening(doc):
    kicker = doc.add_paragraph()
    kicker.paragraph_format.space_after = Pt(6)
    rk = kicker.add_run("VIDEO NOTES / RISK EDUCATION")
    set_run_font(rk, size=8.6, color="7A5A00", bold=True)

    title = doc.add_paragraph(style="Title")
    title.alignment = WD_ALIGN_PARAGRAPH.LEFT
    title.add_run("OpenClaw 自愈交易 Bot：中文记录与风险解读")
    for run in title.runs:
        set_run_font(run, size=22, color="0B2545")

    sub = doc.add_paragraph(style="Subtitle")
    rs = sub.add_run("基于 Sharbel A. 的 YouTube 视频《How I Built a Self-Healing Trading Bot That Fixes Its Own Losses (OpenClaw Tutorial)》整理，面向有一点编程基础的 crypto 新手。")
    set_run_font(rs, size=10.3, color="555555")

    meta = doc.add_table(rows=4, cols=2)
    meta.alignment = WD_TABLE_ALIGNMENT.CENTER
    rows = [
        ("视频", VIDEO["title"]),
        ("频道 / 时长", f"{VIDEO['channel']} / 18 分 55 秒"),
        ("发布日期", "2026-03-19"),
        ("资料状态", "BrowserOS YouTube MCP 成功读取公开视频详情与完整 transcript；本文为中文摘要、工程解读与风险教育，不复制完整字幕。"),
    ]
    for i, (label, value) in enumerate(rows):
        add_cell_text(meta.rows[i].cells[0], label, bold=True, size=8.7, color="0B2545", fill="F2F4F7")
        add_cell_text(meta.rows[i].cells[1], value, size=8.7)
    set_table_geometry(meta, [1.25, 5.03])
    spacer(doc, 5)

    add_callout(
        doc,
        "先读这个边界",
        "视频标题中的“self-healing / fixes its own losses”应理解为自动研究、自动回测、自动筛掉坏策略的工程流程，不等于自动把实盘亏损修回来，更不等于稳赚。本文把它当作自动化交易系统设计、回测纪律和风控教育案例。",
        fill="FDEBD0",
        color="6B3D00",
    )


def add_summary(doc):
    doc.add_heading("1. 视频主旨", level=1)
    para(doc, "视频从一次真实失败讲起：作者原先的 AI trading bot 曾把 50 美元跑到约 500 美元，但在没有持续检查的几天里，账户最后归零。问题不是单笔爆仓，而是策略在市场状态变化后持续小额亏损，加上大量交易手续费，慢慢把账户磨掉。")
    para(doc, "作者随后把 Karpathy 的 auto research 思路迁移到交易策略上：让 AI 反复提出策略或参数变更，用历史数据做回测，用 Sharpe 等指标比较结果，检测 look-ahead bias，并只保留优于当前最佳版本的策略。")

    add_callout(
        doc,
        "一句话结论",
        "这不是“亏了自动回血”的魔法，而是“把研究循环自动化”：生成候选策略 -> 回测 -> 检查作弊/过拟合 -> 记录分数 -> 只接受更好的版本。真正有价值的是纪律化实验和失败记录，不是收益承诺。",
        fill="E8F4FD",
        color="0B2545",
    )

    doc.add_heading("2. 关键时间线", level=1)
    add_table(
        doc,
        ["时间点", "视频内容", "给新手的含义"],
        [
            ("00:00", "50 美元跑到 500 美元后归零，作者强调不是单笔大亏，而是 bot 缓慢失血。", "自动交易最危险的失败之一是“看起来没出事，但每天都在被手续费和小亏损磨损”。"),
            ("03:17", "累计 814 笔交易；closed P&L 约 -193 美元，手续费约 -115 美元，合计损失约 -386 美元。", "交易频率越高，手续费、滑点和错误下单越重要；胜率不够时，手续费会放大亏损。"),
            ("05:10", "人工改参数后胜率从 19% 降到 12%。", "没有验证框架的手动调参常常只是情绪化修补。"),
            ("05:36", "引入 Karpathy auto research：AI 提变更、测试、保留有效结果。", "把“修复”理解成实验自动化，而不是实盘自动加仓挽回。"),
            ("06:40", "系统使用 2024/2025 的 BTC、ETH、SOL 分钟级数据，生成策略并回测。", "需要训练/验证切分，不能让模型看见要测试的未来结果。"),
            ("08:16", "自动检查 look-ahead bias，拒绝“好得不真实”的策略。", "回测太漂亮经常是数据泄漏或过拟合。"),
            ("10:59", "演示如何用 Claude Code / OpenClaw / miniAutoResearch 思路搭建。", "学习重点是项目结构和询问澄清，不是照搬策略上实盘。"),
            ("16:22", "作者承认仍不知道真实实时数据上是否有效，几周后再看结果。", "这是视频里最重要的诚实边界：回测改善不等于实盘盈利。"),
        ],
        [0.72, 3.12, 2.44],
        size=7.9,
    )


def add_architecture(doc):
    doc.add_heading("3. Bot 系统结构", level=1)
    para(doc, "按视频描述，这个系统不是单一交易脚本，而是一个自动研究循环。可以把它拆成七层：")
    add_table(
        doc,
        ["层", "作用", "新手应重点检查"],
        [
            ("数据层", "准备 BTC、ETH、SOL 等两年分钟级行情；视频里提到 2024 与 2025 数据。", "数据缺口、交易所差异、时区、K 线拼接、手续费和滑点是否同步建模。"),
            ("训练/研究集", "用一段历史数据让 AI 设计候选策略或参数。", "不能把验证集结果泄露给生成策略的上下文。"),
            ("策略文件", "视频中对应 train.py，是 agent 反复编辑和测试的核心。", "限制 agent 只能改允许的文件；每次变更可回滚。"),
            ("指令文件", "视频中对应 program.md，约束 agent 的目标、规则和评价方式。", "写清楚目标不是最高收益，而是风险调整后的稳健表现。"),
            ("回测引擎", "对候选策略跑完整历史 candle，得到 P&L、Sharpe 等指标。", "必须包含手续费、滑点、爆仓/保证金规则、最小下单量。"),
            ("验证/拒绝器", "检测 look-ahead bias，拒绝结果好得不真实或差于当前最佳的策略。", "加入样本外、walk-forward、交易次数下限和最大回撤约束。"),
            ("日志与排名", "记录每一代策略、参数、分数、拒绝原因、当前最佳版本。", "没有可审计日志，就无法知道系统是在学习还是在碰运气。"),
        ],
        [0.95, 3.0, 2.33],
        size=7.85,
    )

    doc.add_heading("4. “自愈/修复亏损”到底是什么", level=1)
    label_para(doc, "不是", "亏损后自动加倍下注、自动追回本金、保证下一轮盈利、或者让 AI 直接拿交易权限去实盘试错。")
    label_para(doc, "是", "当旧策略表现变差时，系统自动提出小的策略变更，离线回测并验证，只有在指标更好且没有作弊/过拟合迹象时，才把它标记为更优候选。")
    label_para(doc, "关键边界", "视频中的收益数字主要来自历史回测或早期小资金实验。它能证明一个研究流程可以运行，不能证明任何策略未来会赚钱。")

    add_table(
        doc,
        ["说法", "工程上可接受的理解", "危险误读"],
        [
            ("self-healing", "自动发现当前策略问题并生成候选修复。", "亏损会自动消失。"),
            ("fixes its own losses", "通过回测和筛选减少重复犯错。", "实盘亏了能自动追回。"),
            ("learns the market", "从实验结果中更新策略候选和日志。", "AI 真正理解市场并能预测未来。"),
            ("keeps winners", "用预先定义指标保留更好回测版本。", "只要回测好就能实盘赚钱。"),
        ],
        [1.22, 3.0, 2.06],
        size=8.2,
    )


def add_execution_and_risk(doc):
    doc.add_heading("5. 策略执行流程", level=1)
    numbered_list(doc, [
        "准备历史市场数据，并固定训练集、验证集和样本外测试集。",
        "让 agent 根据 program.md 的目标生成或修改策略文件。",
        "运行回测，计算收益、回撤、Sharpe、交易次数、手续费占比等指标。",
        "自动检查异常：look-ahead bias、结果过于完美、交易次数太少、最大回撤超限、费用吞噬收益。",
        "把候选策略与当前最佳策略比较；只有通过所有门槛才进入候选榜。",
        "记录完整日志：版本、代码 diff、指标、拒绝原因、数据区间、运行时间。",
        "实盘前只允许 dry-run 或 paper trading；真正上线需要人工审批和小资金限额。",
    ])

    doc.add_heading("6. 风控要点", level=1)
    add_callout(
        doc,
        "最低风险线",
        "新手绝不应该把这种系统直接接到有大额资金的实盘账户。先用只读 API、历史回测、paper trading，再用极小资金；并且任何自动系统都必须能自动停机。",
        fill="FDEBD0",
        color="6B3D00",
    )
    add_table(
        doc,
        ["风险", "视频里的线索", "应该怎么防"],
        [
            ("手续费磨损", "814 笔交易中手续费约 -115 美元。", "设置交易频率上限、最小期望收益、手续费敏感性测试。"),
            ("市场状态切换", "作者怀疑原策略在市场反转后失效。", "分市场 regime 评估；上涨、下跌、震荡分别测试。"),
            ("手动乱调参", "胜率从 19% 掉到 12%。", "先定义实验目标与拒绝规则，再让系统自动测试，不凭感觉改。"),
            ("数据泄漏", "视频明确检查 look-ahead bias。", "严格切分数据；策略生成时不可看到验证结果；检测异常完美收益。"),
            ("过拟合", "回测上变好不代表未来有效。", "样本外测试、walk-forward、参数敏感性、少量实盘观察。"),
            ("权限过大", "教程涉及工具链搭建和可能的自动化交易。", "API key 只开必要权限；不开提现；IP 白名单；可随时撤销。"),
        ],
        [1.18, 2.08, 3.1],
        size=7.9,
    )

    doc.add_heading("7. 日志与监控", level=1)
    para(doc, "视频里的失败核心之一，是 bot 在几天内慢慢失血而人没有及时发现。一个合格的自动化交易系统至少需要这些监控：")
    for item in [
        "账户余额、净值、保证金使用率、未实现盈亏、已实现盈亏。",
        "每笔订单的信号来源、下单时间、成交时间、手续费、滑点、撤单状态。",
        "策略层指标：最近 N 笔胜率、平均盈亏比、最大连续亏损、手续费占净值比例。",
        "系统层指标：API 错误、超时、重复下单、交易所限速、数据延迟。",
        "停机条件：单日亏损超过阈值、手续费异常、连续失败订单、回撤超限、策略分数低于门槛。",
        "通知渠道：Telegram/Slack/邮件只是提醒，真正关键是自动降级和自动停止。"
    ]:
        bullet(doc, item)


def add_learning_and_warnings(doc):
    doc.add_heading("8. API 权限与实盘安全", level=1)
    add_table(
        doc,
        ["阶段", "API 权限", "理由"],
        [
            ("学习 / 数据抓取", "只读 public 或只读账户权限。", "先学行情、余额和日志，不暴露下单风险。"),
            ("回测", "不需要交易所私钥。", "历史数据可离线处理；不要为了回测开交易权限。"),
            ("paper / dry-run", "可使用模拟环境或只读真实行情。", "验证延迟、数据和策略生命周期，不真实下单。"),
            ("极小资金试运行", "交易权限可开；提现权限必须关闭。", "降低被盗、误下单和策略失控的损失上限。"),
            ("任何实盘阶段", "IP 白名单、权限最小化、可撤销、密钥不进代码仓库。", "自动化系统一旦泄露 key，损失通常不可逆。"),
        ],
        [1.32, 2.42, 2.06],
        size=8.1,
    )

    doc.add_heading("9. 哪些内容值得学习", level=1)
    for item in [
        "把“策略失败”拆成可观察指标：P&L、手续费、交易次数、胜率、回撤，而不是只看余额。",
        "用自动化研究循环替代情绪化调参：生成、测试、记录、拒绝、保留。",
        "理解训练集/验证集/样本外测试，尤其是 look-ahead bias 的危害。",
        "学习 agent 如何通过明确的 program.md 约束任务，而不是让它自由发挥。",
        "把 Claude Code / OpenClaw 当作工程助手：让它问清楚需求、搭建项目、解释代码，而不是直接托管资金。"
    ]:
        bullet(doc, item)

    doc.add_heading("10. 哪些话术要警惕", level=1)
    add_table(
        doc,
        ["话术", "为什么危险", "更稳妥的问法"],
        [
            ("“自动修复亏损”", "容易让人以为实盘亏损可被系统追回。", "它具体修复什么：参数、策略、数据问题，还是执行错误？"),
            ("“AI 正在学习市场”", "可能掩盖模型只是不断搜索历史上更好看的策略。", "是否有样本外、实时、低资金验证？"),
            ("“30 分钟搭好”", "搭好项目不等于搭好风控、监控和实盘系统。", "最小可运行版本有哪些停机条件？"),
            ("“回测 89% P&L”", "单一回测收益无法说明稳健性。", "最大回撤、手续费、滑点、交易次数和样本外表现是什么？"),
            ("“让 bot 自己迭代”", "如果没有限制，agent 可能改错文件、过拟合或扩大风险。", "agent 的可编辑范围、审批点和回滚机制是什么？"),
        ],
        [1.42, 2.58, 2.0],
        size=8.0,
    )

    doc.add_heading("11. 给 crypto 新手的实操路线", level=1)
    numbered_list(doc, [
        "只看公开视频和 repo，不连接交易所账户；把系统画成数据、策略、回测、验证、日志五部分。",
        "用历史数据复现一个最简单策略，并确认手续费和滑点已经进入回测。",
        "写一份 program.md，明确目标是风险调整收益和最大回撤控制，而不是最高收益。",
        "让 agent 只修改策略文件，并保存每次 diff、指标和拒绝原因。",
        "做 paper trading；观察真实行情延迟、数据缺口、订单失败和通知是否正常。",
        "如果一定要实盘，用极小资金、无提现权限 API key、每日亏损上限和手动总开关。",
    ])


def add_sources(doc):
    doc.add_heading("12. 来源与资料边界", level=1)
    para(doc, "本记录使用 BrowserOS YouTube MCP 读取了公开视频详情和完整 transcript。视频详情显示：频道 Sharbel A.，发布时间 2026-03-19，时长 18 分 55 秒；公开视频描述列出了 miniAutoResearch、OpenClaw、Claude Code 等链接。")
    para(doc, "本文没有验证作者的私有账户、真实交易流水或后续实盘结果；所有收益、亏损、手续费、胜率等数字均按视频 transcript 和公开视频描述记录。它们适合作为案例学习，不应作为投资建议或收益证明。")
    for label, url in SOURCES:
        source_para(doc, label, url)


def build():
    OUT.parent.mkdir(parents=True, exist_ok=True)
    doc = Document()
    setup_styles(doc)
    setup_header_footer(doc)
    add_opening(doc)
    add_summary(doc)
    add_architecture(doc)
    add_execution_and_risk(doc)
    add_learning_and_warnings(doc)
    add_sources(doc)
    doc.save(OUT)
    return OUT


if __name__ == "__main__":
    path = build()
    print(path)
