#!/usr/bin/env python3
"""Demo script to test the trading system setup"""

import sys
sys.path.insert(0, 'src')

def main():
    print("=" * 50)
    print("Polymarket Auto Trading System")
    print("=" * 50)
    print("\nModules loaded successfully!")
    print("\nProject structure:")
    print("  src/trading    - 交易执行模块")
    print("  src/analysis   - 数据分析模块")
    print("  src/simulation - 策略模拟模块")
    print("  src/utils      - 工具函数")
    print("\n下一步:")
    print("  1. 配置 .env 文件添加 API keys")
    print("  2. 查看 docs/ 目录下的文档")
    print("  3. 开始开发你的交易策略!")

if __name__ == "__main__":
    main()
