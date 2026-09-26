import json
from pathlib import Path


OUT = Path("/Users/yitwah/Documents/Codex/2026-06-07/video-youtube-video-open-document-recently/outputs/chain_relationships_explained.excalidraw")


elements = []
seed = 1000


def next_seed():
    global seed
    seed += 1
    return seed


def rect(id_, x, y, w, h, stroke="#1e3a5f", fill="#dbeafe", sw=2, dash=False):
    e = {
        "id": id_, "type": "rectangle", "x": x, "y": y, "width": w, "height": h,
        "angle": 0, "strokeColor": stroke, "backgroundColor": fill, "fillStyle": "solid",
        "strokeWidth": sw, "strokeStyle": "dashed" if dash else "solid",
        "roughness": 0, "opacity": 100, "seed": next_seed(), "version": 1, "versionNonce": next_seed(),
        "isDeleted": False, "groupIds": [], "boundElements": [], "link": None, "locked": False,
        "roundness": {"type": 3},
    }
    elements.append(e)
    return e


def ellipse(id_, x, y, w, h, stroke="#1e3a5f", fill="#dbeafe", sw=2):
    e = {
        "id": id_, "type": "ellipse", "x": x, "y": y, "width": w, "height": h,
        "angle": 0, "strokeColor": stroke, "backgroundColor": fill, "fillStyle": "solid",
        "strokeWidth": sw, "strokeStyle": "solid", "roughness": 0, "opacity": 100,
        "seed": next_seed(), "version": 1, "versionNonce": next_seed(), "isDeleted": False,
        "groupIds": [], "boundElements": [], "link": None, "locked": False,
    }
    elements.append(e)
    return e


def text(id_, x, y, w, h, content, size=18, color="#374151", align="center", container=None, bold=False):
    e = {
        "id": id_, "type": "text", "x": x, "y": y, "width": w, "height": h,
        "angle": 0, "strokeColor": color, "backgroundColor": "transparent", "fillStyle": "solid",
        "strokeWidth": 1, "strokeStyle": "solid", "roughness": 0, "opacity": 100,
        "seed": next_seed(), "version": 1, "versionNonce": next_seed(), "isDeleted": False,
        "groupIds": [], "boundElements": None, "link": None, "locked": False,
        "text": content, "originalText": content, "fontSize": size, "fontFamily": 3,
        "textAlign": align, "verticalAlign": "middle", "containerId": container,
        "lineHeight": 1.25,
    }
    if bold:
        e["fontSize"] = size
    elements.append(e)
    if container:
        for el in elements:
            if el["id"] == container:
                el.setdefault("boundElements", []).append({"id": id_, "type": "text"})
                break
    return e


def arrow(id_, x, y, points, color="#1e3a5f", dashed=False, label=None, label_pos=None):
    xs = [p[0] for p in points]
    ys = [p[1] for p in points]
    e = {
        "id": id_, "type": "arrow", "x": x, "y": y, "width": max(xs) - min(xs),
        "height": max(ys) - min(ys), "angle": 0, "strokeColor": color,
        "backgroundColor": "transparent", "fillStyle": "solid", "strokeWidth": 2,
        "strokeStyle": "dashed" if dashed else "solid", "roughness": 0, "opacity": 100,
        "seed": next_seed(), "version": 1, "versionNonce": next_seed(), "isDeleted": False,
        "groupIds": [], "boundElements": None, "link": None, "locked": False,
        "points": points, "startBinding": None, "endBinding": None,
        "startArrowhead": None, "endArrowhead": "arrow",
    }
    elements.append(e)
    if label and label_pos:
        text(id_ + "_label", label_pos[0], label_pos[1], label_pos[2], label_pos[3], label, 14, color, "center")
    return e


# Title and legend
text("title", 70, 35, 1100, 48, "不同链之间是什么关系？为什么转账会有不同手续费？", 34, "#1e40af", "left")
text("subtitle", 73, 86, 1040, 34, "核心心智模型：每条链都是一套独立账本；跨链不是“搬同一个币”，而是锁定/销毁/发行映射资产。", 18, "#64748b", "left")

# Left: CEX
rect("cex_box", 70, 160, 285, 270, "#c2410c", "#fed7aa")
text("cex_title", 90, 178, 245, 34, "中心化交易所 CEX", 22, "#7c2d12", "center")
text("cex_detail", 95, 225, 235, 140, "例：Bitget / Binance / OKX\n\n你看到的是平台账户余额\n撮合、风控、提现由平台控制\n\n优点：快、流动性强、适合新手\n风险：平台冻结/破产/被盗", 15, "#374151", "center")
rect("cex_internal", 112, 365, 200, 42, "#c2410c", "#fff7ed")
text("cex_internal_text", 123, 374, 178, 25, "平台内部转账 ≠ 链上交易", 14, "#7c2d12", "center")

# Center wallet
ellipse("wallet", 505, 210, 190, 110, "#047857", "#a7f3d0")
text("wallet_text", 530, 230, 140, 62, "你的钱包\n私钥/助记词\n控制资产", 18, "#064e3b", "center")

# Chain rooms
rect("eth_room", 820, 145, 290, 150, "#1e3a5f", "#dbeafe")
text("eth_title", 840, 160, 250, 30, "Ethereum 主网", 20, "#1e3a5f", "center")
text("eth_body", 845, 197, 240, 70, "安全和生态强\ngas 通常较高\n适合大额/DeFi/NFT", 15, "#374151", "center")

rect("sol_room", 820, 330, 290, 150, "#047857", "#a7f3d0")
text("sol_title", 840, 345, 250, 30, "Solana", 20, "#064e3b", "center")
text("sol_body", 845, 382, 240, 70, "高吞吐、低手续费\n账户模型不同\n适合高频小额应用", 15, "#374151", "center")

rect("l2_room", 1160, 145, 295, 150, "#6d28d9", "#ddd6fe")
text("l2_title", 1180, 160, 255, 30, "Ethereum L2", 20, "#4c1d95", "center")
text("l2_body", 1182, 197, 250, 70, "Arbitrum / Optimism / Base\n继承部分 ETH 生态\n手续费通常低于主网", 15, "#374151", "center")

rect("bnb_room", 1160, 330, 295, 150, "#b45309", "#fef3c7")
text("bnb_title", 1180, 345, 255, 30, "BNB Chain / Polygon 等", 20, "#78350f", "center")
text("bnb_body", 1182, 382, 250, 70, "更便宜、更快\n生态和安全模型不同\n假币/钓鱼更多见", 15, "#374151", "center")

# Arrows from wallet to chains
arrow("w_eth", 690, 250, [[0, 0], [125, -35]], "#047857", label="链上转账：付 gas", label_pos=(705, 198, 150, 24))
arrow("w_sol", 690, 280, [[0, 0], [125, 100]], "#047857", label="选择 Solana 网络", label_pos=(700, 332, 155, 24))
arrow("w_l2", 690, 245, [[0, 0], [462, -45]], "#047857", dashed=True, label="跨到 L2 通常要桥", label_pos=(900, 176, 180, 24))
arrow("w_bnb", 690, 290, [[0, 0], [462, 110]], "#047857", dashed=True, label="不同链地址格式/规则不同", label_pos=(885, 438, 240, 24))

# CEX interactions
arrow("cex_to_wallet", 355, 260, [[0, 0], [145, 0]], "#c2410c", label="提现：选择网络", label_pos=(372, 220, 125, 24))
arrow("wallet_to_cex", 505, 305, [[0, 0], [-150, 85]], "#c2410c", dashed=True, label="充值：地址 + 网络必须匹配", label_pos=(352, 355, 165, 24))

# Bridge / wrapped assets
rect("bridge", 510, 500, 520, 125, "#b45309", "#fef3c7")
text("bridge_title", 535, 512, 470, 28, "跨链桥 Bridge / 交易所提现网络", 21, "#78350f", "center")
text("bridge_body", 540, 550, 455, 52, "把 A 链资产锁定/销毁，再在 B 链释放/铸造映射资产\n跨链成本 = A 链 gas + 桥/平台费用 + 等待时间 + 桥安全风险", 15, "#374151", "center")
arrow("eth_bridge", 960, 295, [[0, 0], [-80, 200]], "#b45309", dashed=True)
arrow("sol_bridge", 920, 480, [[0, 0], [-50, 25]], "#b45309", dashed=True)
arrow("bridge_l2", 1030, 530, [[0, 0], [150, -265]], "#b45309", dashed=True)
arrow("bridge_bnb", 1030, 570, [[0, 0], [150, -135]], "#b45309", dashed=True)

# Fee explanation bottom
rect("fee_panel", 70, 675, 1385, 205, "#1e3a5f", "#f8fafc")
text("fee_title", 95, 692, 330, 32, "为什么手续费不同？", 24, "#1e40af", "left")

rect("fee_1", 105, 745, 255, 90, "#1e3a5f", "#dbeafe")
text("fee_1_text", 120, 758, 225, 62, "1. 每条链的规则不同\n区块空间、共识机制、账户模型\n决定基础成本", 15, "#374151", "center")

rect("fee_2", 405, 745, 255, 90, "#1e3a5f", "#dbeafe")
text("fee_2_text", 420, 758, 225, 62, "2. 拥堵程度不同\n同一条链在牛市/NFT/铭文/空投时\n费用会突然变高", 15, "#374151", "center")

rect("fee_3", 705, 745, 255, 90, "#1e3a5f", "#dbeafe")
text("fee_3_text", 720, 758, 225, 62, "3. 跨链更复杂\n不是单次转账\n可能经过桥、包装资产、确认等待", 15, "#374151", "center")

rect("fee_4", 1005, 745, 405, 90, "#dc2626", "#fee2e2")
text("fee_4_text", 1020, 758, 375, 62, "4. 网络选错会丢资产\nUSDT-ERC20 / USDT-TRC20 / USDT-Solana\n名字像，但账本不是同一个", 15, "#7f1d1d", "center")

# Summary rule
rect("rule", 70, 925, 1385, 95, "#047857", "#ecfdf5")
text("rule_text", 95, 943, 1335, 45, "最实用判断：如果你在 Bitget App 内部买卖，那是中心化账本；如果你从 Bitget 提现到钱包，并在某条链上签名交易，那就是链上操作。链选错、地址错、memo/tag 漏填，都可能造成资产损失。", 20, "#064e3b", "center")

data = {
    "type": "excalidraw",
    "version": 2,
    "source": "https://excalidraw.com",
    "elements": elements,
    "appState": {"gridSize": None, "viewBackgroundColor": "#ffffff"},
    "files": {},
}

OUT.parent.mkdir(parents=True, exist_ok=True)
OUT.write_text(json.dumps(data, ensure_ascii=False, indent=2))
print(OUT)
