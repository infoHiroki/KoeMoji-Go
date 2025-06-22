# KoeMoji-Go モックテスト設計書

## 1. 外部依存関係の特定

### 現在のハードコーディングされた外部依存関係

#### whisper package
- **os/exec.Command**: Whisperコマンド実行
- **os.Stat**: ファイル存在確認
- **os.Getenv**: 環境変数取得
- **filepath.Glob**: ファイルパス検索

#### llm package
- **http.Client**: OpenAI API呼び出し
- **http.NewRequest**: HTTPリクエスト作成
- **io.ReadAll**: レスポンス読み取り

#### config package  
- **os.Open/os.Create**: 設定ファイルI/O
- **json.NewDecoder/json.NewEncoder**: JSON処理
- **os/exec.Command**: フォルダ選択ダイアログ
- **portaudio**: オーディオデバイス取得

#### recorder package
- **portaudio**: 音声録音機能
- **os.CreateTemp**: 一時ファイル作成
- **os.File**: ファイル操作

#### processor package
- **filepath.Glob**: ディレクトリスキャン
- **os**: ファイル操作全般

## 2. インターフェース設計の提案

### 2.1 コマンド実行インターフェース

```go
// CommandExecutor - コマンド実行の抽象化
type CommandExecutor interface {
    Execute(name string, args ...string) *CommandResult
    ExecuteWithContext(ctx context.Context, name string, args ...string) *CommandResult
    LookPath(file string) (string, error)
}

type CommandResult interface {
    Run() error
    Start() error
    Wait() error
    StdoutPipe() (io.ReadCloser, error)
    StderrPipe() (io.ReadCloser, error)
    SetStdout(w io.Writer)
    SetStderr(w io.Writer)
}
```

### 2.2 ファイルシステムインターフェース

```go
// FileSystem - ファイル操作の抽象化
type FileSystem interface {
    Open(name string) (File, error)
    Create(name string) (File, error)
    CreateTemp(dir, pattern string) (File, error)
    Stat(name string) (os.FileInfo, error)
    Remove(name string) error
    Glob(pattern string) ([]string, error)
    Getenv(key string) string
}

type File interface {
    io.ReadWriteCloser
    io.Seeker
    Name() string
    Stat() (os.FileInfo, error)
}
```

### 2.3 HTTPクライアントインターフェース

```go
// HTTPClient - HTTP通信の抽象化
type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
    NewRequest(method, url string, body io.Reader) (*http.Request, error)
    SetTimeout(timeout time.Duration)
}

type HTTPResponse interface {
    StatusCode() int
    Body() io.ReadCloser
    Header() http.Header
}
```

### 2.4 オーディオインターフェース

```go
// AudioSystem - PortAudio操作の抽象化
type AudioSystem interface {
    Initialize() error
    Terminate() error
    Devices() ([]AudioDevice, error)
    DefaultInputDevice() (AudioDevice, error)
    OpenStream(params StreamParameters, callback func([]int16)) (AudioStream, error)
    OpenDefaultStream(inputChannels, outputChannels int, sampleRate float64, framesPerBuffer int, callback func([]int16)) (AudioStream, error)
}

type AudioDevice interface {
    Index() int
    Name() string
    MaxInputChannels() int
    MaxOutputChannels() int
    DefaultLowInputLatency() time.Duration
    HostApi() AudioHostAPI
}

type AudioStream interface {
    Start() error
    Stop() error
    Close() error
    IsActive() bool
}
```

## 3. モックライブラリの選定と比較

### 3.1 選択肢の比較

| ライブラリ | メリット | デメリット | 適用場面 |
|-----------|---------|-----------|---------|
| **testify/mock** | 手動モック、軽量、単純 | コード生成なし、手動メンテナンス | シンプルなインターフェース |
| **gomock** | 自動生成、強力な検証 | 生成が必要、複雑 | 複雑な相互作用 |
| **手動モック** | 完全制御、依存なし | 開発工数大、保守コスト高 | 特殊要件 |

### 3.2 推奨選択: testify/mock + 部分的gomock

**理由:**
- testify/mockは学習コストが低く、KoeMoji-Goの開発チームに適している
- 複雑なインターフェース（AudioSystem等）はgomockで自動生成
- プロジェクトの複雑さに応じて段階的に導入可能

## 4. 依存性注入パターンの実装戦略

### 4.1 コンストラクタ注入パターン

```go
// whisper/service.go
type Service struct {
    cmdExecutor CommandExecutor
    fileSystem  FileSystem
    config     *config.Config
}

func NewService(cmdExecutor CommandExecutor, fileSystem FileSystem, config *config.Config) *Service {
    return &Service{
        cmdExecutor: cmdExecutor,
        fileSystem:  fileSystem,
        config:     config,
    }
}

func NewDefaultService(config *config.Config) *Service {
    return NewService(
        &RealCommandExecutor{},
        &RealFileSystem{},
        config,
    )
}
```

### 4.2 ファクトリーパターン

```go
// internal/factory/factory.go
type ServiceFactory struct {
    cmdExecutor CommandExecutor
    fileSystem  FileSystem
    httpClient  HTTPClient
    audioSystem AudioSystem
}

func NewServiceFactory() *ServiceFactory {
    return &ServiceFactory{
        cmdExecutor: &RealCommandExecutor{},
        fileSystem:  &RealFileSystem{},
        httpClient:  &RealHTTPClient{},
        audioSystem: &RealAudioSystem{},
    }
}

func (f *ServiceFactory) CreateWhisperService(config *config.Config) *whisper.Service {
    return whisper.NewService(f.cmdExecutor, f.fileSystem, config)
}

func (f *ServiceFactory) CreateLLMService(config *config.Config) *llm.Service {
    return llm.NewService(f.httpClient, config)
}
```

### 4.3 段階的移行戦略

1. **Phase 1**: 新しいインターフェースを定義し、既存コードと並行稼働
2. **Phase 2**: 各packageで徐々にインターフェースを利用するよう変更
3. **Phase 3**: テストカバレッジを向上させながら既存の直接呼び出しを削除
4. **Phase 4**: 本格的なモックテストを追加

## 5. 具体的な実装例

### 5.1 whisper packageのモック化

```go
// internal/mocks/command_executor.go
package mocks

import (
    "context"
    "io"
    "github.com/stretchr/testify/mock"
)

type MockCommandExecutor struct {
    mock.Mock
}

func (m *MockCommandExecutor) Execute(name string, args ...string) *MockCommandResult {
    argList := append([]interface{}{name}, argsToInterface(args)...)
    callArgs := m.Called(argList...)
    return callArgs.Get(0).(*MockCommandResult)
}

func (m *MockCommandExecutor) LookPath(file string) (string, error) {
    args := m.Called(file)
    return args.String(0), args.Error(1)
}

type MockCommandResult struct {
    mock.Mock
}

func (m *MockCommandResult) Run() error {
    args := m.Called()
    return args.Error(0)
}

func (m *MockCommandResult) StdoutPipe() (io.ReadCloser, error) {
    args := m.Called()
    return args.Get(0).(io.ReadCloser), args.Error(1)
}

// whisper/service_test.go
func TestTranscribeAudio(t *testing.T) {
    // Arrange
    mockCmd := &mocks.MockCommandExecutor{}
    mockFS := &mocks.MockFileSystem{}
    mockResult := &mocks.MockCommandResult{}
    
    config := &config.Config{
        WhisperModel: "large-v3",
        Language:     "ja",
        OutputDir:    "/tmp/output",
        OutputFormat: "txt",
        ComputeType:  "int8",
    }
    
    service := whisper.NewService(mockCmd, mockFS, config)
    
    // Mock setup
    mockCmd.On("Execute", "whisper-ctranslate2", mock.AnythingOfType("string"), 
        mock.AnythingOfType("string"), mock.AnythingOfType("string"),
        mock.AnythingOfType("string"), mock.AnythingOfType("string"),
        mock.AnythingOfType("string"), mock.AnythingOfType("string"),
        mock.AnythingOfType("string")).Return(mockResult)
    
    mockResult.On("Start").Return(nil)
    mockResult.On("Wait").Return(nil)
    mockResult.On("StdoutPipe").Return(io.NopCloser(strings.NewReader("")), nil)
    mockResult.On("StderrPipe").Return(io.NopCloser(strings.NewReader("")), nil)
    
    // Act
    err := service.TranscribeAudio("/test/input.wav")
    
    // Assert
    assert.NoError(t, err)
    mockCmd.AssertExpectations(t)
    mockResult.AssertExpectations(t)
}
```

### 5.2 llm packageのHTTPクライアントモック化

```go
// internal/mocks/http_client.go  
package mocks

import (
    "bytes"
    "io"
    "net/http"
    "github.com/stretchr/testify/mock"
)

type MockHTTPClient struct {
    mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
    args := m.Called(req)
    return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockHTTPClient) NewRequest(method, url string, body io.Reader) (*http.Request, error) {
    args := m.Called(method, url, body)
    return args.Get(0).(*http.Request), args.Error(1)
}

func (m *MockHTTPClient) SetTimeout(timeout time.Duration) {
    m.Called(timeout)
}

// llm/service_test.go
func TestSummarizeText(t *testing.T) {
    // Arrange
    mockClient := &mocks.MockHTTPClient{}
    config := &config.Config{
        LLMSummaryEnabled: true,
        LLMAPIKey:        "test-key",
        LLMAPIProvider:   "openai",
        LLMModel:         "gpt-4o",
        LLMMaxTokens:     4096,
        SummaryPromptTemplate: "Summarize: {text}",
        SummaryLanguage: "ja",
    }
    
    service := llm.NewService(mockClient, config)
    
    // Mock response
    responseBody := `{
        "choices": [{
            "message": {
                "content": "テスト要約です"
            }
        }]
    }`
    
    mockResponse := &http.Response{
        StatusCode: 200,
        Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
        Header:     make(http.Header),
    }
    
    mockClient.On("NewRequest", "POST", "https://api.openai.com/v1/chat/completions", mock.Anything).Return(&http.Request{}, nil)
    mockClient.On("Do", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil)
    
    // Act
    summary, err := service.SummarizeText("テストテキストです")
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "テスト要約です", summary)
    mockClient.AssertExpectations(t)
}
```

### 5.3 config packageのファイルI/Oモック化

```go
// internal/mocks/filesystem.go
package mocks

import (
    "os"
    "github.com/stretchr/testify/mock"
)

type MockFileSystem struct {
    mock.Mock
}

func (m *MockFileSystem) Open(name string) (File, error) {
    args := m.Called(name)
    return args.Get(0).(File), args.Error(1)
}

func (m *MockFileSystem) Create(name string) (File, error) {
    args := m.Called(name)
    return args.Get(0).(File), args.Error(1)
}

func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
    args := m.Called(name)
    return args.Get(0).(os.FileInfo), args.Error(1)
}

type MockFile struct {
    mock.Mock
    content []byte
    pos     int
}

func (m *MockFile) Read(p []byte) (n int, err error) {
    args := m.Called(p)
    return args.Int(0), args.Error(1)
}

func (m *MockFile) Write(p []byte) (n int, err error) {
    args := m.Called(p)
    return args.Int(0), args.Error(1)
}

func (m *MockFile) Close() error {
    args := m.Called()
    return args.Error(0)
}

// config/config_test.go
func TestLoadConfig(t *testing.T) {
    // Arrange
    mockFS := &mocks.MockFileSystem{}
    mockFile := &mocks.MockFile{}
    
    service := config.NewService(mockFS)
    
    configJSON := `{
        "whisper_model": "large-v3",
        "language": "ja",
        "ui_language": "ja"
    }`
    
    mockFS.On("Open", "config.json").Return(mockFile, nil)
    mockFile.On("Read", mock.Anything).Return(len(configJSON), nil).Run(func(args mock.Arguments) {
        copy(args[0].([]byte), configJSON)
    })
    mockFile.On("Close").Return(nil)
    
    // Act
    config, err := service.LoadConfig("config.json")
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "large-v3", config.WhisperModel)
    assert.Equal(t, "ja", config.Language)
    mockFS.AssertExpectations(t)
    mockFile.AssertExpectations(t)
}
```

## 6. テスト用ヘルパー関数とユーティリティの設計

### 6.1 テストヘルパー

```go
// internal/testutil/helpers.go
package testutil

import (
    "io"
    "strings"
    "testing"
    "github.com/hirokitakamura/koemoji-go/internal/config"
)

// CreateTestConfig テスト用の設定を生成
func CreateTestConfig() *config.Config {
    return &config.Config{
        WhisperModel:        "base",
        Language:            "ja",
        UILanguage:          "ja",
        ScanIntervalMinutes: 1,
        MaxCpuPercent:       95,
        ComputeType:         "int8",
        UseColors:           true,
        OutputFormat:        "txt",
        InputDir:            "./test/input",
        OutputDir:           "./test/output",
        ArchiveDir:          "./test/archive",
    }
}

// MockReader モックI/Oリーダー
func MockReader(content string) io.ReadCloser {
    return io.NopCloser(strings.NewReader(content))
}

// AssertLogContains ログ内容の検証
func AssertLogContains(t *testing.T, logs []string, expected string) {
    t.Helper()
    for _, log := range logs {
        if strings.Contains(log, expected) {
            return
        }
    }
    t.Errorf("Expected log to contain '%s', but it was not found", expected)
}
```

### 6.2 モックファクトリー

```go
// internal/testutil/mock_factory.go
package testutil

import (
    "github.com/hirokitakamura/koemoji-go/internal/mocks"
)

type MockFactory struct {
    CmdExecutor *mocks.MockCommandExecutor
    FileSystem  *mocks.MockFileSystem
    HTTPClient  *mocks.MockHTTPClient
    AudioSystem *mocks.MockAudioSystem
}

func NewMockFactory() *MockFactory {
    return &MockFactory{
        CmdExecutor: &mocks.MockCommandExecutor{},
        FileSystem:  &mocks.MockFileSystem{},
        HTTPClient:  &mocks.MockHTTPClient{},
        AudioSystem: &mocks.MockAudioSystem{},
    }
}

func (f *MockFactory) CreateWhisperService(config *config.Config) *whisper.Service {
    return whisper.NewService(f.CmdExecutor, f.FileSystem, config)
}

func (f *MockFactory) CreateLLMService(config *config.Config) *llm.Service {
    return llm.NewService(f.HTTPClient, config)
}

func (f *MockFactory) AssertAllExpectations(t *testing.T) {
    f.CmdExecutor.AssertExpectations(t)
    f.FileSystem.AssertExpectations(t)
    f.HTTPClient.AssertExpectations(t)
    f.AudioSystem.AssertExpectations(t)
}
```

## 7. 段階的実装プラン

### Phase 1: インターフェース導入 (1-2週間)
1. インターフェース定義の追加
2. 実装クラスの作成（既存コードのラッパー）
3. 新しいコンストラクタの追加

### Phase 2: 徐々のリファクタリング (2-3週間) 
1. whisper packageの変更
2. llm packageの変更
3. config packageの変更

### Phase 3: テストの拡充 (2-3週間)
1. モッククラスの作成
2. 単体テストの追加
3. 統合テストの改善

### Phase 4: カバレッジ向上 (1-2週間)
1. エッジケースのテスト追加
2. エラーハンドリングのテスト
3. パフォーマンステストの追加

## 8. 期待される効果

### テストの改善
- **カバレッジ向上**: 現在のテストでは困難な外部依存部分をモック化
- **実行速度向上**: 実際のコマンド実行やHTTP通信を避けることで高速化
- **信頼性向上**: 外部環境に依存しない安定したテスト

### 開発効率の向上
- **並行開発**: 外部サービスの準備を待たずに開発可能
- **デバッグ改善**: 特定の条件やエラー状況を容易に再現
- **CI/CD改善**: 外部依存を排除して安定したビルドパイプライン

### コード品質の向上
- **疎結合**: インターフェースによる依存性の明確化
- **保守性**: テスタブルなコード構造による保守性向上
- **拡張性**: 新しい実装の追加が容易