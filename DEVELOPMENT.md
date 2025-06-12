# KoeMoji-Go 開発メモ

## 開発履歴

### Phase 1: 日本語版実装（完了）
- **ブランチ**: `feature/japanese-version`
- **完了日**: 2025/06/12
- **内容**:
  - エラーメッセージの日本語化（19箇所）
  - ログメッセージの日本語化（18箇所）
  - UI/ヘルプテキストの日本語化（33箇所）
  - コマンドラインフラグの日本語化（4箇所）
  - 合計74箇所を翻訳

### 字幕機能実装（完了）
- **ブランチ**: `feature/subtitle-support`
- **PR**: #1
- **完了日**: 2025/06/12
- **内容**:
  - `OutputFormat`設定を追加
  - txt, srt, vtt, tsv, json, all形式をサポート
  - デフォルトは'txt'で後方互換性を維持
  - 日本語版にもマージ済み

## 今後の開発プラン

### 短期（1週間以内）

#### 1. 処理完了通知
```go
// macOS通知の例
cmd := exec.Command("osascript", "-e", 
    fmt.Sprintf(`display notification "%s" with title "KoeMoji-Go"`, message))
```
- ファイル処理完了時の通知
- 設定で有効/無効を切り替え可能
- 音声再生オプション

#### 2. 簡易Web UI
- `net/http`でシンプルなWebサーバー
- リアルタイム進捗表示（WebSocket）
- 結果のプレビューとダウンロード

#### 3. バッチ処理の最適化
- ワーカープールパターンで並列処理
- GPU使用時の自動調整
- 進捗バーの表示

### 中期（1ヶ月以内）

#### Phase 2: 多言語対応
```go
type Messages struct {
    Processing string
    Completed  string
    // 他のメッセージ
}

var messages = map[string]Messages{
    "ja": { /* 日本語 */ },
    "en": { /* 英語 */ },
}
```
- 設定ベースの言語切り替え
- 翻訳の外部ファイル化（JSON）
- 動的な言語追加対応

### 長期（3ヶ月以内）

#### 1. クラウド連携
- Google Drive API統合
- Dropbox API統合
- WebHook通知機能

#### 2. 話者分離
- pyannoteライブラリとの連携
- 話者ごとのテキスト分離
- 議事録形式での出力

#### 3. 品質向上機能
- 自動リトライメカニズム
- 文字起こし精度の統計表示
- カスタム辞書機能（固有名詞対応）

## 技術的な課題と解決策

### 文字化け問題
- **現状**: GoのUTF-8処理により問題なし
- **注意点**: Windows環境でのファイル名
- **対策**: 現状維持（必要に応じて対応）

### セキュリティ対策
- **実装済み**: inputディレクトリ制限（コマンドインジェクション対策）
- **今後**: ファイル名のサニタイズ強化

### パフォーマンス
- **課題**: 大量ファイル処理時の効率
- **対策案**: 
  - ワーカープール実装
  - ファイルキャッシュ機構
  - 処理済みファイルのDB管理

## ブランチ戦略

```
main (英語版・安定版)
├── feature/subtitle-support (字幕機能) → PR#1
├── feature/japanese-version (日本語版・独立運用)
└── feature/web-ui (今後実装予定)
```

### マージ方針
1. 新機能は`main`ブランチで開発
2. 日本語版には選択的にマージ
3. 破壊的変更は慎重に検討

## コマンド備忘録

### ビルド
```bash
# 英語版
go build -o koemoji-go main.go

# 日本語版
git checkout feature/japanese-version
go build -o koemoji-go-ja main.go
```

### テスト実行
```bash
# デバッグモードで実行
./koemoji-go-ja --debug

# 設定確認
./koemoji-go-ja --help
```

### Git操作
```bash
# PRの作成（GitHub CLI）
gh pr create --title "タイトル" --body "説明"

# ブランチ間のマージ
git checkout feature/japanese-version
git merge feature/subtitle-support
```

## 設定ファイル例

### 基本設定
```json
{
    "whisper_model": "medium",
    "language": "ja",
    "output_format": "txt",
    "scan_interval_minutes": 10,
    "max_cpu_percent": 95,
    "compute_type": "int8",
    "use_colors": true,
    "ui_mode": "enhanced"
}
```

### 字幕生成設定
```json
{
    "output_format": "all",  // 全形式を出力
    "whisper_model": "large",  // 高精度モード
    "language": "ja"
}
```

## リリースノート案

### v1.1.0（予定）
- 字幕形式（SRT/VTT）のサポート
- 出力形式の設定機能
- 設定表示の改善

### v1.0.1-ja（日本語版）
- 完全な日本語ローカライズ
- エラーメッセージの翻訳
- ヘルプテキストの日本語化

## 参考リンク

- [whisper-ctranslate2 ドキュメント](https://github.com/guillaumekln/faster-whisper)
- [Go並行処理パターン](https://go.dev/blog/pipelines)
- [GitHub CLI ドキュメント](https://cli.github.com/manual/)