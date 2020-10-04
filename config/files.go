package config

import (
	"os"
	"path/filepath"
)

func pathsToFiles(paths []string) ([]string, error) {
	files := make([]string, 0)
	for _, p := range paths {
		i, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if i.IsDir() {
			fs, err := getFolderFiles(p)
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

func getFolderFiles(folder string) ([]string, error) {
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
