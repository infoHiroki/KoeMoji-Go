# 開発進捗

## 現在の作業状況 (2025-10-31)

### ✅ 完了したタスク

#### 1. Issue #19: faster-whisper インストール時の requests 不足
- **PR**: #21
- **ブランチ**: `fix/issue-19-requests-module`
- **ステータス**: PR作成済み、手動テスト待ち
- **変更内容**:
  - `internal/whisper/whisper.go:136-157` 修正
  - pipアップグレード追加
  - requestsを明示的にインストール
- **テスト結果**: 全テスト合格 ✅

#### 2. Issue #20: デュアル録音のエラーログ改善
- **PR**: #22
- **ブランチ**: `fix/issue-20-dual-recording-logging`
- **ステータス**: PR作成済み、手動テスト待ち
- **変更内容**:
  - `internal/recorder/dual_recorder.go` (Windows版) 修正
  - `internal/recorder/dual_recorder_darwin.go` (macOS版) 修正
  - GUI/CLI呼び出し側修正
  - fmt.Printf → logger.LogError（5箇所）
- **テスト結果**: 全テスト合格 ✅

---

### 🚧 次のステップ

#### 1. 手動テスト
- [ ] PR #21: Windows環境でfaster-whisperインストールテスト
- [ ] PR #22: デュアル録音エラーログの動作確認

#### 2. PRマージ
- [ ] PR #21をmainにマージ
- [ ] PR #22をmainにマージ

#### 3. Windowsビルド
- [ ] v1.8.2のWindowsビルド作成
  - 場所: `build/windows/build.bat`
  - 成果物: `koemoji-go-1.8.2.zip`
  - 必要環境: MSYS2/MinGW64

#### 4. リリース準備
- [ ] version.goのバージョン更新（v1.8.1 → v1.8.2）
- [ ] CHANGELOG更新
- [ ] GitHub Release作成

---

### 📋 関連ドキュメント

- **ビルド手順**: `CLAUDE.md` - ビルドシステムセクション
- **リリース手順**: `CLAUDE.md` - リリースプロセスセクション
- **テスト手順**: `test/manual-test-commands.md`

---

### 💡 備考

#### Windowsビルドについて
- Windows環境が必要（MSYS2/MinGW64）
- macOSからのクロスコンパイルは非対応（CGO依存のため）
- ビルドコマンド:
  ```cmd
  cd build\windows
  build.bat clean
  build.bat
  ```

#### 命名規則 (v1.8.1以降)
- Windows: `koemoji-go-{VERSION}.zip`（"windows"という単語を排除）
- macOS: `koemoji-go-macos-{VERSION}.tar.gz`

---

### 🐛 既知の問題

- Issue #18: エコーキャンセレーション機能（将来対応）
  - デバッグ機能改善後に着手推奨

---

最終更新: 2025-10-31
