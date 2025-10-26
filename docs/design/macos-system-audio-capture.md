# macOSシステム音声自動キャプチャ機能

## 概要

macOS 14.4+でCATap APIを使用し、BlackHole等の手動設定不要でシステム音声を自動キャプチャする機能を実装する。Windows版DualRecorderと同等の体験をmacOSで提供する。

## 背景と課題

### 現状（v1.7.2まで）
- macOSでシステム音声を録音するには**手動設定が必要**
  - BlackHoleなどの仮想オーディオデバイスをインストール
  - Audio MIDI設定で集約デバイスを作成
  - KoeMoji-Goの設定でデバイス名を指定
- ユーザーにとって**ハードルが高い**
- Windows版DualRecorderとの機能差

### 理想の状態
- ユーザーは**何も設定せず録音ボタンを押すだけ**
- 自動的にシステム音声がキャプチャされる
- Windows版と同等のユーザー体験

## 技術調査結果

### macOSでのシステム音声キャプチャ方法

#### オプションA: ScreenCaptureKit API（macOS 13 Ventura+）
- **用途**: 画面録画 + システム音声キャプチャ
- **特徴**:
  - 公式API、Appleが積極的にサポート
  - 最大48kHz ステレオ音声対応
  - macOS 13+が必須
  - **デメリット**: 画面録画とセットのため音声のみには過剰

#### オプションB: CATap API（macOS 14.4+） ⭐ 採用
- **用途**: システム音声のみのキャプチャ
- **特徴**:
  - Core Audio専用、音声のみに最適
  - マイク+システム音声の同期録音が可能
  - ドリフト補正機能付き
  - **デメリット**: ドキュメントが少ない（"poorly documented"）、macOS 14.4+が必須

### WindowsのWASAPI Loopbackとの比較

| 項目 | Windows (WASAPI Loopback) | macOS (CATap) |
|------|---------------------------|---------------|
| **公式サポート** | ✅ 完全サポート | ✅ 公式API存在 |
| **実装難易度** | 🟢 比較的簡単 | 🔴 高難度 |
| **ドキュメント** | 🟢 充実 | 🔴 不足 |
| **最小OS要件** | Windows Vista+ | macOS 14.4+ |
| **権限要求** | なし | マイクアクセス権限必須 |

### Go言語からの利用可能性

#### 直接的なGoバインディング
- **現状**: 存在しない
- **問題点**:
  - CATap APIはObjective-C/Swiftフレームワーク
  - CGOでのラッパー作成が必要
  - 非同期コールバック処理が複雑
  - メモリ管理（Core Foundation）が困難
  - **推定工数**: 120〜180時間（3〜4.5週間フルタイム）

#### 採用アプローチ: Swiftハイブリッド方式 ⭐
- **Swift製CLIツール**を作成し、Goから呼び出す
- **推定工数**: 26〜46時間（フルCGO実装の1/4）
- **利点**:
  - SwiftネイティブにAPIを呼べる
  - 実装がシンプル
  - メンテナンスしやすい
  - ユーザーから透過的（1つのバイナリとして配布）

## アーキテクチャ設計

### システム構成

```
┌─────────────────────────────────────────────────┐
│ KoeMoji-Go (Goバイナリ)                          │
│                                                 │
│ ┌─────────────────────────────────────────┐    │
│ │ 埋め込みリソース (go:embed)              │    │
│ │ └── audio-capture (Swiftバイナリ)        │    │
│ └─────────────────────────────────────────┘    │
│                                                 │
│ 実行時:                                          │
│ 1. 埋め込んだSwiftバイナリを/tmpに展開           │
│ 2. exec.Commandで実行                           │
│ 3. システム音声をWAVファイルに録音               │
│ 4. 処理完了後、一時バイナリを削除               │
└─────────────────────────────────────────────────┘
```

### ユーザー視点の動作

```bash
# ユーザーは今まで通り使うだけ
./koemoji-go --gui
# 録音ボタンを押す → 自動的にシステム音声がキャプチャされる
```

### 配布パッケージ構造

```
koemoji-go-macos-1.8.0.tar.gz
├── koemoji-go          (Goメインバイナリ、Swift CLIを内包)
├── config.json
└── README.md
```

**ユーザーには1つのバイナリとして配布**され、内部でSwift CLIを自動的に呼び出す。

## 実装計画

### Phase 1: Swift CLIツール開発（12〜20時間）

#### ディレクトリ構造
```
cmd/audio-capture/
├── main.swift              # エントリーポイント
├── AudioCapture.swift      # CATap API実装（今後追加）
└── README.md              # Swift CLI専用ドキュメント（今後追加）
```

#### 機能仕様

**コマンドライン引数**:
```bash
audio-capture \
  --output <path>         # 出力ファイルパス (必須)
  --duration <seconds>    # 録音時間（秒）、0=手動停止 (デフォルト: 0)
  --sample-rate <rate>    # サンプルレート (デフォルト: 44100)
```

**実装ステップ**:
1. ✅ コマンドライン引数処理
2. ✅ macOSバージョンチェック（14.4未満ならエラー）
3. 🚧 CATap APIでシステム音声キャプチャ
4. 🚧 WAVファイルに書き込み
5. 🚧 シグナルハンドリング（Ctrl+Cで終了）
6. 🚧 エラーハンドリング

**参考実装**:
- [insidegui/AudioCap](https://github.com/insidegui/AudioCap) - CATap API使用例
- [makeusabrew/audiotee](https://github.com/makeusabrew/audiotee) - システム全体の音声キャプチャ

### Phase 2: Go側の統合（6〜10時間）

#### 新規ファイル

**`internal/recorder/system_audio_recorder_darwin.go`**
```go
//go:build darwin

package recorder

import (
    _ "embed"
    "os"
    "os/exec"
    "path/filepath"
)

//go:embed ../../bin/audio-capture-darwin
var audioCaptureBin []byte

type SystemAudioRecorder struct {
    outputPath  string
    process     *exec.Cmd
    isRecording bool
}

func NewSystemAudioRecorder() (*SystemAudioRecorder, error) {
    // macOSバージョンチェック（14.4未満なら警告）
    return &SystemAudioRecorder{}, nil
}

func (r *SystemAudioRecorder) Start(outputPath string) error {
    // 1. 埋め込んだSwiftバイナリを一時展開
    tmpPath := filepath.Join(os.TempDir(), "koemoji-audio-capture")
    if err := os.WriteFile(tmpPath, audioCaptureBin, 0755); err != nil {
        return err
    }

    // 2. Swift CLIを実行
    r.process = exec.Command(tmpPath, "--output", outputPath)
    return r.process.Start()
}

func (r *SystemAudioRecorder) Stop() error {
    // プロセス停止とクリーンアップ
    // ...
}
```

#### GUIコンポーネントの修正

**`internal/gui/components_darwin.go`**
```go
func (app *GUIApp) initializeRecorder() error {
    // システム音声キャプチャ機能が使えるかチェック
    if isSystemAudioCaptureAvailable() {
        app.recorder, err = recorder.NewSystemAudioRecorder()
    } else {
        // フォールバック: 従来の方式（集約デバイス）
        app.recorder, err = recorder.NewRecorder()
    }
    return err
}
```

### Phase 3: ビルドシステムの更新（2〜4時間）

**`build/macos/build.sh`** 修正:
```bash
# Swift audio-capture CLIをビルド
build_audio_capture() {
    echo "🎵 Building audio-capture (Swift)..."
    mkdir -p "$SCRIPT_DIR/../../bin"

    swiftc \
        -o "$SCRIPT_DIR/../../bin/audio-capture-darwin" \
        "$SCRIPT_DIR/../../cmd/audio-capture/"*.swift
}

build() {
    build_audio_capture  # Swift CLIをビルド
    build_arch "$arch"   # Goバイナリをビルド（Swift CLI埋め込み）
    # パッケージング...
}
```

**`.gitignore`** 更新:
```
# Swift build artifacts
bin/audio-capture-darwin
cmd/audio-capture/audio-capture
```

### Phase 4: テスト・デバッグ（4〜8時間）

#### Swift CLI単体テスト
```bash
# コンパイル
swiftc -o audio-capture cmd/audio-capture/*.swift

# 動作確認
./audio-capture --output test.wav --duration 5

# システム音声（YouTube等）を流しながら録音して確認
```

#### 統合テスト
1. GUIモードで録音ボタンを押してシステム音声が録音されるか
2. TUIモードで録音コマンドが動作するか
3. エラーハンドリング（権限なし、macOS古いバージョン等）
4. フォールバック動作（Swift CLI失敗時に集約デバイス方式に切り替わるか）

### Phase 5: ドキュメント・リリース（2〜4時間）

#### README.md更新
- macOS 14.4+で自動システム音声キャプチャに対応
- macOS 14.3以前は集約デバイス方式を使用

#### CLAUDE.md更新
- Swift CLIツールのビルド方法
- アーキテクチャ説明
- トラブルシューティング

#### リリースノート（v1.8.0）
```markdown
## 🎉 v1.8.0の主な変更

### ✨ 新機能
- **macOS 14.4+でシステム音声自動キャプチャに対応**
  - BlackHole等の手動設定が不要に
  - CATap APIを使用した内部実装
  - ユーザーは録音ボタンを押すだけでシステム音声をキャプチャ可能

### 🔧 技術的変更
- Swift製音声キャプチャCLIツールを統合
- Go埋め込みリソース（go:embed）で配布
- macOS 14.3以前は従来の集約デバイス方式にフォールバック

### 📝 その他
- Windows版DualRecorderと同等の機能をmacOSで実現
```

## 推定工数

| フェーズ | タスク | 工数 | 進捗 |
|---------|--------|------|------|
| Phase 1 | Swift CLI開発 | 12〜20時間 | 🚧 30% |
| Phase 2 | Go側統合 | 6〜10時間 | ⏳ 0% |
| Phase 3 | ビルドシステム | 2〜4時間 | ⏳ 0% |
| Phase 4 | テスト・デバッグ | 4〜8時間 | ⏳ 0% |
| Phase 5 | ドキュメント・リリース | 2〜4時間 | ⏳ 0% |
| **合計** | | **26〜46時間** | **🚧 10%** |

## リスクと対策

### リスク1: CATap API実装の複雑さ
- **対策**: AudioCapサンプルコードを忠実に参考にする
- **フォールバック**: 実装困難な場合は集約デバイス方式のドキュメント充実で対応

### リスク2: macOSバージョン互換性
- **対策**: macOSバージョンチェックを実装し、古いバージョンでは従来方式にフォールバック
- **影響**: macOS 14.3以前のユーザーは手動設定が必要（現状維持）

### リスク3: 権限関連のエラー
- **対策**: 詳細なエラーメッセージと解決方法の提示
- **要求権限**: `NSAudioCaptureUsageDescription`（マイクアクセス）

## 参考資料

### Apple公式ドキュメント
- [Capturing system audio with Core Audio taps](https://developer.apple.com/documentation/coreaudio/capturing-system-audio-with-core-audio-taps)
- [ScreenCaptureKit](https://developer.apple.com/documentation/screencapturekit)

### 参考実装
- [insidegui/AudioCap](https://github.com/insidegui/AudioCap) - CATap APIサンプル（特定プロセス用）
- [makeusabrew/audiotee](https://github.com/makeusabrew/audiotee) - システム全体の音声キャプチャ

### 技術記事
- [AudioTee: capture system audio output on macOS](https://stronglytyped.uk/articles/audiotee-capture-system-audio-output-macos)

## 現在の進捗（2025-10-26）

### ✅ 完了
- 技術調査完了
- アーキテクチャ設計
- 新ブランチ作成（`feature/macos-system-audio-capture`）
- Swift CLI骨格実装
  - コマンドライン引数処理
  - macOSバージョンチェック
  - コンパイルテスト成功（76KB バイナリ）

### 🚧 進行中
- CATap API実装（次のステップ）

### ⏳ 未着手
- WAVファイル書き込み
- Go側統合
- ビルドシステム更新
- テスト・ドキュメント
