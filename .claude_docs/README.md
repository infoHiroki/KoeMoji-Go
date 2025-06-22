# Claude Docs - カスタムコマンド

このディレクトリには、KoeMoji-Goプロジェクトで使用するClaude Code用のカスタムコマンドが含まれています。

## /release コマンド

手動ビルドプロセスを自動化し、GitHubリリースを作成するコマンドです。

### 使い方

```bash
# Claude Codeで以下のコマンドを実行
/release
```

### 機能

1. **Git状態の確認**: 未コミットの変更を警告
2. **バージョン管理**: 自動的にバージョン番号を更新
3. **ビルド実行**: build.shスクリプトを実行
4. **成果物確認**: 生成されたzipファイルを表示
5. **GitHubリリース作成**: タグ付けとリリース作成（オプション）

### 前提条件

- macOS環境
- PortAudioがインストール済み
- GitHub CLIがインストール・認証済み
- mingw-w64がインストール済み（Windowsクロスコンパイル用）

### 出力ファイル

- `koemoji-go-windows-{version}.zip`
- `koemoji-go-macos-intel-{version}.tar.gz`
- `koemoji-go-macos-arm64-{version}.tar.gz`