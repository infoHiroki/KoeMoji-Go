package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (app *App) startProcessing() {
	// Initial scan
	app.scanAndProcess()

	// Periodic scan
	ticker := time.NewTicker(time.Duration(app.config.ScanIntervalMinutes) * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		app.scanAndProcess()
	}
}

func (app *App) scanAndProcess() {
	app.lastScanTime = time.Now()
	app.logInfo("Scanning for new files...")

	files, err := filepath.Glob("input/*")
	if err != nil {
		app.logError("Failed to scan input directory: %v", err)
		return
	}

	newFiles := app.filterNewAudioFiles(files)
	if len(newFiles) == 0 {
		app.logDebug("No new files found")
		return
	}

	app.logInfo("Found %d new file(s) to process", len(newFiles))

	// Add files to queue
	app.mu.Lock()
	app.queuedFiles = append(app.queuedFiles, newFiles...)
	app.mu.Unlock()

	// Start processing if not already processing
	if !app.isProcessing {
		app.wg.Add(1)
		go app.processQueue()
	}
}

func (app *App) filterNewAudioFiles(files []string) []string {
	app.mu.Lock()
	defer app.mu.Unlock()

	var newFiles []string
	for _, file := range files {
		if app.isAudioFile(file) && !app.processedFiles[file] {
			app.processedFiles[file] = true
			newFiles = append(newFiles, file)
		}
	}
	return newFiles
}

func (app *App) isAudioFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	audioExts := []string{".mp3", ".wav", ".m4a", ".flac", ".ogg", ".aac", ".mp4", ".mov", ".avi"}
	for _, audioExt := range audioExts {
		if ext == audioExt {
			return true
		}
	}
	return false
}

func (app *App) processQueue() {
	defer app.wg.Done()

	for {
		app.mu.Lock()
		if len(app.queuedFiles) == 0 {
			app.isProcessing = false
			app.processingFile = ""
			app.mu.Unlock()
			return
		}

		// Get next file from queue
		filePath := app.queuedFiles[0]
		app.queuedFiles = app.queuedFiles[1:]
		app.processingFile = filepath.Base(filePath)
		app.isProcessing = true
		app.mu.Unlock()

		// Process the file
		app.logProc("Processing: %s", app.processingFile)
		startTime := time.Now()

		if err := app.transcribeAudio(filePath); err != nil {
			app.logError("Failed to process %s: %v", app.processingFile, err)
		} else {
			duration := time.Since(startTime)
			app.logDone("Completed: %s (%s)", app.processingFile, app.formatDuration(duration))
			app.totalProcessed++

			// Move to archive
			if err := app.moveToArchive(filePath); err != nil {
				app.logError("Failed to archive %s: %v", app.processingFile, err)
			}
		}
	}
}

func (app *App) moveToArchive(sourcePath string) error {
	filename := filepath.Base(sourcePath)
	destPath := filepath.Join("archive", filename)

	// Handle duplicate filenames
	if _, err := os.Stat(destPath); err == nil {
		timestamp := time.Now().Format("20060102_150405")
		ext := filepath.Ext(filename)
		name := strings.TrimSuffix(filename, ext)
		destPath = filepath.Join("archive", fmt.Sprintf("%s_%s%s", name, timestamp, ext))
	}

	if err := os.Rename(sourcePath, destPath); err != nil {
		return err
	}
	return nil
}

func (app *App) ensureDirectories() {
	dirs := []string{"input", "output", "archive"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			app.logError("Failed to create directory %s: %v", dir, err)
			os.Exit(1)
		}
	}
}