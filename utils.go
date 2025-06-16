package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func (app *App) formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	if minutes >= 60 {
		hours := minutes / 60
		minutes = minutes % 60
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	return fmt.Sprintf("%dm%ds", minutes, seconds)
}

func (app *App) updateFileCounts() {
	app.inputCount = app.countFiles(app.config.InputDir)
	app.outputCount = app.countFiles(app.config.OutputDir)
	app.archiveCount = app.countFiles(app.config.ArchiveDir)
}

func (app *App) countFiles(dir string) int {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return 0
	}

	count := 0
	for _, file := range files {
		if info, err := os.Stat(file); err == nil && !info.IsDir() {
			count++
		}
	}
	return count
}
