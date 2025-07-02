# Windows環境でのFasterWhisper問題と修正方針

最終更新: 2025-07-02

## 問題概要

Windows環境（特にGPU搭載PC）でKoeMoji-GoのFasterWhisper統合が正常に動作しない問題を確認しました。

### 確認された問題

1. **Windows実行ファイルパスの未対応**
   - `getWhisperCommand()` がWindows環境での whisper-ctranslate2.exe を検索できない
   - Pythonのインストール方法により実行ファイルの場所が異なる

2. **GPU環境でのcompute_type不整合**
   - `compute_type: int8` (CPU向け) 設定時でもGPUが自動選択される
   - `--device` パラメータ未実装により明示的なデバイス指定ができない

## 修正方針（KISS原則）

### 1. Windowsパス検索の追加
- `exec.LookPath` で見つからない場合のみ追加検索
- `.exe` 拡張子を考慮した検索

### 2. デバイス指定の自動化
- `compute_type` が `int8` の場合は `--device cpu` を自動追加
- それ以外は whisper-ctranslate2 の自動選択に任せる

既存の設定を変更せず、最小限の修正で問題を解決します。