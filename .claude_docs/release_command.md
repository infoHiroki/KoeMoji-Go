# /release カスタムコマンド仕様

## 概要
`/release` コマンドは、KoeMoji-Goの手動ビルドプロセスを自動化するカスタムコマンドです。

## 実行手順

### 1. 事前確認
```bash
# 現在のブランチとクリーンな状態を確認
git status
```

### 2. ビルドスクリプトの実行
```bash
cd build
./build.sh
```

### 3. ビルド成果物の確認
```bash
# 生成されたファイルを確認
ls -la koemoji-go-*.zip
```

### 4. リリースタグの作成（オプション）
```bash
# バージョンタグを作成
git tag -a v1.x.x -m "Release version 1.x.x"
git push origin v1.x.x
```

### 5. GitHubリリースの作成
```bash
# GitHub CLIを使用してリリースを作成
gh release create v1.x.x \
  koemoji-go-windows.zip \
  koemoji-go-macos-intel.zip \
  koemoji-go-macos-apple-silicon.zip \
  --title "KoeMoji-Go v1.x.x" \
  --notes "Release notes here"
```

## エラーハンドリング

- ビルドが失敗した場合は、エラーログを確認
- goversioninfoのインストールが必要な場合：
  ```bash
  go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
  ```

## 注意事項

- macOSでの実行が前提
- CGO_ENABLED=1でビルドするため、PortAudioのインストールが必要
- Windows向けクロスコンパイルにはmingw-w64が必要