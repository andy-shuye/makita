package file

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	defaultLocalStorageBaseDir  = "/data/files"
	fallbackLocalStorageBaseDir = "./data/files"
)

// ResolveLocalBaseDir returns a writable local storage base directory.
// Priority:
//  1. explicit baseDir argument
//  2. LOCAL_STORAGE_BASE_DIR environment variable
//  3. defaultLocalStorageBaseDir (/data/files)
//
// If the implicit default directory is not writable (common in non-container runs),
// it falls back to ./data/files.
func ResolveLocalBaseDir(baseDir string) string {
	explicit := normalizeBaseDirInput(strings.TrimSpace(baseDir))
	if explicit == "" {
		explicit = normalizeBaseDirInput(strings.TrimSpace(os.Getenv("LOCAL_STORAGE_BASE_DIR")))
	}
	isImplicitDefault := explicit == ""
	if explicit == "" {
		explicit = defaultLocalStorageBaseDir
	}

	if ensureWritableDir(explicit) == nil {
		return explicit
	}

	if isImplicitDefault && explicit == defaultLocalStorageBaseDir {
		if ensureWritableDir(fallbackLocalStorageBaseDir) == nil {
			return fallbackLocalStorageBaseDir
		}
	}

	return explicit
}

// normalizeBaseDirInput converts common Windows drive paths (e.g. D:\weknora\data)
// to WSL mount paths (/mnt/d/weknora/data) when running on non-Windows systems.
func normalizeBaseDirInput(dir string) string {
	if dir == "" || runtime.GOOS == "windows" {
		return dir
	}
	if len(dir) >= 3 && dir[1] == ':' && (dir[2] == '\\' || dir[2] == '/') {
		drive := strings.ToLower(string(dir[0]))
		rest := strings.ReplaceAll(dir[2:], "\\", "/")
		rest = strings.TrimLeft(rest, "/")
		return filepath.Clean("/mnt/" + drive + "/" + rest)
	}
	return dir
}

func ensureWritableDir(dir string) error {
	absDir, err := filepath.Abs(strings.TrimSpace(dir))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp(absDir, ".weknora-write-test-*")
	if err != nil {
		return err
	}
	name := f.Name()
	_ = f.Close()
	_ = os.Remove(name)
	return nil
}
