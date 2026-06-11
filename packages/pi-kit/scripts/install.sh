#!/usr/bin/env bash
#
# 安装 @xkit/pi-kit 到 pi 全局
#
# Usage:
#   bash scripts/install.sh              # 本地安装
#   bash scripts/install.sh --remove     # 卸载

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PKG_NAME="@xkit/pi-kit"

if [ "${1:-}" = "--remove" ]; then
  echo "🗑️  Removing $PKG_NAME ..."
  pi remove "$ROOT"
  echo "✅ 已移除"
  exit 0
fi

echo "🔧 Installing $PKG_NAME from $ROOT ..."
pi install "$ROOT"
echo "✅ 安装成功"
echo ""
echo "包含的扩展:"
echo "  safe-run   — 命令确认与路径保护"
echo "  daily      — daily memo 写入工具"
echo ""
echo "日志文件: ~/.pi/agent/safe-run.log"
