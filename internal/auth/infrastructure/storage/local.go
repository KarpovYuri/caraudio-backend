package storage

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type LocalStorage struct {
	rootDir       string
	publicPath    string
	publicBaseURL string
}

func NewLocalStorage(rootDir, publicPath, publicBaseURL string) (*LocalStorage, error) {
	rootDir = filepath.Clean(rootDir)
	publicPath = "/" + strings.Trim(publicPath, "/")
	publicBaseURL = strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	if err := os.MkdirAll(filepath.Join(rootDir, "avatars"), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create uploads dir: %w", err)
	}
	return &LocalStorage{
		rootDir:       rootDir,
		publicPath:    publicPath,
		publicBaseURL: publicBaseURL,
	}, nil
}

func (s *LocalStorage) RootDir() string {
	return s.rootDir
}

func (s *LocalStorage) PublicPath() string {
	return s.publicPath
}

func (s *LocalStorage) SaveUserAvatar(ext string, r io.Reader) (string, error) {
	ext = normalizeExt(ext)
	if ext == "" {
		return "", fmt.Errorf("unsupported image extension")
	}

	name := uuid.NewString() + ext
	rel := filepath.Join("avatars", name)
	abs := filepath.Join(s.rootDir, rel)

	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", fmt.Errorf("failed to create avatar uploads dir: %w", err)
	}

	f, err := os.OpenFile(abs, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return "", fmt.Errorf("failed to create avatar file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		_ = os.Remove(abs)
		return "", fmt.Errorf("failed to write avatar file: %w", err)
	}

	publicURL := path.Join(s.publicPath, "avatars", name)
	if s.publicBaseURL != "" {
		return s.publicBaseURL + publicURL, nil
	}
	return publicURL, nil
}

func (s *LocalStorage) DeletePublicURL(publicURL string) error {
	rel, ok := s.relativePath(publicURL)
	if !ok {
		return nil
	}
	abs := filepath.Join(s.rootDir, filepath.FromSlash(rel))
	if err := os.Remove(abs); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *LocalStorage) relativePath(publicURL string) (string, bool) {
	publicURL = strings.TrimSpace(publicURL)
	if publicURL == "" {
		return "", false
	}
	if idx := strings.Index(publicURL, s.publicPath+"/"); idx >= 0 {
		publicURL = publicURL[idx:]
	}
	prefix := s.publicPath + "/"
	if !strings.HasPrefix(publicURL, prefix) {
		return "", false
	}
	rel := strings.TrimPrefix(publicURL, prefix)
	if rel == "" || strings.Contains(rel, "..") {
		return "", false
	}
	return rel, true
}

func normalizeExt(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if ext == "" {
		return ""
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	switch ext {
	case ".webp", ".svg":
		return ext
	default:
		return ""
	}
}
