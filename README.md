# Polymarket Auto Trading System

> 自动化套利系统 | 数据分析 | 策略模拟

## 项目概述

本项目是一个 Polymarket 自动交易系统，包含以下核心模块：

- **Trading System** - 市场数据获取、订单执行、仓位管理
- **Data Analysis** - 市场分析、概率聚合、相关性分析
- **Strategy Simulation** - 策略回测、模拟交易、绩效评估

## 快速开始

```bash
# 克隆仓库
git clone https://github.com/YOUR_USERNAME/ai4trading.git
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

## 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/your-feature`)
3. 提交更改 (`git commit -m 'Add some feature'`)
4. 推送到分支 (`git push origin feature/your-feature`)
5. 创建 Pull Request

### 开发规范

- 代码风格：遵循 PEP 8
- 提交信息：使用中文描述，格式 `type: description`
- 所有新功能需要附带测试
- PR 需要通过 CI 检查

## 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 免责声明

本项目仅供教育和研究目的。实际交易存在风险，请自行承担后果。
