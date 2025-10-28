//go:build ignore
// +build ignore

package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"
	"time"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/gordonklaus/portaudio"
	"github.com/moutend/go-wca/pkg/wca"
)

const (
	TestDuration = 5 * time.Second
	MicSampleRate = 44100
	MicChannels   = 1
	MicBufferSize = 4096
)

func main() {
	fmt.Println("=== Phase 0: デュアル録音技術検証 ===")
	fmt.Println("目的: WASAPI Loopback + PortAudio マイクの並行実行")
	fmt.Println()

	// チャンネル
	systemSamples := make(chan int16, 96000) // システム音声用
	micSamples := make(chan int16, 88200)    // マイク用
	done := make(chan bool)
	var wg sync.WaitGroup

	// 1. WASAPI Loopback（システム音声）を別goroutineで起動
	fmt.Println("[1/3] WASAPI Loopback（システム音声）起動中...")
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := captureSystemAudio(systemSamples, done); err != nil {
			fmt.Printf("❌ システム音声キャプチャエラー: %v\n", err)
		}
	}()
	time.Sleep(500 * time.Millisecond) // 起動待ち
	fmt.Println("✅ システム音声キャプチャ開始")

	// 2. PortAudio（マイク）を別goroutineで起動
	fmt.Println("[2/3] PortAudio（マイク）起動中...")
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := captureMicrophone(micSamples, done); err != nil {
			fmt.Printf("❌ マイクキャプチャエラー: %v\n", err)
		}
	}()
	time.Sleep(500 * time.Millisecond) // 起動待ち
	fmt.Println("✅ マイクキャプチャ開始")

	// 3. ミキサー
	fmt.Println("[3/3] リアルタイムミキシング開始...")
	fmt.Printf("   録音時間: %v\n", TestDuration)
	fmt.Println("   ※ システム音声とマイクに向かって話してください")
	fmt.Println()

	mixed := make([]int16, 0, 480000)
	systemBuf := make([]int16, 0, 10000)
	micBuf := make([]int16, 0, 10000)

	startTime := time.Now()
	systemCount := 0
	micCount := 0

	// ミキシングループ
	for time.Since(startTime) < TestDuration {
		select {
		case s := <-systemSamples:
			systemBuf = append(systemBuf, s)
			systemCount++
		case m := <-micSamples:
			micBuf = append(micBuf, m)
			micCount++
		case <-time.After(10 * time.Millisecond):
			// タイムアウト - バッファをミックス
		}

		// バッファが両方溜まったらミックス
		minLen := len(systemBuf)
		if len(micBuf) < minLen {
			minLen = len(micBuf)
		}

		if minLen > 100 {
			for i := 0; i < minLen; i++ {
				// 簡易ミキシング: 加算合成 + クリッピング防止
				mixedSample := int32(systemBuf[i]) + int32(micBuf[i])
				if mixedSample > 32767 {
					mixedSample = 32767
				} else if mixedSample < -32768 {
					mixedSample = -32768
				}
				mixed = append(mixed, int16(mixedSample))
			}

			// バッファをクリア
			systemBuf = systemBuf[minLen:]
			micBuf = micBuf[minLen:]
		}
	}

	// 録音goroutineに停止シグナル
	close(done)
	time.Sleep(100 * time.Millisecond)

	// チャンネルをクローズして終了待ち
	wg.Wait()
	close(systemSamples)
	close(micSamples)

	fmt.Println("✅ 録音完了")
	fmt.Println()

	// 結果表示
	fmt.Println("=== 検証結果 ===")
	fmt.Printf("システム音声サンプル数: %d\n", systemCount)
	fmt.Printf("マイクサンプル数: %d\n", micCount)
	fmt.Printf("ミックス後サンプル数: %d\n", len(mixed))
	fmt.Println()

	if systemCount == 0 {
		fmt.Println("❌ FAIL: システム音声が取得できませんでした")
		return
	}

	if micCount == 0 {
		fmt.Println("⚠️  WARNING: マイクサンプルが取得できませんでした")
		fmt.Println("   → マイクが正しく設定されているか確認してください")
	}

	if len(mixed) == 0 {
		fmt.Println("❌ FAIL: ミキシングに失敗しました")
		return
	}

	// 音量チェック
	maxAmp := int16(0)
	for _, s := range mixed {
		abs := s
		if abs < 0 {
			abs = -abs
		}
		if abs > maxAmp {
			maxAmp = abs
		}
	}

	fmt.Printf("✅ ミキシング成功\n")
	fmt.Printf("   最大振幅: %d / 32767\n", maxAmp)

	// タイミングのずれチェック
	expectedSamples := int(TestDuration.Seconds() * 44100)
	timingError := float64(abs(len(mixed)-expectedSamples)) / float64(expectedSamples) * 100

	fmt.Printf("\n⏱️  同期評価:\n")
	fmt.Printf("   期待サンプル数: %d\n", expectedSamples)
	fmt.Printf("   実際のサンプル数: %d\n", len(mixed))
	fmt.Printf("   誤差: %.2f%%\n", timingError)

	if timingError < 5.0 {
		fmt.Println("   ✅ 同期品質: 良好")
	} else if timingError < 10.0 {
		fmt.Println("   ⚠️  同期品質: やや不安定")
	} else {
		fmt.Println("   ❌ 同期品質: 不良（バッファ同期に課題あり）")
	}

	// WAV保存
	outputFile := "test_dual_recording.wav"
	fmt.Printf("\n💾 WAVファイル保存: %s\n", outputFile)
	if err := saveWAV(outputFile, mixed, 44100, 1); err != nil {
		fmt.Printf("❌ 保存失敗: %v\n", err)
		return
	}

	fmt.Println("✅ 保存完了")
	fmt.Println()
	fmt.Println("=== Phase 0結論 ===")

	if timingError < 10.0 && len(mixed) > 0 {
		fmt.Println("✅ GO: デュアル録音は技術的に実現可能")
		fmt.Println("   次のステップ: 本格実装を検討")
	} else {
		fmt.Println("⚠️  CAUTION: バッファ同期に課題あり")
		fmt.Println("   → リングバッファ実装が必要")
		fmt.Println("   → 実装コストが高い可能性")
	}
}

func captureSystemAudio(samples chan<- int16, done <-chan bool) error {
	// COM初期化
	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		return err
	}
	defer ole.CoUninitialize()

	var deviceEnumerator *wca.IMMDeviceEnumerator
	if err := wca.CoCreateInstance(
		wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL,
		wca.IID_IMMDeviceEnumerator, &deviceEnumerator,
	); err != nil {
		return err
	}
	defer deviceEnumerator.Release()

	var defaultDevice *wca.IMMDevice
	if err := deviceEnumerator.GetDefaultAudioEndpoint(
		wca.ERender, wca.EConsole, &defaultDevice,
	); err != nil {
		return err
	}
	defer defaultDevice.Release()

	var audioClient *wca.IAudioClient
	if err := defaultDevice.Activate(
		wca.IID_IAudioClient, wca.CLSCTX_ALL, nil, &audioClient,
	); err != nil {
		return err
	}
	defer audioClient.Release()

	var mixFormat *wca.WAVEFORMATEX
	if err := audioClient.GetMixFormat(&mixFormat); err != nil {
		return err
	}

	var defaultPeriod wca.REFERENCE_TIME
	if err := audioClient.GetDevicePeriod(&defaultPeriod, nil); err != nil {
		return err
	}

	if err := audioClient.Initialize(
		wca.AUDCLNT_SHAREMODE_SHARED,
		wca.AUDCLNT_STREAMFLAGS_LOOPBACK|wca.AUDCLNT_STREAMFLAGS_EVENTCALLBACK,
		defaultPeriod, 0, mixFormat, nil,
	); err != nil {
		return err
	}

	audioReadyEvent := wca.CreateEventExA(0, 0, 0, wca.EVENT_MODIFY_STATE|wca.SYNCHRONIZE)
	defer wca.CloseHandle(audioReadyEvent)

	if err := audioClient.SetEventHandle(audioReadyEvent); err != nil {
		return err
	}

	var captureClient *wca.IAudioCaptureClient
	if err := audioClient.GetService(wca.IID_IAudioCaptureClient, &captureClient); err != nil {
		return err
	}
	defer captureClient.Release()

	if err := audioClient.Start(); err != nil {
		return err
	}
	defer audioClient.Stop()

	// キャプチャループ
	for {
		select {
		case <-done:
			return nil
		default:
		}

		wca.WaitForSingleObject(audioReadyEvent, 100)

		var numFramesToRead uint32
		var flags uint32
		var data *byte

		if err := captureClient.GetBuffer(&data, &numFramesToRead, &flags, nil, nil); err != nil {
			continue
		}

		if numFramesToRead > 0 {
			bytesPerSample := int(mixFormat.WBitsPerSample / 8)
			p := unsafe.Pointer(data)

			// 32bit float → 16bit、ステレオ→モノラル変換
			if bytesPerSample == 4 && mixFormat.NChannels == 2 {
				for i := 0; i < int(numFramesToRead); i++ {
					leftFloat := *(*float32)(unsafe.Pointer(uintptr(p) + uintptr(i*8)))
					rightFloat := *(*float32)(unsafe.Pointer(uintptr(p) + uintptr(i*8+4)))
					monoFloat := (leftFloat + rightFloat) / 2.0
					intSample := int16(monoFloat * 32767.0)

					select {
					case samples <- intSample:
					default:
						// チャンネルが満杯の場合はスキップ
					}
				}
			}
		}

		captureClient.ReleaseBuffer(numFramesToRead)
	}

	return nil
}

func captureMicrophone(samples chan<- int16, done <-chan bool) error {
	if err := portaudio.Initialize(); err != nil {
		return err
	}
	defer portaudio.Terminate()

	buffer := make([]int16, MicBufferSize)

	stream, err := portaudio.OpenDefaultStream(
		MicChannels, 0, float64(MicSampleRate), MicBufferSize,
		func(in []int16) {
			for _, s := range in {
				select {
				case samples <- s:
				default:
					// チャンネルが満杯の場合はスキップ
				}
			}
		},
	)
	if err != nil {
		return err
	}
	defer stream.Close()

	if err := stream.Start(); err != nil {
		return err
	}
	defer stream.Stop()

	// 録音継続（doneシグナルを待つ）
	<-done

	_ = buffer // unused
	return nil
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func saveWAV(filename string, samples []int16, sampleRate, channels int) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	dataSize := len(samples) * 2
	fileSize := 36 + dataSize

	file.WriteString("RIFF")
	binary.Write(file, binary.LittleEndian, uint32(fileSize))
	file.WriteString("WAVE")
	file.WriteString("fmt ")
	binary.Write(file, binary.LittleEndian, uint32(16))
	binary.Write(file, binary.LittleEndian, uint16(1))
	binary.Write(file, binary.LittleEndian, uint16(channels))
	binary.Write(file, binary.LittleEndian, uint32(sampleRate))
	binary.Write(file, binary.LittleEndian, uint32(sampleRate*channels*2))
	binary.Write(file, binary.LittleEndian, uint16(channels*2))
	binary.Write(file, binary.LittleEndian, uint16(16))
	file.WriteString("data")
	binary.Write(file, binary.LittleEndian, uint32(dataSize))
	binary.Write(file, binary.LittleEndian, samples)

	return nil
}
