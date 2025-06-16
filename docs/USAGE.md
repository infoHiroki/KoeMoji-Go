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

## トラブルシューティング

### 依存関係エラー
```bash
pip install faster-whisper whisper-ctranslate2
```

### 権限エラー
```bash
chmod +x koemoji-go
```

### ログ確認
`koemoji.log`でエラー詳細を確認