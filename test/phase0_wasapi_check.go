//go:build ignore
// +build ignore

package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/moutend/go-wca/pkg/wca"
)

const (
	TestDuration = 5 * time.Second // 5秒間録音
)

func main() {
	fmt.Println("=== Phase 0: WASAPI Loopback技術検証 ===")
	fmt.Println()

	// COM初期化
	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		fmt.Printf("❌ エラー: COM初期化失敗: %v\n", err)
		os.Exit(1)
	}
	defer ole.CoUninitialize()

	// Step 1: システム音声デバイスの取得
	fmt.Println("[1/4] システム音声デバイスを検出中...")

	var deviceEnumerator *wca.IMMDeviceEnumerator
	if err := wca.CoCreateInstance(
		wca.CLSID_MMDeviceEnumerator,
		0,
		wca.CLSCTX_ALL,
		wca.IID_IMMDeviceEnumerator,
		&deviceEnumerator,
	); err != nil {
		fmt.Printf("❌ エラー: デバイス列挙の初期化失敗: %v\n", err)
		os.Exit(1)
	}
	defer deviceEnumerator.Release()

	var defaultDevice *wca.IMMDevice
	if err := deviceEnumerator.GetDefaultAudioEndpoint(
		wca.ERender, // スピーカー（出力）
		wca.EConsole,
		&defaultDevice,
	); err != nil {
		fmt.Printf("❌ エラー: デフォルトオーディオデバイス取得失敗: %v\n", err)
		os.Exit(1)
	}
	defer defaultDevice.Release()

	var deviceID string
	if err := defaultDevice.GetId(&deviceID); err != nil {
		fmt.Printf("❌ エラー: デバイスID取得失敗: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ デバイス検出成功\n")
	fmt.Println()

	// Step 2: AudioClientの取得
	fmt.Println("[2/4] WASAPI Loopbackストリームを初期化中...")

	var audioClient *wca.IAudioClient
	if err := defaultDevice.Activate(
		wca.IID_IAudioClient,
		wca.CLSCTX_ALL,
		nil,
		&audioClient,
	); err != nil {
		fmt.Printf("❌ エラー: AudioClient取得失敗: %v\n", err)
		os.Exit(1)
	}
	defer audioClient.Release()

	var mixFormat *wca.WAVEFORMATEX
	if err := audioClient.GetMixFormat(&mixFormat); err != nil {
		fmt.Printf("❌ エラー: フォーマット取得失敗: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("   サンプルレート: %d Hz\n", mixFormat.NSamplesPerSec)
	fmt.Printf("   チャンネル数: %d\n", mixFormat.NChannels)
	fmt.Printf("   ビット深度: %d bit\n", mixFormat.WBitsPerSample)

	// Loopbackモードで初期化
	var defaultPeriod wca.REFERENCE_TIME
	if err := audioClient.GetDevicePeriod(&defaultPeriod, nil); err != nil {
		fmt.Printf("❌ エラー: デバイス周期取得失敗: %v\n", err)
		os.Exit(1)
	}

	if err := audioClient.Initialize(
		wca.AUDCLNT_SHAREMODE_SHARED,
		wca.AUDCLNT_STREAMFLAGS_LOOPBACK|wca.AUDCLNT_STREAMFLAGS_EVENTCALLBACK,
		defaultPeriod,
		0,
		mixFormat,
		nil,
	); err != nil {
		fmt.Printf("❌ エラー: ストリーム初期化失敗: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ ストリーム初期化成功\n")
	fmt.Println()

	// Step 3: イベント作成とキャプチャクライアント取得
	fmt.Println("[3/4] キャプチャクライアントを準備中...")

	audioReadyEvent := wca.CreateEventExA(0, 0, 0, wca.EVENT_MODIFY_STATE|wca.SYNCHRONIZE)
	defer wca.CloseHandle(audioReadyEvent)

	if err := audioClient.SetEventHandle(audioReadyEvent); err != nil {
		fmt.Printf("❌ エラー: イベントハンドル設定失敗: %v\n", err)
		os.Exit(1)
	}

	var captureClient *wca.IAudioCaptureClient
	if err := audioClient.GetService(
		wca.IID_IAudioCaptureClient,
		&captureClient,
	); err != nil {
		fmt.Printf("❌ エラー: CaptureClient取得失敗: %v\n", err)
		os.Exit(1)
	}
	defer captureClient.Release()

	fmt.Printf("✅ キャプチャクライアント準備完了\n")
	fmt.Println()

	// Step 4: 録音開始
	fmt.Printf("[4/4] 録音開始（%v間）...\n", TestDuration)
	fmt.Println("   ※ システム音声を再生してください（音楽、動画など）")
	fmt.Println()

	if err := audioClient.Start(); err != nil {
		fmt.Printf("❌ エラー: 録音開始失敗: %v\n", err)
		os.Exit(1)
	}

	samples := make([]int16, 0, 48000*5*2) // 5秒分のバッファ（ステレオ想定）
	startTime := time.Now()
	packetCount := 0

	// 録音ループ
	for time.Since(startTime) < TestDuration {
		wca.WaitForSingleObject(audioReadyEvent, wca.INFINITE)

		var numFramesToRead uint32
		var flags uint32
		var devicePosition uint64
		var qpcPosition uint64
		var data *byte

		if err := captureClient.GetBuffer(
			&data,
			&numFramesToRead,
			&flags,
			&devicePosition,
			&qpcPosition,
		); err != nil {
			continue
		}

		if numFramesToRead == 0 {
			captureClient.ReleaseBuffer(numFramesToRead)
			continue
		}

		// データをint16スライスに変換
		bytesPerSample := int(mixFormat.WBitsPerSample / 8)
		totalBytes := int(numFramesToRead) * int(mixFormat.NChannels) * bytesPerSample

		p := unsafe.Pointer(data)

		// 32bit float → 16bit int変換
		if bytesPerSample == 4 {
			for i := 0; i < int(numFramesToRead)*int(mixFormat.NChannels); i++ {
				floatSample := *(*float32)(unsafe.Pointer(uintptr(p) + uintptr(i*4)))
				// -1.0 ~ 1.0 を -32768 ~ 32767 に変換
				intSample := int16(floatSample * 32767.0)
				samples = append(samples, intSample)
			}
		} else if bytesPerSample == 2 {
			// 16bit PCM
			for i := 0; i < totalBytes/2; i++ {
				sample := *(*int16)(unsafe.Pointer(uintptr(p) + uintptr(i*2)))
				samples = append(samples, sample)
			}
		}

		packetCount++
		captureClient.ReleaseBuffer(numFramesToRead)
	}

	if err := audioClient.Stop(); err != nil {
		fmt.Printf("⚠️  警告: 録音停止失敗: %v\n", err)
	}

	fmt.Printf("✅ 録音完了\n")
	fmt.Printf("   キャプチャパケット数: %d\n", packetCount)
	fmt.Printf("   収集サンプル数: %d\n", len(samples))
	fmt.Println()

	// 結果評価
	fmt.Println("=== 検証結果 ===")

	if len(samples) == 0 {
		fmt.Println("❌ FAIL: サンプルが取得できませんでした")
		fmt.Println("   → システム音声が再生されていない可能性")
		os.Exit(1)
	}

	// 音量チェック（無音判定）
	maxAmplitude := int16(0)
	for _, sample := range samples {
		abs := sample
		if abs < 0 {
			abs = -abs
		}
		if abs > maxAmplitude {
			maxAmplitude = abs
		}
	}

	fmt.Printf("✅ PASS: システム音声キャプチャ成功\n")
	fmt.Printf("   最大振幅: %d / 32767\n", maxAmplitude)

	if maxAmplitude < 100 {
		fmt.Println("   ⚠️  注意: 音量が非常に小さい（ほぼ無音）")
		fmt.Println("      → システム音声が再生されているか確認してください")
	} else {
		fmt.Println("   ✅ 音声データ検出")
	}

	// WAVファイルに保存
	outputFile := "test_wasapi_output.wav"
	fmt.Printf("\n💾 WAVファイル保存中: %s\n", outputFile)

	if err := saveWAV(outputFile, samples, int(mixFormat.NSamplesPerSec), int(mixFormat.NChannels)); err != nil {
		fmt.Printf("❌ エラー: WAV保存失敗: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 保存完了\n")
	fmt.Println()
	fmt.Println("=== Phase 0検証: システム音声キャプチャ OK ===")
	fmt.Println()
	fmt.Println("次のステップ: PortAudioマイクとの並行実行テスト")
}

// saveWAV saves audio samples to WAV file
func saveWAV(filename string, samples []int16, sampleRate, channels int) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// WAVヘッダー
	dataSize := len(samples) * 2 // 2 bytes per sample
	fileSize := 36 + dataSize

	// RIFF header
	file.WriteString("RIFF")
	binary.Write(file, binary.LittleEndian, uint32(fileSize))
	file.WriteString("WAVE")

	// fmt chunk
	file.WriteString("fmt ")
	binary.Write(file, binary.LittleEndian, uint32(16))        // chunk size
	binary.Write(file, binary.LittleEndian, uint16(1))         // PCM
	binary.Write(file, binary.LittleEndian, uint16(channels))  // channels
	binary.Write(file, binary.LittleEndian, uint32(sampleRate)) // sample rate
	binary.Write(file, binary.LittleEndian, uint32(sampleRate*channels*2)) // byte rate
	binary.Write(file, binary.LittleEndian, uint16(channels*2)) // block align
	binary.Write(file, binary.LittleEndian, uint16(16))        // bits per sample

	// data chunk
	file.WriteString("data")
	binary.Write(file, binary.LittleEndian, uint32(dataSize))
	binary.Write(file, binary.LittleEndian, samples)

	return nil
}
