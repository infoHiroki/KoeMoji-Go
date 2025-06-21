# 使用方法

## クイックスタート

### 1. 実行
```bash
./koemoji-go
```
初回実行時は自動的にデフォルト設定で起動

### 2. 設定
実行後に`c`キーで設定変更

### 3. ファイル処理
1. `input/`に音声・動画ファイルを配置
2. 自動的に文字起こし開始
3. 結果は`output/`に保存
4. 処理済みファイルは`archive/`に移動

## 対応形式
**入力**: mp3, wav, m4a, flac, ogg, aac, mp4, mov, avi
**出力**: txt, vtt, srt, tsv, json

## コマンド

### 起動オプション
```bash
./koemoji-go -version     # バージョン表示
./koemoji-go -help        # ヘルプ表示
./koemoji-go -configure   # 設定モード
./koemoji-go -debug       # デバッグモード
```

### 実行中コマンド
- `c` - 設定変更
- `l` - ログ表示
- `r` - 録音開始/停止（v1.4.0新機能）
- `s` - 手動スキャン
- `i` - 入力ディレクトリを開く
- `o` - 出力ディレクトリを開く
- `q` - 終了

## 設定項目

### 基本設定
- **whisper_model**: tiny〜large-v3 (medium推奨)
- **language**: 認識言語 (ja, en, zh等)
- **ui_language**: UI言語 (en/ja)
- **output_format**: 出力形式

### 処理設定
- **scan_interval_minutes**: 監視間隔
- **max_cpu_percent**: CPU使用率制限
- **compute_type**: 計算精度 (int8推奨)

### UI設定
- **ui_mode**: enhanced/simple
- **use_colors**: 色表示の有無

### AI要約設定（v1.2.0新機能）
- **llm_summary_enabled**: AI要約機能の有効/無効
- **llm_api_provider**: APIプロバイダー（openai）
- **llm_api_key**: OpenAI APIキー
- **llm_model**: 使用モデル（gpt-4o/gpt-4-turbo/gpt-3.5-turbo）
- **llm_max_tokens**: 最大トークン数（要約の長さ）
- **summary_prompt_template**: 要約プロンプトテンプレート

### 録音設定（v1.4.0新機能）
- **recording_device_id**: 録音デバイスID（-1=既定デバイス）
- **recording_device_name**: 録音デバイス名（表示用）

## AI要約機能の使い方（v1.2.0新機能）

### 初期設定
1. OpenAI APIキーを取得（https://platform.openai.com/）
2. `c`キーで設定画面を開く
3. 項目14でAPIキーを設定
4. 項目15でモデルを選択（gpt-4o推奨）

### 使用方法
1. `c`キーで設定画面を開き、LLMタブでAI要約をオン
2. 音声ファイルを`input/`に配置
3. 文字起こし完了後、自動的に要約生成
4. `output/ファイル名_summary.txt`に保存

### プロンプトカスタマイズ
設定項目17で要約プロンプトを編集可能：
- `{text}`: 文字起こしテキスト
- `{language}`: 要約言語

## 録音機能の使い方（v1.4.0新機能）

### 初期設定
1. 必要に応じて録音デバイスを設定
   - `c`キーで設定画面を開く
   - 項目18で録音デバイスを選択
   - 仮想デバイス（BlackHole等）も利用可能

### 使用方法
1. `r`キーで録音開始
2. 録音中は画面上部に🔴録音中 - 時間表示
3. 再度`r`キーで録音停止
4. `recording_YYYYMMDD_HHMM.wav`として自動保存
5. 次回スキャンで自動的に文字起こし開始

### 対応デバイス
- **既定マイク**: -1設定で自動選択
- **特定デバイス**: 設定画面でID指定
- **仮想デバイス**: BlackHole（macOS）、Stereo Mix（Windows）
- **システム音声**: 仮想デバイス経由で録音可能

## トラブルシューティング

### 依存関係エラー
```bash
# FasterWhisper
pip install faster-whisper whisper-ctranslate2

# 録音機能（macOS）
brew install portaudio pkg-config

# 録音機能（Windows）
# PortAudio DLLが必要（通常は自動インストール）
```

### 権限エラー
```bash
chmod +x koemoji-go
```

### ログ確認
`koemoji.log`でエラー詳細を確認