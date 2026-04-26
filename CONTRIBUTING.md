# 贡献指南 | Contribution Guide

## 开发环境设置

```bash
# 1. Fork 并克隆仓库
git clone https://github.com/easygl1der/ai4trading.git
cd ai4trading

# 2. 创建虚拟环境
python -m venv venv
source venv/bin/activate  # Linux/Mac
.\venv\Scripts\activate   # Windows

# 3. 安装依赖
pip install -r requirements.txt

# 4. 创建 .env 文件
cp .env.example .env
```

## 项目技能清单 (Project Skills)

贡献者可以通过认领和完成技能来推进项目进度。

### Trading 模块技能

| 技能ID | 技能名称 | 描述 | 状态 |
|--------|----------|------|------|
| `T-001` | `fetch_markets` | 获取 Polymarket 活跃市场列表 | ⬜ |
| `T-002` | `get_order_book` | 获取指定市场的订单簿数据 | ⬜ |
| `T-003` | `place_order` | 执行市价/限价下单 | ⬜ |
| `T-004` | `get_positions` | 查询当前仓位 | ⬜ |
| `T-005` | `cancel_order` | 取消 pending 订单 | ⬜ |

### Analysis 模块技能

| 技能ID | 技能名称 | 描述 | 状态 |
|--------|----------|------|------|
| `A-001` | `aggregate_probs` | 聚合同一事件的多市场概率 | ⬜ |
| `A-002` | `calc_correlation` | 计算市场间相关性矩阵 | ⬜ |
| `A-003` | `spread_analysis` | 检测价差套利机会 | ⬜ |
| `A-004` | `liquidity_analysis` | 流动性评估 | ⬜ |

### Simulation 模块技能

| 技能ID | 技能名称 | 描述 | 状态 |
|--------|----------|------|------|
| `S-001` | `backtest` | 策略回测框架 | ⬜ |
| `S-002` | `paper_trade` | 模拟交易执行器 | ⬜ |
| `S-003` | `perf_metrics` | 夏普比率、最大回撤等指标 | ⬜ |

## 分支命名规范

```
feature/T-001-fetch-markets      # 认领技能 T-001
feature/add-spread-detection     # 功能开发
fix/order-execution-delay        # Bug修复
refactor/analysis-module         # 重构
docs/update-api-reference        # 文档更新
```

## 提交规范 (Conventional Commits)

```
type(scope): description

# Example:
feat(trading): add fetch_markets function
fix(analysis): correct probability aggregation logic
docs(readme): update project skills checklist
```

### Type 类型

| Type | 说明 |
|------|------|
| `feat` | 新功能 |
| `fix` | Bug修复 |
| `docs` | 文档更新 |
| `refactor` | 重构 |
| `test` | 测试相关 |
| `chore` | 构建/工具更新 |

## Pull Request 流程

1. **Fork** - 从主仓库 Fork 到个人账户
2. **Clone** - `git clone https://github.com/YOUR_USERNAME/ai4trading.git`
3. **创建分支** - `git checkout -b feature/T-001-fetch-markets`
4. **开发** - 完成技能对应的功能代码和测试
5. **提交** - `git commit -m 'feat(trading): implement fetch_markets'`
6. **推送** - `git push origin feature/T-001-fetch-markets`
7. **PR** - 创建 Pull Request 到 `easygl1der:main`

### PR 描述模板

```markdown
## 技能关联
- 关联技能: T-001 `fetch_markets`

## 改动内容
- 新增 `src/trading/markets.py`
- 实现 `fetch_markets()` 函数
- 添加单元测试 `tests/test_markets.py`

## 测试结果
- [ ] 本地测试通过
- [ ] 代码风格检查通过

## 截图/输出 (如有)
```

## 代码规范

- 遵循 **PEP 8** 代码风格
- 使用 **type hints** 类型注解
- 所有公共函数需要 **docstring**
- 新功能必须附带 **单元测试**

## 测试要求

```bash
# 运行所有测试
pytest tests/

# 运行带覆盖率的测试
pytest tests/ --cov=src --cov-report=html
```

## Review 清单

PR 创建后请确认：

- [ ] 代码风格检查通过 (`flake8`)
- [ ] 单元测试全部通过 (`pytest`)
- [ ] 文档已更新（如有 API 变更）
- [ ] PR 描述清晰完整

## 协作沟通

- 使用 GitHub Issues 讨论技术问题
- 使用 Pull Request 进行代码 review
- Tag 项目 Owner 进行最终合并

---

**认领技能 → 完成开发 → 提交 PR → 合并上线**
