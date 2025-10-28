package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// FileInfo represents a file with its metadata
type FileInfo struct {
	Name    string
	Path    string
	Size    int64
	ModTime time.Time
	IsDir   bool
}

// CreateFileList creates a tview.List populated with files from the specified directory
func CreateFileList(dir string, app *tview.Application) (*tview.List, error) {
	list := tview.NewList().ShowSecondaryText(false)

	// Read directory
	files, err := ReadDirFiles(dir)
	if err != nil {
		return nil, err
	}

	// Add files to list
	for _, file := range files {
		displayText := fmt.Sprintf("%-40s %10s", file.Name, FormatFileSize(file.Size))
		list.AddItem(displayText, "", 0, nil)
	}

	// Add summary at the bottom
	totalSize := int64(0)
	for _, file := range files {
		totalSize += file.Size
	}
	summaryText := fmt.Sprintf("\n合計: %d ファイル (%s)", len(files), FormatFileSize(totalSize))
	list.AddItem(summaryText, "", 0, nil)

	// Add help text
	helpText := "\n[Enter: フォルダを開く]  [r: 再読み込み]"
	list.AddItem(helpText, "", 0, nil)

	// Handle Enter key to open folder
	list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		// Ignore if summary or help item
		if index >= len(files) {
			return
		}
		// Open folder
		if err := OpenFolder(dir); err != nil {
			// TODO: Show error dialog
			fmt.Printf("フォルダを開けません: %v\n", err)
		}
	})

	// Handle 'r' key to refresh
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'r' || event.Rune() == 'R' {
			// Refresh list
			newList, err := CreateFileList(dir, app)
			if err == nil {
				list.Clear()
				for i := 0; i < newList.GetItemCount(); i++ {
					mainText, secondaryText := newList.GetItemText(i)
					list.AddItem(mainText, secondaryText, 0, nil)
				}
			}
			return nil
		}
		return event
	})

	return list, nil
}

// ReadDirFiles reads all files from a directory and returns FileInfo slice
func ReadDirFiles(dir string) ([]FileInfo, error) {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("ディレクトリが存在しません: %s", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		// Skip directories
		if entry.IsDir() {
			continue
		}

		// Skip hidden files (starting with .)
		if len(entry.Name()) > 0 && entry.Name()[0] == '.' {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, FileInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(dir, entry.Name()),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   entry.IsDir(),
		})
	}

	// Sort by modification time (newest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})

	return files, nil
}

// FormatFileSize formats a file size in bytes to a human-readable string
func FormatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// OpenFolder opens a folder in the system's default file manager
func OpenFolder(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("explorer", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	default:
		return fmt.Errorf("サポートされていないOS: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// GetFileListTitle returns the title for a file list based on directory type
func GetFileListTitle(dirType, dirPath string) string {
	switch dirType {
	case "input":
		return fmt.Sprintf(" 入力フォルダ (%s) ", dirPath)
	case "output":
		return fmt.Sprintf(" 出力フォルダ (%s) ", dirPath)
	case "archive":
		return fmt.Sprintf(" アーカイブフォルダ (%s) ", dirPath)
	default:
		return fmt.Sprintf(" フォルダ (%s) ", dirPath)
	}
}
