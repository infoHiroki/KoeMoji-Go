package processor

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hirokitakamura/koemoji-go/internal/config"
	"github.com/hirokitakamura/koemoji-go/internal/llm"
	"github.com/hirokitakamura/koemoji-go/internal/logger"
	"github.com/hirokitakamura/koemoji-go/internal/ui"
	"github.com/hirokitakamura/koemoji-go/internal/whisper"
)

func StartProcessing(ctx context.Context, config *config.Config, log *log.Logger, logBuffer *[]logger.LogEntry, logMutex *sync.RWMutex,
	lastScanTime *time.Time, queuedFiles *[]string, processingFile *string, isProcessing *bool,
	processedFiles *map[string]bool, mu *sync.Mutex, wg *sync.WaitGroup, debugMode bool) {

	// Initial scan
	ScanAndProcess(config, log, logBuffer, logMutex, lastScanTime, queuedFiles, processingFile,
		isProcessing, processedFiles, mu, wg, debugMode)

	// Periodic scan with context cancellation
	ticker := time.NewTicker(time.Duration(config.ScanIntervalMinutes) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.LogInfo(log, logBuffer, logMutex, "File processing stopped by context cancellation")
			return
		case <-ticker.C:
			ScanAndProcess(config, log, logBuffer, logMutex, lastScanTime, queuedFiles, processingFile,
				isProcessing, processedFiles, mu, wg, debugMode)
		}
	}
}

func ScanAndProcess(config *config.Config, log *log.Logger, logBuffer *[]logger.LogEntry, logMutex *sync.RWMutex,
	lastScanTime *time.Time, queuedFiles *[]string, processingFile *string, isProcessing *bool,
	processedFiles *map[string]bool, mu *sync.Mutex, wg *sync.WaitGroup, debugMode bool) {

	*lastScanTime = time.Now()
	msg := ui.GetMessages(config)
	logger.LogInfo(log, logBuffer, logMutex, msg.ScanningDir)

	files, err := filepath.Glob(filepath.Join(config.InputDir, "*"))
	if err != nil {
		logger.LogError(log, logBuffer, logMutex, "Failed to scan input directory: %v", err)
		return
	}

	newFiles := filterNewAudioFiles(files, processedFiles, mu)
	if len(newFiles) == 0 {
		logger.LogDebug(log, logBuffer, logMutex, debugMode, "No new files found")
		return
	}

	logger.LogInfo(log, logBuffer, logMutex, msg.FoundFiles, len(newFiles))

	// Add files to queue
	mu.Lock()
	*queuedFiles = append(*queuedFiles, newFiles...)
	
	// Phase 1: Periodic cleanup of processed files map
	if len(*processedFiles) > 5000 {
		cleanupProcessedFiles(processedFiles, mu, log, logBuffer, logMutex)
	}
	mu.Unlock()

	// Start processing if not already processing (with proper locking)
	mu.Lock()
	if !*isProcessing {
		*isProcessing = true
		mu.Unlock()
		if wg != nil {
			wg.Add(1)
		}
		go processQueue(config, log, logBuffer, logMutex, queuedFiles, processingFile,
			isProcessing, mu, wg, debugMode)
	} else {
		mu.Unlock()
	}
}

func filterNewAudioFiles(files []string, processedFiles *map[string]bool, mu *sync.Mutex) []string {
	mu.Lock()
	defer mu.Unlock()

	var newFiles []string
	for _, file := range files {
		if ui.IsAudioFile(file) && !(*processedFiles)[file] {
			(*processedFiles)[file] = true
			newFiles = append(newFiles, file)
		}
	}
	return newFiles
}

func processQueue(config *config.Config, log *log.Logger, logBuffer *[]logger.LogEntry, logMutex *sync.RWMutex,
	queuedFiles *[]string, processingFile *string, isProcessing *bool, mu *sync.Mutex,
	wg *sync.WaitGroup, debugMode bool) {

	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()

	for {
		mu.Lock()
		if len(*queuedFiles) == 0 {
			*isProcessing = false
			*processingFile = ""
			mu.Unlock()
			return
		}

		// Get next file from queue
		filePath := (*queuedFiles)[0]
		*queuedFiles = (*queuedFiles)[1:]
		*processingFile = filepath.Base(filePath)
		*isProcessing = true
		mu.Unlock()

		// Process the file
		msg := ui.GetMessages(config)
		logger.LogProc(log, logBuffer, logMutex, msg.ProcessingFile, *processingFile)
		startTime := time.Now()

		if err := whisper.TranscribeAudio(config, log, logBuffer, logMutex, debugMode, filePath); err != nil {
			logger.LogError(log, logBuffer, logMutex, msg.ProcessFailed, *processingFile, err)
		} else {
			duration := time.Since(startTime)
			logger.LogDone(log, logBuffer, logMutex, msg.ProcessComplete, *processingFile, formatDuration(duration))

			// Generate summary if enabled
			if config.LLMSummaryEnabled {
				if err := generateSummary(config, log, logBuffer, logMutex, debugMode, filePath); err != nil {
					logger.LogError(log, logBuffer, logMutex, "Summary generation failed for %s: %v", *processingFile, err)
				}
			}

			// Move to archive
			msg2 := ui.GetMessages(config)
			logger.LogProc(log, logBuffer, logMutex, msg2.MovingToArchive, *processingFile)
			if err := moveToArchive(config, filePath); err != nil {
				logger.LogError(log, logBuffer, logMutex, msg.ProcessFailed, *processingFile, err)
			}
		}
	}
}

func moveToArchive(config *config.Config, sourcePath string) error {
	filename := filepath.Base(sourcePath)
	destPath := filepath.Join(config.ArchiveDir, filename)

	// Handle duplicate filenames
	if _, err := os.Stat(destPath); err == nil {
		timestamp := time.Now().Format("20060102_150405")
		ext := filepath.Ext(filename)
		name := strings.TrimSuffix(filename, ext)
		destPath = filepath.Join(config.ArchiveDir, fmt.Sprintf("%s_%s%s", name, timestamp, ext))
	}

	if err := os.Rename(sourcePath, destPath); err != nil {
		return err
	}
	return nil
}

func EnsureDirectories(config *config.Config, log *log.Logger) {
	dirs := []string{config.InputDir, config.OutputDir, config.ArchiveDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			msg := ui.GetMessages(config)
			log.Printf("[ERROR] "+msg.DirCreateError, dir, err)
			os.Exit(1)
		}
	}
}

func generateSummary(config *config.Config, log *log.Logger, logBuffer *[]logger.LogEntry,
	logMutex *sync.RWMutex, debugMode bool, originalFilePath string) error {

	// Find the corresponding transcription file
	basename := strings.TrimSuffix(filepath.Base(originalFilePath), filepath.Ext(originalFilePath))
	transcriptionFile := filepath.Join(config.OutputDir, basename+"."+config.OutputFormat)

	// Check if transcription file exists
	if _, err := os.Stat(transcriptionFile); os.IsNotExist(err) {
		return fmt.Errorf("transcription file not found: %s", transcriptionFile)
	}

	logger.LogProc(log, logBuffer, logMutex, "Generating summary for %s...", basename)

	// Read transcription content
	content, err := readTranscriptionFile(transcriptionFile)
	if err != nil {
		return fmt.Errorf("failed to read transcription file: %w", err)
	}

	// Generate summary using LLM
	summary, err := llm.SummarizeText(config, log, logBuffer, logMutex, debugMode, content)
	if err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	// Save summary to file
	summaryFile := filepath.Join(config.OutputDir, basename+"_summary.txt")
	if err := saveSummaryFile(summaryFile, summary); err != nil {
		return fmt.Errorf("failed to save summary: %w", err)
	}

	logger.LogDone(log, logBuffer, logMutex, "Summary saved to %s", filepath.Base(summaryFile))
	return nil
}

func readTranscriptionFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func saveSummaryFile(filePath, summary string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(summary)
	return err
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	} else {
		return fmt.Sprintf("%ds", seconds)
	}
}

// Phase 1: Helper function for processed files cleanup
func cleanupProcessedFiles(processedFiles *map[string]bool, mu *sync.Mutex, log *log.Logger, logBuffer *[]logger.LogEntry, logMutex *sync.RWMutex) {
	// Keep only the most recent 2500 entries (half of the threshold)
	if len(*processedFiles) <= 2500 {
		return
	}
	
	// Note: This function is called while mu is already locked in ScanAndProcess
	// No additional locking needed here as the caller already holds the lock
	
	// Simple approach: reset the map when it gets too large
	// In a production system, you might want to keep recent entries based on timestamp
	newMap := make(map[string]bool)
	count := 0
	
	// Keep approximately half of the entries
	for file := range *processedFiles {
		if count < 2500 {
			newMap[file] = true
			count++
		}
	}
	
	*processedFiles = newMap
	logger.LogInfo(log, logBuffer, logMutex, "Cleaned up processed files map, kept %d entries", len(*processedFiles))
}
