package vcsutil

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// FindProjectDirectory will try to find the project directory based on repository folders (.git)
func FindProjectDirectory() (string, error) {
	currentDirectory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	directoryParts := filepath.SplitList(currentDirectory)
	for {
		// check: current directory is a git repository
		if _, err := os.Stat(filepath.Join(currentDirectory, ".git")); err == nil {
			return currentDirectory, nil
		}

		// cancel at root path
		if directoryParts[0]+"\\" == currentDirectory || currentDirectory == "/" {
			return "", errors.New("didn't find any repositories for the current working directory")
		}

		// cancel when we reach the root directory, no repository found
		if isRootPath(currentDirectory) {
			break
		}

		// check parent directory in next iteration
		currentDirectory = filepath.Dir(currentDirectory)
	}

	return "", errors.New("didn't find any repositories for the current working directory")
}

// isRootPath checks if the path is the root directory
func isRootPath(path string) bool {
	if os.PathSeparator == '/' {
		return path == "/"
	} else if os.PathSeparator == '\\' {
		return filepath.VolumeName(path)+`\` == path
	}

	return false
}
