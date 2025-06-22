package ui

import "github.com/hirokitakamura/koemoji-go/internal/config"

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
	RecordingDeviceLabel string
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
	RecordingDeviceLabel: "Recording Device",
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
	RecordingDeviceLabel: "録音デバイス",
}

// GetMessages returns the messages for the current UI language
func GetMessages(config *config.Config) *Messages {
	if config != nil && config.UILanguage == "ja" {
		return &messagesJA
	}
	return &messagesEN
}
