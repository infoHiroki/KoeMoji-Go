# サウンドレコーダー設計書 v4.0（TUI統合完了版）

## 概要
KISS原則に基づいたクロスプラットフォーム音声録音ライブラリ。
KoeMoji-GoのTUIに完全統合され、デバイス選択機能・システム音声録音・仮想デバイス自動検出に対応。

## プロジェクト状況
- **実装状況**: 100%完了・TUI統合完了
- **対応プラットフォーム**: macOS/Windows
- **統合バージョン**: v1.5.5
- **最終更新**: 2025-10-03（デバイス名ベース設定に変更）

## 基本要件
- ✅ 音声録音の開始/停止（インタラクティブ制御）
- ✅ WAVファイルとして保存（16bit, 44.1kHz, モノラル）
- ✅ デバイス選択機能（マイク・システム音声対応）
- ✅ クロスプラットフォーム対応（macOS/Windows）
- ✅ 仮想デバイス自動検出（BlackHole/Stereo Mix/集約デバイス）
- ✅ シンプルなAPI・エラーハンドリング

## 技術スタック

### 言語・フレームワーク
- **言語**: Go 1.23+
- **音声ライブラリ**: gordonklaus/portaudio v0.0.0-20250206071425-98a94950218b

### 音声仕様
- **出力形式**: WAV
- **サンプリングレート**: 44100Hz
- **ビット深度**: 16bit（int16）
- **チャンネル数**: モノラル（1チャンネル）
- **バッファサイズ**: 4096フレーム（約93ms@44.1kHz）

### 依存関係
- **macOS**: `brew install portaudio pkg-config`
- **Windows**: PortAudio DLLが必要

## プラットフォーム対応

### macOS
- **マイク録音**: Core Audio経由でデフォルト対応
- **システム音声録音**: BlackHole仮想デバイス経由
- **集約デバイス**: Audio MIDI設定で作成可能
- **マルチ出力**: 音声を聞きながら録音可能
- **制限**: セキュリティによりシステム音声直接録音不可

### Windows  
- **マイク録音**: WASAPI/DirectSound経由でデフォルト対応
- **システム音声録音**: Stereo Mix機能使用
- **名称バリエーション**: "Stereo Mix", "What U Hear", "Rec. Playback"
- **優位性**: 標準機能でシステム音声録音可能（BlackHole不要）

## アーキテクチャ設計

### 核となる構造体

```go
// ユーザー向けデバイス情報
type DeviceInfo struct {
    ID           int     // PortAudio内部のデバイスIndex（環境依存・保存には使用しない）
    Name         string  // デバイス表示名（設定保存に使用）
    IsDefault    bool    // デフォルトデバイスかどうか
    MaxChannels  int     // 最大入力チャンネル数
    HostAPI      string  // ホストAPI名（Core Audio, WASAPI等）
    IsVirtual    bool    // 仮想デバイスかどうか
    VirtualType  string  // 仮想デバイスの種類
}

// 録音エンジン本体
type Recorder struct {
    stream     *portaudio.Stream      // PortAudioストリーム
    samples    []int16               // 録音データバッファ
    recording  bool                  // 録音状態フラグ
    sampleRate float64              // サンプリングレート
    deviceInfo *portaudio.DeviceInfo // 使用中のデバイス情報
    mutex      sync.Mutex           // スレッドセーフ制御
}
```

### 主要API設計

```go
// 基本的な初期化（デフォルトデバイス使用）
func NewRecorder() (*Recorder, error)

// デバイス名指定初期化（v1.5.5+推奨）
func NewRecorderWithDeviceName(deviceName string) (*Recorder, error)

// デバイスID指定初期化（非推奨：環境依存のため）
func NewRecorderWithDevice(deviceID int) (*Recorder, error)

// 利用可能デバイス一覧取得
func ListDevices() ([]DeviceInfo, error)

// 録音制御
func (r *Recorder) Start() error
func (r *Recorder) Stop() error
func (r *Recorder) IsRecording() bool
func (r *Recorder) GetDuration() float64

// ファイル操作
func (r *Recorder) SaveToFile(filename string) error

// リソース管理
func (r *Recorder) Close() error
```

## 実装詳細

### 1. プロジェクト初期化

```bash
# モジュール作成
go mod init soundrecorder

# 依存関係追加
go get github.com/gordonklaus/portaudio
```

```go
// go.mod
module soundrecorder

go 1.23.2

require github.com/gordonklaus/portaudio v0.0.0-20250206071425-98a94950218b
```

### 2. 録音エンジン実装（recorder.go）

```go
package soundrecorder

import (
    "errors"
    "fmt"
    "runtime"
    "strings"
    "sync"

    "github.com/gordonklaus/portaudio"
)

const (
    SampleRate  = 44100
    Channels    = 1
    BufferSize  = 4096
)

type DeviceInfo struct {
    ID           int
    Name         string
    IsDefault    bool
    MaxChannels  int
    HostAPI      string
    IsVirtual    bool
    VirtualType  string
}

type Recorder struct {
    stream     *portaudio.Stream
    samples    []int16
    recording  bool
    sampleRate float64
    deviceInfo *portaudio.DeviceInfo
    mutex      sync.Mutex
}

// デフォルトデバイスでレコーダー初期化
func NewRecorder() (*Recorder, error) {
    err := portaudio.Initialize()
    if err != nil {
        return nil, err
    }

    return &Recorder{
        samples:    make([]int16, 0),
        recording:  false,
        sampleRate: SampleRate,
    }, nil
}

// 指定デバイスでレコーダー初期化
func NewRecorderWithDevice(deviceID int) (*Recorder, error) {
    err := portaudio.Initialize()
    if err != nil {
        return nil, err
    }

    devices, err := portaudio.Devices()
    if err != nil {
        portaudio.Terminate()
        return nil, err
    }

    var selectedDevice *portaudio.DeviceInfo
    for _, device := range devices {
        if device.Index == deviceID && device.MaxInputChannels > 0 {
            selectedDevice = device
            break
        }
    }

    if selectedDevice == nil {
        portaudio.Terminate()
        return nil, fmt.Errorf("device not found or has no input channels: index %d", deviceID)
    }

    return &Recorder{
        samples:    make([]int16, 0),
        recording:  false,
        sampleRate: SampleRate,
        deviceInfo: selectedDevice,
    }, nil
}

// 録音開始
func (r *Recorder) Start() error {
    r.mutex.Lock()
    defer r.mutex.Unlock()

    if r.recording {
        return errors.New("recording already in progress")
    }

    r.samples = make([]int16, 0)

    var stream *portaudio.Stream
    var err error

    if r.deviceInfo != nil {
        // 特定デバイス指定の場合
        params := portaudio.StreamParameters{
            Input: portaudio.StreamDeviceParameters{
                Device:   r.deviceInfo,
                Channels: Channels,
                Latency:  r.deviceInfo.DefaultLowInputLatency,
            },
            SampleRate:      r.sampleRate,
            FramesPerBuffer: BufferSize,
        }
        stream, err = portaudio.OpenStream(params, r.recordCallback)
    } else {
        // デフォルトデバイスの場合
        stream, err = portaudio.OpenDefaultStream(
            Channels,
            0,
            r.sampleRate,
            BufferSize,
            r.recordCallback,
        )
    }

    if err != nil {
        return err
    }

    err = stream.Start()
    if err != nil {
        stream.Close()
        return err
    }

    r.stream = stream
    r.recording = true
    return nil
}

// コールバック関数（リアルタイム音声データ処理）
func (r *Recorder) recordCallback(in []int16) {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    if r.recording {
        r.samples = append(r.samples, in...)
    }
}

// 録音停止
func (r *Recorder) Stop() error {
    r.mutex.Lock()
    defer r.mutex.Unlock()

    if !r.recording {
        return errors.New("no recording in progress")
    }

    err := r.stream.Stop()
    if err != nil {
        return err
    }

    err = r.stream.Close()
    if err != nil {
        return err
    }

    r.recording = false
    r.stream = nil
    return nil
}

// WAVファイル保存
func (r *Recorder) SaveToFile(filename string) error {
    r.mutex.Lock()
    samples := make([]int16, len(r.samples))
    copy(samples, r.samples)
    r.mutex.Unlock()

    if len(samples) == 0 {
        return errors.New("no audio data to save")
    }

    return SaveWAV(filename, samples, int(r.sampleRate), Channels)
}

// リソース解放
func (r *Recorder) Close() error {
    if r.recording {
        err := r.Stop()
        if err != nil {
            return err
        }
    }
    
    portaudio.Terminate()
    return nil
}

// 録音状態確認
func (r *Recorder) IsRecording() bool {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    return r.recording
}

// 録音時間取得
func (r *Recorder) GetDuration() float64 {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    return float64(len(r.samples)) / r.sampleRate
}

// 仮想デバイス検出
func detectVirtualDevice(device *portaudio.DeviceInfo) (bool, string) {
    name := strings.ToLower(device.Name)
    
    switch runtime.GOOS {
    case "darwin":
        if strings.Contains(name, "blackhole") {
            return true, "BlackHole"
        }
        if strings.Contains(name, "aggregate") || strings.Contains(name, "集約") {
            return true, "Aggregate"
        }
        if strings.Contains(name, "multi-output") || strings.Contains(name, "マルチ出力") {
            return true, "Multi-Output"
        }
    case "windows":
        if strings.Contains(name, "stereo mix") || 
           strings.Contains(name, "what u hear") || 
           strings.Contains(name, "rec. playback") {
            return true, "Stereo Mix"
        }
    }
    
    return false, ""
}

// デバイス一覧取得
func ListDevices() ([]DeviceInfo, error) {
    err := portaudio.Initialize()
    if err != nil {
        return nil, err
    }
    defer portaudio.Terminate()

    devices, err := portaudio.Devices()
    if err != nil {
        return nil, err
    }

    defaultInput, err := portaudio.DefaultInputDevice()
    if err != nil {
        defaultInput = nil
    }

    var deviceList []DeviceInfo
    for _, device := range devices {
        if device.MaxInputChannels > 0 {
            isVirtual, virtualType := detectVirtualDevice(device)
            isDefault := defaultInput != nil && device.Index == defaultInput.Index
            
            deviceInfo := DeviceInfo{
                ID:          device.Index,
                Name:        device.Name,
                IsDefault:   isDefault,
                MaxChannels: device.MaxInputChannels,
                HostAPI:     device.HostApi.Name,
                IsVirtual:   isVirtual,
                VirtualType: virtualType,
            }
            deviceList = append(deviceList, deviceInfo)
        }
    }

    return deviceList, nil
}
```

### 3. WAVファイル出力実装（wav.go）

```go
package soundrecorder

import (
    "encoding/binary"
    "os"
)

type WAVHeader struct {
    ChunkID       [4]byte
    ChunkSize     uint32
    Format        [4]byte
    Subchunk1ID   [4]byte
    Subchunk1Size uint32
    AudioFormat   uint16
    NumChannels   uint16
    SampleRate    uint32
    ByteRate      uint32
    BlockAlign    uint16
    BitsPerSample uint16
    Subchunk2ID   [4]byte
    Subchunk2Size uint32
}

func SaveWAV(filename string, samples []int16, sampleRate, channels int) error {
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    bitsPerSample := 16
    dataSize := len(samples) * 2  // 16bit = 2bytes per sample
    fileSize := 36 + dataSize     // WAVヘッダーサイズ + データサイズ

    header := WAVHeader{
        ChunkID:       [4]byte{'R', 'I', 'F', 'F'},
        ChunkSize:     uint32(fileSize),
        Format:        [4]byte{'W', 'A', 'V', 'E'},
        Subchunk1ID:   [4]byte{'f', 'm', 't', ' '},
        Subchunk1Size: 16,
        AudioFormat:   1,  // PCM
        NumChannels:   uint16(channels),
        SampleRate:    uint32(sampleRate),
        ByteRate:      uint32(sampleRate * channels * bitsPerSample / 8),
        BlockAlign:    uint16(channels * bitsPerSample / 8),
        BitsPerSample: uint16(bitsPerSample),
        Subchunk2ID:   [4]byte{'d', 'a', 't', 'a'},
        Subchunk2Size: uint32(dataSize),
    }

    // ヘッダー書き込み
    err = binary.Write(file, binary.LittleEndian, header)
    if err != nil {
        return err
    }

    // 音声データ書き込み
    err = binary.Write(file, binary.LittleEndian, samples)
    if err != nil {
        return err
    }

    return nil
}
```

### 4. インタラクティブUI実装（example/main.go）

```go
package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
    "strconv"
    "strings"
    
    "soundrecorder"
)

// デバイス選択UI
func selectDevice() (*soundrecorder.Recorder, error) {
    devices, err := soundrecorder.ListDevices()
    if err != nil {
        return nil, fmt.Errorf("failed to list devices: %v", err)
    }

    if len(devices) == 0 {
        return nil, fmt.Errorf("no input devices found")
    }

    fmt.Println("Available input devices:")
    for i, device := range devices {
        prefix := fmt.Sprintf("%d: %s", i, device.Name)
        if device.IsDefault {
            prefix += " (default)"
        }
        if device.IsVirtual {
            switch device.VirtualType {
            case "Aggregate":
                prefix += " [集約デバイス]"
            case "Multi-Output":
                prefix += " [マルチ出力]"
            default:
                prefix += fmt.Sprintf(" [%s]", device.VirtualType)
            }
        }
        fmt.Printf("  %s (%d channels, %s)\n", prefix, device.MaxChannels, device.HostAPI)
    }

    fmt.Print("Select device (number) or press Enter for default: ")
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    input := strings.TrimSpace(scanner.Text())

    if input == "" {
        return soundrecorder.NewRecorder()
    }

    deviceIndex, err := strconv.Atoi(input)
    if err != nil || deviceIndex < 0 || deviceIndex >= len(devices) {
        return nil, fmt.Errorf("invalid device selection: %s", input)
    }

    return soundrecorder.NewRecorderWithDevice(devices[deviceIndex].ID)
}

func main() {
    filename := "recording.wav"
    if len(os.Args) > 1 {
        filename = os.Args[1]
    }

    fmt.Println("Sound Recorder v2.0")
    fmt.Printf("Output: %s\n", filename)
    fmt.Println()

    recorder, err := selectDevice()
    if err != nil {
        log.Fatal("Device selection failed:", err)
    }
    defer recorder.Close()

    fmt.Println()
    fmt.Println("Controls:")
    fmt.Println("  Enter - Start/Stop recording")
    fmt.Println("  q     - Quit")
    fmt.Println()

    scanner := bufio.NewScanner(os.Stdin)
    recording := false
    
    for {
        if recording {
            fmt.Printf("Recording (%.1fs) - Enter to stop: ", recorder.GetDuration())
        } else {
            fmt.Print("Ready - Enter to start, 'q' to quit: ")
        }

        if !scanner.Scan() {
            break
        }

        input := strings.TrimSpace(scanner.Text())
        
        if input == "q" {
            break
        }

        if input == "" {
            if recording {
                recorder.Stop()
                err = recorder.SaveToFile(filename)
                if err != nil {
                    log.Printf("Save failed: %v", err)
                    continue
                }
                fmt.Printf("Saved: %s (%.1fs)\n", filename, recorder.GetDuration())
                recording = false
            } else {
                err = recorder.Start()
                if err != nil {
                    log.Printf("Start failed: %v", err)
                    continue
                }
                fmt.Println("Recording started")
                recording = true
            }
        }
    }

    if recording {
        recorder.Stop()
        recorder.SaveToFile(filename)
        fmt.Printf("Saved: %s\n", filename)
    }
}
```

## ビルド・デプロイメント

### macOS環境構築

```bash
# PortAudio依存関係インストール
brew install portaudio pkg-config

# プロジェクト初期化
go mod init soundrecorder
go get github.com/gordonklaus/portaudio

# ビルド
go build -o recorder example/main.go

# 実行
./recorder [出力ファイル名]
```

### Windows環境構築

```bash
# 1. PortAudio for Windowsをダウンロード
# http://www.portaudio.com/download.html

# 2. 必要なDLLをシステムパスに配置

# 3. ビルド
go build -o recorder.exe example/main.go

# 4. 実行
./recorder.exe [出力ファイル名]
```

### クロスコンパイル

```bash
# Windows向けビルド（macOSから）
GOOS=windows GOARCH=amd64 go build -o recorder.exe example/main.go

# macOS向けビルド（Windowsから）
GOOS=darwin GOARCH=amd64 go build -o recorder example/main.go
```

## 重要な実装パターン

### 1. スレッドセーフな録音制御

```go
func (r *Recorder) Start() error {
    r.mutex.Lock()           // 排他制御開始
    defer r.mutex.Unlock()   // 関数終了時に自動解放

    if r.recording {
        return errors.New("recording already in progress")
    }
    // ... 録音開始処理
}
```

### 2. コールバック関数の実装

```go
func (r *Recorder) recordCallback(in []int16) {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    // リアルタイムで呼ばれる処理
    // 高速性が要求されるため最小限の処理のみ
    if r.recording {
        r.samples = append(r.samples, in...)
    }
}
```

### 3. リソース管理パターン

```go
func NewRecorder() (*Recorder, error) {
    err := portaudio.Initialize()  // 初期化
    if err != nil {
        return nil, err
    }
    
    // 成功時のみRecorderを返す
    return &Recorder{...}, nil
}

func (r *Recorder) Close() error {
    if r.recording {
        r.Stop()  // 録音中なら停止
    }
    portaudio.Terminate()  // 必ずリソース解放
    return nil
}
```

### 4. エラーハンドリングパターン

```go
func (r *Recorder) Start() error {
    // ... ストリーム作成
    stream, err := portaudio.OpenStream(params, r.recordCallback)
    if err != nil {
        return err  // エラーをそのまま返す
    }

    err = stream.Start()
    if err != nil {
        stream.Close()  // 失敗時のクリーンアップ
        return err
    }
    
    // 成功時のみ状態更新
    r.stream = stream
    r.recording = true
    return nil
}
```

## プラットフォーム固有の設定

### macOS - BlackHole設定

```bash
# BlackHoleインストール
brew install blackhole-2ch

# Audio MIDI設定での操作：
# 1. アプリケーション → ユーティリティ → Audio MIDI設定
# 2. マルチ出力デバイス作成（音を聞きながら録音）
# 3. 集約デバイス作成（マイク＋システム音声同時録音）
```

### Windows - Stereo Mix設定

```bash
# Windows設定での操作：
# 1. 設定 → システム → サウンド → サウンドの詳細オプション
# 2. 録音タブで右クリック → 「無効なデバイスの表示」
# 3. Stereo Mixを右クリック → 有効
```

## トラブルシューティング

### よくある問題と解決方法

#### 1. ビルドエラー

```bash
# pkg-config not found (macOS)
brew install pkg-config

# PortAudio not found (Windows)
# PortAudio DLLをシステムパスに配置

# Go module not found
go mod tidy
```

#### 2. 実行時エラー

```bash
# Permission denied (macOS)
# システム環境設定 → セキュリティとプライバシー → マイク

# Device not found
# デバイスが物理的に接続されているか確認
# Audio MIDI設定でデバイスが認識されているか確認

# No audio data
# マイクの音量設定を確認
# 他のアプリがデバイスを占有していないか確認
```

#### 3. 音質問題

```bash
# ノイズが多い
# バッファサイズを調整（BufferSize定数）
# サンプリングレートを変更（SampleRate定数）

# 音が途切れる
# システムの負荷を下げる
# バッファサイズを大きくする
```

## 拡張可能性

### 将来の機能拡張案

1. **ステレオ録音対応**
   - Channels定数を2に変更
   - WAVヘッダーの調整

2. **リアルタイム音声処理**
   - コールバック内でフィルタ処理
   - ノイズキャンセリング機能

3. **複数フォーマット対応**
   - MP3エンコーダー追加
   - FLACエンコーダー追加

4. **音声レベル表示**
   - VU メーター実装
   - しきい値録音機能

5. **ネットワーク機能**
   - ストリーミング配信
   - リモート録音制御

## TUI統合実装（v1.4.0）

### 統合仕様
- **操作キー**: `r`キーで録音開始/停止のトグル
- **状態表示**: 🔴録音中 - 経過時間のリアルタイム表示
- **設定統合**: 設定画面（cキー）の項目18でデバイス選択
- **ファイル管理**: `recording_YYYYMMDD_HHMM.wav`で自動保存

### 実装箇所
- **Config**: `RecordingDeviceName`設定（v1.5.5で`RecordingDeviceID`を削除）
- **Messages**: 録音関連メッセージの日英対応
- **App**: 録音状態管理フィールド追加
- **UI**: リアルタイム状態表示とコマンド追加

### ワークフロー
1. `r`キー押下で録音開始
2. 🔴録音中表示でリアルタイム時間更新
3. 再度`r`キー押下で録音停止・保存
4. 次回スキャンで自動的に文字起こし対象に

## デバイス設定システム（v1.5.5更新）

### デバイス名ベース設定の採用理由

**問題点（v1.5.4以前）**:
- `recording_device_id`（デバイスID）はPortAudio内部のインデックス番号
- デバイスID値は環境によって異なる（接続順序、OS、デバイス構成に依存）
- 同じconfig.jsonを別環境で使用すると誤ったデバイスを選択

**解決策（v1.5.5以降）**:
- `recording_device_name`（デバイス名）のみを保存
- 起動時にデバイス名で検索して実際のデバイスを特定
- 環境が変わってもデバイス名が同じなら正しく動作

### 設定フィールド

```json
{
  "recording_device_name": "",  // 空文字列 = デフォルトデバイス
  "recording_max_hours": 0,     // 最大録音時間（0 = 無制限）
  "recording_max_file_mb": 0    // 最大ファイルサイズ（0 = 無制限）
}
```

### デバイス選択の流れ

1. **設定画面でデバイス選択**:
   - `ListDevices()`で利用可能デバイス一覧を取得
   - ユーザーがデバイス名を選択
   - `recording_device_name`に保存

2. **録音開始時のデバイス検索**:
   ```go
   if app.Config.RecordingDeviceName != "" {
       // デバイス名で検索
       app.recorder, err = recorder.NewRecorderWithDeviceName(app.Config.RecordingDeviceName)
   } else {
       // デフォルトデバイス使用
       app.recorder, err = recorder.NewRecorder()
   }
   ```

3. **デバイス名による検索処理**:
   ```go
   func NewRecorderWithDeviceName(deviceName string) (*Recorder, error) {
       devices, err := portaudio.Devices()
       if err != nil {
           return nil, err
       }

       // 完全一致でデバイスを検索
       for _, device := range devices {
           if device.Name == deviceName && device.MaxInputChannels > 0 {
               return &Recorder{
                   deviceInfo: device,
                   // ...
               }, nil
           }
       }

       return nil, fmt.Errorf("recording device not found: '%s'", deviceName)
   }
   ```

### 環境非依存性のメリット

1. **ポータビリティ**: 同じconfig.jsonを異なるマシンで使用可能
2. **ユーザーフレンドリー**: デバイス名は人間が読める形式
3. **保守性**: デバイスID番号の管理が不要
4. **柔軟性**: デバイス接続順序が変わっても影響を受けない

### 注意事項

- デバイス名を変更すると再設定が必要（意図的な仕様）
- デバイスが見つからない場合は起動時にエラー表示
- 空文字列の場合は常にシステムデフォルトデバイスを使用

## ライセンス・配布

```
MIT License

Copyright (c) 2025

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

[標準MIT License条文]
```

---

この設計書により、同等の機能を持つサウンドレコーダーを1から実装することが可能です。