# KoeMoji-Go Web UI 設計書

## 1. 概要
既存のCLI機能をWebブラウザから操作できるようにする。
またいくつかの追加機能を実装する。

## 2. 機能一覧

### 2.1 設定管理（config.json）
- 現在の設定表示（CLIの`c`コマンド相当）
- 設定の編集・保存
- フォルダパスの設定（新機能）

### 2.2 基本操作
- 手動スキャン実行（CLIの`s`コマンド相当）
- ログ表示（CLIの`l`コマンド相当）

## 3. API設計

```
GET  /api/config      # 設定取得
PUT  /api/config      # 設定更新

POST /api/scan        # 手動スキャン実行
GET  /api/logs        # ログファイル内容取得
```

## 4. UI画面

### 4.1 設定画面
```
Whisperモデル: [medium ▼]
言語: [ja ▼]
出力形式: [txt ▼]
スキャン間隔: [10] 分
最大CPU使用率: [95] %
計算タイプ: [int8 ▼]

--- フォルダ設定 ---
Inputフォルダ:  [./input] [選択]
Outputフォルダ: [./output] [選択]
Archiveフォルダ: [./archive] [選択]

[保存] [リセット]
```

### 4.2 操作画面
```
[手動スキャン実行]
[ログを表示]
```

## 5. config.json の拡張

既存の設定項目に加えて、以下の3つのフィールドを追加：

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
  "input_dir": "./input",      // 新規追加
  "output_dir": "./output",     // 新規追加
  "archive_dir": "./archive"    // 新規追加
}
```

## 6. 実装方針

### 6.1 技術スタック
- Go標準ライブラリのみ使用（net/http、embed）
- フロントエンド: Vanilla HTML/CSS/JavaScript
- 外部依存なし
- Go 1.16以降のembed機能で静的ファイルを埋め込み

### 6.2 起動オプション
```bash
koemoji-go --web        # Web UIを有効化（デフォルトポート: 8080）
koemoji-go --web-port 8081  # ポート指定
```

### 6.3 セキュリティ
- デフォルトでlocalhostのみリッスン
- フォルダ選択時のパストラバーサル対策
- 相対パス・絶対パスの両方に対応

### 6.4 後方互換性
- フォルダ設定が存在しない場合はデフォルト値を使用
- 既存のCLI動作に影響なし

## 7. 実装優先順位

1. **Phase 1**: 基本実装
   - HTTPサーバー追加
   - 設定の読み取り・保存API
   - 最小限のHTML UI

2. **Phase 2**: 機能追加
   - フォルダ選択機能
   - 手動スキャン実行
   - ログ表示

3. **Phase 3**: UI改善
   - エラーハンドリング
   - 設定値のバリデーション
   - レスポンシブデザイン

## 8. ファイル構成とビルド

### 8.1 開発時のファイル構成
```
KoeMoji-Go/
├── main.go          # Web サーバー機能を追加
├── web/             # 静的ファイル（embedで埋め込まれる）
│   ├── index.html
│   ├── style.css
│   └── script.js
└── config.json      # 拡張されたフィールドを含む
```

### 8.2 ビルドと配布
```go
//go:embed web/*
var webFiles embed.FS
```

- `web/` ディレクトリの内容は実行ファイルに埋め込まれる
- 配布時は単一の実行ファイル（koemoji-go）のみで動作
- config.jsonは外部ファイルとして残す（ユーザーが編集可能）

### 8.3 実装例
```go
// Webサーバーの静的ファイル配信
http.Handle("/", http.FileServer(http.FS(webFiles)))
```