package main

import (
	"fmt"
	"os"

	"github.com/gordonklaus/portaudio"
)

func main() {
	fmt.Println("=== PortAudio デバイス情報 ===\n")

	err := portaudio.Initialize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "PortAudio初期化エラー: %v\n", err)
		os.Exit(1)
	}
	defer portaudio.Terminate()

	devices, err := portaudio.Devices()
	if err != nil {
		fmt.Fprintf(os.Stderr, "デバイス取得エラー: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("検出されたデバイス数: %d\n\n", len(devices))

	for i, device := range devices {
		fmt.Printf("【デバイス %d】\n", i)
		fmt.Printf("  名前: %s\n", device.Name)
		fmt.Printf("  入力チャンネル: %d\n", device.MaxInputChannels)
		fmt.Printf("  出力チャンネル: %d\n", device.MaxOutputChannels)
		fmt.Printf("  デフォルトサンプルレート: %.0f Hz\n", device.DefaultSampleRate)
		fmt.Printf("  HostAPI: %s\n", device.HostApi.Name)
		fmt.Println()

		// VoiceMeeter関連デバイスを強調表示
		nameLower := device.Name
		if contains(nameLower, "voicemeeter") || contains(nameLower, "vb-audio") || contains(nameLower, "vaio") || contains(nameLower, "cable") {
			fmt.Printf("  ⚡ VoiceMeeter関連デバイスの可能性\n\n")
		}
	}

	fmt.Println("=== 検出完了 ===")
}

func contains(s, substr string) bool {
	// 大文字小文字を区別せずに部分一致検索
	sLower := toLower(s)
	substrLower := toLower(substr)
	return indexInString(sLower, substrLower) >= 0
}

func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + ('a' - 'A')
		} else {
			result[i] = r
		}
	}
	return string(result)
}

func indexInString(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
