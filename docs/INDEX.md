# KoeMoji-Go ドキュメント索引

## 📁 ドキュメント構成

このディレクトリには、KoeMoji-Goプロジェクトの全ドキュメントが整理されています。

### 🎯 [user/](./user/) - ユーザー向けドキュメント
エンドユーザーが使用するための詳細なガイドライン。

- **[BASIC_USAGE.md](./user/BASIC_USAGE.md)** - 基本的な操作方法とコマンド
- **[AI_FEATURES.md](./user/AI_FEATURES.md)** - AI要約機能の詳細設定
- **[RECORDING_SETUP.md](./user/RECORDING_SETUP.md)** - 録音機能セットアップガイド
- **[SYSTEM_AUDIO_RECORDING_MACOS.md](./user/SYSTEM_AUDIO_RECORDING_MACOS.md)** - macOSシステム音声録音ガイド
- **[TROUBLESHOOTING.md](./user/TROUBLESHOOTING.md)** - 問題解決とFAQ

### 💻 [developer/](./developer/) - 開発者向けドキュメント
コードベースの理解と開発に必要な技術文書。

- **[ARCHITECTURE.md](./developer/ARCHITECTURE.md)** - システムアーキテクチャ、パッケージ構成、設計思想
- **[DEVELOPMENT.md](./developer/DEVELOPMENT.md)** - 開発環境構築、ビルドプロセス、テスト手順
- **[MACOS_BUILD_GUIDE.md](./developer/MACOS_BUILD_GUIDE.md)** - macOS環境でのビルド手順
- **[WINDOWS_BUILD_GUIDE.md](./developer/WINDOWS_BUILD_GUIDE.md)** - Windows環境でのビルド手順（MSYS2使用）
- **[GITHUB_CLI.md](./developer/GITHUB_CLI.md)** - GitHub CLI自動化、リリース管理
- **[PROGRAM_FLOW.md](./developer/PROGRAM_FLOW.md)** - プログラムフロー詳細
- **[VERSION_UPDATE_CHECKLIST.md](./developer/VERSION_UPDATE_CHECKLIST.md)** - バージョン更新チェックリスト

### 📐 [design/](./design/) - 設計ドキュメント
システム設計と仕様に関する詳細な設計書。

- **[dual-recording-design.md](./design/dual-recording-design.md)** - デュアル録音設計（システム音声+マイク）
- **[macos-system-audio-capture.md](./design/macos-system-audio-capture.md)** - macOSシステム音声キャプチャ設計
- **[SoundRecorderDesign.md](./design/SoundRecorderDesign.md)** - 音声録音システム設計書
- **[VoiceMeeterIntegration.md](./design/VoiceMeeterIntegration.md)** - VoiceMeeter統合設計
- **[KISS-Design-Principles.md](./design/KISS-Design-Principles.md)** - シンプル設計原則
- **[MOCK_TESTING_DESIGN.md](./design/MOCK_TESTING_DESIGN.md)** - モックテスト設計
- **[UI_DEVELOPMENT_POLICY.md](./design/UI_DEVELOPMENT_POLICY.md)** - UI開発ポリシー

### 🔧 [technical/](./technical/) - 技術ノート
実装時の技術的課題と解決策を記録したテクニカルノート。

- **[FASTER_WHISPER.md](./technical/FASTER_WHISPER.md)** - FasterWhisper技術文書
- **[fyne-implementation.md](./technical/fyne-implementation.md)** - Fyneフレームワーク実装における課題と解決策
- **[Fyne-Deep-Research.md](./technical/Fyne-Deep-Research.md)** - Fyneフレームワークの詳細調査
- **[windows-process-execution.md](./technical/windows-process-execution.md)** - Windows環境での外部プロセス起動とコンソールウィンドウ制御

### 💼 [business/](./business/) - ビジネス・営業向けドキュメント
商用利用ガイドと営業資料。

- **[COMMERCIAL_USE.md](./business/COMMERCIAL_USE.md)** - 商用利用ガイド
- **[ONE_PAGER.md](./business/ONE_PAGER.md)** - 1ページ営業資料
- **[SECURITY_WHITEPAPER.md](./business/SECURITY_WHITEPAPER.md)** - セキュリティホワイトペーパー
- **[SERVICE_INTRODUCTION.md](./business/SERVICE_INTRODUCTION.md)** - サービス紹介
- **[SUPPORT_PLANS.md](./business/SUPPORT_PLANS.md)** - サポートプラン詳細

### 🧪 [testing/](./testing/) - テスト関連ドキュメント
テスト手順とチェックリスト。

- **[MANUAL_TEST_CHECKLIST.md](./testing/MANUAL_TEST_CHECKLIST.md)** - 手動テストチェックリスト

### 📦 [archive/](./archive/) - アーカイブドキュメント
過去のバージョンや非推奨ドキュメントの保管場所。

- **[creative/](./archive/creative/)** - 創作・教育コンテンツ（3ファイル）
- **[tasks/](./archive/tasks/)** - 完了済みタスク（2ファイル）
- **[design/](./archive/design/)** - 古い設計文書（4ファイル）
- **[progress/](./archive/progress/)** - v1.6.x開発進捗記録（7ファイル）
- **[CODE_REVIEW_2025-10-27.md](./archive/CODE_REVIEW_2025-10-27.md)** - コードレビュー結果
- **[USAGE.md](./archive/USAGE.md)** - 旧使用方法ガイド

## 🗂️ その他の重要なドキュメント

プロジェクトルートには以下の重要なドキュメントがあります：

- **[README.md](../README.md)** - プロジェクト概要とクイックスタート
- **[TROUBLESHOOTING.md](user/TROUBLESHOOTING.md)** - 問題解決とFAQ
- **[CLAUDE.md](../CLAUDE.md)** - AI開発アシスタント向けプロジェクト情報

## 📝 ドキュメント利用ガイド

### 初めてのユーザー
1. [README.md](../README.md) でプロジェクト概要とクイックスタートを確認
2. [user/BASIC_USAGE.md](./user/BASIC_USAGE.md) で詳細な使用方法を学習
3. 問題が発生したら [user/TROUBLESHOOTING.md](./user/TROUBLESHOOTING.md) を確認

### 開発者・コントリビューター
1. [developer/ARCHITECTURE.md](./developer/ARCHITECTURE.md) でシステム構成を理解
2. [developer/DEVELOPMENT.md](./developer/DEVELOPMENT.md) で開発環境を構築
3. [technical/fyne-implementation.md](./technical/fyne-implementation.md) で実装課題を確認

### システム設計者
1. [design/dual-recording-design.md](./design/dual-recording-design.md) でデュアル録音設計を確認
2. [design/SoundRecorderDesign.md](./design/SoundRecorderDesign.md) で音声システム設計を理解
3. [design/macos-system-audio-capture.md](./design/macos-system-audio-capture.md) でmacOS音声キャプチャ設計を確認

## 🔄 ドキュメント更新方針

- **構造変更**: 新しい機能や重要な変更時にドキュメント構造を見直し
- **内容更新**: 機能追加・変更時に該当ドキュメントを同時更新
- **言語対応**: 重要なユーザー向けドキュメントは日英両言語で提供
- **クロスリファレンス**: 関連ドキュメント間の相互参照を充実

## 📞 フィードバック・問い合わせ

ドキュメントに関する改善提案や質問は、GitHubのIssueまたはPull Requestでお知らせください。