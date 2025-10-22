package main

import (
	"fmt"
	"os"

	"github.com/hirokitakamura/koemoji-go/internal/recorder"
)

func main() {
	fmt.Println("=== VoiceMeeter検出テスト ===")

	deviceName, err := recorder.DetectVoiceMeeter()
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		os.Exit(1)
	}

	if deviceName == "" {
		fmt.Println("❌ VoiceMeeterが見つかりませんでした")
		fmt.Println("\n対処方法:")
		fmt.Println("  1. VoiceMeeterアプリが起動しているか確認")
		fmt.Println("  2. VoiceMeeterがインストールされているか確認")
		fmt.Println("  3. システムを再起動してみる")
		os.Exit(1)
	}

	fmt.Println("✅ VoiceMeeter検出成功！")
	fmt.Println()
	fmt.Printf("検出されたデバイス名: %s\n", deviceName)
	fmt.Println()
	fmt.Println("=== テスト完了 ===")
}
