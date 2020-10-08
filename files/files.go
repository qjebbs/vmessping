package files

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

// ResolvePath resolves relative path to absolute and convert "~" to home path
func ResolvePath(p string) (string, error) {
	if filepath.HasPrefix(p, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("Cannot resolve path %s: %v", p, err)
		}
		return path.Join(home, p[2:]), nil
	} else if !filepath.IsAbs(p) {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return path.Join(wd, p), nil
	}
	return p, nil
}

// PathsToFiles convert any path (file paths & folder paths) to file paths
func PathsToFiles(paths []string) ([]string, error) {
	files := make([]string, 0)
	for _, p := range paths {
		i, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if i.IsDir() {
			fs, err := GetFolderFiles(p)
			if err != nil {
				return nil, err
			}
			files = append(files, fs...)
			continue
		}
		files = append(files, p)
	}
	return files, nil
}

// GetFolderFiles get files in the folder and it's children
func GetFolderFiles(folder string) ([]string, error) {
	var files []string
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		ext := filepath.Ext(path)
		if ext == ".json" || ext == ".jsonc" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil

}
