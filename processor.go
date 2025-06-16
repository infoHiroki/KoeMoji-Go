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
	msg := app.getMessages()
	app.logInfo(msg.ScanningDir)

	files, err := filepath.Glob(filepath.Join(app.config.InputDir, "*"))
	if err != nil {
		app.logError("Failed to scan input directory: %v", err)
		return
	}

	newFiles := app.filterNewAudioFiles(files)
	if len(newFiles) == 0 {
		app.logDebug("No new files found")
		return
	}

	app.logInfo(msg.FoundFiles, len(newFiles))

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
		if isAudioFile(file) && !app.processedFiles[file] {
			app.processedFiles[file] = true
			newFiles = append(newFiles, file)
		}
	}
	return newFiles
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
		msg := app.getMessages()
		app.logProc(msg.ProcessingFile, app.processingFile)
		startTime := time.Now()

		if err := app.transcribeAudio(filePath); err != nil {
			app.logError(msg.ProcessFailed, app.processingFile, err)
		} else {
			duration := time.Since(startTime)
			app.logDone(msg.ProcessComplete, app.processingFile, app.formatDuration(duration))

			// Move to archive
			msg2 := app.getMessages()
			app.logProc(msg2.MovingToArchive, app.processingFile)
			if err := app.moveToArchive(filePath); err != nil {
				app.logError(msg.ProcessFailed, app.processingFile, err)
			}
		}
	}
}

func (app *App) moveToArchive(sourcePath string) error {
	filename := filepath.Base(sourcePath)
	destPath := filepath.Join(app.config.ArchiveDir, filename)

	// Handle duplicate filenames
	if _, err := os.Stat(destPath); err == nil {
		timestamp := time.Now().Format("20060102_150405")
		ext := filepath.Ext(filename)
		name := strings.TrimSuffix(filename, ext)
		destPath = filepath.Join(app.config.ArchiveDir, fmt.Sprintf("%s_%s%s", name, timestamp, ext))
	}

	if err := os.Rename(sourcePath, destPath); err != nil {
		return err
	}
	return nil
}

func (app *App) ensureDirectories() {
	dirs := []string{app.config.InputDir, app.config.OutputDir, app.config.ArchiveDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			msg := app.getMessages()
			app.logError(msg.DirCreateError, dir, err)
			os.Exit(1)
		}
	}
}
