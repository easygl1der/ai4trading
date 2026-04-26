# Polymarket Auto Trading System

> 自动化套利系统 | 数据分析 | 策略模拟

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 项目概述

本项目是一个 Polymarket 自动交易系统，包含以下核心模块：

| 模块 | 说明 |
|------|------|
| **Trading System** | 市场数据获取、订单执行、仓位管理 |
| **Data Analysis** | 市场分析、概率聚合、相关性分析 |
| **Strategy Simulation** | 策略回测、模拟交易、绩效评估 |

## 核心技能 (Project Skills)

### Trading Module
- [ ] `fetch_markets` - 获取市场数据
- [ ] `get_order_book` - 订单簿查询
- [ ] `place_order` - 下单执行
- [ ] `get_positions` - 仓位查询
- [ ] `cancel_order` - 取消订单

### Analysis Module
- [ ] `aggregate_probs` - 概率聚合分析
- [ ] `calc_correlation` - 相关性计算
- [ ] `spread_analysis` - 价差分析
- [ ] `liquidity_analysis` - 流动性分析

### Simulation Module
- [ ] `backtest` - 策略回测
- [ ] `paper_trade` - 模拟交易
- [ ] `perf_metrics` - 绩效指标计算
- [ ] `equity_curve` - 权益曲线绘制

## 快速开始

```bash
# 克隆仓库
git clone https://github.com/easygl1der/ai4trading.git
cd ai4trading

# 安装依赖
pip install -r requirements.txt

# 配置 API Keys
cp .env.example .env
# 编辑 .env 填写你的 Polymarket API keys

# 运行示例
python scripts/demo.py
```

## 项目结构

```
ai4trading/
├── src/
│   ├── trading/        # 交易执行模块
│   ├── analysis/       # 数据分析模块
│   ├── simulation/     # 策略模拟模块
│   └── utils/          # 工具函数
├── tests/              # 单元测试
├── docs/               # 文档
├── scripts/            # 示例脚本
└── configs/            # 配置文件
```

## 技术栈

- **语言**: Python 3.10+
- **数据处理**: pandas, numpy
- **回测引擎**: backtrader
- **API通信**: requests
- **配置管理**: python-dotenv

## 贡献指南

详细贡献流程请查看 [CONTRIBUTING.md](CONTRIBUTING.md)

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/your-feature`)
3. 提交更改 (`git commit -m 'feat: Add some feature'`)
4. 推送到分支 (`git push origin feature/your-feature`)
5. 创建 Pull Request

## 协作成员

| 角色 | 职责 |
|------|------|
| Owner | 项目主导、技术决策 |
| Contributor | 功能开发、Bug修复 |

## 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 免责声明

本项目仅供教育和研究目的。实际交易存在风险，请自行承担后果。
