package processor

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hirokitakamura/koemoji-go/internal/config"
	"github.com/hirokitakamura/koemoji-go/internal/logger"
	"github.com/hirokitakamura/koemoji-go/internal/ui"
	"github.com/hirokitakamura/koemoji-go/internal/whisper"
)

func StartProcessing(config *config.Config, log *log.Logger, logBuffer *[]logger.LogEntry, logMutex *sync.RWMutex,
	lastScanTime *time.Time, queuedFiles *[]string, processingFile *string, isProcessing *bool,
	processedFiles *map[string]bool, mu *sync.Mutex, wg *sync.WaitGroup, debugMode bool) {
	
	// Initial scan
	ScanAndProcess(config, log, logBuffer, logMutex, lastScanTime, queuedFiles, processingFile, 
		isProcessing, processedFiles, mu, wg, debugMode)

	// Periodic scan
	ticker := time.NewTicker(time.Duration(config.ScanIntervalMinutes) * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ScanAndProcess(config, log, logBuffer, logMutex, lastScanTime, queuedFiles, processingFile, 
			isProcessing, processedFiles, mu, wg, debugMode)
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
	mu.Unlock()

	// Start processing if not already processing
	if !*isProcessing {
		wg.Add(1)
		go processQueue(config, log, logBuffer, logMutex, queuedFiles, processingFile, 
			isProcessing, mu, wg, debugMode)
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
	
	defer wg.Done()

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