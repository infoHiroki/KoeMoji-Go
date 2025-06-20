package processor

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/hirokitakamura/koemoji-go/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestEnsureDirectories(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.GetDefaultConfig()
	cfg.InputDir = filepath.Join(tempDir, "input")
	cfg.OutputDir = filepath.Join(tempDir, "output")
	cfg.ArchiveDir = filepath.Join(tempDir, "archive")

	logger := log.New(os.Stdout, "", log.LstdFlags)

	EnsureDirectories(cfg, logger)

	assert.DirExists(t, cfg.InputDir)
	assert.DirExists(t, cfg.OutputDir)
	assert.DirExists(t, cfg.ArchiveDir)
}

func TestEnsureDirectories_AlreadyExists(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.GetDefaultConfig()
	cfg.InputDir = filepath.Join(tempDir, "input")
	cfg.OutputDir = filepath.Join(tempDir, "output")
	cfg.ArchiveDir = filepath.Join(tempDir, "archive")

	// Pre-create directories
	err := os.MkdirAll(cfg.InputDir, 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(cfg.OutputDir, 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(cfg.ArchiveDir, 0755)
	assert.NoError(t, err)

	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Should not fail if directories already exist
	EnsureDirectories(cfg, logger)

	assert.DirExists(t, cfg.InputDir)
	assert.DirExists(t, cfg.OutputDir)
	assert.DirExists(t, cfg.ArchiveDir)
}

// Note: ScanAndProcess and StartProcessing require complex setup with logger buffers
// and multiple concurrent variables. These are better tested through integration tests
// or by testing individual components separately.
// 
// The core business logic in these functions is handled by unexported helper functions
// which would need to be made exported to be unit tested, or tested indirectly.