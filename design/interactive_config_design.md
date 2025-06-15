# KoeMoji-Go 対話型設定機能 設計書

## 概要
JSONファイルの手動編集を不要にし、対話型UIで全ての設定を変更可能にする。

## コマンド
```bash
./koemoji-go --config
```

## 対話型UI仕様

### メイン画面
```
=== KoeMoji-Go Configuration ===

1. Whisper Model: medium
2. Language: ja
3. Output Format: txt
4. Scan Interval: 10 minutes
5. Max CPU: 95%
6. Compute Type: int8
7. Use Colors: true
8. UI Mode: enhanced
9. Input Directory: ./input
10. Output Directory: ./output
11. Archive Directory: ./archive

s: Save and exit
q: Quit without saving
Select number to change: 
```

### 各設定項目の変更UI

#### 1. Whisper Model
```
Select Whisper Model:
1. tiny (fastest, lowest quality)
2. base (fast, low quality)
3. small (balanced)
4. medium (recommended) ✓
5. large (best quality, slowest)
Choice: 
```

#### 2. Language
```
Select Language:
1. ja (Japanese) ✓
2. en (English)
3. auto (Auto-detect)
4. Other (enter manually)
Choice: 
```

#### 3. Output Format
```
Select Output Format:
1. txt (Plain text) ✓
2. srt (Subtitle)
3. vtt (WebVTT)
4. all (All formats)
Choice: 
```

#### 9-11. Directory設定
```
Input Directory: ./input

1. Enter path manually
2. Browse with file dialog
Choice: 2

[OS標準フォルダ選択ダイアログが開く]

✓ Selected: /Users/user/Documents/audio
```

## 実装詳細

### Config構造体の拡張
```go
type Config struct {
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

### フォルダ選択機能
```go
func selectFolderWithDialog(prompt string) (string, error) {
    switch runtime.GOOS {
    case "darwin":
        return selectFolderMac(prompt)
    case "windows":
        return selectFolderWindows(prompt)
    default:
        return "", fmt.Errorf("GUI not supported on %s", runtime.GOOS)
    }
}
```

### デフォルト値
```json
{
    "whisper_model": "medium",
    "language": "ja",
    "output_format": "txt",
    "scan_interval_minutes": 10,
    "max_cpu_percent": 95,
    "compute_type": "int8",
    "use_colors": true,
    "ui_mode": "enhanced",
    "input_dir": "./input",
    "output_dir": "./output",
    "archive_dir": "./archive"
}
```

## エラーハンドリング

1. **無効な入力**: 再入力を促す
2. **GUI利用不可**: 手動入力にフォールバック
3. **権限エラー**: エラーメッセージ表示して別パス選択を促す

## 実装優先順位

1. 基本的な対話型UI実装
2. Config構造体にフォルダパス追加
3. macOS/Windows向けフォルダ選択ダイアログ
4. 既存コードのフォルダパス対応