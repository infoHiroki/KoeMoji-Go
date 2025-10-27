#!/bin/bash
# KoeMoji-Go Stress Tests
# 目的: コードレビューで指摘された潜在的リスクの検証

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test results
PASSED=0
FAILED=0
TEST_LOG="test/test-results-$(date +%Y%m%d-%H%M%S).log"

log() {
    echo -e "$1" | tee -a "$TEST_LOG"
}

# Test 1: 短時間録音の繰り返し（スレッドセーフティ検証）
test_rapid_start_stop() {
    log "\n${YELLOW}=== Test 1: 短時間録音×50回（スレッドセーフティ検証） ===${NC}"
    log "目的: audioFile競合によるクラッシュを検出"

    local failures=0

    for i in {1..50}; do
        # Swift CLI直接テスト（より厳密）
        timeout 2 ./cmd/audio-capture/audio-capture -o test/output/rapid-test-$i.wav -d 0.5 > /dev/null 2>&1 &
        local pid=$!
        sleep 0.6

        # プロセスが正常終了したか確認
        if ! wait $pid 2>/dev/null; then
            ((failures++))
            log "${RED}  [FAIL] Iteration $i: Process crashed${NC}"
        fi

        # ファイルが生成されたか確認
        if [ ! -f "test/output/rapid-test-$i.caf" ]; then
            log "${YELLOW}  [WARN] Iteration $i: Output file not created${NC}"
        else
            rm -f "test/output/rapid-test-$i.caf"
        fi

        if (( i % 10 == 0 )); then
            log "  Progress: $i/50"
        fi
    done

    if [ $failures -eq 0 ]; then
        log "${GREEN}✓ Test 1 PASSED: 0/50 failures${NC}"
        ((PASSED++))
    else
        log "${RED}✗ Test 1 FAILED: $failures/50 failures${NC}"
        ((FAILED++))
    fi
}

# Test 2: デュアル録音の動作確認（パス処理検証）
test_dual_recording() {
    log "\n${YELLOW}=== Test 2: デュアル録音30秒（パス処理検証） ===${NC}"
    log "目的: CAF→WAV変換とファイルパス解決の検証"

    # Go側でテストする必要があるため、スキップ
    log "${YELLOW}[SKIP] このテストは手動でGUIから実行する必要があります${NC}"
    log "手順:"
    log "  1. ./koemoji-go でGUI起動"
    log "  2. デュアル録音モードで30秒録音"
    log "  3. ファイルが正しく生成されることを確認"
    log "  4. koemoji.logでCAF→WAV変換を確認"
}

# Test 3: シグナル処理のストレステスト
test_signal_handling() {
    log "\n${YELLOW}=== Test 3: SIGTERM連続送信（Signal Handler検証） ===${NC}"
    log "目的: 二重停止によるクラッシュを検出"

    local failures=0

    for i in {1..20}; do
        # バックグラウンドで録音開始
        ./cmd/audio-capture/audio-capture -o test/output/signal-test-$i.wav -d 0 > /dev/null 2>&1 &
        local pid=$!
        sleep 0.5

        # SIGTERM連続送信（0.01秒間隔で3回）
        kill -TERM $pid 2>/dev/null || true
        sleep 0.01
        kill -TERM $pid 2>/dev/null || true
        sleep 0.01
        kill -TERM $pid 2>/dev/null || true

        # プロセスが終了するまで待つ
        sleep 0.5

        # プロセスがまだ残っているか確認
        if ps -p $pid > /dev/null 2>&1; then
            kill -9 $pid 2>/dev/null || true
            ((failures++))
            log "${RED}  [FAIL] Iteration $i: Process hung${NC}"
        fi

        # クリーンアップ
        rm -f "test/output/signal-test-$i.caf" "test/output/signal-test-$i.wav"

        if (( i % 5 == 0 )); then
            log "  Progress: $i/20"
        fi
    done

    if [ $failures -eq 0 ]; then
        log "${GREEN}✓ Test 3 PASSED: 0/20 failures${NC}"
        ((PASSED++))
    else
        log "${RED}✗ Test 3 FAILED: $failures/20 failures${NC}"
        ((FAILED++))
    fi
}

# Test 4: 長時間録音（メモリリーク検証）
test_long_recording() {
    log "\n${YELLOW}=== Test 4: 長時間録音5分（メモリリーク検証） ===${NC}"
    log "目的: メモリリークとバッファオーバーフローの検出"

    # メモリ使用量の記録開始
    ./cmd/audio-capture/audio-capture -o test/output/long-test.wav -d 300 > /dev/null 2>&1 &
    local pid=$!

    log "  Recording started (PID: $pid)..."

    # 30秒ごとにメモリ使用量をチェック
    local start_mem=$(ps -o rss= -p $pid 2>/dev/null || echo "0")
    log "  Initial memory: ${start_mem} KB"

    for i in {1..10}; do
        sleep 30
        local current_mem=$(ps -o rss= -p $pid 2>/dev/null || echo "0")
        log "  ${i}min: ${current_mem} KB"

        # メモリが10倍以上増えたら警告
        if [ $current_mem -gt $((start_mem * 10)) ]; then
            log "${RED}  [WARN] Memory usage increased significantly${NC}"
        fi
    done

    # 録音終了を待つ
    if wait $pid 2>/dev/null; then
        log "${GREEN}✓ Test 4 PASSED: Recording completed successfully${NC}"
        ((PASSED++))

        # ファイルサイズ確認
        if [ -f "test/output/long-test.caf" ]; then
            local filesize=$(ls -lh test/output/long-test.caf | awk '{print $5}')
            log "  Output file size: $filesize"
            rm -f "test/output/long-test.caf"
        fi
    else
        log "${RED}✗ Test 4 FAILED: Recording failed${NC}"
        ((FAILED++))
    fi
}

# Test 5: 連続実行テスト
test_continuous_execution() {
    log "\n${YELLOW}=== Test 5: 連続実行×100回（安定性検証） ===${NC}"
    log "目的: リソースリークと累積エラーの検出"

    local failures=0

    for i in {1..100}; do
        # 3秒録音
        timeout 5 ./cmd/audio-capture/audio-capture -o test/output/continuous-test.wav -d 3 > /dev/null 2>&1
        local result=$?

        if [ $result -ne 0 ]; then
            ((failures++))
            log "${RED}  [FAIL] Iteration $i: Exit code $result${NC}"
        fi

        # クリーンアップ
        rm -f test/output/continuous-test.caf test/output/continuous-test.wav

        if (( i % 20 == 0 )); then
            log "  Progress: $i/100"
        fi
    done

    if [ $failures -eq 0 ]; then
        log "${GREEN}✓ Test 5 PASSED: 0/100 failures${NC}"
        ((PASSED++))
    else
        log "${RED}✗ Test 5 FAILED: $failures/100 failures (${failures}% failure rate)${NC}"
        ((FAILED++))
    fi
}

# Main execution
main() {
    log "${GREEN}==================================================="
    log "KoeMoji-Go Stress Test Suite"
    log "Started at: $(date)"
    log "===================================================${NC}"

    # 出力ディレクトリ作成
    mkdir -p test/output

    # テスト実行
    test_rapid_start_stop
    test_dual_recording
    test_signal_handling
    test_long_recording
    test_continuous_execution

    # 結果サマリ
    log "\n${GREEN}==================================================="
    log "Test Summary"
    log "===================================================${NC}"
    log "Passed: ${GREEN}${PASSED}${NC}"
    log "Failed: ${RED}${FAILED}${NC}"
    log "Results saved to: $TEST_LOG"

    if [ $FAILED -eq 0 ]; then
        log "\n${GREEN}✓ All tests passed!${NC}"
        exit 0
    else
        log "\n${RED}✗ Some tests failed. Review the log for details.${NC}"
        exit 1
    fi
}

main
