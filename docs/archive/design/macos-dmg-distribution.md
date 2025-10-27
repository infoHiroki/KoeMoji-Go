# macOS DMG配布の設計ドキュメント

**作成日**: 2025-01-23
**バージョン**: v1.0
**ステータス**: Draft

---

## 1. 背景と目的

### 1.1 現状の課題

KoeMoji-Go v1.6.1は、macOS向けにtar.gz形式で配布している。この形式には以下の課題がある：

#### ユーザー体験の課題
- **ターミナル操作が必須**: 起動に`./koemoji-go`コマンドの実行が必要
- **技術的ハードル**: 非技術者にとって敷居が高い
- **macOSらしさの欠如**: Launchpadやアプリケーションフォルダに表示されない
- **視認性の低さ**: Finderでのアイコン表示がない

#### 配布上の課題
- **見た目の問題**: GitHub Releasesでの印象が「開発者向けツール」に限定される
- **認知度の課題**: 「macOSアプリ」としての訴求力が弱い
- **ユーザー層の制限**: CLI/TUIに慣れたユーザーしか使えない

### 1.2 目的

macOS標準の.appバンドル形式およびDMG形式での配布を実現し、以下を達成する：

1. **ユーザー体験の向上**
   - ダブルクリックでの起動
   - Launchpad、Dockからのアクセス
   - macOSネイティブな使い心地

2. **ユーザー層の拡大**
   - 非技術者でも使えるようにする
   - GUI中心のユーザーをサポート
   - 技術者向けCLI版も継続提供

3. **プロフェッショナルな印象**
   - 配布形式の見栄え向上
   - 信頼性の向上
   - 将来的な商用展開への布石

---

## 2. 提案する解決策

### 2.1 基本方針

**段階的アプローチ**を採用する：

- **Phase 1（本設計の対象）**: 基本的な.app/DMG生成（署名なし）
- **Phase 2（将来）**: UX改善（設定ファイル自動配置、初回起動ダイアログ等）
- **Phase 3（将来）**: コード署名・公証（ユーザー増加時）

### 2.2 配布戦略

2つの配布形式を並行提供：

#### 1. DMG版（メイン・推奨）
- **ファイル名**: `koemoji-go-{VERSION}-macos.dmg`
- **対象ユーザー**: GUI中心、非技術者
- **内容**:
  - KoeMoji-Go.app（ダブルクリック起動）
  - Applicationsフォルダへのシンボリックリンク
  - README.txt（起動方法、注意事項）

#### 2. CLI版（継続）
- **ファイル名**: `koemoji-go-{VERSION}-macos-cli.tar.gz`
- **対象ユーザー**: 技術者、自動化用途
- **内容**: 従来通りの単一バイナリ + config.json

### 2.3 重要な設計判断

#### 署名なしでの配布
- **理由**: 年間13,800円の継続コストを回避
- **影響**: 初回起動時にセキュリティ警告（右クリック→開くで回避可能）
- **対策**: READMEとアプリ内で詳細に案内

#### CLI/TUIとの互換性維持
同一バイナリで両方をサポート：
```bash
# .appからもCLI/TUI利用可能
/Applications/KoeMoji-Go.app/Contents/MacOS/koemoji-go --tui
```

#### 設定ファイルの配置（Phase 1）
最小限の変更で実装するため、Phase 1では**現状維持**：
- カレントディレクトリの`./config.json`を優先
- 後方互換性を完全保持

---

## 3. 技術的詳細

### 3.1 使用技術

#### Fyne Package Tool
- **役割**: Go製FyneアプリをmacOS .appバンドルに変換
- **インストール**: `go install fyne.io/fyne/v2/cmd/fyne@latest`
- **利点**:
  - クロスプラットフォーム対応
  - アイコンの自動変換（PNG → ICNS）
  - メタデータの自動生成

#### hdiutil（macOS標準）
- **役割**: DMGファイルの作成
- **コマンド例**:
  ```bash
  hdiutil create -volname "KoeMoji-Go" \
    -srcfolder dist -ov -format UDZO output.dmg
  ```

### 3.2 アプリケーションID

**決定**: `com.hirokitakamura.koemoji-go`に統一

**理由**:
1. `internal/gui/app.go`で既に使用されている
2. 開発者名（hirokitakamura）が明示される
3. GitHubアカウント（infoHiroki）との整合性
4. 将来の他プロジェクトとの統一感（`com.hirokitakamura.*`）

**変更箇所**:
- `FyneApp.toml`のIDを修正

### 3.3 ディレクトリ構造

#### .appバンドル構造
```
KoeMoji-Go.app/
├── Contents/
│   ├── Info.plist              # メタデータ
│   ├── MacOS/
│   │   └── koemoji-go          # 実行バイナリ
│   └── Resources/
│       ├── icon.icns           # アプリアイコン
│       └── config.example.json # 設定テンプレート（オプション）
```

#### DMGマウント時の構造
```
KoeMoji-Go（マウント）/
├── KoeMoji-Go.app              # アプリケーション
├── Applications -> /Applications # シンボリックリンク
└── README.txt                  # ユーザー向け案内
```

---

## 4. Phase 1 実装スコープ

### 4.1 実装する機能

#### ✅ 含まれるもの
1. **Fyne-cliのセットアップ**
2. **FyneApp.tomlの修正**（ID統一、メタデータ追加）
3. **.app生成機能**（`build.sh app`）
4. **DMG生成機能**（`build.sh dmg`）
5. **CLI版ビルド機能**（`build.sh cli`）
6. **統合ビルド機能**（`build.sh all`）
7. **README_APP.md**（.app版ユーザー向けドキュメント）
8. **ドキュメント更新**（CLAUDE.md）

#### ❌ 含まれないもの（Phase 2以降）
1. **コード署名・公証**（年間コスト回避）
2. **設定ファイルのApplication Support移行**
3. **作業ディレクトリの自動配置**
4. **初回起動ウェルカムダイアログ**
5. **自動アップデート機能**

### 4.2 設計上の制約

#### 後方互換性
- 既存のCLI版ユーザーに影響を与えない
- 設定ファイルのパスや形式は変更しない
- 既存の動作を完全に保持

#### 最小限の変更
- コアロジック（internal/）は変更しない
- ビルドスクリプトとメタデータのみ修正
- テスト範囲を最小化

---

## 5. リスクと対策

### 5.1 セキュリティ警告（署名なし）

#### リスク
- 初回起動時に「開発元が未確認」の警告
- 非技術者が戸惑う可能性
- ネガティブな第一印象

#### 対策
1. **詳細なREADME作成**
   - スクリーンショット付き手順
   - 「なぜこの警告が出るのか」の説明
   - 安全性の保証

2. **DMGに案内を同梱**
   - README.txtを目立つ位置に配置
   - 簡潔で分かりやすい説明

3. **代替手段の提供**
   - CLI版も継続提供
   - 技術者は従来通り使用可能

### 5.2 設定ファイルの場所

#### リスク
- .app起動時、カレントディレクトリが予測不能
- 設定ファイルが見つからない可能性

#### 対策
1. **複数パスのフォールバック**（Phase 2で実装）
   ```
   1. カレントディレクトリ
   2. ~/Library/Application Support/KoeMoji-Go/
   3. ~/.config/koemoji-go/
   4. デフォルト設定
   ```

2. **Phase 1では最小限の対応**
   - 問題が発生したら即座にPhase 2へ移行

### 5.3 Python依存の問題

#### リスク
- FasterWhisperのPython依存
- .appからPython呼び出しが失敗する可能性

#### 対策
1. **既存の自動インストール機能**
   - 内部的には問題なく動作する想定
   - 初回起動時のエラーメッセージを分かりやすく

2. **動作確認の徹底**
   - Phase 1実装後に重点的にテスト

---

## 6. 段階的実装計画

### Phase 1: 基本的な.app/DMG生成（本設計）
**目標**: ダブルクリック起動可能なDMGを配布
**期間**: 2-3時間
**リリース**: v1.7.0

**成果物**:
- `koemoji-go-1.7.0-macos.dmg`
- `koemoji-go-1.7.0-macos-cli.tar.gz`

### Phase 2: UX改善（将来）
**目標**: よりmacOSらしい体験
**実装内容**:
- 設定ファイルのApplication Support対応
- 初回起動ウェルカムダイアログ
- 作業ディレクトリの自動配置
- GUIから設定変更時のパス自動切り替え

### Phase 3: コード署名（将来）
**目標**: セキュリティ警告の解消
**条件**:
- ユーザー数が一定規模に達した
- 寄付やスポンサーで年間コスト（13,800円）をカバーできる
- または収益化の目処が立った

**実装内容**:
- Apple Developer Program加入
- コード署名設定
- 公証（Notarization）
- 自動ビルドパイプライン

---

## 7. 成功基準

### Phase 1の完了条件

#### 機能要件
- ✅ .appがダブルクリックで起動する
- ✅ DMGからApplicationsフォルダにドラッグ&ドロップできる
- ✅ CLI版（tar.gz）も正常にビルドされる
- ✅ CLI/TUIモードが.appからも利用可能

#### 品質要件
- ✅ 既存機能（文字起こし、録音、AI要約）が正常動作
- ✅ 設定ファイルの読み込み・保存が正常動作
- ✅ Python（FasterWhisper）呼び出しが成功

#### ドキュメント要件
- ✅ README_APP.mdが分かりやすい
- ✅ セキュリティ警告の回避方法が明記されている
- ✅ CLAUDE.mdにビルド手順が追加されている

### ユーザー体験の向上指標（定性的）
- 非技術者がREADMEを見て起動できる
- 「macOSアプリらしい」と感じられる
- CLI版ユーザーが不便を感じない

---

## 8. 参考資料

### 技術ドキュメント
- [Fyne Packaging for Desktop](https://docs.fyne.io/started/packaging.html)
- [macOS Bundle Structure](https://developer.apple.com/library/archive/documentation/CoreFoundation/Conceptual/CFBundles/BundleTypes/BundleTypes.html)
- [hdiutil man page](https://ss64.com/osx/hdiutil.html)

### 類似プロジェクト
- 他のGo製GUIアプリのDMG配布例
- 署名なしで成功しているOSSプロジェクト

---

## 9. 変更履歴

| 日付 | バージョン | 変更内容 |
|------|-----------|---------|
| 2025-01-23 | v1.0 | 初版作成 |

---

## 10. 承認

| 役割 | 氏名 | 日付 | 承認 |
|------|------|------|------|
| 開発者 | Hiroki Takamura | 2025-01-23 | ✅ |
| レビュアー | Claude Code | 2025-01-23 | ✅ |
