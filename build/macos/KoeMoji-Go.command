#!/bin/bash

# KoeMoji-Go 起動スクリプト
# このスクリプトは初回起動時に検疫属性を自動的に削除します

cd "$(dirname "$0")"

# 初回起動時のみ検疫属性を削除
if [ -f ".first_run_done" ]; then
    # 2回目以降は直接起動
    ./koemoji-go
else
    # 初回起動：検疫属性削除
    echo "===================================="
    echo "  KoeMoji-Go 初回起動"
    echo "===================================="
    echo ""
    echo "セキュリティ属性を削除しています..."
    xattr -cr . 2>/dev/null || true
    touch .first_run_done
    echo "完了しました。アプリを起動します..."
    echo ""

    # GUIモードで起動
    ./koemoji-go
fi
