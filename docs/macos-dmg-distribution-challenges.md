# macOS DMG配布の技術的課題と解決策

**作成日**: 2025-10-23  
**状態**: 調査完了・実装保留

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

### 2. 右クリック→「開く」が機能しない

通常、**右クリック→「開く」でQuarantine警告を回避できる**はずだが、今回は**同じエラーが表示された**。

これは、**コード署名がない（Ad-hoc署名のみ）アプリに対するmacOS Sequoia以降の厳しい制限**と推測される。

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

Apple Developer Program（年間13,800円）に加入し、以下を実施：

1. **Developer ID証明書取得**
2. **コード署名**
3. **公証（Notarization）**

#### メリット
- **完全に解決**（ダブルクリックで起動可能）
- ユーザー体験が最高
- macOSの標準的な配布方法

#### デメリット
- **年間13,800円のコスト**
- 公証プロセスに時間がかかる（初回は数十分）
- Apple IDの2要素認証が必要

#### 評価
⭐⭐⭐⭐⭐ （唯一の根本的解決策）

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

### 短期的な対応（即座に実施）

**DMG配布を一旦保留**する。

1. **v1.7.0リリースを削除**（完了）
2. **本ドキュメントを作成**して経緯を記録
3. **CLI版（tar.gz）のみ**を推奨（v1.6.1を継続使用）

### 中期的な対応（検討中）

**Apple Developer Program加入を検討**する。

- **メリット**: 完全に解決、ユーザー体験向上
- **コスト**: 年間13,800円
- **判断基準**: ユーザー数、収益、開発継続意欲

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
- **公証**: Apple のサーバーでマルウェアスキャン

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
