package ui

import "github.com/infoHiroki/KoeMoji-Go/internal/config"

// Messages contains all UI text strings
type Messages struct {
	// Main UI
	Active     string
	Processing string
	Queue      string
	None       string
	Input      string
	Output     string
	Archive    string
	Last       string
	Next       string
	Never      string
	Soon       string
	Uptime     string

	// Commands
	ConfigCmd    string
	LogsCmd      string
	ScanCmd      string
	QuitCmd      string
	InputDirCmd  string
	OutputDirCmd string
	RecordCmd    string

	// Log levels
	LogInfo  string
	LogProc  string
	LogDone  string
	LogError string
	LogDebug string

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
	WhisperNotFound string
	WhisperLocation string

	// Recording messages
	RecordingDevice string
	Recording       string
	RecordingStop   string
	SelectDevice    string

	// Settings dialog messages
	SettingsTitle  string
	BasicTab       string
	DirectoriesTab string
	LLMTab         string
	RecordingTab   string
	SaveBtn        string
	CancelBtn      string

	// Settings form labels
	LanguageLabel        string
	WhisperModelLabel    string
	SpeechLanguageLabel  string
	ScanIntervalLabel    string
	UseColorsLabel       string
	InputDirLabel        string
	OutputDirLabel       string
	ArchiveDirLabel      string
	LLMEnabledLabel      string
	APIKeyLabel          string
	ModelLabel           string
	PromptTemplateLabel  string
	RecordingDeviceLabel string
	BrowseBtn            string

	// Additional GUI messages
	LogPlaceholder           string
	LogTitle                 string
	UsingDefaultConfig       string
	LogFileOpenError         string
	AppStartedGUI            string
	ResourceCleanupComplete  string
	RecordingDeviceListError string
	DeviceLoadError          string
	DefaultDevice            string
	ConfigSaved              string
	Success                  string
	RecordingExitWarning     string
	RecordingInProgress      string
	DependencyError          string
	ConfigError              string
	ConfigLoadErrorDialog    string
}

var messagesEN = Messages{
	// Main UI
	Active:     "Active",
	Processing: "Processing",
	Queue:      "Queue",
	None:       "None",
	Input:      "Input",
	Output:     "Output",
	Archive:    "Archive",
	Last:       "Last",
	Next:       "Next",
	Never:      "Never",
	Soon:       "Soon",
	Uptime:     "Uptime",

	// Commands
	ConfigCmd:    "config",
	LogsCmd:      "logs",
	ScanCmd:      "scan",
	QuitCmd:      "quit",
	InputDirCmd:  "input",
	OutputDirCmd: "output",
	RecordCmd:    "record",

	// Log levels
	LogInfo:  "INFO",
	LogProc:  "PROC",
	LogDone:  "DONE",
	LogError: "ERROR",
	LogDebug: "DEBUG",

	// Processing messages
	AppStarted:      "KoeMoji-Go started",
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
	WhisperNotFound: "whisper-ctranslate2 not found. Install with: pip install whisper-ctranslate2",
	WhisperLocation: "To check installation location: pip show whisper-ctranslate2",

	// Recording messages
	RecordingDevice: "Recording Device",
	Recording:       "Recording",
	RecordingStop:   "Recording stopped: %s",
	SelectDevice:    "Select recording device",

	// Settings dialog messages
	SettingsTitle:  "KoeMoji-Go Settings",
	BasicTab:       "Basic",
	DirectoriesTab: "Directories",
	LLMTab:         "AI Summary",
	RecordingTab:   "Recording",
	SaveBtn:        "Save",
	CancelBtn:      "Cancel",

	// Settings form labels
	LanguageLabel:        "Language",
	WhisperModelLabel:    "Whisper Model",
	SpeechLanguageLabel:  "Speech Recognition Language",
	ScanIntervalLabel:    "Scan Interval (min)",
	UseColorsLabel:       "Use Colors",
	InputDirLabel:        "Input Folder",
	OutputDirLabel:       "Output Folder",
	ArchiveDirLabel:      "Archive Folder",
	LLMEnabledLabel:      "Enable AI Summary",
	APIKeyLabel:          "API Key",
	ModelLabel:           "Model",
	PromptTemplateLabel:  "Prompt Template",
	RecordingDeviceLabel: "Recording Device",
	BrowseBtn:            "Browse...",

	// Additional GUI messages
	LogPlaceholder:           "**Waiting for log entries...**",
	LogTitle:                 "Logs",
	UsingDefaultConfig:       "Using default configuration due to config load error",
	LogFileOpenError:         "Failed to open log file: %v",
	AppStartedGUI:            "KoeMoji-Go started (GUI mode)",
	ResourceCleanupComplete:  "Application resources cleaned up",
	RecordingDeviceListError: "Failed to get recording devices: %v",
	DeviceLoadError:          "Failed to load recording devices",
	DefaultDevice:            "Default Device",
	ConfigSaved:              "Configuration saved",
	Success:                  "Success",
	RecordingExitWarning:     "Recording in progress (%s elapsed)\nRecording data will be lost. Do you want to exit?",
	RecordingInProgress:      "Recording in Progress",
	DependencyError:          "Audio recognition engine (Whisper) not found: %v\n\nRecording and file management are available, but audio transcription is not possible.\n\nSolution:\npip install faster-whisper whisper-ctranslate2",
	ConfigError:              "Configuration Error",
	ConfigLoadErrorDialog:    "Failed to load configuration: %v\n\nUsing default configuration.",
}

var messagesJA = Messages{
	// Main UI
	Active:     "稼働中",
	Processing: "処理中",
	Queue:      "待機",
	None:       "なし",
	Input:      "入力",
	Output:     "出力",
	Archive:    "アーカイブ",
	Last:       "最終",
	Next:       "次回",
	Never:      "未実行",
	Soon:       "まもなく",
	Uptime:     "稼働時間",

	// Commands
	ConfigCmd:    "設定",
	LogsCmd:      "ログ",
	ScanCmd:      "スキャン",
	QuitCmd:      "終了",
	InputDirCmd:  "入力",
	OutputDirCmd: "出力",
	RecordCmd:    "録音",

	// Log levels
	LogInfo:  "情報",
	LogProc:  "処理",
	LogDone:  "完了",
	LogError: "エラー",
	LogDebug: "デバッグ",

	// Processing messages
	AppStarted:      "KoeMoji-Go を開始しました",
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
	WhisperNotFound: "whisper-ctranslate2が見つかりません。インストール: pip install whisper-ctranslate2",
	WhisperLocation: "インストール場所の確認: pip show whisper-ctranslate2",

	// Recording messages
	RecordingDevice: "録音デバイス",
	Recording:       "録音中",
	RecordingStop:   "録音を停止しました: %s",
	SelectDevice:    "録音デバイスを選択",

	// Settings dialog messages
	SettingsTitle:  "KoeMoji-Go 設定",
	BasicTab:       "基本設定",
	DirectoriesTab: "フォルダ設定",
	LLMTab:         "AI要約",
	RecordingTab:   "録音設定",
	SaveBtn:        "保存",
	CancelBtn:      "キャンセル",

	// Settings form labels
	LanguageLabel:        "言語",
	WhisperModelLabel:    "Whisperモデル",
	SpeechLanguageLabel:  "音声認識言語",
	ScanIntervalLabel:    "スキャン間隔（分）",
	UseColorsLabel:       "色を使用",
	InputDirLabel:        "入力フォルダ",
	OutputDirLabel:       "出力フォルダ",
	ArchiveDirLabel:      "アーカイブフォルダ",
	LLMEnabledLabel:      "AI要約を有効化",
	APIKeyLabel:          "APIキー",
	ModelLabel:           "モデル",
	PromptTemplateLabel:  "プロンプトテンプレート",
	RecordingDeviceLabel: "録音デバイス",
	BrowseBtn:            "参照...",

	// Additional GUI messages
	LogPlaceholder:           "**ログをここに表示します...**",
	LogTitle:                 "ログ",
	UsingDefaultConfig:       "設定読み込みエラーのためデフォルト設定を使用します",
	LogFileOpenError:         "ログファイルのオープンに失敗しました: %v",
	AppStartedGUI:            "KoeMoji-Goを開始しました (GUIモード)",
	ResourceCleanupComplete:  "アプリケーションリソースをクリーンアップしました",
	RecordingDeviceListError: "録音デバイスの取得に失敗しました: %v",
	DeviceLoadError:          "録音デバイスの読み込みに失敗しました",
	DefaultDevice:            "デフォルトデバイス",
	ConfigSaved:              "設定を保存しました",
	Success:                  "成功",
	RecordingExitWarning:     "録音中です（%s経過）\n録音データが失われますが終了しますか？",
	RecordingInProgress:      "録音中",
	DependencyError:          "音声認識エンジン（Whisper）が見つかりません: %v\n\n録音とファイル管理は利用できますが、音声ファイルの文字起こしはできません。\n\n解決方法:\npip install faster-whisper whisper-ctranslate2",
	ConfigError:              "設定エラー",
	ConfigLoadErrorDialog:    "設定の読み込みに失敗しました: %v\n\nデフォルト設定を使用します。",
}

// GetMessages returns the messages for the current UI language
func GetMessages(config *config.Config) *Messages {
	if config != nil && config.UILanguage == "ja" {
		return &messagesJA
	}
	return &messagesEN
}
