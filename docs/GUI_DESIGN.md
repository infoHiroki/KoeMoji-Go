# KoeMoji-Go GUI 設計書

## 設計方針

### 1. 基本原則
- **機能互換性**: 既存TUIと完全に同じ機能を提供
- **シンプル設計**: 必要最小限の実装で直感的なUI
- **既存コード活用**: processor/logger/whisper/configパッケージはそのまま使用

### 2. 終了仕様
- **即座終了**: ウィンドウクローズ、終了ボタンで即座にアプリケーション終了
- **クリーンアップなし**: グレースフルシャットダウンは実装しない
- **実用性重視**: 処理中でも強制終了（ユーザーが終了したい時は処理中が多いため）

### 3. UI更新方式
- **定期更新**: 5秒間隔でUI全体を更新
- **シンプル実装**: イベント駆動ではなく定期ポーリング
- **軽量化**: ログは12エントリ制限を維持

## アーキテクチャ設計

### パッケージ構成
```
internal/gui/
├── app.go          # GUI アプリケーション本体
├── window.go       # ウィンドウレイアウト
├── components.go   # UI コンポーネント（ステータス、ログ、ボタン）
└── icons/
    ├── icon.go     # アイコンリソース
    └── icon.png    # アプリケーションアイコン
```

### 起動方式
```bash
# GUI モード起動
./koemoji-go --gui

# 従来通りTUI起動（デフォルト）
./koemoji-go
```

### データフロー
```
既存パッケージ → App構造体（TUIと共通） → 5秒間隔更新 → GUI表示
```

## UI設計

### ウィンドウ仕様
- **サイズ**: 800x700 ピクセル
- **レイアウト**: BorderLayout（上：ステータス、中央：ログ、下：ボタン）
- **起動位置**: 画面中央
- **リサイズ**: 可能

### コンポーネント設計

#### 1. ステータスパネル（上部）
```
🟢 稼働中 | 待機: 0 | 処理中: なし
📁 入力: 2 → 出力: 5 → アーカイブ: 10  
⏰ 最終: 15:04:05 | 次回: 15:05:05 | 稼働: 2h
```

#### 2. ログビューア（中央）
- スクロール可能なログ表示
- 12エントリの循環表示（TUIと同じ）
- フォーマット: `[レベル] 時刻 メッセージ`
- 実装: `widget.RichText` + `container.NewVScroll()` 推奨

#### 3. ボタンパネル（下部）
```
[設定] [ログ] [スキャン] [入力] [出力] [終了]
```

### 設定ダイアログ
- タブ構成: 基本設定 / ディレクトリ / LLM設定
- フォーム形式の入力
- ディレクトリ選択ダイアログ対応
- 入力値検証

## 技術仕様

### フレームワーク
- **Fyne v2**: クロスプラットフォームGUIフレームワーク
- **Go 1.21+**: 既存要件と同じ

### Fyne固有の注意事項
- **スクロール使い分け**: フォーム要素では避ける、ログ表示では`container.NewVScroll()`使用可能
- **レイアウト優先**: `container.NewVBox()`、`container.NewBorder()`を活用
- **アプリID必須**: `app.NewWithID("com.hirokitakamura.koemoji-go")`で警告回避

### 実装方針

#### 1. main.go の変更
```go
var guiMode = flag.Bool("gui", false, "Run in GUI mode")

func main() {
    // 既存のフラグ処理
    
    if *guiMode {
        gui.Run(configPath, debugMode)
    } else {
        // 既存のTUI実行
    }
}
```

#### 2. GUI App 構造
```go
type App struct {
    // 既存のTUI App構造体をベースにする
    // GUI固有の要素のみ追加
    fyneApp fyne.App
    window  fyne.Window
}
```

#### 3. 定期更新
```go
func (a *App) startPeriodicUpdate() {
    go func() {
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                a.updateUI()
            }
        }
    }()
}
```

#### 4. ボタンアクション
```go
// 既存のTUIコマンド処理をそのまま呼び出し
configBtn.OnTapped = func() {
    // TUIの設定処理を再利用 → GUI設定ダイアログに変更
}

scanBtn.OnTapped = func() {
    // TUIのスキャン処理をそのまま使用
}
```

## 多言語対応

### メッセージシステム
- 既存の`ui.GetMessages()`をそのまま活用
- GUI固有のメッセージが必要な場合のみ追加

### 対応言語
- 日本語（デフォルト）
- 英語

## ビルド仕様

### 依存関係追加
```bash
go get fyne.io/fyne/v2/app
go get fyne.io/fyne/v2/widget
go get fyne.io/fyne/v2/container
go get fyne.io/fyne/v2/dialog
```

### ビルドコマンド
```bash
# Windows GUI版（コンソール非表示）
GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui -s -w" -o koemoji-go-gui.exe

# macOS版
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o koemoji-go-gui

# Linux版  
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o koemoji-go-gui
```

### 配布パッケージ
```
koemoji-go-v1.3.0-windows/
├── koemoji-go.exe      # TUI版
├── koemoji-go-gui.exe  # GUI版
├── config.json         # 設定例
└── README.md           # 使用方法
```

## 実装優先順位

### Phase 1: 基本GUI
1. ウィンドウとレイアウト作成
2. ステータス表示（静的）
3. ログ表示（静的）
4. ボタン配置

### Phase 2: 動的機能
1. 定期更新実装
2. 既存処理との連携
3. ボタンアクション実装

### Phase 3: 設定機能
1. GUI設定ダイアログ
2. 多言語対応
3. エラーハンドリング

## 制約事項

### 技術的制約
- **バイナリサイズ**: GUI版は+30-40MB（Fyne込み）
- **メモリ使用**: GUI版は+30-40MB
- **起動時間**: GUI版は若干遅くなる

### 機能制約
- **即座終了**: 処理中でも強制終了
- **シンプルUI**: 複雑なアニメーションや高度なUI要素は実装しない
- **TUI互換**: TUIで不可能な機能はGUIでも実装しない

## 期待効果

### ユーザビリティ向上
- 視覚的な状態確認
- マウス操作対応
- ウィンドウ管理（最小化/復元）

### 既存機能保持
- TUIと完全同等の機能
- 設定ファイル互換性
- 処理ロジック不変

### 将来拡張性
- ファイルリスト表示
- プログレスバー
- ドラッグ&ドロップ

---

この設計書に基づいて、シンプルで実用的なGUI実装を行う。