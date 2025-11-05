# Changelog

All notable changes to KoeMoji-Go will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.8.4] - 2025-11-05

### Added
- **診断機能実装**（`--doctor`）
  - クライアント環境のトラブルシューティング用診断ツール
  - `internal/diagnostics/` パッケージ新規作成（7ファイル、553行追加）
  - システム情報収集（OS、バージョン、実行パス）
  - オーディオデバイス列挙（PortAudio経由、全デバイス検出）
  - 仮想デバイス検出（VoiceMeeter、Virtual Cable、Stereo Mix等）
  - デュアル録音対応チェック（Windows WASAPI/COM、macOS ScreenCaptureKit）
  - config.json検証（デバイス名存在確認）
  - サマリーレポート生成（チェック合否、警告、エラー）
- **ダブルクリック実行可能な診断バッチファイル**
  - Windows: `診断実行.bat`
  - macOS: `診断実行.command`（実行権限付き）
  - 診断結果を`診断結果.txt`に自動保存
  - クライアントが簡単にサポートに送付可能

### Changed
- **バージョン情報の動的取得**
  - `internal/diagnostics/system.go`でバージョンがハードコーディングされていた問題を修正
  - `main.version`から動的に取得するように変更
  - `diagnostics.SetVersion(version)`で設定
- **ビルドスクリプト更新**
  - `build/windows/build.bat`: 診断実行.batをパッケージに追加（184行目）
  - `build/macos/build.sh`: 診断実行.commandをパッケージに追加（219-220行目）

### Files
- `internal/diagnostics/diagnostics.go` - エントリーポイント
- `internal/diagnostics/system.go` - システム情報収集
- `internal/diagnostics/audio.go` - デバイス列挙
- `internal/diagnostics/audio_windows.go` - Windows固有チェック（COM/WASAPI）
- `internal/diagnostics/audio_darwin.go` - macOS固有チェック（ScreenCaptureKit）
- `internal/diagnostics/config.go` - 設定検証
- `internal/diagnostics/output.go` - サマリー生成

## [1.8.3] - 2025-11-03

### Changed
- **CPU使用を無条件で強制**
  - GPU初期化問題を完全回避
  - 全てのcompute_typeで`--device cpu`を自動追加
  - Windows環境での文字起こし速度問題を解決
  - ファイル：`internal/whisper/whisper.go:204-205`
  - 変更内容：`compute_type`による条件分岐を削除し、無条件でCPU使用

### Added
- **処理時間ログ追加**
  - 文字起こし完了時に処理時間を記録
  - パフォーマンス問題の診断が容易に
  - 例：「Transcription completed in 25m30s」

### Fixed
- **テスト更新**
  - `TestTranscribeAudio_DeviceParameter`を更新
  - 全てのcompute_typeでCPU使用を検証

### Background
- ユーザー環境で1時間の音声が5時間かかる問題が報告
- GPU使用時の初期化失敗・リトライが原因と推定
- CPU専用化により、処理時間を大幅短縮（期待値：5時間→30分）

## [1.8.2] - 2025-11-02

### Added
- **データ消失防止機能**: `validateOutputFile()` 関数追加（`internal/whisper/whisper.go:421-449`）
  - 0バイト出力ファイルを検出
  - 破損した音声/動画ファイル処理時に元ファイルを`input/`フォルダに保持
  - これまで：0バイト出力でも「成功」と判断し、`archive/`に移動してデータ消失
  - 改善後：0バイト検出でエラー返却、元ファイルは`input/`に保持

### Fixed
- **faster-whisperインストール時の依存関係問題**（Issue #19 → PR #21）
  - pipをアップグレードして依存関係解決を改善
  - `requests`モジュールを明示的にインストール
  - 古いpipバージョンでの間接依存関係エラーを解決
  - ファイル：`internal/whisper/whisper.go:136-157`
- **デュアル録音エラーロギング**（Issue #20 → PR #22）
  - デュアル録音エラーがGUI/TUI/ログファイルに記録されるように修正
  - `fmt.Printf` → `logger.LogError`に変更（5箇所）
  - DualRecorder構造体にロギングフィールド追加
  - Windows/macOS両対応
  - ファイル：`internal/recorder/dual_recorder.go`、`dual_recorder_darwin.go`

### Changed
- **エラーメッセージ改善**
  - ユーザーフレンドリーなエラーメッセージに改善
  - 技術用語（FFmpeg、コマンドライン例）を削除
  - 確証のない推測情報を削除
  - シンプルで分かりやすい表現に統一
  - 「音声ファイル」→「音声/動画ファイル」に汎用化

### Closed Issues
- Issue #19: faster-whisper requests不足 → 解決 ✅
- Issue #20: デュアル録音エラーログ → 解決 ✅
- Issue #15: installFasterWhisper冗長 → 方針転換によりクローズ
- Issue #16: Draft issue → プレースホルダーのためクローズ
- Issue #17: 設定ファイル読み込み問題 → 無関係のためクローズ

### Known Limitations
- PR #22（デュアル録音エラーロギング）の実動作検証は未実施（Issue #23で追跡中）
- Windows版はビルド・検証後に別途リリース予定

## [1.8.1] - 2025-10-29

### Changed
- **TUI正式版化（Phase 14完了）**
  - Rich TUIを正式版に昇格（`tui_rich.go` → `tui_tview.go`）
  - `RichTUI` → `TUI`に命名を簡素化
  - `--tui-rich`フラグを削除、`--tui`に統一（6文字削減）
  - 旧Simple TUI実装を削除（~110行削除）
  - テスト全面書き直し（247行、全合格）
  - 未使用コードクリーンアップ（正味-97行）
- **Windows版ビルド成果物の命名規則変更**
  - セキュリティフィルタ回避のため「windows」という単語を完全排除
  - 旧: `koemoji-go-windows-{VERSION}.zip`
  - 新: `koemoji-go-{VERSION}.zip`
  - macOS版は変更なし: `koemoji-go-macos-{VERSION}.tar.gz`
  - 背景: 「windows」という単語がセキュリティソフトやブラウザのフィルタリングに引っかかる問題

### Fixed
- **TUI設定画面のバグ修正**
  - キャンセルボタンが動作しない問題を修正
  - 録音設定追加後のインデックスずれに対応
  - case文のインデックスを正しく更新（case 5→6に修正）

### Added
- **デュアル録音の注意喚起追加**
  - ヘッドホン/イヤホン推奨の警告をドキュメントに追加
  - GUI設定画面に情報ラベル追加（macOS向け）
  - 物理的な音響結合問題をわかりやすく説明
  - GitHub Issue #18作成（将来のエコーキャンセレーション機能）

### Documentation
- **TUI制限のドキュメント明記**
  - TUIモード（`--tui`）はmacOS専用であることを明記
  - WindowsではGUIモードのみサポート
  - 理由: tview/tcellライブラリのWindows互換性問題（"The handle is invalid."エラー）

## [1.8.0] - 2025-10-27

### Added
- **macOSデュアル録音機能実装**（Phase 0-4完了）
  - ScreenCaptureKit APIを使用したシステム音声キャプチャ
  - Swift CLIバイナリ（`cmd/audio-capture`）の統合
  - 2ストリーム方式（システム音声とマイク音声を別ファイルで保存）
  - macOS 13以降対応、画面収録権限が必要
- **実装ファイル**
  - `internal/recorder/system_audio_darwin.go` (242行) - Swift CLI ラッパー
  - `internal/recorder/dual_recorder_darwin.go` (386行) - デュアル録音実装
  - `cmd/audio-capture/AudioCapture.swift` - ScreenCaptureKit統合
  - `cmd/audio-capture/main.swift` - CLI エントリーポイント
- **GUIデュアル録音切り替え機能**
  - 設定画面に録音モード選択ラジオボタン追加
  - "シングル録音（マイクのみ）" / "デュアル録音（システム音声+マイク）"
  - macOS 13以降と画面収録権限が必要な旨の情報ラベル表示
  - Windows版と統一されたUIパターン
- **FFmpeg不要の1ファイルミキシング機能**
  - `internal/recorder/mixer.go` (366行) - Go標準ライブラリのみでWAVミキシング実装
  - 線形補間による44.1kHz→48kHzリサンプリング
  - ステレオ+モノラルのミキシング（システム70%, マイク100%）
  - ソフトクリッピングによる歪み防止
  - afconvert互換性対応（FLLR等の余分なチャンク処理）
  - デフォルトで1ファイル出力（`recording_*.wav`）、重複文字起こし防止
  - `SaveSeparateFiles()`で2ファイル出力も可能（話者分離用）

### Changed
- **ビルドシステム更新**
  - `build/macos/build.sh`: Swift CLIバイナリを自動パッケージング
  - リリース版に`audio-capture`バイナリ（171KB）を同梱

### Technical Specifications
- システム音声: 48kHz Float32 Stereo (CAF → WAV自動変換)
- マイク音声: 44.1kHz Int16 Mono
- SIGTERMによる優雅な停止処理（DispatchSourceSignal）
- バイナリ検索パス: 実行ファイルディレクトリ優先

### Documentation
- `docs/user/SYSTEM_AUDIO_RECORDING_MACOS.md` - macOS版ユーザーガイド追加
- `README.md` - デュアル録音機能のmacOS対応を反映
- `TEST_RESULTS.md` - 全自動テスト合格（8/8）、手動テスト合格（3/3）

### Tests
- 自動テスト: ビルド、環境確認、デバイス検出、設定シナリオ、統合テスト（全合格）
- 手動テスト: GUI/TUIモードでのシングル/デュアル録音動作確認（全合格）
- `internal/recorder/mixer_test.go` (290行) - 8つのユニットテスト（全合格）
- `internal/recorder/test_afconvert_wav_test.go` - afconvert互換性テスト

## [1.7.2] - 2025-06-23

### Changed
- macOS配布形式をDMGからtar.gz単一バイナリに変更
- 理由: Apple Developer Program年間費用削減のため

### Removed
- macOS DMG版の配布を廃止

## [1.7.0] - 2025-01-22

### Changed
- **ビルド成果物の命名規則変更**
  - ウイルス検知回避強化のためプラットフォーム名を中央に配置
  - macOS: `koemoji-go-macos-{VERSION}.tar.gz`
  - Windows: `koemoji-go-windows-{VERSION}.zip`（v1.8.1で再変更）

## [1.6.1] - 2025-01-22

### Fixed
- **Windows GUI 日本語文字化け修正**
  - Fyne + Windows + 日本語 + Boldスタイルの組み合わせで文字化けが発生
  - `internal/gui/components.go`, `internal/gui/window.go`からBoldスタイルを削除
  - 通常フォントでの表示に変更

### Changed
- **ビルド成果物の命名規則変更**
  - 旧: `KoeMoji-Go-v1.6.0-win.zip`
  - 新: `koemoji-go-1.6.1-windows.zip`
  - 理由: `-win`サフィックスがアンチウイルスソフトに誤検知される問題への対応
  - 全て小文字、プラットフォーム名を明確に（`windows`, `macos`）

## [1.6.0] - 2025-01-21

### Added
- **VoiceMeeter統合機能**（Windows専用）
  - システム音声+マイクの同時録音対応
  - VoiceMeeter自動検出機能
  - 音量自動正規化機能（閾値5000、目標20000）
  - GUI設定ダイアログに「VoiceMeeter設定を適用」ボタン追加

### Fixed
- **Windowsビルドシステム修正**
  - goversioninfo実行時のエラー修正
  - アイコン埋め込み機能の復元
  - ビルドスクリプトの自動化改善

## [1.5.5] - 2025-01-21

### Changed
- **録音デバイス設定の改善**
  - デバイスIDから名前ベースの指定に変更
  - 環境非依存の設定を実現
  - 空文字列でデフォルトデバイスを使用

## [1.5.4] - 2025-01-21

### Changed
- **フォルダ命名規則の統一**
  - 旧: `koemoji-go-windows-1.5.4`、`koemoji-go-macos-arm64-1.5.4`
  - 新: `KoeMoji-Go-v1.5.4`（全プラットフォーム共通）
  - 効果: アップデート時のユーザーデータ保護
- **Windows DLL処理の簡素化**
  - 旧: MSYS2パス検出の複雑な条件分岐（14行）
  - 新: ワイルドカード`*.dll`使用（4行）
  - 効果: 保守性向上、新DLL自動対応

### Improved
- **FasterWhisper自動インストール改善**
  - Windows実行ファイルパス対応
  - GPU環境での`--device cpu`自動追加
  - クロスプラットフォーム対応強化

---

## Version Numbering

KoeMoji-Go follows [Semantic Versioning](https://semver.org/):
- **MAJOR** version: 互換性のない変更
- **MINOR** version: 後方互換性のある機能追加
- **PATCH** version: 後方互換性のあるバグ修正

## Release Notes

各バージョンの詳細なリリースノートは、[GitHub Releases](https://github.com/infoHiroki/KoeMoji-Go/releases)を参照してください。
