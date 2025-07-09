# KoeMoji-Go ドキュメント索引

## 📁 ドキュメント構成

このディレクトリには、KoeMoji-Goプロジェクトの全ドキュメントが整理されています。

### 🎯 [user/](./user/) - ユーザー向けドキュメント
エンドユーザーが使用するための詳細なガイドライン。

- **[USAGE.md](./user/USAGE.md)** - 使用方法ガイドの目次・ナビゲーション
- **[BASIC_USAGE.md](./user/BASIC_USAGE.md)** - 基本的な操作方法とコマンド
- **[AI_FEATURES.md](./user/AI_FEATURES.md)** - AI要約機能の詳細設定
- **[RECORDING_SETUP.md](./user/RECORDING_SETUP.md)** - 録音機能セットアップガイド
- **[TROUBLESHOOTING.md](./user/TROUBLESHOOTING.md)** - 問題解決とFAQ

### 💻 [developer/](./developer/) - 開発者向けドキュメント
コードベースの理解と開発に必要な技術文書。

- **[ARCHITECTURE.md](./developer/ARCHITECTURE.md)** - システムアーキテクチャ、パッケージ構成、設計思想
- **[DEVELOPMENT.md](./developer/DEVELOPMENT.md)** - 開発環境構築、ビルドプロセス、テスト手順
- **[WINDOWS_BUILD_GUIDE.md](./developer/WINDOWS_BUILD_GUIDE.md)** - Windows環境でのビルド手順（MSYS2使用）
- **[GITHUB_CLI.md](./developer/GITHUB_CLI.md)** - GitHub CLI自動化、リリース管理
- **[CLAUDE.md](./developer/CLAUDE.md)** - AI開発アシスタント向けプロジェクト情報

### 📐 [design/](./design/) - 設計ドキュメント
システム設計と仕様に関する詳細な設計書。

- **[DASHBOARD_GUI_DESIGN.md](./design/DASHBOARD_GUI_DESIGN.md)** - GUI設計仕様、ユーザーインターフェース設計原則
- **[SoundRecorderDesign.md](./design/SoundRecorderDesign.md)** - 音声録音システム設計書（v3.0完全実装版）
- **[KISS-Design-Principles.md](./design/KISS-Design-Principles.md)** - シンプル設計原則

### 🔧 [technical/](./technical/) - 技術ノート
実装時の技術的課題と解決策を記録したテクニカルノート。

- **[fyne-implementation.md](./technical/fyne-implementation.md)** - Fyneフレームワーク実装における課題と解決策
- **[windows-process-execution.md](./technical/windows-process-execution.md)** - Windows環境での外部プロセス起動とコンソールウィンドウ制御
- **[Fyne-Deep-Research.md](./technical/Fyne-Deep-Research.md)** - Fyneフレームワークの詳細調査

### 🎨 [creative/](./creative/) - 創作・教育コンテンツ
技術学習や創作的なアプローチでのドキュメント。

- **[GO_LANGUAGE_ESSAY.md](./creative/GO_LANGUAGE_ESSAY.md)** - Go言語学習エッセイ
- **[LYNCH_NARRATIVE.md](./creative/LYNCH_NARRATIVE.md)** - 芸術的解釈によるコードベース説明

## 🗂️ その他の重要なドキュメント

プロジェクトルートには以下の重要なドキュメントがあります：

- **[README.md](../README.md)** - プロジェクト概要とクイックスタート
- **[TROUBLESHOOTING.md](user/TROUBLESHOOTING.md)** - 問題解決とFAQ
- **[CLAUDE.md](../CLAUDE.md)** - AI開発アシスタント向けプロジェクト情報

## 📝 ドキュメント利用ガイド

### 初めてのユーザー
1. [README.md](../README.md) でプロジェクト概要とクイックスタートを確認
2. [user/USAGE.md](./user/USAGE.md) で詳細な使用方法を学習
3. 問題が発生したら [TROUBLESHOOTING.md](user/TROUBLESHOOTING.md) を確認

### 開発者・コントリビューター
1. [developer/ARCHITECTURE.md](./developer/ARCHITECTURE.md) でシステム構成を理解
2. [developer/DEVELOPMENT.md](./developer/DEVELOPMENT.md) で開発環境を構築
3. [technical/fyne-implementation.md](./technical/fyne-implementation.md) で実装課題を確認

### システム設計者
1. [design/DASHBOARD_GUI_DESIGN.md](./design/DASHBOARD_GUI_DESIGN.md) でGUI設計を確認
2. [design/SoundRecorderDesign.md](./design/SoundRecorderDesign.md) で音声システム設計を理解

## 🔄 ドキュメント更新方針

- **構造変更**: 新しい機能や重要な変更時にドキュメント構造を見直し
- **内容更新**: 機能追加・変更時に該当ドキュメントを同時更新
- **言語対応**: 重要なユーザー向けドキュメントは日英両言語で提供
- **クロスリファレンス**: 関連ドキュメント間の相互参照を充実

## 📞 フィードバック・問い合わせ

ドキュメントに関する改善提案や質問は、GitHubのIssueまたはPull Requestでお知らせください。