# KoeMoji-Go プログラム流れ設計ドキュメント

このドキュメントでは、KoeMoji-Goアプリケーションの全体的なプログラムフローと設計を詳しく説明します。

## 目次

1. [アプリケーション起動フロー](#1-アプリケーション起動フロー)
2. [メイン処理ループ](#2-メイン処理ループ)
3. [ファイル処理パイプライン](#3-ファイル処理パイプライン)
4. [録音機能フロー](#4-録音機能フロー)
5. [GUI vs TUI モード](#5-gui-vs-tui-モード)
6. [設定管理フロー](#6-設定管理フロー)
7. [エラーハンドリングフロー](#7-エラーハンドリングフロー)

---

## 1. アプリケーション起動フロー

### 1.1 エントリーポイント (`cmd/koemoji-go/main.go`)

```go
func main() {
    // フェーズ1: コマンドライン引数解析
    configPath, debugMode, showVersion, showHelp, configMode, guiMode := parseFlags()
    
    // フェーズ2: 基本コマンド処理（バージョン表示、ヘルプ）
    if showVersion { /* バージョン表示して終了 */ }
    if showHelp { /* ヘルプ表示して終了 */ }
    
    // フェーズ3: モード分岐
    if guiMode {
        gui.Run(configPath, debugMode)  // GUI モードに分岐
        return
    }
    
    // フェーズ4: TUI モード初期化
    app := &App{
        configPath:     configPath,
        debugMode:      debugMode,
        processedFiles: make(map[string]bool),
        startTime:      time.Now(),
        logBuffer:      make([]logger.LogEntry, 0, 12),
        queuedFiles:    make([]string, 0),
    }
    
    // フェーズ5: アプリケーション初期化
    app.initLogger()                    // ログシステム初期化
    cfg := config.LoadConfig(...)       // 設定ファイル読み込み
    app.Config = cfg
    
    // フェーズ6: 設定モードチェック
    if configMode {
        config.ConfigureSettings(...)   // 対話式設定モード
        return
    }
    
    // フェーズ7: ディレクトリと依存関係準備
    processor.EnsureDirectories(...)    // 必要ディレクトリ作成
    whisper.EnsureDependencies(...)     // FasterWhisper依存関係確認
    
    // フェーズ8: メイン処理開始
    app.run()
}
```

### 1.2 初期化プロセス詳細

**ログシステム初期化 (`initLogger`)**
```go
func (app *App) initLogger() {
    // ファイルベースログ設定
    logFile, err := os.OpenFile("koemoji.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    app.logger = log.New(io.MultiWriter(logFile), "", log.LstdFlags)
    
    // 循環バッファに初期ログエントリ追加
    logger.LogInfo(app.logger, &app.logBuffer, &app.logMutex, "KoeMoji-Go v%s started", version)
}
```

**依存関係確認 (`whisper.EnsureDependencies`)**
```go
func EnsureDependencies() {
    // FasterWhisper利用可能性チェック
    if !isFasterWhisperAvailable() {
        // 自動インストール試行
        installFasterWhisper(...)
    }
}
```

---

## 2. メイン処理ループ

### 2.1 TUI モードメインループ (`app.run()`)

```go
func (app *App) run() {
    // フェーズ1: コンテキスト設定（グレースフルシャットダウン用）
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // フェーズ2: シグナルハンドリング設定
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    
    // フェーズ3: バックグラウンド処理開始
    go processor.StartProcessing(ctx, ...)    // ファイル監視・処理ゴルーチン
    go app.handleUserInput(ctx)               // ユーザー入力処理ゴルーチン
    
    // フェーズ4: 初期ダッシュボード表示
    time.Sleep(100 * time.Millisecond)       // 初期化待機
    ui.RefreshDisplay(...)                   // ダッシュボード描画
    
    // フェーズ5: シャットダウン待機
    <-sigChan                                // シグナル受信まで待機
    
    // フェーズ6: グレースフルシャットダウン
    cancel()                                 // 全ゴルーチンに停止指示
    
    // タイムアウト付きでゴルーチン終了待機
    select {
    case <-done:
        // 正常終了
    case <-time.After(10 * time.Second):
        // タイムアウト強制終了
    }
}
```

### 2.2 ユーザー入力処理 (`handleUserInput`)

```go
func (app *App) handleUserInput(ctx context.Context) {
    reader := bufio.NewReader(os.Stdin)
    for {
        select {
        case <-ctx.Done():
            return  // コンテキストキャンセル時の終了
        default:
            // ブロッキング入力読み取り
        }
        
        input, err := reader.ReadString('\n')
        
        switch strings.TrimSpace(strings.ToLower(input)) {
        case "":      // Enter = 手動リフレッシュ
        case "c":     // 設定画面
        case "l":     // ログ全表示
        case "s":     // 手動スキャン
        case "i":     // 入力フォルダ開く
        case "o":     // 出力フォルダ開く
        case "r":     // 録音開始/停止
        case "q":     // 終了
        }
    }
}
```

---

## 3. ファイル処理パイプライン

### 3.1 ファイル監視・処理フロー (`processor.StartProcessing`)

```go
func StartProcessing(ctx context.Context, ...) {
    // フェーズ1: 初回スキャン実行
    ScanAndProcess(...)
    
    // フェーズ2: 定期スキャンループ
    ticker := time.NewTicker(time.Duration(config.ScanIntervalMinutes) * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return  // コンテキストキャンセル時終了
        case <-ticker.C:
            ScanAndProcess(...)  // 定期スキャン実行
        }
    }
}
```

### 3.2 スキャン・処理フロー (`ScanAndProcess`)

```go
func ScanAndProcess(...) {
    // フェーズ1: ディレクトリスキャン
    files, err := filepath.Glob(filepath.Join(config.InputDir, "*"))
    
    // フェーズ2: 新規オーディオファイルフィルタリング
    newFiles := filterNewAudioFiles(files, processedFiles, mu)
    if len(newFiles) == 0 {
        return  // 新規ファイルなし
    }
    
    // フェーズ3: ファイルをキューに追加
    mu.Lock()
    *queuedFiles = append(*queuedFiles, newFiles...)
    mu.Unlock()
    
    // フェーズ4: 処理開始（処理中でない場合のみ）
    if !*isProcessing {
        *isProcessing = true
        go processQueue(...)  // 非同期でキュー処理開始
    }
}
```

### 3.3 ファイル処理キュー (`processQueue`)

```go
func processQueue(...) {
    defer wg.Done()  // WaitGroup完了通知
    
    for {
        mu.Lock()
        if len(*queuedFiles) == 0 {
            *isProcessing = false
            *processingFile = ""
            mu.Unlock()
            return  // キューが空なら終了
        }
        
        // フェーズ1: キューから次のファイル取得
        filePath := (*queuedFiles)[0]
        *queuedFiles = (*queuedFiles)[1:]
        *processingFile = filepath.Base(filePath)
        mu.Unlock()
        
        // フェーズ2: 音声転写処理
        startTime := time.Now()
        if err := whisper.TranscribeAudio(..., filePath); err != nil {
            // エラー処理
        } else {
            duration := time.Since(startTime)
            
            // フェーズ3: LLM要約生成（有効な場合）
            if config.LLMSummaryEnabled {
                generateSummary(..., filePath)
            }
            
            // フェーズ4: アーカイブ移動
            moveToArchive(config, filePath)
        }
    }
}
```

---

## 4. 録音機能フロー

### 4.1 録音開始フロー (`startRecording`)

```go
func (app *App) startRecording() {
    // フェーズ1: レコーダー初期化（初回のみ）
    if app.recorder == nil {
        if app.Config.RecordingDeviceID == -1 {
            app.recorder, err = recorder.NewRecorder()  // デフォルトデバイス
        } else {
            app.recorder, err = recorder.NewRecorderWithDevice(...)  // 指定デバイス
        }
        
        // 録音制限設定
        app.recorder.SetLimits(maxDuration, maxFileSize)
    }
    
    // フェーズ2: 録音開始
    err := app.recorder.Start()
    app.isRecording = true
    app.recordingStartTime = time.Now()
    
    // フェーズ3: UI更新
    ui.RefreshDisplay(...)
}
```

### 4.2 録音中のデータフロー (`recorder.recordCallback`)

```go
func (r *Recorder) recordCallback(in []int16) {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    // フェーズ1: 録音制限チェック
    if r.exceedsLimits() {
        r.recording = false
        return
    }
    
    // フェーズ2: メモリバッファに追加
    r.samples = append(r.samples, in...)
    r.totalSamples += int64(len(in))
    
    // フェーズ3: バッファフラッシュ判定
    if len(r.samples) >= FlushThreshold || time.Since(r.lastFlush) > 2*time.Second {
        r.flushToTempFile()  // 一時ファイルにフラッシュ
    }
}
```

### 4.3 録音停止・保存フロー (`stopRecording`)

```go
func (app *App) stopRecording() {
    // フェーズ1: 録音停止
    err := app.recorder.Stop()
    
    // フェーズ2: ファイル名生成
    filename := fmt.Sprintf("recording_%s.wav", now.Format("20060102_1504"))
    outputPath := filepath.Join(app.Config.InputDir, filename)
    
    // フェーズ3: WAVファイル保存
    err = app.recorder.SaveToFile(outputPath)
    
    // フェーズ4: 状態更新
    app.isRecording = false
    duration := time.Since(app.recordingStartTime)
    
    // フェーズ5: UI更新
    ui.RefreshDisplay(...)
}
```

---

## 5. GUI vs TUI モード

### 5.1 モード分岐ポイント

```go
func main() {
    // コマンドライン解析後の分岐
    if guiMode {
        gui.Run(configPath, debugMode)  // GUI実行
        return
    }
    // TUI実行継続
    app := &App{...}
    app.run()
}
```

### 5.2 GUI初期化フロー (`gui.Run`)

```go
func Run(configPath string, debugMode bool) {
    // フェーズ1: コンテキスト作成
    ctx, cancel := context.WithCancel(context.Background())
    
    // フェーズ2: GUIアプリ構造体初期化
    guiApp := &GUIApp{
        configPath:     configPath,
        debugMode:      debugMode,
        processedFiles: make(map[string]bool),
        startTime:      time.Now(),
        logBuffer:      make([]logger.LogEntry, 0, 12),
        ctx:            ctx,
        cancelFunc:     cancel,
    }
    
    // フェーズ3: Fyneアプリケーション作成
    guiApp.fyneApp = app.NewWithID("com.hirokitakamura.koemoji-go")
    
    // フェーズ4: 設定読み込み
    guiApp.loadConfig()
    
    // フェーズ5: ウィンドウ作成・表示
    guiApp.createWindow()
    guiApp.window.ShowAndRun()  // ブロッキング実行
}
```

### 5.3 GUI vs TUI 共通処理

両モードで共通利用される機能：

- **設定管理**: `internal/config/`
- **ログシステム**: `internal/logger/`
- **ファイル処理**: `internal/processor/`
- **音声転写**: `internal/whisper/`
- **録音機能**: `internal/recorder/`
- **LLM統合**: `internal/llm/`

---

## 6. 設定管理フロー

### 6.1 設定読み込みフロー (`config.LoadConfig`)

```go
func LoadConfig(configPath string, logger *log.Logger) *Config {
    // フェーズ1: デフォルト設定作成
    config := GetDefaultConfig()
    
    // フェーズ2: 設定ファイル存在確認
    file, err := os.Open(configPath)
    if err != nil {
        if os.IsNotExist(err) {
            return config  // デフォルト設定を返す
        }
        os.Exit(1)  // 他のエラーは致命的
    }
    defer file.Close()
    
    // フェーズ3: JSON解析
    if err := json.NewDecoder(file).Decode(config); err != nil {
        os.Exit(1)  // パースエラーは致命的
    }
    
    return config
}
```

### 6.2 対話式設定フロー (`config.ConfigureSettings`)

```go
func ConfigureSettings(...) {
    reader := bufio.NewReader(os.Stdin)
    modified := false
    
    for {
        // フェーズ1: 設定メニュー表示
        displayConfigMenu(config)
        
        // フェーズ2: ユーザー選択処理
        input, _ := reader.ReadString('\n')
        choice := strings.TrimSpace(input)
        
        switch choice {
        case "1": // Whisperモデル設定
            if configureWhisperModel(...) { modified = true }
        case "2": // 言語設定
            if configureLanguage(...) { modified = true }
        // ... 他の設定項目
        case "s": // 保存して終了
            if modified {
                SaveConfig(config, configPath)
            }
            return
        case "q": // 保存せずに終了
            return
        }
    }
}
```

---

## 7. エラーハンドリングフロー

### 7.1 階層化エラーハンドリング

```go
// レベル1: 致命的エラー（アプリケーション終了）
func criticalError() {
    log.Printf("[ERROR] Critical error occurred")
    os.Exit(1)
}

// レベル2: 処理継続可能エラー（ログ記録）
func handleableError(err error) {
    logger.LogError(log, logBuffer, logMutex, "Processing failed: %v", err)
    // 処理は継続
}

// レベル3: 情報レベル（デバッグ用）
func debugInfo(info string) {
    if debugMode {
        logger.LogDebug(log, logBuffer, logMutex, debugMode, info)
    }
}
```

### 7.2 グレースフルシャットダウン

```go
func gracefulShutdown() {
    // フェーズ1: 全ゴルーチンに停止指示
    cancel()
    
    // フェーズ2: タイムアウト付き待機
    done := make(chan bool, 1)
    go func() {
        wg.Wait()  // 全ゴルーチン終了待機
        done <- true
    }()
    
    select {
    case <-done:
        logger.LogInfo(..., "Shutdown completed successfully")
    case <-time.After(10 * time.Second):
        logger.LogInfo(..., "Shutdown timeout, forcing exit")
    }
}
```

---

## 設計原則とベストプラクティス

### KISS原則の適用例

1. **単一責任原則**: 各パッケージが明確な責任を持つ
2. **状態管理の単純化**: レコーダー状態は `recorder.Recorder` のみで管理
3. **エラー処理の統一**: `internal/logger` による一貫したログ処理

### パフォーマンス最適化

1. **メモリ効率**: 録音データの段階的フラッシュ
2. **並行処理**: ファイル処理とUI更新の分離
3. **リソース管理**: 適切なクリーンアップ処理

### セキュリティ考慮事項

1. **パス検証**: 入力ディレクトリ外アクセスの防止
2. **入力検証**: 設定値の適切な検証
3. **リソース制限**: 録音時間とファイルサイズの制限

この設計により、保守性、拡張性、セキュリティを確保しながら、効率的なオーディオ転写処理を実現しています。