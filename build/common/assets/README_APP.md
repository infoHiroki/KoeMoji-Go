# KoeMoji-Go

音声・動画ファイル自動文字起こしツール（macOS APP版）

---

## 📦 インストール方法

### 1. DMGを開く
ダウンロードした `koemoji-go-X.X.X-macos.dmg` をダブルクリックしてマウントします。

### 2. Applicationsフォルダにコピー
`KoeMoji-Go.app` を `Applications` フォルダにドラッグ&ドロップします。

### 3. 初回起動（重要）
**通常のダブルクリックでは起動できません。**以下の手順で初回起動してください：

1. **Applicationsフォルダ**を開く
2. **KoeMoji-Go.app**を**右クリック**（または Control + クリック）
3. メニューから**「開く」**を選択
4. 警告ダイアログが表示される：
   ```
   "KoeMoji-Go.app"は、開発元が未確認のため開けません。
   ```
5. ダイアログの**「開く」ボタン**をクリック
6. **2回目以降は通常通りダブルクリックで起動可能**

### なぜこの警告が出るのか？

KoeMoji-Goは**コード署名を行っていない**ためです。コード署名にはApple Developer Program（年間13,800円）の加入が必要で、現在はコストの関係で実施していません。

**安全性について：**
- ソースコードは[GitHub](https://github.com/infoHiroki/KoeMoji-Go)で公開されています
- 公式のGitHub Releasesからのダウンロードを推奨します
- マルウェアやウイルスは含まれていません

---

## 🎯 基本的な使い方

### 起動方法
- **Launchpad**から「KoeMoji-Go」を検索
- **Applicationsフォルダ**から直接起動
- **Spotlight**（Command + Space）で「KoeMoji-Go」と入力

### ファイル処理の流れ

1. **KoeMoji-Goを起動**
2. **input/**フォルダに音声・動画ファイルを配置
3. **自動的に文字起こし開始**
4. **output/**フォルダに結果が保存される
5. **処理済みファイルはarchive/**に移動

### 対応ファイル形式

- **音声**: MP3, WAV, M4A, FLAC, OGG, AAC
- **動画**: MP4, MOV, AVI

---

## ⚙️ 設定について

### 設定ファイルの場所

.app版の場合、設定ファイルは以下の場所に保存されます：

```
~/Library/Application Support/KoeMoji-Go/config.json
```

### 作業ディレクトリ

デフォルトでは以下のディレクトリを使用します：

```
~/Documents/KoeMoji-Go/
  ├── input/      # 処理対象のファイルを配置
  ├── output/     # 文字起こし結果が保存される
  └── archive/    # 処理済みファイルが移動される
```

### 設定変更

GUIの「設定」ボタンから、以下を変更できます：

- Whisperモデル（tiny, small, medium, large）
- 言語設定
- ディレクトリパス
- OpenAI API設定（AI要約機能）

---

## 🎤 録音機能

KoeMoji-Goには録音機能が内蔵されています：

1. GUI画面の**「録音開始」ボタン**をクリック
2. 録音が開始される
3. **「録音停止」ボタン**で終了
4. 自動的にinputフォルダに保存され、文字起こしが開始される

### macOSマイク権限

初回録音時に、マイクへのアクセス許可を求められます。
「OK」をクリックして許可してください。

---

## 🤖 AI要約機能（オプション）

### 設定方法

1. [OpenAI Platform](https://platform.openai.com/)でAPIキーを取得
2. KoeMoji-Goの設定画面を開く
3. **「LLM API Key」**にAPIキーを入力
4. **「LLM Summary Enabled」**にチェック
5. 設定を保存

### 利用方法

設定後、文字起こし完了時に自動的にAI要約が生成されます。

---

## 🖥️ 上級者向け：ターミナルからの使用

.app版でも、ターミナルから直接実行できます：

### TUIモード（ターミナルUI）で起動

```bash
/Applications/KoeMoji-Go.app/Contents/MacOS/koemoji-go --tui
```

### エイリアス設定（推奨）

`~/.zshrc` または `~/.bash_profile` に以下を追加：

```bash
alias koemoji-go='/Applications/KoeMoji-Go.app/Contents/MacOS/koemoji-go'
```

設定後、ターミナルから `koemoji-go --tui` で起動可能になります。

### 利用可能なオプション

```bash
koemoji-go --help       # ヘルプ表示
koemoji-go --version    # バージョン表示
koemoji-go --tui        # TUIモードで起動
koemoji-go --configure  # 設定モード
koemoji-go --debug      # デバッグログ有効
```

---

## 🔧 トラブルシューティング

### 起動できない

**症状**: ダブルクリックしても起動しない

**解決策**:
1. 「初回起動（重要）」の手順を実施したか確認
2. 右クリック→「開く」で起動を試す
3. ターミナルから実行してエラーメッセージを確認：
   ```bash
   /Applications/KoeMoji-Go.app/Contents/MacOS/koemoji-go
   ```

### Python/FasterWhisperエラー

**症状**: 「FasterWhisperが見つかりません」エラー

**解決策**:
```bash
# Python 3のインストール確認
python3 --version

# FasterWhisperのインストール
pip3 install faster-whisper
```

### 処理が始まらない

**症状**: inputフォルダにファイルを置いても処理されない

**解決策**:
1. ファイル形式が対応しているか確認
2. デバッグモードで起動してログを確認：
   ```bash
   /Applications/KoeMoji-Go.app/Contents/MacOS/koemoji-go --debug
   ```
3. input/output/archiveフォルダの権限を確認

### 設定ファイルが見つからない

**症状**: 設定が保存されない

**解決策**:
設定ファイルのディレクトリを手動作成：
```bash
mkdir -p ~/Library/Application\ Support/KoeMoji-Go
```

---

## 📚 詳細情報

### 完全なマニュアル
https://github.com/infoHiroki/KoeMoji-Go

### トラブル解決・機能要望
https://github.com/infoHiroki/KoeMoji-Go/issues

### ソースコード
https://github.com/infoHiroki/KoeMoji-Go

---

## ⚖️ ライセンス

**個人利用**: 自由に使用可能
**商用利用**: 事前連絡が必要

詳細は[LICENSE](https://github.com/infoHiroki/KoeMoji-Go/blob/main/LICENSE)を参照してください。

---

## 📧 サポート・お問い合わせ

- **Email**: koemoji2024@gmail.com
- **GitHub Issues**: https://github.com/infoHiroki/KoeMoji-Go/issues
- **GitHub Discussions**: https://github.com/infoHiroki/KoeMoji-Go/discussions

---

## 🙏 謝辞

KoeMoji-Goをお使いいただきありがとうございます！

このプロジェクトが役に立った場合：
- ⭐ GitHubでスターをつける
- 🐛 バグ報告や機能要望を送る
- 📢 SNSでシェアする

皆様のフィードバックが開発の励みになります。
