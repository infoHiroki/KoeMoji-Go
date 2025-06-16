# KoeMoji-Go 対話型設定機能 - 最終設計

## 実装範囲
1. **対話型設定UI** (`--config`フラグ)
2. **フォルダパス設定のJSON対応**
3. **OS標準フォルダ選択ダイアログ** (macOS/Windows)
4. **Linux非対応**

## 解決される問題
- ❌ JSON手動編集 → ✅ 対話型UI
- ❌ パス入力エラー → ✅ フォルダ選択ダイアログ
- ❌ 設定が難しい → ✅ 直感的な選択式

## 実装する機能

### 1. Config構造体の拡張
```go
type Config struct {
    // 既存フィールド
    WhisperModel        string `json:"whisper_model"`
    Language            string `json:"language"`
    OutputFormat        string `json:"output_format"`
    ScanIntervalMinutes int    `json:"scan_interval_minutes"`
    MaxCpuPercent       int    `json:"max_cpu_percent"`
    ComputeType         string `json:"compute_type"`
    UseColors           bool   `json:"use_colors"`
    UIMode              string `json:"ui_mode"`
    
    // 新規追加
    InputDir            string `json:"input_dir"`
    OutputDir           string `json:"output_dir"`
    ArchiveDir          string `json:"archive_dir"`
}
```

### 2. コマンドライン
```bash
# 対話型設定
./koemoji-go --config

# 従来の監視モード（変更なし）
./koemoji-go

# ヘルプ表示に--config追加
./koemoji-go --help
```

### 3. 対話型UIフロー
```
=== KoeMoji-Go Configuration ===

現在の設定を番号で選択して変更:
1-11: 各設定項目
s: 保存して終了
q: 保存せずに終了

フォルダ選択時:
- macOS/Windows: GUIダイアログ or 手動入力
- 無効なパス: エラー表示して再入力
```

### 4. フォルダ選択実装
```go
// macOS: osascript (標準搭載)
// Windows: PowerShell (標準搭載)
// Linux: 非対応

func selectFolderWithDialog(prompt string) (string, error) {
    switch runtime.GOOS {
    case "darwin":
        // osascript実装
    case "windows":
        // PowerShell実装
    default:
        return "", errors.New("not supported")
    }
}
```

## 実装しない機能
- ❌ Web UI
- ❌ ファイル引数での単発処理
- ❌ Linux対応
- ❌ 複雑なCLIフラグ群

## 期待される効果
1. **ライトユーザー**: JSON編集不要で設定可能
2. **パス設定**: GUIダイアログで確実に設定
3. **学習コスト**: 対話型で直感的
4. **後方互換性**: 既存のconfig.jsonも動作

## 実装の複雑度
- **低**: 既存コードへの影響最小限
- **中**: 対話型UI部分の実装
- **低**: OS標準機能の呼び出し

## 次のステップ
1. Config構造体にフォルダパス追加
2. 対話型設定関数の実装
3. フォルダ選択ダイアログの実装
4. 既存コードのパス参照を更新