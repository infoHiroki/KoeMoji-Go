# KoeMoji-Go トラブルシューティング・FAQ

このドキュメントでは、KoeMoji-Go使用時に発生する可能性のある問題と解決方法をまとめています。

## 環境・インストール関連

### Q: OneDriveフォルダでビルドエラーが発生する
```
A: OneDriveの同期によるファイルロックが原因です
- 解決策: OneDrive外のフォルダ（例: C:\Dev\KoeMoji-Go）にプロジェクトを移動
- xcopyコマンドでプロジェクトをコピー:
  xcopy /E /I /Y "OneDrive内のパス" "C:\Dev\KoeMoji-Go"
```

### Q: "Python not found" エラーが出る
```
A: Pythonがインストールされていません
- Python 3.8以上をインストールしてください
- インストール後、ターミナル/コマンドプロンプトを再起動
- Windowsの場合、インストール時に「Add Python to PATH」をチェック
```

### Q: Pythonはあるが古いバージョン（3.7以下）
```bash
# バージョン確認
python --version

# 新しいバージョンをインストール（推奨: 3.11以上）
# Windows: 公式サイトから最新版をダウンロード
# macOS: brew install python
```

### Q: FasterWhisperのインストールに失敗する
```bash
# 手動インストール
pip install faster-whisper whisper-ctranslate2

# pipが古い場合
pip install --upgrade pip
pip install faster-whisper whisper-ctranslate2

# 権限エラーの場合
pip install --user faster-whisper whisper-ctranslate2

# 仮想環境での実行
python -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate
pip install faster-whisper whisper-ctranslate2
```

### Q: "whisper-ctranslate2 not found" エラー

#### 基本的な確認手順
```bash
# パッケージ確認
pip show whisper-ctranslate2
pip list | grep whisper

# パス確認
which whisper-ctranslate2        # macOS/Linux
where whisper-ctranslate2        # Windows

# 再インストール（推奨）
pip uninstall whisper-ctranslate2 faster-whisper
pip install whisper-ctranslate2
```

#### Windows環境での詳細なトラブルシューティング

**1. whisper-ctranslate2がインストールされているが見つからない場合**

```bash
# コマンドプロンプトで実行場所を確認
where whisper-ctranslate2

# 見つからない場合、Pythonのインストール場所を確認
where python
python -c "import sys; print(sys.executable)"

# Pythonのスクリプトフォルダを確認
python -c "import site; print(site.USER_BASE)"
```

**2. PATHが通っていない場合の対処**

```bash
# Pythonスクリプトのパスを確認
python -m site --user-site

# 通常、以下のようなパスにインストールされます：
# C:\Users\[ユーザー名]\AppData\Local\Programs\Python\Python312\Scripts\
# C:\Users\[ユーザー名]\AppData\Roaming\Python\Python312\Scripts\
```

PATHに追加する方法：
1. Windowsキー + 「環境変数」で検索
2. 「環境変数を編集」を選択
3. ユーザー環境変数の「Path」を編集
4. 上記のScriptsフォルダのパスを追加
5. コマンドプロンプトを再起動

**3. 複数のPython環境がある場合**

```bash
# すべてのPythonを確認
where python

# 各Pythonでwhisper-ctranslate2を確認
C:\Python311\python.exe -m pip show whisper-ctranslate2
C:\Python312\python.exe -m pip show whisper-ctranslate2

# 正しいPythonにインストール
C:\Python312\python.exe -m pip install whisper-ctranslate2
```

**4. KoeMoji-Goが検索する標準的なパス**

KoeMoji-Goは以下のパスを自動的に検索します：
- `%LOCALAPPDATA%\Programs\Python\Python3XX\Scripts\`
- `%APPDATA%\Python\Python3XX\Scripts\`
- `%APPDATA%\Roaming\Python\Python3XX\Scripts\`
- `C:\Python3XX\Scripts\`
- Anaconda/Miniconda環境のScriptsフォルダ

**5. それでも解決しない場合**

```bash
# 手動でフルパスを確認して実行
dir C:\Users\%USERNAME%\AppData\Local\Programs\Python\Python*\Scripts\whisper-ctranslate2.exe /s

# 権限の問題を確認（管理者権限で実行）
# アンチウイルスソフトがブロックしていないか確認
```

## 実行・操作関連

### Q: アプリケーションが起動しない
```
A: 以下を順番に確認：
1. 実行権限: chmod +x koemoji-go
2. Python環境: python --version
3. ログ確認: cat koemoji.log
4. 設定ファイル: config.jsonが正しい形式か確認
```

### Q: 音声ファイルが処理されない
```
A: 以下を確認：
- 対応形式: MP3, WAV, M4A, FLAC, OGG, AAC, MP4, MOV, AVI
- ファイル名: 日本語・特殊文字を避ける
- ファイルサイズ: 極端に大きなファイル（>2GB）は処理できない場合あり
- ファイル破損: 他のプレーヤーで再生可能か確認
- 権限: ファイルの読み取り権限があるか確認
```

### Q: ディレクトリが開かない（i/oコマンド）
```
A: プラットフォーム別の対応：
- Windows: explorerが利用可能か確認
- macOS: Finderが利用可能か確認
```

### Q: GUI版で入力/出力ボタンが反応しない（v1.5.1で修正済み）
```
A: v1.5.1以前のバージョンで発生する問題：
- 原因: WindowsでHideWindowフラグがexplorer.exeと互換性がない
- 解決策: v1.5.1以降にアップデート
- 一時的な回避策: TUIモード（--tui）を使用
```

### Q: 設定変更（cキー）が反映されない
```
A: 以下を確認：
1. 設定保存の確認: config.jsonファイルが更新されているか
2. アプリ再起動: 一度終了（q）して再実行
3. 設定ファイル権限: config.jsonに書き込み権限があるか
4. JSON形式: 設定ファイルが正しいJSON形式か確認
```

## パフォーマンス・品質関連

### Q: 処理が非常に遅い
```
A: 以下の最適化を試す：
1. モデル変更: large-v3 → medium → small → tiny
2. compute_type確認: "int8"が最速（デフォルト）
3. CPU使用率: タスクマネージャーでCPU使用状況確認
4. メモリ不足: 8GB以上のRAM推奨
5. 他のプロセス: 重いアプリケーションを終了
```

### Q: 文字起こし結果が不正確
```
A: 品質向上のために：
1. モデル変更: tiny/small → medium → large-v3
2. 音声品質確認: ノイズ除去、音量調整
3. 言語設定: config.jsonのlanguageが"ja"になっているか
4. 音声クリアさ: 発話がはっきりしているか
5. 背景音: 音楽・雑音が少ない環境で録音
```

### Q: 日本語が正しく認識されない
```
A: 以下を確認・設定：
1. language設定: "ja"に設定
2. モデル: large-v3を推奨（日本語最適化）
3. 音声品質: クリアな日本語発話
4. 方言: 標準的な発話の方が認識精度が高い
```

## エラー・異常終了関連

### Q: 突然終了する・クラッシュする
```
A: 原因調査方法：
1. ログ確認: tail -f koemoji.log
2. メモリ不足: システムモニターでメモリ使用量確認
3. ディスク容量: 出力先の空き容量確認
4. 権限エラー: input/output/archiveディレクトリの権限確認
5. 破損ファイル: 問題のある音声ファイルを特定・除外
```

### Q: "Config file not found" が表示される
```
A: 正常な動作です
- config.jsonがない場合、デフォルト設定で起動
- 設定を変更したい場合は、実行中にcキーで設定可能
- または: cp config.example.json config.json
```

### Q: ログファイルが肥大化する
```bash
# ログファイルのクリア
> koemoji.log

# またはファイル削除
rm koemoji.log

# 次回実行時に新しいログファイルが作成されます
```

## 高度な設定・カスタマイズ

### Q: 複数の設定を使い分けたい
```bash
# 設定ファイルを分ける
cp config.json config-fast.json    # 高速処理用
cp config.json config-quality.json # 高品質処理用

# 使い分け
./koemoji-go -config config-fast.json
./koemoji-go -config config-quality.json
```

### Q: 監視間隔を変更したい
```json
{
  "scan_interval_minutes": 5  // 5分間隔（デフォルト:1分）
}
```

### Q: 出力フォーマットを変更したい
```json
{
  "output_format": "srt"  // txt, vtt, srt, tsv, json
}
```

### Q: 入力・出力ディレクトリを変更したい
```json
{
  "input_dir": "/Users/username/Desktop/音声",
  "output_dir": "/Users/username/Desktop/文字起こし",
  "archive_dir": "/Users/username/Desktop/処理済み"
}
```

## 録音機能関連（v1.4.0+）

### Q: 録音ボタン（r）が反応しない
```
A: 以下を確認：
1. PortAudio依存関係: 必要なライブラリがインストールされているか
2. マイク権限: macOSの場合、マイクアクセス許可が必要
3. 録音デバイス: 設定画面（c → 項目18）でデバイスが正しく選択されているか
4. ログ確認: エラーメッセージがkoemoji.logに記録されているか
```

### Q: 録音した音声が認識されない
```
A: 録音品質を確認：
1. 音量レベル: 録音時の音声レベルが適切か
2. ノイズ: 背景音や雑音が多すぎないか
3. デバイス設定: 正しいマイクデバイスが選択されているか
4. ファイル確認: 録音されたWAVファイルが正常に再生できるか
```

## AI要約機能関連（v1.2.0+）

### Q: AI要約が生成されない
```
A: 以下を確認：
1. APIキー設定: OpenAI APIキーが正しく設定されているか
2. 要約機能: `a`キーでAI要約がオンになっているか
3. API制限: OpenAIアカウントの利用制限に達していないか
4. ネットワーク: インターネット接続が正常か
```

### Q: AI要約の品質が良くない
```
A: 設定を調整：
1. モデル変更: gpt-3.5-turbo → gpt-4o
2. トークン数: llm_max_tokensを増やす（4096 → 8192）
3. 元文章品質: 文字起こし結果の精度を向上させる
```

## ログの確認

問題が発生した場合は`koemoji.log`を確認してください：
```bash
# ログファイルの確認
cat koemoji.log

# 最新のログのみ確認
tail -f koemoji.log
```

## サポート

上記で解決しない場合は、以下の情報と共にお問い合わせください：
- OS及びバージョン
- Python及びpipのバージョン
- エラーメッセージの全文
- koemoji.logの内容
- 再現手順

連絡先: koemoji2024@gmail.com