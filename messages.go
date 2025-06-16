package main

// Messages contains all UI text strings
type Messages struct {
	// Main UI
	Active          string
	Processing      string
	Queue           string
	None            string
	Input           string
	Output          string
	Archive         string
	Last            string
	Next            string
	Never           string
	Soon            string
	Uptime          string
	
	// Commands
	ConfigCmd       string
	LogsCmd         string
	ScanCmd         string
	QuitCmd         string
	InputDirCmd     string
	OutputDirCmd    string
	
	// Config menu
	ConfigTitle     string
	WhisperModel    string
	Language        string
	UILanguage      string
	ScanInterval    string
	MaxCPUPercent   string
	ComputeType     string
	UseColors       string
	UIMode          string
	OutputFormat    string
	InputDirectory  string
	OutputDirectory string
	ArchiveDirectory string
	ResetDefaults   string
	SaveAndExit     string
	QuitWithoutSave string
	SelectOption    string
	Minutes         string
	Current         string
	
	// Config prompts
	SelectModel     string
	EnterLanguage   string
	SelectUILang    string
	EnterInterval   string
	EnterCPU        string
	SelectCompute   string
	EnableColors    string
	SelectUIMode    string
	SelectFormat    string
	SelectFolder    string
	TypePath        string
	KeepCurrent     string
	ResetConfirm    string
	UnsavedChanges  string
	
	// Config messages
	ModelSet        string
	LanguageSet     string
	UILanguageSet   string
	IntervalSet     string
	CPUSet          string
	ComputeSet      string
	ColorsEnabled   string
	ColorsDisabled  string
	UIModeSet       string
	FormatSet       string
	InputDirSet     string
	OutputDirSet    string
	ArchiveDirSet   string
	ConfigReset     string
	ConfigSaved     string
	NoChanges       string
	InvalidOption   string
	InvalidInput    string
	FolderSelectFail string
	
	// Log levels
	LogInfo         string
	LogProc         string
	LogDone         string
	LogError        string
	LogDebug        string
	
	// Processing messages
	AppStarted      string
	AppRunning      string
	MonitoringDir   string
	ShuttingDown    string
	ScanningDir     string
	FoundFiles      string
	ProcessingFile  string
	ProcessComplete string
	ProcessFailed   string
	MovingToArchive string
	
	// Error messages
	LogFileError    string
	ConfigLoadError string
	ConfigSaveError string
	DirCreateError  string
	DirNotExist     string
	FileNotFound    string
	InvalidPath     string
	TranscribeFail  string
	UnsupportedOS   string
}

var messagesEN = Messages{
	// Main UI
	Active:          "Active",
	Processing:      "Processing",
	Queue:           "Queue",
	None:            "None",
	Input:           "Input",
	Output:          "Output",
	Archive:         "Archive",
	Last:            "Last",
	Next:            "Next",
	Never:           "Never",
	Soon:            "Soon",
	Uptime:          "Uptime",
	
	// Commands
	ConfigCmd:       "config",
	LogsCmd:         "logs",
	ScanCmd:         "scan",
	QuitCmd:         "quit",
	InputDirCmd:     "input",
	OutputDirCmd:    "output",
	
	// Config menu
	ConfigTitle:     "KoeMoji-Go Configuration",
	WhisperModel:    "Whisper Model",
	Language:        "Language",
	UILanguage:      "UI Language",
	ScanInterval:    "Scan Interval",
	MaxCPUPercent:   "Max CPU Percent",
	ComputeType:     "Compute Type",
	UseColors:       "Use Colors",
	UIMode:          "UI Mode",
	OutputFormat:    "Output Format",
	InputDirectory:  "Input Directory",
	OutputDirectory: "Output Directory",
	ArchiveDirectory: "Archive Directory",
	ResetDefaults:   "Reset to defaults",
	SaveAndExit:     "Save and exit",
	QuitWithoutSave: "Quit without saving",
	SelectOption:    "Select option",
	Minutes:         "minutes",
	Current:         "current",
	
	// Config prompts
	SelectModel:     "Select model (1-%d) or press Enter to keep current:",
	EnterLanguage:   "Enter new language code (e.g., ja, en, zh) or press Enter to keep current:",
	SelectUILang:    "Select UI language (1-2) or press Enter to keep current:",
	EnterInterval:   "Enter new scan interval (minutes) or press Enter to keep current:",
	EnterCPU:        "Enter new max CPU percent (1-100) or press Enter to keep current:",
	SelectCompute:   "Select compute type (1-%d) or press Enter to keep current:",
	EnableColors:    "Enable colors? (y/n) or press Enter to keep current:",
	SelectUIMode:    "Select UI mode (1-2) or press Enter to keep current:",
	SelectFormat:    "Select output format (1-%d) or press Enter to keep current:",
	SelectFolder:    "Press Enter to select folder with dialog, or type path manually:",
	TypePath:        "Enter new %s path or press Enter to keep current:",
	KeepCurrent:     "or press Enter to keep current",
	ResetConfirm:    "Are you sure you want to reset all settings to defaults? (y/N):",
	UnsavedChanges:  "You have unsaved changes. Are you sure you want to quit? (y/N):",
	
	// Config messages
	ModelSet:        "Whisper model set to: %s",
	LanguageSet:     "Language set to: %s",
	UILanguageSet:   "UI language set to: %s",
	IntervalSet:     "Scan interval set to: %d minutes",
	CPUSet:          "Max CPU percent set to: %d%%",
	ComputeSet:      "Compute type set to: %s",
	ColorsEnabled:   "Colors enabled",
	ColorsDisabled:  "Colors disabled",
	UIModeSet:       "UI mode set to: %s",
	FormatSet:       "Output format set to: %s",
	InputDirSet:     "Input directory set to: %s",
	OutputDirSet:    "Output directory set to: %s",
	ArchiveDirSet:   "Archive directory set to: %s",
	ConfigReset:     "Configuration reset to defaults.",
	ConfigSaved:     "Configuration saved successfully!",
	NoChanges:       "No changes to save.",
	InvalidOption:   "Invalid option. Please try again.",
	InvalidInput:    "Invalid input.",
	FolderSelectFail: "Folder selection failed: %v",
	
	// Log levels
	LogInfo:         "INFO",
	LogProc:         "PROC",
	LogDone:         "DONE",
	LogError:        "ERROR",
	LogDebug:        "DEBUG",
	
	// Processing messages
	AppStarted:      "KoeMoji-Go v%s started",
	AppRunning:      "KoeMoji-Go is running. Use commands below to interact.",
	MonitoringDir:   "Monitoring %s directory every %d minutes",
	ShuttingDown:    "Shutting down KoeMoji-Go...",
	ScanningDir:     "Scanning directory for audio files...",
	FoundFiles:      "Found %d audio files to process",
	ProcessingFile:  "Processing %s",
	ProcessComplete: "Completed %s in %s",
	ProcessFailed:   "Failed to process %s: %v",
	MovingToArchive: "Moving %s to archive",
	
	// Error messages
	LogFileError:    "Failed to open log file: %v",
	ConfigLoadError: "Failed to load config: %v",
	ConfigSaveError: "Failed to save config: %v",
	DirCreateError:  "Failed to create directory %s: %v",
	DirNotExist:     "Directory does not exist: %s",
	FileNotFound:    "File not found: %s",
	InvalidPath:     "Invalid file path: %v",
	TranscribeFail:  "Transcription failed: %v",
	UnsupportedOS:   "Log viewing not supported on this platform",
}

var messagesJA = Messages{
	// Main UI
	Active:          "稼働中",
	Processing:      "処理中",
	Queue:           "待機",
	None:            "なし",
	Input:           "入力",
	Output:          "出力",
	Archive:         "アーカイブ",
	Last:            "最終",
	Next:            "次回",
	Never:           "未実行",
	Soon:            "まもなく",
	Uptime:          "稼働時間",
	
	// Commands
	ConfigCmd:       "設定",
	LogsCmd:         "ログ",
	ScanCmd:         "スキャン",
	QuitCmd:         "終了",
	InputDirCmd:     "入力",
	OutputDirCmd:    "出力",
	
	// Config menu
	ConfigTitle:     "KoeMoji-Go 設定",
	WhisperModel:    "Whisperモデル",
	Language:        "認識言語",
	UILanguage:      "UI言語",
	ScanInterval:    "スキャン間隔",
	MaxCPUPercent:   "最大CPU使用率",
	ComputeType:     "計算タイプ",
	UseColors:       "色を使用",
	UIMode:          "UIモード",
	OutputFormat:    "出力フォーマット",
	InputDirectory:  "入力ディレクトリ",
	OutputDirectory: "出力ディレクトリ",
	ArchiveDirectory: "アーカイブディレクトリ",
	ResetDefaults:   "デフォルトに戻す",
	SaveAndExit:     "保存して終了",
	QuitWithoutSave: "保存せずに終了",
	SelectOption:    "オプションを選択",
	Minutes:         "分",
	Current:         "現在",
	
	// Config prompts
	SelectModel:     "モデルを選択 (1-%d) またはEnterで現在の設定を維持:",
	EnterLanguage:   "新しい言語コード (例: ja, en, zh) を入力またはEnterで現在の設定を維持:",
	SelectUILang:    "UI言語を選択 (1-2) またはEnterで現在の設定を維持:",
	EnterInterval:   "新しいスキャン間隔（分）を入力またはEnterで現在の設定を維持:",
	EnterCPU:        "新しい最大CPU使用率 (1-100) を入力またはEnterで現在の設定を維持:",
	SelectCompute:   "計算タイプを選択 (1-%d) またはEnterで現在の設定を維持:",
	EnableColors:    "色を有効にしますか？ (y/n) またはEnterで現在の設定を維持:",
	SelectUIMode:    "UIモードを選択 (1-2) またはEnterで現在の設定を維持:",
	SelectFormat:    "出力フォーマットを選択 (1-%d) またはEnterで現在の設定を維持:",
	SelectFolder:    "Enterでフォルダ選択ダイアログを開く、または手動でパスを入力:",
	TypePath:        "新しい%sパスを入力またはEnterで現在の設定を維持:",
	KeepCurrent:     "またはEnterで現在の設定を維持",
	ResetConfirm:    "本当にすべての設定をデフォルトに戻しますか？ (y/N):",
	UnsavedChanges:  "未保存の変更があります。本当に終了しますか？ (y/N):",
	
	// Config messages
	ModelSet:        "Whisperモデルを設定: %s",
	LanguageSet:     "言語を設定: %s",
	UILanguageSet:   "UI言語を設定: %s",
	IntervalSet:     "スキャン間隔を設定: %d分",
	CPUSet:          "最大CPU使用率を設定: %d%%",
	ComputeSet:      "計算タイプを設定: %s",
	ColorsEnabled:   "色を有効にしました",
	ColorsDisabled:  "色を無効にしました",
	UIModeSet:       "UIモードを設定: %s",
	FormatSet:       "出力フォーマットを設定: %s",
	InputDirSet:     "入力ディレクトリを設定: %s",
	OutputDirSet:    "出力ディレクトリを設定: %s",
	ArchiveDirSet:   "アーカイブディレクトリを設定: %s",
	ConfigReset:     "設定をデフォルトに戻しました。",
	ConfigSaved:     "設定を保存しました！",
	NoChanges:       "変更はありません。",
	InvalidOption:   "無効なオプションです。もう一度お試しください。",
	InvalidInput:    "無効な入力です。",
	FolderSelectFail: "フォルダ選択に失敗: %v",
	
	// Log levels
	LogInfo:         "情報",
	LogProc:         "処理",
	LogDone:         "完了",
	LogError:        "エラー",
	LogDebug:        "デバッグ",
	
	// Processing messages
	AppStarted:      "KoeMoji-Go v%s を開始しました",
	AppRunning:      "KoeMoji-Goが実行中です。以下のコマンドを使用してください。",
	MonitoringDir:   "%sディレクトリを%d分ごとに監視しています",
	ShuttingDown:    "KoeMoji-Goを終了しています...",
	ScanningDir:     "音声ファイルをスキャンしています...",
	FoundFiles:      "%d個の音声ファイルを検出しました",
	ProcessingFile:  "%sを処理中",
	ProcessComplete: "%sの処理を完了 (処理時間: %s)",
	ProcessFailed:   "%sの処理に失敗: %v",
	MovingToArchive: "%sをアーカイブに移動",
	
	// Error messages
	LogFileError:    "ログファイルを開けません: %v",
	ConfigLoadError: "設定の読み込みに失敗: %v",
	ConfigSaveError: "設定の保存に失敗: %v",
	DirCreateError:  "ディレクトリ%sの作成に失敗: %v",
	DirNotExist:     "ディレクトリが存在しません: %s",
	FileNotFound:    "ファイルが見つかりません: %s",
	InvalidPath:     "無効なファイルパス: %v",
	TranscribeFail:  "文字起こしに失敗: %v",
	UnsupportedOS:   "このプラットフォームではログ表示はサポートされていません",
}

// getMessages returns the messages for the current UI language
func (app *App) getMessages() *Messages {
	if app.config != nil && app.config.UILanguage == "ja" {
		return &messagesJA
	}
	return &messagesEN
}