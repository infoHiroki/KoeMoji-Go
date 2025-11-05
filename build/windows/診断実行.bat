@echo off
chcp 65001 >nul
echo ====================================
echo   KoeMoji-Go 診断ツール
echo ====================================
echo.

"%~dp0koemoji-go.exe" --doctor > "%~dp0診断結果.txt" 2>&1

if %errorlevel% equ 0 (
    echo [成功] 診断が完了しました
) else (
    echo [エラー] 診断中に問題が発生しました
)

echo.
echo 診断結果は「診断結果.txt」に保存されました
echo このファイルをサポートに送付してください
echo.
pause
