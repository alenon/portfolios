package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitHomeDir(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		homePath  string
		wantError bool
	}{
		{
			name:      "Create home directory with custom path",
			homePath:  filepath.Join(tmpDir, "test-portfolios"),
			wantError: false,
		},
		{
			name:      "Create home directory with empty path (uses default)",
			homePath:  "",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home, err := InitHomeDir(tt.homePath)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify home directory was created
			if !home.Exists() {
				t.Error("Home directory was not created")
			}

			// Verify logs directory was created
			if info, err := os.Stat(home.LogsDir); err != nil || !info.IsDir() {
				t.Errorf("Logs directory was not created: %v", err)
			}

			// Verify paths are set correctly
			if home.Root == "" {
				t.Error("Root path is empty")
			}
			if home.ConfigPath == "" {
				t.Error("Config path is empty")
			}
			if home.LogsDir == "" {
				t.Error("Logs directory path is empty")
			}
			if home.ServerLog == "" {
				t.Error("Server log path is empty")
			}
			if home.RequestLog == "" {
				t.Error("Request log path is empty")
			}

			// Verify config path is under root using filepath.Rel
			relConfig, err := filepath.Rel(home.Root, home.ConfigPath)
			if err != nil || strings.HasPrefix(relConfig, "..") || filepath.IsAbs(relConfig) {
				t.Error("Config path is not under root directory")
			}

			// Verify logs directory is under root using filepath.Rel
			relLogs, err := filepath.Rel(home.Root, home.LogsDir)
			if err != nil || strings.HasPrefix(relLogs, "..") || filepath.IsAbs(relLogs) {
				t.Error("Logs directory is not under root directory")
			}
		})
	}
}

func TestHomeDir_Exists(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		setupFunc  func() *HomeDir
		wantExists bool
	}{
		{
			name: "Home directory exists",
			setupFunc: func() *HomeDir {
				home, _ := InitHomeDir(filepath.Join(tmpDir, "existing"))
				return home
			},
			wantExists: true,
		},
		{
			name: "Home directory does not exist",
			setupFunc: func() *HomeDir {
				return &HomeDir{
					Root: filepath.Join(tmpDir, "nonexistent"),
				}
			},
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := tt.setupFunc()
			exists := home.Exists()

			if exists != tt.wantExists {
				t.Errorf("Exists() = %v, want %v", exists, tt.wantExists)
			}
		})
	}
}

func TestHomeDir_ConfigExists(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		setupFunc  func() *HomeDir
		wantExists bool
	}{
		{
			name: "Config file exists",
			setupFunc: func() *HomeDir {
				home, _ := InitHomeDir(filepath.Join(tmpDir, "with-config"))
				// Create config file
				_ = os.WriteFile(home.ConfigPath, []byte("test"), 0600)
				return home
			},
			wantExists: true,
		},
		{
			name: "Config file does not exist",
			setupFunc: func() *HomeDir {
				home, _ := InitHomeDir(filepath.Join(tmpDir, "no-config"))
				return home
			},
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := tt.setupFunc()
			exists := home.ConfigExists()

			if exists != tt.wantExists {
				t.Errorf("ConfigExists() = %v, want %v", exists, tt.wantExists)
			}
		})
	}
}

func TestHomeDir_EnsureLogFiles(t *testing.T) {
	tmpDir := t.TempDir()
	home, err := InitHomeDir(filepath.Join(tmpDir, "test-logs"))
	if err != nil {
		t.Fatalf("Failed to initialize home directory: %v", err)
	}

	// Ensure log files are created
	if err := home.EnsureLogFiles(); err != nil {
		t.Fatalf("EnsureLogFiles() error = %v", err)
	}

	// Verify server log file exists
	if info, err := os.Stat(home.ServerLog); err != nil || info.IsDir() {
		t.Errorf("Server log file was not created: %v", err)
	}

	// Verify request log file exists
	if info, err := os.Stat(home.RequestLog); err != nil || info.IsDir() {
		t.Errorf("Request log file was not created: %v", err)
	}

	// Test calling EnsureLogFiles again (should not error)
	if err := home.EnsureLogFiles(); err != nil {
		t.Errorf("EnsureLogFiles() second call error = %v", err)
	}
}

func TestGetDefaultHomePath(t *testing.T) {
	path, err := GetDefaultHomePath()
	if err != nil {
		t.Fatalf("GetDefaultHomePath() error = %v", err)
	}

	if path == "" {
		t.Error("GetDefaultHomePath() returned empty path")
	}

	// Verify path ends with .portfolios
	if filepath.Base(path) != DefaultHomeDirName {
		t.Errorf("GetDefaultHomePath() path does not end with %s, got %s", DefaultHomeDirName, filepath.Base(path))
	}
}

func TestHomeDir_createDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	home := &HomeDir{
		Root:    filepath.Join(tmpDir, "test-create"),
		LogsDir: filepath.Join(tmpDir, "test-create", "logs"),
	}

	// Create directories
	if err := home.createDirectories(); err != nil {
		t.Fatalf("createDirectories() error = %v", err)
	}

	// Verify root directory exists
	if info, err := os.Stat(home.Root); err != nil || !info.IsDir() {
		t.Errorf("Root directory was not created: %v", err)
	}

	// Verify logs directory exists
	if info, err := os.Stat(home.LogsDir); err != nil || !info.IsDir() {
		t.Errorf("Logs directory was not created: %v", err)
	}
}

func TestHomeDir_ensureFile(t *testing.T) {
	tmpDir := t.TempDir()
	home, _ := InitHomeDir(filepath.Join(tmpDir, "test-ensure"))

	testFile := filepath.Join(home.Root, "test.log")

	// First call should create the file
	if err := home.ensureFile(testFile); err != nil {
		t.Fatalf("ensureFile() error = %v", err)
	}

	// Verify file exists
	if info, err := os.Stat(testFile); err != nil || info.IsDir() {
		t.Errorf("File was not created: %v", err)
	}

	// Second call should not error
	if err := home.ensureFile(testFile); err != nil {
		t.Errorf("ensureFile() second call error = %v", err)
	}
}
