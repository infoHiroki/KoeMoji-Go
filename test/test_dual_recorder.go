package main

import (
	"fmt"
	"os"
	"time"

	"github.com/hirokitakamura/koemoji-go/internal/recorder"
)

func main() {
	fmt.Println("=== DualRecorder 動作テスト ===")
	fmt.Println()

	// DualRecorder作成
	fmt.Println("[1/5] DualRecorderを初期化中...")
	dr, err := recorder.NewDualRecorder()
	if err != nil {
		fmt.Printf("❌ エラー: 初期化失敗: %v\n", err)
		os.Exit(1)
	}
	defer dr.Close()

	fmt.Println("✅ 初期化成功")
	fmt.Println()

	// 音量設定
	fmt.Println("[2/5] 音量バランスを設定中...")
	dr.SetVolumes(0.7, 1.0) // システム70%, マイク100%
	fmt.Println("   システム音声: 70%")
	fmt.Println("   マイク: 100%")
	fmt.Println("✅ 音量設定完了")
	fmt.Println()

	// 録音開始
	fmt.Println("[3/5] 録音開始...")
	fmt.Println("   ※ システム音声（音楽・動画など）を再生してください")
	fmt.Println("   ※ マイクに向かって話してください")
	fmt.Println("   録音時間: 10秒")
	fmt.Println()

	if err := dr.Start(); err != nil {
		fmt.Printf("❌ エラー: 録音開始失敗: %v\n", err)
		os.Exit(1)
	}

	// 録音中の状態表示
	startTime := time.Now()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for elapsed := 0; elapsed < 10; elapsed++ {
		<-ticker.C
		duration := dr.GetDuration()
		fmt.Printf("   ⏱️  経過時間: %ds / 録音データ: %.1f秒分\n", elapsed+1, duration)
	}

	fmt.Println()
	fmt.Println("[4/5] 録音停止中...")

	if err := dr.Stop(); err != nil {
		fmt.Printf("❌ エラー: 録音停止失敗: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ 録音停止完了")
	fmt.Println()

	// 結果確認
	duration := dr.GetDuration()
	fmt.Println("📊 録音結果:")
	fmt.Printf("   録音時間: %.2f秒\n", duration)

	if duration < 1.0 {
		fmt.Println("   ⚠️  警告: 録音データが非常に短い")
		fmt.Println("   → システム音声・マイクが正しく動作しているか確認してください")
	} else {
		fmt.Println("   ✅ 十分な録音データ")
	}

	fmt.Println()

	// ファイル保存
	fmt.Println("[5/5] WAVファイル保存中...")
	outputFile := "test_dual_recorder_output.wav"

	if err := dr.SaveToFileWithNormalization(outputFile, true); err != nil {
		fmt.Printf("❌ エラー: ファイル保存失敗: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 保存完了: %s\n", outputFile)
	fmt.Println()

	// 最終評価
	fmt.Println("=== テスト結果 ===")

	if duration >= 8.0 {
		fmt.Println("✅ PASS: DualRecorderは正常に動作しています")
		fmt.Println()
		fmt.Println("次のステップ:")
		fmt.Println("1. WAVファイルを再生して音質を確認")
		fmt.Println("2. システム音声とマイクの両方が聞こえるか確認")
		fmt.Println("3. 音量バランスが適切か確認")
	} else if duration >= 5.0 {
		fmt.Println("⚠️  WARNING: 録音時間がやや短い")
		fmt.Println("   → 一部のストリームが正常に動作していない可能性")
	} else {
		fmt.Println("❌ FAIL: 録音が正常に完了していません")
		fmt.Println("   → ログを確認してエラーの原因を特定してください")
	}

	fmt.Println()
	fmt.Printf("実行時間: %v\n", time.Since(startTime))
}
