# KoeMoji-Go 設計書

## 1. プロジェクト概要

### 1.1 プロジェクト名
**KoeMoji-Go** - 音声・動画ファイル自動文字起こしツール

### 1.2 目的
Python版 KoeMojiAuto-cli の Go言語移植により、以下を実現する：
- **処理管理の効率化**: Go言語による安定したプロセス制御と監視機能
- **運用の簡素化**: シングルバイナリによる配布と起動
- **保守性の向上**: シンプルなアーキテクチャによる引き継ぎ容易性

### 1.3 対象ユーザー
- 不特定多数（一般ユーザー）
- Windows/Mac/Linux 各種OS対応

### 1.4 開発方針
- **AI駆動開発**: Claude等との協力による効率的な開発
- **個人開発**: 1人での開発・保守を前提とした設計
- **シンプル第一**: 複雑さを避け、理解しやすい構造を優先

## 2. 技術選定・方針

### 2.1 言語選択理由
**Go言語** を選択した理由：
- **シングルプロセス**: FasterWhisperの特性に合わせた順次処理
- **クロスコンパイル**: 1つの環境で全OS向けバイナリ作成
- **シンプルさ**: 学習コスト低、AI駆動開発に適している
- **外部プロセス管理**: exec.Commandによる安定したプロセス制御

### 2.2 基本方針

#### アーキテクチャ
- **シンプル構成**: main.go 1ファイルで完結
- **機能分離なし**: パッケージ分離による複雑化を避ける
- **約400行以下**: 1ファイルで全体把握可能な規模

#### 設定管理
- **config.json互換**: Python版からの移行容易性
- **下位互換性**: 既存設定ファイルをそのまま利用可能

#### エラーハンドリング
- **シンプル戦略**: FasterWhisperの安定性を前提とした最小限の処理
- **明確なエラーメッセージ**: ユーザーが対処可能な情報を提供

#### テスト戦略
- **単体テスト**: 基本機能（設定読み込み、ファイル操作等）
- **手動テスト**: FasterWhisper連携部分は実際の音声ファイルで確認

#### ログ方針
- **英語ログ**: 文字化け回避、国際対応
- **基本情報**: 処理開始/完了/エラーのみ
- **ファイル出力**: koemoji.log
- **ローテーション**: なし（年間数MB程度のため不要）

## 3. アーキテクチャ設計

### 3.1 全体構成

```
koemoji-go/
├── main.go                 # メイン実装（全機能）
├── config.json            # 設定ファイル（Python版互換）
├── go.mod                 # Go modules
├── go.sum                 # 依存関係
├── README.md              # 利用者向けドキュメント
├── input/                 # 処理対象ファイル置き場
├── output/                # 処理結果出力先
├── archive/               # 処理済みファイル保管
└── koemoji.log           # ログファイル
```

### 3.2 データフロー

```
[input/] → [File Watcher] → [Queue Manager] → [Audio Processor] → [FasterWhisper] → [output/] → [archive/]
                                                      ↓
                                                 [Logging]
```

### 3.3 主要コンポーネント

#### 3.3.1 ファイル監視 (File Watcher)
- input/ディレクトリを監視
- 新規ファイル検出時に処理キューに追加
- 対応フォーマット判定

#### 3.3.2 音声処理 (Audio Processor)
- FasterWhisper（whisper-ctranslate2）コマンドの実行
- キューに基づく順次処理
- ファイル移動（input → archive）

#### 3.3.3 設定管理 (Config Manager)
- config.json読み込み（Python版完全互換）
- デフォルト値設定
- 設定値検証

#### 3.3.4 ログ管理 (Logger)
- コンソール出力（リアルタイム表示）
- ファイル出力（koemoji.log）
- レベル別ログ（INFO/ERROR/DEBUG）
- タイムスタンプ付きフォーマット

#### 3.3.5 UI管理 (Enhanced Interactive UI)
- **リアルタイムステータス表示**: 処理状況、キュー、ファイル数の4行ヘッダー
- **リアルタイムログ表示**: 最新12行のログをリングバッファで管理
- **色付きログ表示**: ANSIエスケープシーケンスによる視認性向上
- **シンプル対話機能**: c=config, l=logs, s=scan, q=quit の4コマンド
- **一時表示機能**: コマンド出力後のEnter待ち復帰方式
- **環境対応**: 色非対応環境での自動フォールバック
- **自動実行**: 起動と同時に監視開始

## 4. 実装仕様

### 4.1 設定ファイル仕様

#### config.json（Python版互換 + UI拡張）
```json
{
  "whisper_model": "medium",
  "language": "ja",
  "scan_interval_minutes": 10,
  "max_cpu_percent": 95,
  "compute_type": "int8",
  "use_colors": true,
  "ui_mode": "enhanced"
}
```

#### Go実装での扱い
```go
type Config struct {
    WhisperModel        string `json:"whisper_model"`
    Language           string `json:"language"`
    ScanIntervalMinutes int    `json:"scan_interval_minutes"`
    MaxCpuPercent      int    `json:"max_cpu_percent"`
    ComputeType        string `json:"compute_type"` // 量子化タイプ（int8/float16等）
    UseColors          bool   `json:"use_colors"`   // 色付きログ表示
    UIMode             string `json:"ui_mode"`      // enhanced/simple
}
```

### 4.2 対応ファイル形式
- **音声**: MP3, WAV, M4A, FLAC, OGG, AAC
- **動画**: MP4, MOV, AVI

### 4.3 FasterWhisper連携

#### 依存関係管理戦略
**基本方針**: FasterWhisperを使用し、高速処理を実現

**パッケージ構成:**
- `faster-whisper`: Pythonライブラリ（CTranslate2ベースの高速実装）
- `whisper-ctranslate2`: コマンドラインツール（faster-whisperのCLI版）

**PATH問題への対応:**
pipでインストールされるコマンドは標準的な場所に配置されるが、ユーザー環境によってはPATHに含まれていない場合がある。

```go
func ensureDependencies() error {
    // FasterWhisperチェック
    if !isFasterWhisperAvailable() {
        logger.info("FasterWhisper not found. Attempting to install...")
        if err := installFasterWhisper(); err != nil {
            return fmt.Errorf("FasterWhisper installation failed: %w\nPlease install manually: pip install faster-whisper whisper-ctranslate2", err)
        }
    }
    
    return nil
}

func getWhisperCommand() string {
    // 1. 通常のPATHで試す
    if _, err := exec.LookPath("whisper-ctranslate2"); err == nil {
        return "whisper-ctranslate2"
    }
    
    // 2. 標準的なインストール場所を検索
    standardPaths := []string{
        filepath.Join(os.Getenv("HOME"), ".local", "bin", "whisper-ctranslate2"),  // Linux/macOS user install
        "/usr/local/bin/whisper-ctranslate2",                                      // Linux/macOS system
    }
    
    for _, path := range standardPaths {
        if _, err := os.Stat(path); err == nil {
            return path
        }
    }
    
    return "whisper-ctranslate2" // フォールバック
}

func isFasterWhisperAvailable() bool {
    cmd := exec.Command(getWhisperCommand(), "--help")
    cmd.Stdout = nil
    cmd.Stderr = nil
    return cmd.Run() == nil
}

func installFasterWhisper() error {
    logger.info("Installing faster-whisper and whisper-ctranslate2...")
    cmd := exec.Command("pip", "install", "faster-whisper", "whisper-ctranslate2")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("pip install failed: %w\nOutput: %s", err, string(output))
    }
    return nil
}
```

#### FasterWhisper実行
```bash
# 音声ファイルを直接処理（WAV変換不要）
whisper-ctranslate2 \
  --model medium \
  --language ja \
  --output_dir ./output \
  --output_format txt \
  --verbose False \
  ./input/audio.mp3
```

#### Go実装
```go
func transcribeAudio(inputFile string) error {
    outputDir := "./output"
    
    // getWhisperCommand()でPATH問題に対応
    whisperCmd := getWhisperCommand()
    
    cmd := exec.Command(whisperCmd,
        "--model", config.WhisperModel,
        "--language", config.Language,
        "--output_dir", outputDir,
        "--output_format", "txt",
        "--verbose", "False",
        inputFile,
    )
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("whisper execution failed: %w\nOutput: %s", err, string(output))
    }
    
    return nil
}
```

### 4.4 処理キュー管理

#### 順次処理の実装
```go
type App struct {
    config         *Config
    logger         *log.Logger
    debugMode      bool
    wg             sync.WaitGroup
    processedFiles map[string]bool
    mu             sync.Mutex
    
    // UI related fields
    startTime       time.Time
    lastScanTime    time.Time
    logBuffer       []LogEntry
    logMutex        sync.RWMutex
    totalProcessed  int
    inputCount      int
    outputCount     int
    archiveCount    int
    
    // Queue management for sequential processing
    queuedFiles     []string      // 処理待ちファイルキュー
    processingFile  string        // 現在処理中のファイル名（表示用）
    isProcessing    bool          // 処理中フラグ
}

func (app *App) processQueue() {
    for {
        app.mu.Lock()
        if len(app.queuedFiles) == 0 {
            app.isProcessing = false
            app.processingFile = ""
            app.mu.Unlock()
            return
        }
        
        // キューから次のファイルを取得
        filePath := app.queuedFiles[0]
        app.queuedFiles = app.queuedFiles[1:]
        app.processingFile = filepath.Base(filePath)
        app.isProcessing = true
        app.mu.Unlock()
        
        // FasterWhisper処理を実行
        if err := app.transcribeAudio(filePath); err != nil {
            app.logError("Failed to process %s: %v", app.processingFile, err)
        }
        app.totalProcessed++
    }
}
```

### 4.5 ログフォーマット

#### Enhanced UI表示レイアウト
```
=== KoeMoji-Go v1.0.0 ===                    ← ヘッダー
🟡 Processing | Queue: 2 | Processing: meeting.mp3   ← ステータス
📁 Input: 3 → Output: 127 → Archive: 89      ← ファイル数
⏰ Last: 12:34 | Next: 12:44 | Uptime: 1h24m ← 時間情報

INFO  12:34:15 Scanning for new files...     ← リアルタイム
PROC  12:34:16 Processing: meeting.mp3       ← ログ表示
DONE  12:36:45 Completed: intro.m4a (1m23s)  ← (最新12行)
ERROR 12:37:01 Failed: broken_file.mp4       ← 自動スクロール
... (12行まで表示) ...                       ← 古いログ削除

c=config l=logs s=scan q=quit                ← コマンド
> xyz
Invalid command 'xyz' (use c/l/s/q or Enter to refresh)
> [Enter]                                    ← 手動更新
[画面クリア + 最新状態で再描画]
> c                                          ← 入力例
--- Configuration ---                        ← 出力（下に表示）
Whisper model: medium
Language: ja
...
Press Enter to continue...                   ← Enter待ち
```

#### 対話機能の動作仕様
- **コマンド実行時**: リアルタイム更新を一時停止
- **出力表示**: プロンプト下に自然に表示
- **復帰方法**: Enter キーでリアルタイム表示に戻る
- **画面制御**: Enter後に画面クリア+再描画
```

#### 色定義（ANSIエスケープシーケンス）
- **INFO**: 青 (`\033[34m`) - 標準情報
- **PROC**: 黄 (`\033[33m`) - 処理中
- **DONE**: 緑 (`\033[32m`) - 処理完了
- **ERROR**: 赤 (`\033[31m`) - エラー
- **DEBUG**: グレー (`\033[37m`) - デバッグ情報

#### 従来ログ（Simple UI / 色非対応環境）
```
=== KoeMoji-Go v1.0.0 ===
Status: Active | Queue: 2 | Processing: meeting.mp3
Input: 3 -> Output: 127 -> Archive: 89
Last: 12:34 | Next: 12:44 | Uptime: 1h24m

[INFO ] 12:34:15 Scanning for new files...
[PROC ] 12:34:16 Processing: meeting.mp3
[DONE ] 12:36:45 Completed: intro.m4a (1m23s)
[ERROR] 12:37:01 Failed: broken_file.mp4

c=config l=logs s=scan q=quit
> _
```

### 4.6 動作モード

#### 通常実行（自動開始）
```bash
./koemoji-go                    # 起動と同時に監視開始、ログ表示
```

起動後の対話機能：
- `c` - 設定詳細表示（Enter で復帰）
- `l` - 全ログファイル表示（ページング、Enter で復帰）
- `s` - 即座にスキャン実行
- `q` - 即座に終了（os.Exit(0)）
- `Enter` - 画面手動更新（リフレッシュ）
- `Ctrl+C` - 即座に終了

#### エラーハンドリング
- **不明コマンド**: `Invalid command 'x' (use c/l/s/q or Enter to refresh)`
- **空文字入力**: Enter のみでリアルタイム表示を手動更新

#### 対話機能の動作フロー
1. コマンド入力 → リアルタイム更新停止
2. 出力をプロンプト下に表示
3. "Press Enter to continue..." 表示
4. Enter 入力待ち
5. 画面クリア → リアルタイム表示復帰

#### コマンドライン引数
```bash
./koemoji-go --config custom.json  # カスタム設定ファイル使用
./koemoji-go --debug               # デバッグモード（詳細ログ）
./koemoji-go --version             # バージョン表示して終了
./koemoji-go --help                # ヘルプ表示して終了
```

### 4.7 UI制御の実装詳細

#### リアルタイム表示制御
- **更新トリガー**: 各ログ出力時に `refreshDisplay()` 実行
- **一時停止**: コマンド実行時の更新停止
- **復帰処理**: Enter後の画面再構築

#### コマンドエラーハンドリング
```go
switch strings.TrimSpace(strings.ToLower(input)) {
case "":
    // 空Enter = 手動画面更新
    if app.config.UIMode == "enhanced" {
        app.refreshDisplay()
    }
case "c", "l":
    // 設定・ログ表示 + Enter待ち復帰
case "s", "q":
    // 即座実行
default:
    fmt.Printf("Invalid command '%s' (use c/l/s/q or Enter to refresh)\n", 
               strings.TrimSpace(input))
}
```

#### 画面更新戦略
- **イベント駆動**: ログ出力時に自動更新
- **手動更新**: Enter キーによる任意更新
- **自動更新なし**: タイマーによる定期更新は行わない

### 4.8 設定表示形式

```
--- Configuration (config.json) ---
Whisper model: medium
Language: ja
Scan interval: 30 minutes
Max CPU percent: 95%
Compute type: int8

Directories:
  Input: ./input/
  Output: ./output/
  Archive: ./archive/
---
```

### 4.9 処理フロー詳細

#### 起動から処理までの流れ
```
1. 起動
   ├─ 設定ファイル読み込み
   ├─ ディレクトリ確認・作成
   ├─ FasterWhisper依存関係チェック
   └─ UI初期化

2. メインループ開始
   ├─ ファイル監視開始（定期スキャン）
   ├─ ユーザー入力待機（ゴルーチン）
   └─ シグナル監視（Ctrl+C）

3. ファイル検出時
   ├─ 新規ファイルをキューに追加
   ├─ processQueue()を実行
   └─ キューが空になるまで順次処理

4. 各ファイル処理
   ├─ キューから取り出し
   ├─ FasterWhisper実行（ブロッキング）
   ├─ 結果をoutputに保存
   ├─ 元ファイルをarchiveに移動
   └─ 次のファイル処理へ

5. 終了処理
   ├─ 処理中のFasterWhisperを待機
   ├─ ログファイルクローズ
   └─ 終了
```

#### スキャンタイミング
- 起動時に即座にスキャン
- その後、設定された間隔で定期スキャン
- 手動スキャンコマンド（s）でも実行可能

## 5. テスト方針

### 5.1 単体テスト対象

#### 設定管理
```go
func TestLoadConfig(t *testing.T) {
    config, err := loadConfig("test_config.json")
    assert.NoError(t, err)
    assert.Equal(t, "medium", config.WhisperModel)
}

func TestDefaultConfig(t *testing.T) {
    config := getDefaultConfig()
    assert.Equal(t, "ja", config.Language)
}
```

#### ファイル操作
```go
func TestIsAudioFile(t *testing.T) {
    assert.True(t, isAudioFile("test.mp3"))
    assert.True(t, isAudioFile("test.wav"))
    assert.False(t, isAudioFile("test.txt"))
}

func TestMoveToArchive(t *testing.T) {
    // テストファイル作成→移動→確認
}
```

#### ユーティリティ関数
```go
func TestFormatDuration(t *testing.T) {
    assert.Equal(t, "2m30s", formatDuration(150*time.Second))
}
```

### 5.2 手動テストチェックリスト

#### リリース前確認項目
- [ ] 設定ファイル読み込み確認
- [ ] 各種音声ファイル形式での処理確認
- [ ] フォルダ監視機能確認
- [ ] 順次処理確認（キューによるファイル処理）
- [ ] エラー処理確認（存在しないファイル等）
- [ ] ログ出力確認（コンソール・ファイル）
- [ ] クロスプラットフォーム動作確認

#### テスト用データ
- 短時間音声ファイル（30秒程度）
- 長時間音声ファイル（10分程度）
- 各種フォーマット（mp3, wav, mp4等）
- 破損ファイル
- 非対応ファイル

## 6. 配布・運用

### 6.1 ビルド方法

#### 開発環境ビルド
```bash
go build -o koemoji-go main.go
```

#### リリースビルド（全プラットフォーム）
```bash
# Windows 64bit
GOOS=windows GOARCH=amd64 go build -o koemoji-go-windows-amd64.exe main.go

# macOS 64bit (Intel)
GOOS=darwin GOARCH=amd64 go build -o koemoji-go-darwin-amd64 main.go

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o koemoji-go-darwin-arm64 main.go

# Linux 64bit
GOOS=linux GOARCH=amd64 go build -o koemoji-go-linux-amd64 main.go
```

### 6.2 配布パッケージ構成
```
koemoji-go-v1.0.0/
├── koemoji-go(.exe)           # 実行ファイル
├── config.json               # デフォルト設定
└── README.md                 # 使い方説明
```

**注意**: `input/`、`output/`、`archive/`ディレクトリは初回実行時に自動作成されるため、配布パッケージに含める必要はありません。

```

### 6.3 システム要件

#### 動作要件
- **OS**: Windows 10/11, macOS 10.15+, Linux (主要ディストリビューション)
- **CPU**: Intel/AMD 64bit, Apple Silicon
- **メモリ**: 4GB以上推奨（8GB以上でより快適）
- **ストレージ**: 5GB以上（モデルファイル含む）

#### 依存関係
- **Python 3.8以上**: FasterWhisper実行に必要
- **pip**: Pythonパッケージマネージャー
- **FasterWhisper**: 自動インストール試行、失敗時は手動インストールガイド表示

### 6.4 インストール方法

#### ユーザー向け手順
1. GitHubリリースページから対応OS版をダウンロード
2. 任意のフォルダに展開
3. 実行（FasterWhisperが未インストールの場合、自動インストールを試行）
4. 設定ファイル編集（必要に応じて）

#### 自動インストール対応環境
- **全OS共通**: pip経由
  - `pip install faster-whisper whisper-ctranslate2`

#### 手動インストールが必要な場合
- パッケージマネージャーが未インストール
- 管理者権限がない環境
- 対応外のLinuxディストリビューション

上記の場合、詳細なインストールガイドが表示され、以下の手順を案内：
1. 各OS向けのパッケージマネージャーインストール方法
2. Pythonのインストール方法
3. pipでのFasterWhisperインストール手順

## 7. 開発ガイド

### 7.1 開発環境セットアップ

#### 必要ツール
- Go 1.21以上
- Python 3.8以上 + FasterWhisper（テスト用）
- Git

#### セットアップ手順
```bash
git clone https://github.com/[username]/koemoji-go.git
cd koemoji-go
go mod init koemoji-go
go mod tidy
```

### 7.2 AI駆動開発ガイド

#### Claudeとの協力方針
- **1ファイル設計の活用**: main.go全体をClaudeに共有可能
- **段階的実装**: 機能ごとに分けて実装・テスト
- **コード品質保持**: 定期的なリファクタリング依頼

#### 開発フロー
1. **仕様検討**: Claudeと要件整理
2. **実装**: 機能単位での実装
3. **テスト**: 単体テスト + 手動確認
4. **リファクタリング**: コード品質向上
5. **ドキュメント更新**: README等の更新

### 7.3 引き継ぎポイント

#### 重要なファイル
- **main.go**: 全ての実装が含まれる唯一のソースファイル
- **config.json**: 設定仕様（Python版互換）
- **go.mod**: 依存関係管理

#### 設計思想
- **シンプル第一**: 複雑な設計パターンは避ける
- **実用性重視**: 理論より実際の使いやすさを優先
- **AI協力前提**: Claude等のAIと協力して開発することを想定

#### よくある作業
- **新フォーマット対応**: isAudioFile関数の拡張
- **設定項目追加**: Config構造体の拡張
- **ログレベル調整**: ログ出力内容の変更
- **パフォーマンス調整**: 並行処理数の最適化

### 7.4 リリース手順

#### バージョニング
- セマンティックバージョニング（v1.0.0形式）
- 破壊的変更時はメジャーバージョンアップ

#### リリースチェックリスト
1. [ ] 全テスト実行
2. [ ] 手動テスト完了
3. [ ] ドキュメント更新
4. [ ] 全プラットフォームビルド
5. [ ] GitHubリリース作成
6. [ ] リリースノート作成

## 8. 補足情報

### 8.1 Python版からの主な変更点
- **言語**: Python → Go
- **配布**: Pythonインストール必要 → シングルバイナリ
- **ログ**: 日本語 → 英語
- **アーキテクチャ**: マルチファイル → シングルファイル

### 8.2 今後の拡張可能性
- **Web UI**: Goのnet/httpによるWeb インターフェース
- **API化**: REST API提供
- **リアルタイム処理**: ストリーミング音声対応
- **設定UI**: 設定ファイル編集のGUI

### 8.3 制限事項
- Python環境の依存（FasterWhisper実行に必要）
- 大容量ファイルの処理にはメモリ制限あり
- リアルタイム処理は非対応
- **処理中断不可**: whisperプロセスは開始後中断できない
- pipが利用できない環境での初期セットアップ困難
- whisperプロセスの強制終了（graceful shutdownは不可）

### 8.4 設計方針の決定事項
- **FasterWhisper採用**: 処理速度を優先し、Python依存を許容
- **自動インストール**: pip経由でFasterWhisperを自動インストール
- **ポータブル版**: Python依存のため実現困難
- **設定形式**: JSONを維持（シンプルさとPython版互換性を優先）
- **compute_type**: FasterWhisperで使用（CPU: int8、GPU: float16）
- **UI設計**: 自動実行＋最小限の対話機能（ログ表示メイン）
- **終了処理**: 強制終了（Whisperは処理中断不可の仕様）
- **音声処理**: FasterWhisperが内部で自動変換（ffmpeg不要）
- **エラー時動作**: 依存関係不足時は明確なエラーメッセージ
- **ログローテーション**: 不要（長期運用でも問題ないサイズ）
- **処理方式**: FasterWhisperの重い処理特性に合わせて順次処理を採用

---

## 更新履歴
- v1.0 (2024/06/07): 初版作成
- v1.1 (2025/06/07): whisper.cpp連携方針を明確化、自動インストール戦略を追加
- v1.2 (2025/06/07): UI設計を自動実行＋最小限対話型に変更、設定形式をJSON確定
- v1.3 (2025/06/07): ffmpeg依存関係を明確化、音声変換処理を追加
- v1.4 (2025/06/07): エラー時動作、compute_type、ログローテーション方針を明確化
- v2.0 (2025/06/07): whisper.cppからFasterWhisperに変更、高速処理を実現
- v2.1 (2025/06/08): CUI改善 - リアルタイムステータス表示と色付きログを追加、対話機能拡張
- v2.2 (2025/06/08): CUI仕様確定 - リアルタイム12行表示、4コマンド体系、処理制約明確化
- v2.3 (2025/06/08): UI操作性改善 - コマンド出力の自然な表示とEnter復帰方式を実装
- v2.4 (2025/06/08): コマンドエラーハンドリング追加、Enter手動更新機能実装
- v2.5 (2025/06/08): FasterWhisper PATH問題解決 - 標準インストール場所検索機能追加
- v2.6 (2025/06/08): 並行処理から順次処理へ変更 - FasterWhisperの処理特性に最適化
- v2.7 (2025/06/08): リアルタイムテキスト表示機能を削除 - 重複ログと画面更新問題を解決、シンプル化
- v2.8 (2025/06/08): Success表示を削除 - 不要な情報を削除してUI表示をシンプル化
