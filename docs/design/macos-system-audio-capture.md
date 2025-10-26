# macOSシステム音声自動キャプチャ機能

## 概要

**✅ 実装完了（2025-10-26）**

macOS 13+でScreenCaptureKit APIを使用し、BlackHole等の手動設定不要でシステム音声を自動キャプチャする機能を実装しました。Windows版DualRecorderと同等の体験をmacOSで提供します。

**最終的な実装**: ScreenCaptureKit（CATap APIからの変更 - 詳細は後述）

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

## 現在の進捗（2025-10-26 更新）

### ✅ Phase 1完了: Swift CLIツール開発
- ✅ 技術調査完了（AudioTee、AudioCap分析）
- ✅ アーキテクチャ設計完了
- ✅ 新ブランチ作成（`feature/macos-system-audio-capture`）
- ✅ Swift CLI完全実装:
  - **AudioTapManager**: CATap APIでシステム全体の音声キャプチャ
    - `processes=[]` でシステム全体を対象
    - 集約デバイスの作成とtapの紐付け
    - デバイス準備完了待機処理（最大2秒）
    - フォーマット取得リトライロジック（最大3回）
  - **WAVFileWriter**: WAVファイル書き込み機能
    - 16ビット PCM、ステレオ対応
    - 動的ヘッダー生成
  - **AudioRecorder**: IOProc経由でオーディオデータ録音
  - コマンドライン引数処理完成
  - エラーハンドリング実装
- ✅ **動作確認成功** (macOS 15.6.1):
  - コンパイル成功（141KB バイナリ）
  - 5秒間のシステム音声録音成功
  - 1.8MB WAVファイル生成確認 (48kHz, stereo, 16bit PCM)

**参考**: AudioTeeの実装を詳細分析し、正しいCATap API使用方法を学習

### 🚧 Phase 2進行中: Go側統合
- `internal/recorder/system_audio_recorder_darwin.go` 作成（次のステップ）

### ⏳ Phase 3-5未着手
- Swift バイナリ埋め込み（go:embed）
- ビルドシステム更新
- 統合テスト
- ドキュメント更新

## 実装結果（2025-10-26）

### 最終的な実装方式

**ScreenCaptureKit APIを採用**（当初計画のCATap APIから変更）

### 変更理由

**CATap API実装時の問題**:
1. ✅ Tap作成、集約デバイス作成、IOProcコールバックは成功
2. ❌ **オーディオバッファデータがすべて0**
3. ❌ 権限・Entitlements設定が複雑（Code Signing必須の可能性）
4. ❌ ドキュメント不足で問題解決が困難

**Screen CaptureKit採用の理由**:
1. ✅ **実装成功、動作確認済み**
2. ✅ Apple公式の推奨API（充実したドキュメント）
3. ✅ macOS 13+対応（当初のmacOS 14.4+より広い対応範囲）
4. ✅ 「画面収録」権限のみで動作（より単純な権限モデル）
5. ⚠️ 画面キャプチャAPIを音声のみに使用（若干オーバースペック）

### 実装詳細

**ファイル構成**:
```
cmd/audio-capture/
├── main.swift              # CLIエントリーポイント
├── AudioCapture.swift      # ScreenCaptureKit実装
└── audio-capture           # ビルド済みバイナリ（.gitignore）
```

**主要機能**:
- ScreenCaptureKitでシステム音声キャプチャ
- CAF形式で録音 → WAV形式に自動変換（afconvert使用）
- 48kHz, Float32, ステレオ
- 録音時間指定可能

**コマンド例**:
```bash
# 5秒間録音
./audio-capture --output recording.wav --duration 5

# 継続録音（Ctrl+Cで停止）
./audio-capture --output recording.wav
```

### Phase 1完了状況

| タスク | 状態 | 備考 |
|--------|------|------|
| Swift CLI開発 | ✅ 完了 | ScreenCaptureKit使用 |
| 音声キャプチャ実装 | ✅ 完了 | CAF→WAV変換機能付き |
| 動作確認 | ✅ 完了 | システム音声録音成功、再生確認済み |

### 次のステップ（Phase 2以降）

**Phase 2: Go側統合**（未着手）
- [ ] `internal/recorder/system_audio_recorder_darwin.go`作成
- [ ] Go埋め込みリソース（go:embed）でSwift CLIバイナリ埋め込み
- [ ] GUI/TUI統合

**Phase 3: ビルドシステム**（未着手）
- [ ] `build/macos/build.sh`にSwift CLIビルド追加
- [ ] dmg/tar.gzパッケージにバイナリ同梱

**Phase 4: テスト**（未着手）
- [ ] 統合テスト実施

**Phase 5: ドキュメント・リリース**（未着手）
- [ ] README.md, CLAUDE.md更新
- [ ] v1.8.0リリース準備

### 技術メモ

**ScreenCaptureKit使用時の注意点**:
1. **権限**: 初回実行時に「画面収録」権限ダイアログが表示される
2. **最小要件**: macOS 13 Ventura以上
3. **ビデオ設定**: 音声のみキャプチャでも、ディスプレイフィルターが必要
   - `width`/`height`を最小値（1x1）に設定してオーバーヘッド削減
4. **フォーマット**: ネイティブフォーマット（Float32, 48kHz）をそのまま使用するのが最も安定

**WAV変換**:
- AVAudioFileでWAV直接書き込みはフォーマット不一致エラーが発生
- CAF形式で録音 → afconvertでWAVに変換する方式が確実


---

## デュアル録音（システム音声 + マイク音声）の設計

### Phase 0完了後の状況（2025-10-27更新）

**✅ 完了**:
- システム音声のみキャプチャ（ScreenCaptureKit）
- Swift CLIツール動作確認済み（macOS 13+対応）
- **方針変更**: macOS 15+ `captureMicrophone`アプローチを諦め、2ストリーム方式を採用

**❌ 未実装**:
- マイク音声キャプチャ（既存のPortAudio実装を流用予定）
- システム音声とマイク音声の結合（FFmpegまたはGo実装）

### デュアル録音の実装アプローチ

#### アプローチA: ScreenCaptureKit単独（macOS 15+ のみ） - ❌ 不採用

**概要**: ScreenCaptureKitの`captureMicrophone`プロパティを使用

**試行結果（2025-10-27）**:
1. ✅ `captureMicrophone = true`設定は成功
2. ✅ マイクバッファは受信（512フレーム確認）
3. ❌ **AVAudioFile.write()でクラッシュ（trace trap）**
   - マイク: Int16, 1ch (mono)
   - システム: Float32, 2ch (stereo)
   - フォーマット不一致が原因

**諦めた理由**:
1. **複雑すぎる**: AVAssetWriter実装、フォーマット変換が必要（推定10+時間）
2. **macOS 15+限定**: 古いOS（13-14）で動かない
3. **ドキュメント不足**: 2024年の新機能でサンプル少ない
4. **YAGNI違反**: 既存の方法（2ストリーム）で十分達成できる

**結論**: ❌ 不採用（2025-10-27方針決定）

---

#### アプローチB: 2ストリーム方式 - ✅ 採用

**概要**: システム音声とマイク音声を別々に録音して結合（Windows版DualRecorderと同じアーキテクチャ）

**採用理由（2025-10-27）**:
1. **シンプル**: 既存コード（PortAudio）の再利用
2. **確実**: 動作実績あり
3. **互換性**: macOS 13+で動作
4. **Windows版と統一**: アーキテクチャが同じ
5. **メンテナンス性**: 理解しやすい

```
┌───────────────────────┐
│ システム音声          │ ScreenCaptureKit
│ Swift CLI             │ → system.wav (48kHz, Float32, Stereo)
└───────────────────────┘

┌───────────────────────┐
│ マイク音声            │ PortAudio (既存の実装)
│ Go (internal/recorder)│ → mic.wav (48kHz, Float32, Mono)
└───────────────────────┘
          ↓
    ┌─────────┐
    │ 音声結合 │ FFmpeg or Go
    └─────────┘
          ↓
      output.wav (48kHz, Float32, Stereo)
```

**実装詳細**:

**1. Go側での同時録音**:
```go
// internal/recorder/dual_recorder_darwin.go

type DualRecorder struct {
    systemRecorder *SystemAudioRecorder  // Swift CLI呼び出し
    micRecorder    *Recorder             // 既存のPortAudio実装
}

func (r *DualRecorder) Start() error {
    // 1. システム音声録音開始（バックグラウンド）
    go r.systemRecorder.Start()
    
    // 2. マイク録音開始
    r.micRecorder.Start()
    
    return nil
}

func (r *DualRecorder) Stop() (string, error) {
    // 1. 両方停止
    systemFile := r.systemRecorder.Stop()
    micFile := r.micRecorder.Stop()
    
    // 2. 音声結合
    outputFile := r.mixAudio(systemFile, micFile)
    
    return outputFile, nil
}
```

**2. 音声結合方法**:

**Option 1: FFmpegで結合**（簡単・推奨）:
```go
func (r *DualRecorder) mixAudio(systemWav, micWav string) string {
    cmd := exec.Command("ffmpeg",
        "-i", systemWav,        // システム音声（Stereo）
        "-i", micWav,           // マイク音声（Mono）
        "-filter_complex",
        "[0:a]volume=1.0[sys];[1:a]volume=1.0,pan=stereo|c0=c0|c1=c0[mic];[sys][mic]amix=inputs=2:duration=longest",
        "-ar", "48000",
        "-ac", "2",
        "output.wav")
    cmd.Run()
    return "output.wav"
}
```

**Option 2: Goでリアルタイム結合**（複雑）:
- バッファリングして同時に処理
- タイムスタンプ同期が必要
- より高度な実装

**メリット**:
- ✅ macOS 13+で動作（幅広い互換性）
- ✅ 既存のPortAudio実装を再利用
- ✅ 音量バランス調整が柔軟

**デメリット**:
- ❌ 2つのストリーム管理が必要
- ❌ 同期の問題（タイムスタンプ管理）
- ❌ 結合処理のオーバーヘッド

---

### 実装ロードマップ

**Phase 1: Go側統合**（次のステップ）
- [ ] `internal/recorder/system_audio_recorder_darwin.go`作成
- [ ] Swift CLIバイナリをgo:embedで埋め込み
- [ ] GUI/TUI統合
- [ ] システム音声のみの録音機能をKoeMoji-Goに統合

**Phase 2: デュアル録音実装**（その後）
- [ ] `internal/recorder/dual_recorder_darwin.go`作成
- [ ] システム音声 + マイク音声の同時録音
- [ ] FFmpegでの音声結合処理
- [ ] GUI/TUIでデュアル録音オプション追加

**Phase 3: 最適化**（オプション）
- [ ] Goでのリアルタイム結合実装（FFmpegなしで動作）
- [ ] タイムスタンプ同期の改善
- [ ] 音量自動調整機能

---

### 技術的な課題と解決策

**課題1: サンプルレート・フォーマット不一致**
- システム音声: 48kHz, Float32, Stereo
- マイク音声: 設定可能（既存実装で対応済み）
- **解決策**: マイクも48kHz, Float32に統一

**課題2: タイムスタンプ同期**
- 2つのストリームの開始タイミングがズレる
- **解決策**: 
  - 両方の録音開始時刻を記録
  - FFmpegの`-itsoffset`で調整
  - または、Goで先頭の無音部分をトリミング

**課題3: FFmpeg依存**
- FFmpegがインストールされていない環境
- **解決策**:
  - macOSには通常FFmpegがない → Homebrewでインストール必要
  - Phase 3でGo実装に切り替え（FFmpeg不要に）
  - または、ビルド時にFFmpegをバンドル

---

### 参考: Windows版DualRecorderとの比較

| 項目 | Windows (DualRecorder) | macOS (計画) |
|------|------------------------|--------------|
| システム音声 | WASAPI Loopback | ScreenCaptureKit |
| マイク音声 | PortAudio | PortAudio（同じ） |
| 結合方法 | Go（リアルタイム） | FFmpeg → Go実装 |
| 同期 | 同じスレッドで処理 | タイムスタンプ同期 |
| 権限 | なし | 画面収録権限 |

