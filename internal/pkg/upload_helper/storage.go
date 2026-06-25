package upload_helper

import (
	"appliance-recycle/internal/config"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Storage interface {
	Save(fh *multipart.FileHeader) (url string, err error)
}

type LocalStorage struct {
	localPath string
	baseURL   string
}

func NewLocalStorage(cfg config.UploadConfig) *LocalStorage {
	return &LocalStorage{
		localPath: cfg.LocalPath,
		baseURL:   cfg.BaseURL,
	}
}

func (s *LocalStorage) Save(fh *multipart.FileHeader) (string, error) {
	src, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	h := sha256.New()
	if _, err := io.Copy(h, src); err != nil {
		return "", err
	}
	hash := hex.EncodeToString(h.Sum(nil))
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	dateDir := time.Now().Format("20060102")
	dir := filepath.Join(s.localPath, dateDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("%s%s", hash[:16], ext)
	dstPath := filepath.Join(dir, fileName)

	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/%s/%s", s.baseURL, dateDir, fileName)
	return url, nil
}
