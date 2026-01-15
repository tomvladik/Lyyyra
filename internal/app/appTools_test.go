package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUnzip(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test_unzip_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test zip file
	zipPath := filepath.Join(tmpDir, "test.zip")
	extractPath := filepath.Join(tmpDir, "extracted")

	// Create some test files to zip
	testDataDir := filepath.Join(tmpDir, "testdata")
	if err := os.MkdirAll(testDataDir, os.ModePerm); err != nil {
		t.Fatalf("Failed to create test data dir: %v", err)
	}

	testFile := filepath.Join(testDataDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a simple zip file (we'll skip actual zip creation for simplicity)
	// In a real test, you'd create a proper zip file here

	tests := []struct {
		name      string
		zipFile   string
		dest      string
		wantError bool
	}{
		{
			name:      "non-existent zip file",
			zipFile:   filepath.Join(tmpDir, "nonexistent.zip"),
			dest:      extractPath,
			wantError: true,
		},
		{
			name:      "invalid destination",
			zipFile:   zipPath,
			dest:      "/invalid/path",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := unzip(tt.zipFile, tt.dest)
			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_copy_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name      string
		setup     func() (string, string)
		wantError bool
	}{
		{
			name: "successful file copy",
			setup: func() (string, string) {
				src := filepath.Join(tmpDir, "source.txt")
				dst := filepath.Join(tmpDir, "dest.txt")
				_ = os.WriteFile(src, []byte("test content"), 0644)
				return src, dst
			},
			wantError: false,
		},
		{
			name: "non-existent source file",
			setup: func() (string, string) {
				src := filepath.Join(tmpDir, "nonexistent.txt")
				dst := filepath.Join(tmpDir, "dest2.txt")
				return src, dst
			},
			wantError: true,
		},
		{
			name: "invalid destination path",
			setup: func() (string, string) {
				src := filepath.Join(tmpDir, "source2.txt")
				_ = os.WriteFile(src, []byte("test"), 0644)
				dst := "/invalid/path/dest.txt"
				return src, dst
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, dst := tt.setup()
			err := copyFile(src, dst)

			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantError && err == nil {
				// Verify file was copied
				srcData, _ := os.ReadFile(src)
				dstData, _ := os.ReadFile(dst)
				if string(srcData) != string(dstData) {
					t.Error("copied file content doesn't match source")
				}
			}
		})
	}
}

func TestCopyDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_copydir_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name      string
		setup     func() (string, string)
		verify    func(t *testing.T, dst string)
		wantError bool
	}{
		{
			name: "successful directory copy",
			setup: func() (string, string) {
				src := filepath.Join(tmpDir, "srcdir")
				dst := filepath.Join(tmpDir, "dstdir")
				_ = os.MkdirAll(filepath.Join(src, "subdir"), os.ModePerm)
				_ = os.WriteFile(filepath.Join(src, "file1.txt"), []byte("content1"), 0644)
				_ = os.WriteFile(filepath.Join(src, "subdir", "file2.txt"), []byte("content2"), 0644)
				return src, dst
			},
			verify: func(t *testing.T, dst string) {
				if _, err := os.Stat(filepath.Join(dst, "file1.txt")); os.IsNotExist(err) {
					t.Error("file1.txt was not copied")
				}
				if _, err := os.Stat(filepath.Join(dst, "subdir", "file2.txt")); os.IsNotExist(err) {
					t.Error("subdir/file2.txt was not copied")
				}
			},
			wantError: false,
		},
		{
			name: "non-existent source directory",
			setup: func() (string, string) {
				src := filepath.Join(tmpDir, "nonexistent")
				dst := filepath.Join(tmpDir, "dstdir2")
				return src, dst
			},
			verify:    func(t *testing.T, dst string) {},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, dst := tt.setup()
			err := copyDir(src, dst)

			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantError && err == nil {
				tt.verify(t, dst)
			}
		})
	}
}
