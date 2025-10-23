# macOS DMG配布の技術的課題と解決策

**作成日**: 2025-10-23
**状態**: 調査完了・実装保留（現在は解決策3「CLI版のみ配布」を採用中）

---

## 概要

KoeMoji-Go v1.7.0でmacOS DMG配布（.appバンドル形式）を実装したが、**macOSのQuarantine属性により、一般ユーザーが起動できない**という致命的な問題が発覚した。

本ドキュメントは、問題の詳細、調査プロセス、解決策の比較、今後の方針をまとめたものである。

---

## 実装内容（v1.7.0）

### 1. DMG配布システム
- `.appバンドル形式`でのパッケージング
- `PortAudioライブラリ`の自動バンドリング（`@executable_path`参照）
- Ad-hoc署名による動作保証

### 2. パス解決システム
- `IsRunningAsApp()` - .app実行検出
- `GetAppBaseDir()` - 実行環境別のベースディレクトリ取得
  - .app版: `~/Documents/KoeMoji-Go/`
  - CLI版: 実行ファイルのディレクトリ
- データ（input/output/archive/config/log）を一元管理

### 3. macOSマイク許可対応
- Info.plistに`NSMicrophoneUsageDescription`を追加
- 初回起動時にマイクアクセス許可ダイアログを表示
- 録音機能の正常動作を実現

---

## 遭遇した問題

### 症状

GitHubからDMGをダウンロードして起動すると、以下のエラーが表示される：

```
"KoeMoji-Go.app" は壊れているため開けません。

Apple は、"KoeMoji-Go.app" に Mac に害を及ぼす
可能性のあるマルウェアが含まれていないことを確認
できませんでした。
```

- **DMG上から直接起動**: エラー
- **Applicationsフォルダにコピー後に起動**: **同じエラー**
- **右クリック→「開く」**: **同じエラー（予想外）**

### ユーザー体験への影響

このエラーが表示された時点で、**ほとんどのユーザーは諦める**。

- README.txtを読み返す人はほぼいない
- 技術的な対処法（ターミナルコマンド）は一般ユーザーには不可能
- 「壊れている」という表現が誤解を招く（実際は壊れていない）

---

## 詳細な調査プロセスと結果

### 1. Quarantine属性の確認

```bash
$ ls -la@ /Applications/KoeMoji-Go.app
drwxr-xr-x@  3 user  admin    96 Oct 23 15:30 .
	com.apple.provenance	  11 
	com.apple.quarantine	  57  # ← これが原因
```

**結論**: GitHubからダウンロードしたファイルには`com.apple.quarantine`属性が自動付与される。

### 2. .appバンドル構造の検証

```bash
$ ls -la /Applications/KoeMoji-Go.app/Contents/MacOS/
-rwxr-xr-x@ 1 user  admin  23134560 Oct 23 15:30 koemoji-go

$ defaults read /Applications/KoeMoji-Go.app/Contents/Info CFBundleExecutable
koemoji-go

$ file /Applications/KoeMoji-Go.app/Contents/MacOS/koemoji-go
Mach-O 64-bit executable arm64
```

**結論**: .appバンドル構造は完全に正常。

### 3. コード署名の検証

```bash
$ codesign -vvv /Applications/KoeMoji-Go.app
/Applications/KoeMoji-Go.app: valid on disk
/Applications/KoeMoji-Go.app: satisfies its Designated Requirement
```

**結論**: Ad-hoc署名は有効。

### 4. ライブラリ依存関係の確認

```bash
$ otool -L /Applications/KoeMoji-Go.app/Contents/MacOS/koemoji-go | grep portaudio
@executable_path/../Frameworks/libportaudio.2.dylib (compatibility version 3.0.0, current version 3.0.0)

$ ls -la /Applications/KoeMoji-Go.app/Contents/Frameworks/
-r--r--r--@ 1 user  admin  139104 Oct 23 15:30 libportaudio.2.dylib
```

**結論**: PortAudioライブラリのバンドリングは正常。

### 5. 実行ファイルの動作確認

```bash
$ /Applications/KoeMoji-Go.app/Contents/MacOS/koemoji-go --version
KoeMoji-Go v1.7.0
```

**結論**: 実行ファイル自体は正常に動作する。

### 6. 解決策の検証

```bash
# Quarantine属性を削除
$ xattr -d com.apple.quarantine /Applications/KoeMoji-Go.app

# 再起動
$ open /Applications/KoeMoji-Go.app
# → 正常に起動！
```

**結論**: **Quarantine属性が原因**だった。

---

## なぜ一般ユーザーには対処できないのか

### 1. ターミナルコマンドの障壁

```bash
xattr -d com.apple.quarantine /Applications/KoeMoji-Go.app
```

- 一般ユーザーはターミナルを使わない
- コマンドの意味が理解できない
- タイプミスでエラーになる

### 2. 右クリック→「開く」が機能しない（macOS Sequoia の変更）

通常、**右クリック→「開く」でQuarantine警告を回避できる**はずだが、今回は**同じエラーが表示された**。

**原因**: macOS Sequoia（2024年リリース）から、Gatekeeperのポリシーが大幅に変更された。

- **変更前（Sonoma以前）**: 右クリック→「開く」でQuarantine警告を回避可能
- **変更後（Sequoia以降）**: 未公証アプリの右クリック回避機能が**完全に削除**
- **現在の回避方法**: システム設定 > プライバシーとセキュリティで個別に例外追加するしか方法がない

この変更により、**Ad-hoc署名のみのアプリは、一般ユーザーにとって事実上起動不可能**になった。

### 3. READMEを読まない

エラーが出た時点で、ユーザーは：
1. 「壊れている」と判断
2. アプリをゴミ箱に入れる
3. 他のソフトを探す

README.txtに戻って対処法を読む人は**ほぼいない**。

---

## 解決策の比較

### 解決策1: インストールスクリプト提供

#### 内容

```bash
#!/bin/bash
# install.sh
echo "KoeMoji-Goをインストールしています..."
xattr -dr com.apple.quarantine KoeMoji-Go.app
cp -R KoeMoji-Go.app /Applications/
echo "✅ インストール完了！Applicationsフォルダから起動してください。"
```

DMGに`install.sh`を同梱し、以下の手順をREADMEに記載：

```
1. DMGを開く
2. ターミナルを開く
3. cd /Volumes/KoeMoji-Go
4. bash install.sh
```

#### メリット
- コスト0円
- 技術的には確実

#### デメリット
- **技術者向け**（一般ユーザーには不可能）
- ターミナルを使えないユーザーは脱落
- セキュリティ警告が出る可能性

#### 評価
⭐⭐☆☆☆ （技術者のみ対応可能）

---

### 解決策2: Apple Developer Program加入

#### 内容

Apple Developer Program（年間$99 = 約12,000-14,800円）に加入し、以下を実施：

1. **Developer ID証明書取得**
2. **コード署名**（Hardened Runtime有効化）
3. **公証（Notarization）**

#### なぜこれで解決するのか

**Gatekeeperの動作原理**:
```
ダウンロードされたファイル
→ ブラウザが com.apple.quarantine 属性を付与
→ 初回起動時にGatekeeperがチェック
→ 公証済み: ✅ 起動許可（ダブルクリックで即起動）
→ 未公証: ❌ 起動拒否（Sequoia以降）
```

**重要**: 公証されたアプリは、Quarantine属性があっても**Gatekeeperがブロックしません**。GitHubからダウンロードしても、普通にダブルクリックで起動できます。

#### 具体的な実装手順

既存の`build/macos/build.sh`に以下を追加するだけ：

```bash
# 1. Developer ID証明書でコード署名（--options runtimeが重要）
codesign --force --options runtime --deep \
  --sign "Developer ID Application: Your Name (TEAM_ID)" \
  -i "com.hirokitakamura.koemoji-go" KoeMoji-Go.app

# 2. DMGを作成
hdiutil create -srcfolder KoeMoji-Go.app -volname "KoeMoji-Go" KoeMoji-Go.dmg

# 3. DMGに署名
codesign --sign "Developer ID Application: Your Name (TEAM_ID)" KoeMoji-Go.dmg

# 4. 公証申請（平均30秒で完了！）
xcrun notarytool submit KoeMoji-Go.dmg \
  --keychain-profile "notary-profile" \
  --wait

# 5. チケットを添付
xcrun stapler staple KoeMoji-Go.dmg
```

**初回セットアップ**（一度だけ）:
```bash
# 認証情報を保存
xcrun notarytool store-credentials "notary-profile" \
  --apple-id "your-apple-id@example.com" \
  --team-id "TEAM_ID" \
  --password "app-specific-password"
```

#### メリット
- **完全に解決**（ダブルクリックで起動可能）
- ユーザー体験が最高
- macOSの標準的な配布方法
- **公証は高速**（平均30秒）
- 実装は既存のビルドスクリプトに追加するだけ

#### デメリット
- **年間コスト**: $99（約12,000-14,800円、為替次第）
- 初回セットアップに手間がかかる
- Apple IDの2要素認証が必要
- Developer ID証明書の取得が必要（Account Holderのみ）

#### 評価
⭐⭐⭐⭐⭐ （唯一の根本的解決策）

**技術的には完全に実装可能で、コストのみが判断ポイント**

---

### 解決策3: DMG配布を諦める

#### 内容

- `.app配布を中止`
- `CLI版（tar.gz）のみ`を推奨
- 技術者向けのツールと割り切る

#### メリット
- コスト0円
- 問題が発生しない
- 技術者には問題ない

#### デメリット
- **GUIユーザーが使えない**
- ターミナルから起動する必要がある
- 一般ユーザーへの普及が困難

#### 評価
⭐⭐⭐☆☆ （現実的だが、UX的には妥協）

---

## 結論と今後の方針

### 現在の状況（2025-10-23時点）

**採用している解決策: 解決策3（CLI版のみ配布）**

1. **v1.7.0リリースを削除**（完了）
2. **本ドキュメントを作成**して経緯を記録
3. **CLI版（tar.gz）のみ**を配布中（v1.6.1ベース）

### 今後の判断基準

**Apple Developer Program加入を検討する際のポイント**:

1. **ターゲットユーザー**
   - 技術者のみ → CLI版で十分
   - 一般ユーザーも含む → Apple Developer Program必須

2. **年間コストの許容度**
   - $99（約12,000-14,800円/年）が負担 → CLI版継続
   - 許容できる → 最高のUX提供可能

3. **開発継続意欲**
   - 長期的に開発・サポート → 投資価値あり
   - 短期的なプロジェクト → CLI版で十分

4. **ユーザー数の見込み**
   - 多数のユーザーが見込める → 投資回収しやすい
   - 少数ユーザー → CLI版で十分

### 中期的な対応（検討中）

**Apple Developer Program加入を検討**する。

- **メリット**: 完全に解決、ユーザー体験向上、公証は高速（30秒）
- **コスト**: 年間$99（約12,000-14,800円）
- **実装難易度**: 低（既存のbuild.shに追加するだけ）

### 長期的な展望

将来的に以下のいずれかを選択：

1. **Apple Developer Program加入** → macOS DMG配布再開
2. **CLI版のみ** → 技術者向けツールとして継続
3. **クロスプラットフォーム配布** → Windowsに注力

---

## 技術的な学び

### macOSのセキュリティモデル

- **Gatekeeper**: ダウンロードされたアプリを検証
- **Quarantine属性**: インターネットからのファイルに自動付与
- **コード署名**: Developer ID証明書が必要
- **公証（Notarization）**: Apple のサーバーでマルウェアスキャン
  - 最新ツール: `xcrun notarytool`（2023年11月以降）
  - 廃止されたツール: `xcrun altool`（2023年11月廃止）
  - 処理速度: 平均30秒
- **macOS Sequoia（2024年〜）**: 右クリック→「開く」回避機能を削除

### Ad-hoc署名の限界

- 開発中のテストには十分
- **配布には不十分**（Quarantine属性で起動不可）
- `codesign --sign -` では解決しない

### Fyneのパッケージング

- `fyne package` でInfo.plistが自動生成される
- カスタマイズには`FyneApp.toml`が必要
- しかし、**Info.plistに直接追記する方が確実**

---

## 参考資料

### Apple公式ドキュメント

- [Notarizing macOS Software Before Distribution](https://developer.apple.com/documentation/security/notarizing_macos_software_before_distribution)
- [Gatekeeper and runtime protection](https://support.apple.com/guide/security/gatekeeper-and-runtime-protection-sec5599b66df/web)
- [Signing Mac Software with Developer ID](https://developer.apple.com/developer-id/)

### macOS Sequoiaの変更

- [Gatekeeper and notarization in Sequoia – The Eclectic Light Company](https://eclecticlight.co/2024/08/10/gatekeeper-and-notarization-in-sequoia/)

### Fyne + macOS公証の実装例

- [Build, package, sign, image, sign again, notarize, and staple a fyne.io app for MacOS](https://gist.github.com/blockpane/fe03eb0839fac417b92cd7eb98cdf356)

### 関連Issue

- GitHub Issue: (未作成)
- 参考: [electron/electron#7476](https://github.com/electron/electron/issues/7476)

---

## まとめ

**macOSでコード署名なしの.appを配布することは、技術的には可能だが、実用的ではない。**

一般ユーザーにとって「壊れている」エラーは致命的で、対処法を提示しても実行できない。

**根本的な解決策はApple Developer Program加入のみ**であり、それ以外の方法は全て妥協を伴う。

今後の方針は、**ユーザー層とツールの位置づけ**によって決定すべきである。

---

**作成者**: Claude Code + Hiroki Takamura
**更新履歴**:
- 2025-10-23: 初版作成
- 2025-10-23: 追加調査完了、Apple Developer Program加入による解決方法を詳細化
  - macOS Sequoiaでの右クリック回避機能削除を確認
  - 公証の速度（平均30秒）、notarytool使用を明記
  - Gatekeeperの動作原理を追加
  - 具体的な実装手順とコマンド例を追加
  - 現在の配布方法（CLI版のみ）を明記
