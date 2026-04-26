# 贡献指南

## 开发环境设置

```bash
# 1. Fork 并克隆仓库
git clone https://github.com/YOUR_USERNAME/ai4trading.git
cd ai4trading

# 2. 创建虚拟环境
python -m venv venv
source venv/bin/activate  # Linux/Mac
# or
.\venv\Scripts\activate   # Windows

# 3. 安装依赖
pip install -r requirements.txt

# 4. 创建 .env 文件
cp .env.example .env
```

## 分支命名规范

- `feature/` - 新功能 (e.g., `feature/mean-reversion-strategy`)
- `fix/` - Bug 修复 (e.g., `fix/order-execution-delay`)
- `refactor/` - 代码重构
- `docs/` - 文档更新

## 提交规范

```
type: 简短描述

可选的详细描述
```

Type 类型:
- `feat:` 新功能
- `fix:` Bug 修复
- `docs:` 文档更新
- `refactor:` 重构
- `test:` 测试相关
- `chore:` 构建/工具更新

## Pull Request 流程

1. 确保所有测试通过
2. 更新相关文档
3. PR 描述清楚改动内容和动机
4. 等待 code review

## 测试要求

- 所有新功能必须附带单元测试
- 测试覆盖率不应低于 80%
- 运行测试: `pytest tests/`
