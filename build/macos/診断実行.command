#!/bin/bash
cd "$(dirname "$0")"
echo "===================================="
echo "  KoeMoji-Go 診断ツール"
echo "===================================="
echo ""

./koemoji-go --doctor > 診断結果.txt 2>&1

if [ $? -eq 0 ]; then
    echo "[成功] 診断が完了しました"
else
    echo "[エラー] 診断中に問題が発生しました"
fi

echo ""
echo "診断結果は「診断結果.txt」に保存されました"
echo "このファイルをサポートに送付してください"
echo ""
read -p "Enterキーで終了..."
