# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

KoeMoji-Goは、Goで書かれた音声・動画ファイル自動文字起こしツールです。GUI（Fyne-based）とTUIの両インターフェースを提供し、FasterWhisperによる音声認識と、オプションのOpenAI API統合によるAI要約機能を備えています。

### 主な特徴
- **クロスプラットフォーム**: Windows、macOS対応（Linux未テスト）
- **二つのUI**: GUI（Fyne）とTUI（Terminal UI）
- **高精度音声認識**: FasterWhisper（OpenAI Whisperの高速版）
- **Python 3.12（推奨）**: FasterWhisperバックエンド用（3.13以降は非対応）
- **デュアル録音**: システム音声+マイク同時録音（Windows: v1.7.0～、macOS: v1.8.0～）
  - Windows: VoiceMeeter/ステレオミキサー方式
  - macOS: ScreenCaptureKit API方式（macOS 13+）
- **リアルタイム録音**: PortAudio統合による録音機能
- **AI要約**: OpenAI API連携でテキスト要約生成
- **ポータブル設計**: 単一実行ファイルで動作（Python依存を除く）

## 開発コマンド

### ビルド
```bash
# 開発用ビルド
go build -o koemoji-go ./cmd/koemoji-go

# macOS配布用ビルド（v1.8.0以降）
cd build/macos
./build.sh        # デフォルトでtar.gz版をビルド → koemoji-go-macos-1.8.0.tar.gz
./build.sh build  # 明示的にビルド（上記と同じ）
./build.sh clean  # ビルド成果物のクリーンアップ

# Windows配布用ビルド（MSYS2/MinGW64が必要）
cd build/windows && build.bat  # → koemoji-go-1.8.0.zip

# ビルド成果物のクリーンアップ
cd build/macos && ./build.sh clean
cd build/windows && build.bat clean
```

### リリース成果物（v1.7.0以降の命名規則）
**macOS**:
- `koemoji-go-macos-{VERSION}.tar.gz` (GUI/TUI両対応)

**Windows**:
- `koemoji-go-{VERSION}.zip`

**変更履歴**:
- v1.8.1（未リリース）: Windows版命名規則変更（`koemoji-go-windows-{VERSION}.zip` → `koemoji-go-{VERSION}.zip`、「windows」単語を排除してセキュリティフィルタ回避）
- v1.8.0: macOSデュアル録音機能実装（FFmpeg不要の1ファイルミキシング）
- v1.7.2: macOS版はtar.gz形式のみ（DMG版は廃止、Apple Developer Programコスト削減のため）
- v1.7.0: 命名規則変更（プラットフォーム名を中央に配置、ウイルス検知回避強化）
- v1.6.1: 命名規則変更（`-win`→`-windows`）
- v1.6.0以前: `KoeMoji-Go-v{VERSION}-win.zip`形式

### テスト
```bash
# 単体テスト実行
go test ./...

# 特定パッケージのテスト実行
go test ./internal/config
go test ./internal/logger
go test ./internal/processor

# 統合テスト実行
go test ./test

# 手動テスト（詳細な手順はtest/manual-test-commands.mdを参照）
./koemoji-go --version
./koemoji-go --help
./koemoji-go --gui     # GUIモード
./koemoji-go --tui     # ターミナルUIモード
./koemoji-go --debug   # デバッグログ
```

### アプリケーション実行
```bash
# GUIモード（デフォルト）
./koemoji-go

# ターミナルUIモード
./koemoji-go --tui

# 設定モード
./koemoji-go --configure

# デバッグモード（詳細ログを記録）
./koemoji-go --debug
./koemoji-go --tui --debug  # TUIでもデバッグモード可能
```

**デバッグモードについて**:
- 通常モード: `[INFO]`, `[PROC]`, `[DONE]`, `[ERROR]`レベルのログのみ記録
- デバッグモード: 上記に加え、`[DEBUG]`レベルの詳細ログを記録
  - Whisper実行コマンドと出力
  - OpenAI APIリクエスト/レスポンス内容
  - ファイル処理の詳細な進行状況
- エラー調査やトラブルシューティング時はデバッグモードを推奨

## アーキテクチャ概要

### 中核設計パターン
- **単一責任**: 各internalパッケージは専門機能に特化
- **順次処理**: 安定性のための1ファイルずつ処理方式
- **コンテキストベース停止**: Goのcontextパッケージによる優雅な終了処理
- **依存性注入**: 設定とロガーは関数パラメータ経由で受け渡し

### ディレクトリ構造（詳細）
```
KoeMoji-Go/
├── cmd/koemoji-go/         # アプリケーションエントリーポイント
│   └── main.go            # メイン関数、CLI引数処理
├── internal/              # 内部パッケージ（外部公開しない）
│   ├── config/           # 設定管理
│   │   ├── config.go    # 設定構造体と読み込み
│   │   └── validate.go  # 設定バリデーション
│   ├── gui/             # GUIコンポーネント
│   │   ├── app.go      # Fyneアプリケーション
│   │   ├── components.go # UIコンポーネント更新
│   │   ├── window.go   # ウィンドウレイアウト
│   │   ├── resources.go # 埋め込みリソース（アイコン）
│   │   ├── dialogs.go  # ダイアログ機能
│   │   └── icon.png    # アプリケーションアイコン
│   ├── ui/              # TUIコンポーネント
│   │   └── tui.go      # ターミナルUI実装
│   ├── processor/       # ファイル処理エンジン
│   │   └── processor.go # ファイル監視と処理
│   ├── recorder/        # 録音システム
│   │   └── recorder.go  # PortAudio統合
│   ├── whisper/         # 音声認識
│   │   └── whisper.go   # FasterWhisper呼び出し
│   ├── llm/            # AI統合
│   │   └── openai.go   # OpenAI API連携
│   └── logger/         # ログシステム
│       └── logger.go   # バッファード・ロガー
├── build/              # ビルドスクリプトとアセット
│   ├── common/assets/  # 共通リソース
│   │   ├── config.example.json
│   │   └── README_RELEASE.md
│   ├── macos/         # macOSビルド
│   │   └── build.sh
│   └── windows/       # Windowsビルド
│       ├── build.bat
│       ├── *.dll      # 必要なDLLファイル
│       └── icon.ico   # アプリケーションアイコン
├── test/              # テストファイル
│   └── manual-test-commands.md
├── version.go         # バージョン定義
├── go.mod            # Go依存関係
├── README.md         # ユーザー向けドキュメント
├── CLAUDE.md         # このファイル
└── config.json       # ユーザー設定（実行時生成）
```

### 主要技術
- **GUIフレームワーク**: クロスプラットフォームデスクトップインターフェース用Fyne v2.6.1
  - アプリアイコン統合（go:embed）
  - ダークモード自動対応
  - 統一されたアイコンデザイン
- **音声処理**: リアルタイム音声録音用PortAudio
- **音声認識**: FasterWhisper（Python 3.12推奨、3.13以降は非対応）
- **AI統合**: テキスト要約用OpenAI API
- **テスト**: 単体テスト用Testify v1.10.0

### サポート形式
- **入力音声**: MP3, WAV, M4A, FLAC, OGG, AAC
- **入力動画**: MP4, MOV, AVI
- **出力**: TXT, VTT, SRT, TSV, JSON

## 設定システム

設定は`config.json`で管理され、以下の主要領域があります：
- **Whisper設定**: モデル選択（tinyからlarge-v3）、言語設定、計算精度
- **パフォーマンス**: CPU使用制限、処理オプション
- **ディレクトリ**: 入力/出力/アーカイブパスの設定
- **録音**: デバイス名で指定（環境非依存）、録音時間/サイズ制限
- **AI機能**: OpenAI APIキーとカスタマイズ可能プロンプト
- **UI**: カラーサポート、言語設定

新規インストール時は`config.example.json`をテンプレートとして使用してください。

### 設定ファイル例
```json
{
  "whisper_model": "large-v3",
  "language": "ja",
  "input_dir": "./input",
  "output_dir": "./output",
  "archive_dir": "./archive",
  "recording_device_name": "",
  "llm_summary_enabled": false,
  "llm_api_key": ""
}
```

### 録音デバイス設定

録音デバイスは**デバイス名**で指定します（v1.5.5以降）：

```json
{
  "recording_device_name": "koemoji"  // 集約デバイスの名前
}
```

**重要な仕様**:
- デバイスIDではなく**名前**で管理（環境非依存）
- 空文字列の場合はデフォルトデバイスを使用
- 起動時に名前から自動的にデバイスを検索
- macOS: Audio MIDI設定で作成した集約デバイス名
- Windows: ステレオミキサーやサウンドデバイス名

## ビルドシステム

### 前提条件
- **Go 1.21+** コンパイル用
- **Python 3.12（推奨）** FasterWhisperバックエンド用（3.13以降は非対応）
- **MSYS2/MinGW64** Windows ビルド用（CGO依存）
- **PortAudio** 開発ライブラリ

### プラットフォーム固有の考慮事項
- **macOS**: Apple Silicon最適化、マイクアクセス許可が必要
- **Windows**: 必要なDLL（libportaudio.dll、GCCランタイム）を同梱
- **CGO**: PortAudio統合に必須

### 配布パッケージ化

**macOS（v1.7.2以降）**:
- `koemoji-go-macos-{VERSION}.tar.gz`
  - 単一バイナリ形式（GUI/TUI両対応）
  - ビルド: `cd build/macos && ./build.sh`
  - **注**: DMG版は廃止（Apple Developer Program年間費用削減のため）

**Windows**:
- `koemoji-go-{VERSION}.zip`
- 解凍後フォルダ: `koemoji-go-{VERSION}/`

### DLL処理（Windows）
`build.bat`はワイルドカード`*.dll`を使用して自動的にDLLをコピー：
```batch
copy /Y *.dll "%DIST_DIR%\" >nul
```

## よくある問題と解決策

### 1. FasterWhisperインストール問題
**症状**: 「whisper-ctranslate2が見つかりません」エラー
**解決策**:
```bash
pip install faster-whisper
# または
pip3 install faster-whisper
```

### 2. Windows GPU環境での問題
**症状**: `compute_type: int8`設定でもGPUが使用される
**解決策**: config.jsonで明示的にCPU使用を指定（内部で`--device cpu`を自動追加）

### 3. macOS録音許可
**症状**: 録音開始時にエラー
**解決策**: システム環境設定 > セキュリティとプライバシー > マイクでアプリを許可

### 4. ファイル処理が始まらない
**症状**: inputフォルダにファイルを置いても処理されない
**解決策**:
- ファイル形式を確認（サポート形式のみ）
- ログで詳細確認: `./koemoji-go --debug`
- 権限を確認

### 5. Windowsビルドが途中で落ちる（v1.6.0で修正済み）
**症状**: `build.bat`実行時にウィンドウが即座に閉じる
**原因**: goversioninfo実行時のエラー、パス指定の問題
**解決策**:
```cmd
# コマンドプロンプトから実行してエラー確認
cd build\windows
build.bat

# 環境チェック
check_env.bat

# 段階的テスト
test_go_build.bat        # Goビルドのみ
test_packaging_only.bat  # パッケージングのみ
```

**詳細**: [docs/progress/v1.6.0-build-system-fix.md](docs/progress/v1.6.0-build-system-fix.md)

## デバッグとトラブルシューティング

### ログレベル
```bash
# 通常ログ
./koemoji-go

# デバッグログ（詳細）
./koemoji-go --debug
```

### ログ出力場所
- **GUI**: アプリケーション内のログエリア（一時的、最大12エントリ）
- **TUI**: ターミナル画面下部（一時的、最大12エントリ）
- **ファイル**: `koemoji.log`（永続的、実行ディレクトリに保存）

**重要**: GUIの「ログ」ボタンで`koemoji.log`ファイルを開けます。エラー調査時は必ずログファイルを確認してください。

### ログレベルの詳細
- **通常モード** (`./koemoji-go`):
  - `[INFO]`: 一般情報（起動、設定読み込み、録音開始/停止など）
  - `[PROC]`: ファイル処理開始、要約生成開始
  - `[DONE]`: 処理完了
  - `[ERROR]`: エラー発生時

- **デバッグモード** (`./koemoji-go --debug`):
  - 上記に加え、`[DEBUG]`レベルの詳細情報:
    - Whisper実行コマンドと標準出力/エラー出力
    - OpenAI APIリクエストJSON（プロンプト、パラメータ）
    - OpenAI APIレスポンス詳細（ステータス、生成文字数）
    - ファイルスキャン詳細、処理フロー詳細

### よく使うデバッグコマンド
```bash
# Whisperコマンドの確認
which whisper-ctranslate2

# Python環境確認
python --version
pip list | grep faster-whisper

# PortAudio確認（macOS）
brew list portaudio

# DLL確認（Windows）
dir build\windows\*.dll
```

## 重要な実装詳細

### 録音状態管理
録音機能は、GUIコンポーネントとレコーダーバックエンド間の慎重な状態同期を使用します。以下に注意：
- チャネルを使ったスレッドセーフな状態更新
- 録音開始/停止操作でのUI状態一貫性
- 特に録音中のアプリケーション終了時の適切なクリーンアップ

### ファイル処理パイプライン
リソース競合を防ぐため、ファイルは順次処理されます：
1. ディレクトリ監視が新ファイルを検出
2. 単一ファイル処理でメモリ問題を防止
3. 処理成功後、ファイルをアーカイブに移動
4. エラーハンドリングで元ファイルを保護

### テストアプローチ
重要なテストシナリオには以下が含まれます：
- 録音状態管理とUI同期
- 録音中終了警告
- クロスプラットフォームビルド検証
- FasterWhisper Pythonバックエンドとの統合
- 長時間録音でのメモリ使用量

包括的なテスト手順については`test/manual-test-commands.md`を参照してください。

## 最近の重要な変更

### v1.8.0での変更（2025-10-27）
1. **macOSデュアル録音機能実装**（Phase 0-4完了）
   - ScreenCaptureKit APIを使用したシステム音声キャプチャ
   - Swift CLIバイナリ（`cmd/audio-capture`）の統合
   - 2ストリーム方式（システム音声とマイク音声を別ファイルで保存）
   - macOS 13以降対応、画面収録権限が必要

2. **実装ファイル**
   - `internal/recorder/system_audio_darwin.go` (242行) - Swift CLI ラッパー
   - `internal/recorder/dual_recorder_darwin.go` (386行) - デュアル録音実装
   - `cmd/audio-capture/AudioCapture.swift` - ScreenCaptureKit統合
   - `cmd/audio-capture/main.swift` - CLI エントリーポイント

3. **GUIデュアル録音切り替え機能**
   - 設定画面に録音モード選択ラジオボタン追加
   - "シングル録音（マイクのみ）" / "デュアル録音（システム音声+マイク）"
   - macOS 13以降と画面収録権限が必要な旨の情報ラベル表示
   - Windows版と統一されたUIパターン

4. **技術仕様**
   - システム音声: 48kHz Float32 Stereo (CAF → WAV自動変換)
   - マイク音声: 44.1kHz Int16 Mono
   - SIGTERMによる優雅な停止処理（DispatchSourceSignal）
   - バイナリ検索パス: 実行ファイルディレクトリ優先

5. **ビルドシステム更新**
   - `build/macos/build.sh`: Swift CLIバイナリを自動パッケージング
   - リリース版に`audio-capture`バイナリ（171KB）を同梱

6. **ドキュメント追加**
   - `docs/user/SYSTEM_AUDIO_RECORDING_MACOS.md` - macOS版ユーザーガイド
   - `README.md` - デュアル録音機能のmacOS対応を反映
   - `TEST_RESULTS.md` - 全自動テスト合格（8/8）、手動テスト合格（3/3）

7. **テスト結果**
   - 自動テスト: ビルド、環境確認、デバイス検出、設定シナリオ、統合テスト（全合格）
   - 手動テスト: GUI/TUIモードでのシングル/デュアル録音動作確認（全合格）

8. **FFmpeg不要の1ファイルミキシング機能**
   - `internal/recorder/mixer.go` (366行) - Go標準ライブラリのみでWAVミキシング実装
   - 線形補間による44.1kHz→48kHzリサンプリング
   - ステレオ+モノラルのミキシング（システム70%, マイク100%）
   - ソフトクリッピングによる歪み防止
   - afconvert互換性対応（FLLR等の余分なチャンク処理）
   - デフォルトで1ファイル出力（`recording_*.wav`）、重複文字起こし防止
   - `SaveSeparateFiles()`で2ファイル出力も可能（話者分離用）

9. **ミキシング機能のテスト**
   - `internal/recorder/mixer_test.go` (290行) - 8つのユニットテスト（全合格）
   - `internal/recorder/test_afconvert_wav_test.go` - afconvert互換性テスト
   - 手動テスト: 503KB、2.68秒、48kHz Stereo Int16 正常動作確認

### 未リリース（2025-10-29）
1. **Windows版ビルド成果物の命名規則変更**
   - セキュリティフィルタ回避のため「windows」という単語を完全排除
   - 旧: `koemoji-go-windows-{VERSION}.zip`
   - 新: `koemoji-go-{VERSION}.zip`
   - macOS版は変更なし: `koemoji-go-macos-{VERSION}.tar.gz`

2. **変更ファイル（計9ファイル）**
   - **実装層**: `build/windows/build.bat`, `scripts/release.sh`
   - **ドキュメント層**: `CLAUDE.md`, `README.md`, `WINDOWS_BUILD_GUIDE.md`, `GITHUB_CLI.md`, `MACOS_BUILD_GUIDE.md`, `VERSION_UPDATE_CHECKLIST.md`

3. **背景**
   - 「windows」という単語がセキュリティソフトやブラウザのフィルタリングに引っかかる問題
   - 企業PCや学校ネットワークでダウンロードがブロックされるケースを回避
   - メール添付時の拒否リスクを最小化

### 未リリース（2025-10-28）
1. **TUI正式版化（Phase 14完了）**
   - Rich TUIを正式版に昇格（`tui_rich.go` → `tui_tview.go`）
   - `RichTUI` → `TUI`に命名を簡素化
   - `--tui-rich`フラグを削除、`--tui`に統一（6文字削減）
   - 旧Simple TUI実装を削除（~110行削除）
   - テスト全面書き直し（247行、全合格）
   - 未使用コードクリーンアップ（正味-97行）

2. **デュアル録音の注意喚起追加**
   - ヘッドホン/イヤホン推奨の警告をドキュメントに追加
   - GUI設定画面に情報ラベル追加（macOS向け）
   - 物理的な音響結合問題をわかりやすく説明
   - GitHub Issue #18作成（将来のエコーキャンセレーション機能）

3. **TUI設定画面のバグ修正**
   - キャンセルボタンが動作しない問題を修正
   - 録音設定追加後のインデックスずれに対応
   - case文のインデックスを正しく更新（case 5→6に修正）

### 未リリース（2025-10-25）
1. **ログシステムの改善**
   - GUIモードでファイル処理ログがファイルに記録されない問題を修正
   - `internal/gui/components.go`: `processor`呼び出し時に`app.logger`を正しく渡すように修正
   - デバッグモード（`--debug`）で詳細ログ（Whisper実行、OpenAI API詳細）を記録
   - 通常モードでは`[INFO]`, `[PROC]`, `[DONE]`, `[ERROR]`のみ記録

2. **エラーメッセージ表示の改善**
   - OpenAI APIエラーレスポンスを1000文字に制限（GUIでの表示切れ対策）
   - ユーザー向けエラーメッセージをシンプル化
   - 詳細情報は`koemoji.log`に記録

3. **コードの簡素化**
   - AI要約プロンプトの後方互換性処理を削除（ユーザー数が少ないため不要）
   - `preparePrompt()`関数をシンプル化
   - 関連テストを新仕様に更新

4. **ログボタンの修正**
   - Windows環境でGitのnotepadラッパーが開く問題を修正
   - `C:\Windows\notepad.exe`を直接指定するように変更

### v1.6.1での変更（2025-01-22）
1. **Windows GUI 日本語文字化け修正**
   - Fyne + Windows + 日本語 + Boldスタイルの組み合わせで文字化けが発生
   - `internal/gui/components.go`, `internal/gui/window.go`からBoldスタイルを削除
   - 通常フォントでの表示に変更

2. **ビルド成果物の命名規則変更**
   - 旧: `KoeMoji-Go-v1.6.0-win.zip`
   - 新: `koemoji-go-1.6.1-windows.zip`
   - 理由: `-win`サフィックスがアンチウイルスソフトに誤検知される問題への対応
   - 全て小文字、プラットフォーム名を明確に（`windows`, `macos`）

### v1.6.0での変更（2025-01-21）
1. **VoiceMeeter統合機能**（Windows専用）
   - システム音声+マイクの同時録音対応
   - VoiceMeeter自動検出機能
   - 音量自動正規化機能（閾値5000、目標20000）
   - GUI設定ダイアログに「VoiceMeeter設定を適用」ボタン追加

2. **Windowsビルドシステム修正**
   - goversioninfo実行時のエラー修正
   - アイコン埋め込み機能の復元
   - ビルドスクリプトの自動化改善

### v1.5.5での変更
1. **録音デバイス設定の改善**
   - デバイスIDから名前ベースの指定に変更
   - 環境非依存の設定を実現
   - 空文字列でデフォルトデバイスを使用

### v1.5.4での変更
1. **フォルダ命名規則の統一**
   - 旧: `koemoji-go-windows-1.5.4`、`koemoji-go-macos-arm64-1.5.4`
   - 新: `KoeMoji-Go-v1.5.4`（全プラットフォーム共通）
   - 効果: アップデート時のユーザーデータ保護

2. **Windows DLL処理の簡素化**
   - 旧: MSYS2パス検出の複雑な条件分岐（14行）
   - 新: ワイルドカード`*.dll`使用（4行）
   - 効果: 保守性向上、新DLL自動対応

3. **FasterWhisper自動インストール改善**
   - Windows実行ファイルパス対応
   - GPU環境での`--device cpu`自動追加
   - クロスプラットフォーム対応強化

## Git操作時の注意事項

### ブランチ戦略
- `main`: 安定版リリース
- `feature/*`: 新機能開発
- `fix/*`: バグ修正

### コミットメッセージ規則
```
🔧 fix: バグ修正の概要を一行で説明

- 変更内容の詳細
- 修正した問題の説明
- 影響範囲

Co-Authored-By: Claude <noreply@anthropic.com>
```

**重要なルール**:
- **アトミックコミット**: 一つのコミットは一つの変更のみ
- **絵文字付き日本語**: 視覚的にわかりやすく
- **概要は一行、詳細は本文**: 必要に応じて詳細説明を追加

**コミット絵文字例**:
- 🔧 `fix:` バグ修正
- ✨ `feat:` 新機能
- 📚 `docs:` ドキュメント
- ♻️ `refactor:` リファクタリング
- 🎨 `style:` コードスタイル
- ✅ `test:` テスト
- 🚀 `perf:` パフォーマンス
- 🔨 `build:` ビルド関連

## 開発原則

### YAGNI (You Aren't Gonna Need It)
必要になるまで実装しない。将来の拡張を予測した過剰な設計を避ける。

### DRY (Don't Repeat Yourself)
同じコードを繰り返さない。共通処理は関数化・モジュール化する。

### KISS (Keep It Simple, Stupid)
シンプルに保つ。複雑な解決策より単純な解決策を選ぶ。

### リリースプロセス

#### **ワンコマンドリリース（v1.7.0以降、推奨）**

```bash
# 1. version.goのバージョンを更新
# 2. 自動リリーススクリプト実行
./scripts/release.sh
```

このスクリプトが自動実行する内容：
- バージョン番号の自動取得（version.goから）
- Windowsビルド実行
- Gitタグの作成とプッシュ
- GitHub Releaseの作成
- ビルド成果物の自動アップロード
- リリースノートの自動生成

#### **手動リリース（従来の方法）**

#### **1. バージョン番号の更新**
```bash
# version.go を編集
const Version = "1.7.0"  # 新しいバージョンに変更
```

#### **2. ビルド実行**
```bash
# macOS
cd build/macos
./build.sh clean
./build.sh  # tar.gz版

# Windows（Windows環境で）
cd build\windows
build.bat clean
build.bat
```

#### **3. GitHub CLIでリリース作成**
```bash
# リリース作成と同時にアセットアップロード
gh release create v1.7.0 \
  --title "v1.7.0" \
  --notes "リリースノート" \
  build/releases/koemoji-go-1.7.0.zip

# リリース確認
gh release view v1.7.0
```

#### **GitHub CLI セットアップ（初回のみ）**
```powershell
# インストール
winget install --id GitHub.cli

# 認証
gh auth login
```

#### **gh CLIの便利なコマンド**
```bash
# リリース一覧
gh release list

# 特定リリースの詳細表示
gh release view v1.7.0

# アセット追加アップロード
gh release upload v1.7.0 build/releases/koemoji-go-macos-1.7.0.dmg

# アセット削除（間違えた場合）
gh release delete-asset v1.7.0 ファイル名.zip

# リリース削除（やり直す場合）
gh release delete v1.7.0
```

#### **6. リリースノートのテンプレート**
```markdown
## 🎉 v1.X.Xの主な変更

### ✨ 新機能
- 機能1
- 機能2

### 🔧 変更内容
- 変更1
- 変更2

### 🐛 バグ修正
- 修正1
- 修正2

---

## 📦 ダウンロード

### macOS
- tar.gz版（GUI/TUI両対応）

### Windows
- ZIP形式（デュアル録音機能対応）

---

**Full Changelog**: https://github.com/infoHiroki/KoeMoji-Go/compare/v1.X.X...v1.Y.Y
```

## 開発のヒント

1. **設定変更時**: 必ずvalidate.goでバリデーション追加
2. **新機能追加時**: config.example.json更新を忘れずに
3. **エラー処理**: ユーザーフレンドリーなメッセージを心がける
4. **クロスプラットフォーム**: OS固有の処理は明確に分離
5. **テスト**: 手動テストコマンドをmanual-test-commands.mdに追加

## 連絡先とサポート

- **GitHub Issues**: バグ報告と機能要望
- **作者**: [@infoHiroki](https://github.com/infoHiroki)
- **ライセンス**: 個人利用自由、商用利用要連絡