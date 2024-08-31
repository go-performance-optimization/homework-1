package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func getRootDirectory(wd string) (string, error) {
	currentDir := wd

	for {
		dirs, err := os.ReadDir(currentDir)

		if err != nil {
			return "", err
		}

		for _, d := range dirs {
			if d.Name() == "cmd" {
				return currentDir, nil
			}
		}

		if currentDir == string(os.PathSeparator) || currentDir == "." {
			return "", fmt.Errorf("invariant error: cmd directory not found")
		}

		currentDir = filepath.Dir(currentDir)
	}
}

func ResolvePath(filename string) (string, error) {
	wd, err := os.Getwd()

	if err != nil {
		return "", err
	}

	root, err := getRootDirectory(wd)

	if err != nil {
		return "", err
	}

	nameWithoutExt := strings.TrimRight(root, filepath.Ext(filename))

	var result string

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		name := d.Name()

		if name == filename || name == nameWithoutExt {
			result = path
			return filepath.SkipAll
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("walk fail tree fail, error: %w", err)
	}

	if result == "" {
		return "", fmt.Errorf("file %s not found in root %s", filename, root)
	}

	return result, nil
}

func GoBuild(ctx context.Context, filepath string, outputPath string) error {
	cmd := exec.CommandContext(ctx, "go", "build", "-o", outputPath, filepath)

	err := cmd.Run()

	if err != nil {
		return err
	}

	return nil
}
