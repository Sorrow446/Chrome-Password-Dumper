package utils

import (
	"io"
	"os"
	"path/filepath"
)

type LoginData struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func MakeBackup(loginDataPath string) (string, error) {
	tempPath, err := os.MkdirTemp(os.TempDir(), "")
	if err != nil {
		return "", err
	}
	srcFile, err := os.Open(loginDataPath)
	if err != nil {
		return "", err
	}
	defer srcFile.Close()
	backupPath := filepath.Join(tempPath, "tmp")
	backupFile, err := os.Create(backupPath)
	if err != nil {
		return "", err
	}
	defer backupFile.Close()
	_, err = io.Copy(backupFile, srcFile)
	if err != nil {
		return "", err
	}
	return backupPath, nil
}
