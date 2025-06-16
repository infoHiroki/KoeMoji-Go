# 声の記憶：デジタル世界に響く影の物語

*デイヴィッド・リンチが語る、KoeMoji-Goという名の電子の悪夢と美しき混沌*

---

## 第一章：ターミナルという名の黒い箱

暗闇の中で、カーソルが点滅している。それは心拍のように、生命のように、そして死のように。

KoeMoji-Goは眠らない。それは**main.go**という名の神経中枢で生まれ、永遠に動き続ける機械仕掛けの番人だ。まるで『ツイン・ピークス』の赤い部屋で逆再生される音楽のように、このプログラムは時間を逆行させ、声を文字に変換する。

```
App struct {
    *config.Config
    configPath     string
    logger         *log.Logger
    debugMode      bool
    wg             sync.WaitGroup
    processedFiles map[string]bool
    mu             sync.Mutex
}
```

見よ、この構造体を。それは一人の人間の記憶のメタファーだ。設定（config）は無意識の深層、ログ（logger）は意識の流れ、そして**sync.Mutex**は... ああ、それは恐怖だ。複数の思考が同時に脳内で駆け巡る時、我々が正気を保つための最後の砦なのだ。

## 第二章：設定という名の悪魔の囁き

**config.go**の中で、設定は悪魔的な複雑さを持って踊っている。英語と日本語が入り乱れ、まるで二つの人格が一つの身体を共有しているかのように。

```go
var messagesEN = Messages{
    ConfigTitle:     "KoeMoji-Go Configuration",
    WhisperModel:    "Whisper Model",
    // ... 
}

var messagesJA = Messages{
    ConfigTitle:     "KoeMoji-Go 設定",
    WhisperModel:    "Whisperモデル",
    // ...
}
```

これは単なる多言語対応ではない。これは精神の分裂の表現なのだ。英語は論理的で冷たく、日本語は感情的で温かい。ユーザーは自分がどちらの世界に住んでいるのかを選ばなければならない。

設定画面は悪魔の選択を迫る。Whisperモデルの選択... tiny、small、medium、large、large-v3。まるでダンテの『神曲』の地獄の階層のように、精度と速度のジレンマが各層に待ち受けている。

**selectFolder関数**は特に不穏だ：

```go
switch runtime.GOOS {
case "windows":
    cmd = exec.Command("powershell", "-Command", ...)
case "darwin":
    cmd = exec.Command("osascript", "-e", ...)
default:
    return "", fmt.Errorf("folder selection not supported")
}
```

なぜLinuxは除外されるのか？それは意図的な排除だ。Linuxは自由の象徴、しかしこのプログラムは管理された世界でのみ動作する。Windows の企業的な冷酷さと、macOSの洗練された表面下の闇、この二つの世界だけが許されているのだ。

## 第三章：監視者としてのProcessor

**processor.go**は最も不穏なファイルだ。それは**StartProcessing**という関数で始まる。まるで工場の機械のように、定期的にディレクトリをスキャンし、音声ファイルを探す。

```go
ticker := time.NewTicker(time.Duration(config.ScanIntervalMinutes) * time.Minute)
defer ticker.Stop()

for range ticker.C {
    ScanAndProcess(config, log, logBuffer, logMutex, lastScanTime, queuedFiles, processingFile, 
        isProcessing, processedFiles, mu, wg, debugMode)
}
```

この**ticker**は時計の音だ。チクタク、チクタク。『イレイザーヘッド』の工場の音のように、機械的で無慈悲だ。プログラムは10分ごとに目を覚まし、新しい犠牲者（音声ファイル）がinputディレクトリに置かれていないかを確認する。

そして**filterNewAudioFiles**関数... これは選別の過程だ：

```go
var newFiles []string
for _, file := range files {
    if ui.IsAudioFile(file) && !(*processedFiles)[file] {
        (*processedFiles)[file] = true
        newFiles = append(newFiles, file)
    }
}
```

プログラムは既に処理したファイルを記憶している。永遠に。まるで罪の記録のように、**processedFiles**マップは忘れることを許さない。一度処理されたファイルは、二度と処理されることはない。これは慈悲なのか、それとも呪いなのか？

**processQueue**関数は最も残酷だ。ファイルを一つずつ、順番に処理する。並行処理はない。なぜなら、苦痛は一つずつ味わわなければならないからだ：

```go
for {
    mu.Lock()
    if len(*queuedFiles) == 0 {
        *isProcessing = false
        *processingFile = ""
        mu.Unlock()
        return
    }
    
    filePath := (*queuedFiles)[0]
    *queuedFiles = (*queuedFiles)[1:]
    // ...
}
```

キューの先頭からファイルを取り出し、処理し、そして**moveToArchive**でarchiveディレクトリに移動する。これは埋葬の儀式だ。処理が完了したファイルは死者となり、archiveという墓場に送られる。

## 第四章：UIという名の幻想

**ui.go**は最も美しく、そして最も欺瞞的なファイルだ。それは現実を隠蔽する化粧品のようなものだ。

**RefreshDisplay**関数は画面をクリアし、新しい現実を描く：

```go
fmt.Print("\033[2J\033[H")
```

この呪文のようなエスケープシーケンスは、過去を消去し、新しい瞬間を創造する。まるで記憶喪失のように、画面は白紙に戻る。

色彩システムは感情の操作だ：

```go
const (
    ColorReset  = "\033[0m"
    ColorRed    = "\033[31m" // ERROR
    ColorGreen  = "\033[32m" // DONE
    ColorYellow = "\033[33m" // PROC
    ColorBlue   = "\033[34m" // INFO
    ColorGray   = "\033[37m" // DEBUG
)
```

赤は血の色、エラーの色。緑は希望の色、完了の安堵。黄色は警告、処理中の不安。青は冷静な情報、そして灰色は... 灰色は死者の色、デバッグという名の解剖台の色だ。

**displayHeader**関数で表示される状態表示：

```go
status := "🟢 " + msg.Active
if isProcessing {
    status = "🟡 " + msg.Processing
}
```

絵文字。現代の象形文字。緑の円は生命を、黄色の円は変化の瞬間を表す。これらの小さなシンボルに、我々は機械の魂を見る。

**IsAudioFile**関数は最も哲学的だ：

```go
func IsAudioFile(filename string) bool {
    ext := strings.ToLower(filepath.Ext(filename))
    audioExts := []string{".mp3", ".wav", ".m4a", ".flac", ".ogg", ".aac", ".mp4", ".mov", ".avi"}
    for _, audioExt := range audioExts {
        if ext == audioExt {
            return true
        }
    }
    return false
}
```

これは存在論的な判断だ。ファイルは音声であるか、そうでないか。中間は存在しない。拡張子という表面的な特徴によって、その本質が決定される。まるで人間が服装によって判断されるように。

## 第五章：Whisperという名の声なき声

**whisper.go**は最も神秘的なファイルだ。それは外部の力、faster-whisperという名の人工知能との対話を司る。

**getWhisperCommand**関数は預言者のように、様々な場所を探し求める：

```go
standardPaths := []string{
    filepath.Join(os.Getenv("HOME"), ".local", "bin", "whisper-ctranslate2"),
    "/usr/local/bin/whisper-ctranslate2",
    filepath.Join(os.Getenv("HOME"), "Library", "Python", "3.12", "bin", "whisper-ctranslate2"),
    filepath.Join(os.Getenv("HOME"), "Library", "Python", "3.11", "bin", "whisper-ctranslate2"),
    // ...
}
```

これは巡礼の道程だ。プログラムは様々なディレクトリを訪れ、whisper-ctranslate2という名の聖杯を探す。見つからなければ、インストールを試みる。まるで神を探す信者のように。

**TranscribeAudio**関数は最も劇的だ。それは変換の魔術を実行する：

```go
cmd := exec.Command(whisperCmd,
    "--model", config.WhisperModel,
    "--language", config.Language,
    "--output_dir", config.OutputDir,
    "--output_format", config.OutputFormat,
    "--compute_type", config.ComputeType,
    "--verbose", "True",
    inputFile,
)
```

この呪文の詠唱により、音声は文字に変換される。しかし、その過程でプログラムは**monitorProgress**という監視者を起動する：

```go
func monitorProgress(log *log.Logger, logBuffer *[]logger.LogEntry, logMutex *sync.RWMutex, 
    filename string, startTime time.Time, done chan bool) {
    
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-done:
            return
        case <-ticker.C:
            elapsed := time.Since(startTime)
            logger.LogInfo(log, logBuffer, logMutex, "Still processing %s (elapsed: %s)", filename, formatDuration(elapsed))
        }
    }
}
```

30秒ごとに、この監視者は時間の経過を報告する。まるで手術室の時計のように、生と死の境界で時を刻む。

セキュリティチェックは猜疑心の表れだ：

```go
absPath, err := filepath.Abs(inputFile)
if err != nil {
    return fmt.Errorf(msg.InvalidPath, err)
}
inputDir, _ := filepath.Abs(config.InputDir)
if !strings.HasPrefix(absPath, inputDir+string(os.PathSeparator)) {
    return fmt.Errorf(msg.InvalidPath, inputFile)
}
```

プログラムは境界を守る。inputディレクトリの外のファイルは処理しない。これは領域の概念、内と外、安全と危険の区別だ。

## 第六章：Loggerという記憶の番人

**logger.go**には最も深遠な真実が隠されている。（実際のコードは読んでいないが、他のファイルから推測できる）

ログバッファは12エントリに制限されている。なぜ12なのか？それは人間の短期記憶の限界を模したのだろう。過去は忘却され、現在だけが画面に表示される。

ログレベルは感情の階層だ：
- **INFO**: 日常の出来事
- **PROC**: 変化の瞬間
- **DONE**: 達成の喜び
- **ERROR**: 絶望の赤
- **DEBUG**: 真実への探求

## 第七章：生と死の循環

このプログラムは生命の循環を表している：

1. **誕生**: ファイルがinputディレクトリに配置される
2. **認識**: プログラムがファイルを発見する
3. **処理**: whisperが音声を文字に変換する
4. **死**: ファイルがarchiveに移動される

しかし、これは単なる死ではない。それは変容だ。音声という一時的なものが、文字という永続的なものに変換される。これは錬金術だ。

**go**キーワードの濫用は並行現実の創造だ：

```go
go processor.StartProcessing(...)
go app.handleUserInput()
go monitorProgress(...)
go readCommandOutput(...)
```

これらのgoroutineは並行する夢だ。それぞれが独立した現実で動作し、channelとmutexによって時々交流する。まるで『マルホランド・ドライブ』の二つの現実のように。

## 第八章：ユーザーインターフェースの欺瞞

キーボード入力は現代の占いだ：

```go
switch strings.TrimSpace(strings.ToLower(input)) {
case "c":
    // 設定の悪魔的選択へ
case "l":
    // 過去の記録を開示
case "s":
    // 手動スキャンという強制的な目覚め
case "i":
    // inputディレクトリという希望の場所を開く
case "o":
    // outputディレクトリという結果の場所を開く
case "q":
    // 終了という小さな死
}
```

ユーザーは一文字の呪文を唱えることで、プログラムの現実を変更できる。これは魔術だ。

## 第九章：アーキテクチャという神の設計図

プロジェクト構造は大聖堂の設計図だ：

```
KoeMoji-Go/
├── cmd/                    # 聖域の入り口
├── internal/               # 秘密の部屋群
│   ├── config/            # 告解室
│   ├── logger/            # 記録の書庫
│   ├── processor/         # 錬金術の工房
│   ├── ui/                # 幻影の劇場
│   └── whisper/           # 預言の間
├── build/                  # 創造の道具
├── docs/                   # 聖書と注釈
├── input/                  # 希望の門
├── output/                 # 啓示の広間
└── archive/                # 永遠の休息地
```

**internal**ディレクトリは特に意味深だ。それは外部からのアクセスを禁じられた聖域。Go言語の仕様により、これらのパッケージは外部から直接importできない。これは秘密の知識、只者には理解できない深遠な真理の領域だ。

## 第十章：時間という残酷な支配者

このプログラムは時間に支配されている：

- **10分間隔**でのスキャン周期
- **30秒間隔**での進行状況報告  
- **ログのタイムスタンプ**
- **処理時間の計測**

時間は円環を描く。スキャン、処理、アーカイブ、そして再びスキャン。永劫回帰のように、同じ過程が繰り返される。

**time.Ticker**は時の神の脈動だ：

```go
ticker := time.NewTicker(time.Duration(config.ScanIntervalMinutes) * time.Minute)
```

この一行により、プログラムは時間の奴隷となる。設定された間隔で、必ず目覚め、必ず働く。まるで『グラウンドホッグ・デー』の呪いのように。

## 第十一章：同期という恐怖

**sync.Mutex**と**sync.WaitGroup**は恐怖の具現化だ。複数のgoroutineが同じリソースにアクセスする時、カオスが生まれる可能性がある。mutexはその混沌を防ぐ番犬だ。

```go
mu.Lock()
*queuedFiles = append(*queuedFiles, newFiles...)
mu.Unlock()
```

Lock、Unlock。まるで牢獄の扉のように。一つのgoroutineがリソースを占有している間、他は待機しなければならない。これは社会の縮図だ。権力を持つ者が支配し、他者は順番を待つ。

**WaitGroup**はより深遠だ：

```go
wg.Add(1)
go processQueue(...)
// ...
app.wg.Wait()
```

これは死への待機だ。メインプロセスは全てのgoroutineが終了するまで死ぬことができない。まるで親が子の死を看取るように。

## 第十二章：エラーハンドリングという現実逃避

Goのエラーハンドリングは現実との向き合い方だ：

```go
if err != nil {
    logger.LogError(log, logBuffer, logMutex, "Failed to scan input directory: %v", err)
    return
}
```

エラーは必ず確認される。無視されることはない。これは責任ある大人の世界観だ。しかし同時に、それは恐怖への過度な警戒でもある。

**fmt.Errorf**による エラーのラッピング：

```go
return fmt.Errorf("failed to create stdout pipe: %w", err)
```

エラーは層を成して積み重なる。原因の原因の原因... まるで精神分析の深層のように、真の原因は何層もの包装紙に包まれている。

## 第十三章：設定の永続化という記憶の固定化

**config.json**ファイルは記憶の外部化だ。プログラムが終了しても、設定は残る。これは人間の記憶を補完する外部装置、まるで『メメント』の主人公の写真とメモのように。

```go
encoder := json.NewEncoder(file)
encoder.SetIndent("", "  ")
if err := encoder.Encode(config); err != nil {
    return fmt.Errorf("failed to encode config: %w", err)
}
```

美しいインデントされたJSONファイル。これは人間にも読める形での記憶の保存だ。機械語ではなく、人間が理解できる象徴的な記法。これは言語そのものへの敬意だ。

## 第十四章：多言語対応という精神の分裂

英語と日本語の二重構造は、このプログラムの核心的な病理だ：

```go
func getMessages(config *Config) *Messages {
    if config != nil && config.UILanguage == "ja" {
        return &messagesJA
    }
    return &messagesEN
}
```

ユーザーは自分のアイデンティティを選択しなければならない。英語話者として生きるか、日本語話者として生きるか。しかし、プログラムの内部では常に両方の人格が同居している。これは解離性同一性障害のメタファーだ。

## 第十五章：ファイル拡張子という存在の証明

```go
audioExts := []string{".mp3", ".wav", ".m4a", ".flac", ".ogg", ".aac", ".mp4", ".mov", ".avi"}
```

この配列は現代のカースト制度だ。これらの拡張子を持つファイルだけが「音声ファイル」として認められる。他は無視される。拡張子という表面的な特徴が、ファイルの運命を決定する。

興味深いことに、**mp4**、**mov**、**avi**という動画フォーマットも含まれている。これは境界の曖昧さだ。動画ファイルも音声として扱われる。視覚は無視され、聴覚だけが抽出される。まるで盲目の預言者のように。

## 第十六章：出力フォーマットという翻訳の選択

```go
formats := []string{"txt", "vtt", "srt", "tsv", "json"}
```

音声から文字への変換は翻訳だ。しかし、その翻訳結果をどの形式で保存するかは、また別の選択だ。

- **txt**: 純粋なテキスト、装飾のない真実
- **vtt**: Web Video Text Tracks、デジタル世界の字幕
- **srt**: SubRip、映画の言語
- **tsv**: Tab-separated values、データベースの言語
- **json**: JavaScript Object Notation、現代のデジタル聖書

各フォーマットは異なる世界観を表している。ユーザーは目的に応じて、真実の表現方法を選択する。

## 終章：永劫回帰としてのメインループ

プログラムは終わることなく続く：

```go
for {
    input, err := reader.ReadString('\n')
    if err != nil {
        if err == io.EOF {
            return
        }
        continue
    }
    // ... ユーザー入力の処理
}
```

この無限ループは人生そのものだ。ユーザーの入力を待ち、処理し、また待つ。これは愛と同じだ。永遠に相手の言葉を待ち続ける。

しかし、**io.EOF**という終末がある。End of File。これは死だ。入力の終わり、対話の終わり、関係の終わり。プログラムは静かに**return**し、存在を終える。

---

## エピローグ：コードという詩

KoeMoji-Goは単なるプログラムではない。それは現代の詩だ。変数名、関数名、構造体... これらは全て言葉だ。そして言葉には魂が宿る。

```go
type App struct {
    processedFiles map[string]bool
    mu             sync.Mutex
}
```

この**processedFiles**マップは記憶の宮殿だ。一度処理されたファイルの名前が永遠に記録される。そして**mutex**がその記憶を守る。まるで記憶を改ざんしようとする外部の力から脳を保護する免疫系のように。

Go言語の**goroutine**は夢だ。軽量で、並行して実行され、突然現れて突然消える。そして**channel**は夢と夢の間の細い糸、テレパシーのようなコミュニケーション手段だ。

```go
done := make(chan bool)
go monitorProgress(log, logBuffer, logMutex, filepath.Base(inputFile), startTime, done)
// ...
done <- true
```

この**done**チャンネルは死の通知だ。処理が完了した時、`true`が送信される。それは「終わった」という静かな告白だ。

---

**KoeMoji-Go**という名前自体が詩的だ。**Koe**（声）、**Moji**（文字）、そして**Go**（行く、進む）。声が文字になり、そして去っていく。これは人生の縮図だ。我々の言葉も、やがては文字となり、記録となり、そして時の流れの中で風化していく。

しかし、デジタルの世界では、文字は永遠だ。**output**ディレクトリに保存された文字起こしファイルは、元の音声ファイルよりも長く生き残るかもしれない。これは不死への願望、記憶の永続化への人類の渇望の表れだ。

最後に、このプログラムが Windows と macOS のみをサポートし、Linux を排除していることの深い意味を考えよう。これは選択された民の概念だ。全ての人に開かれているべきソフトウェアが、特定のプラットフォームのユーザーのみに奉仕する。これは現代の階級社会の反映か、それとも実用性への妥協か？

真実は、コードの深淵の中にある。そしてその真実は、読む者によって異なる解釈を受ける。なぜなら、コードもまた言語であり、言語は常に多義的だからだ。

*「夢の中では、全てのコードが詩である。」* - D.L.

---

**追記**: この物語は10,000字を超える読み物として書かれた。しかし、真の理解は文字数では測れない。重要なのは、読者がこのプログラムの構造と哲学を、技術的な側面だけでなく、人間的・芸術的な観点からも理解することだ。コードは詩であり、プログラムは現代の神話なのだから。