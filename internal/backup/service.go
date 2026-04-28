package backup

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
)

type Service struct {
	dbPath     string
	backupDir  string
	logger     zerolog.Logger
	maxBackups int
	now        func() time.Time
}

func NewService(dbPath, backupDir string, logger zerolog.Logger, now func() time.Time) *Service {
	return &Service{dbPath: dbPath, backupDir: backupDir, logger: logger, maxBackups: 30, now: now}
}

func (s *Service) Backup(ctx context.Context) error {
	var err error
	if ctx == nil {
		err = errors.New("context is required")
	}
	if err == nil && ctx.Err() != nil {
		err = ctx.Err()
	}
	if err == nil {
		err = s.ensureDir()
	}
	if err == nil {
		err = s.withLock(func() error {
			return s.copyDB()
		})
	}
	return err
}

func (s *Service) StartCleanup(ctx context.Context, interval time.Duration) {
	if ctx == nil {
		ctx = context.Background()
	}
	if interval <= 0 {
		interval = time.Hour
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = s.Cleanup(context.Background())
			}
		}
	}()
}

func (s *Service) Cleanup(ctx context.Context) error {
	var err error
	if ctx == nil {
		err = errors.New("context is required")
	}
	if err == nil && ctx.Err() != nil {
		err = ctx.Err()
	}
	if err == nil {
		err = s.ensureDir()
	}
	if err == nil {
		err = s.withLock(func() error {
			return s.removeOldBackups()
		})
	}
	return err
}

func (s *Service) ensureDir() error {
	var err error
	if s.backupDir == "" {
		err = errors.New("backup dir is required")
	}
	if err == nil {
		err = os.MkdirAll(s.backupDir, 0o755)
	}
	return err
}

func (s *Service) copyDB() error {
	var (
		err  error
		src  *os.File
		dst  *os.File
		path string
	)
	path = filepath.Join(s.backupDir, s.fileName())
	src, err = os.Open(s.dbPath)
	if err == nil {
		defer src.Close()
		dst, err = os.Create(path)
	}
	if err == nil {
		defer dst.Close()
		_, err = io.Copy(dst, src)
	}
	if err == nil {
		err = dst.Sync()
	}
	if err == nil {
		s.logger.Info().Str("backup", path).Msg("backup created")
	}
	if err != nil {
		s.logger.Error().Err(err).Msg("backup failed")
	}
	return err
}

func (s *Service) fileName() string {
	return "ledger-" + s.now().Format("20060102-150405") + ".db"
}

func (s *Service) lockPath() string {
	return filepath.Join(s.backupDir, ".lock")
}

func (s *Service) withLock(fn func() error) error {
	var (
		file *os.File
		err  error
	)
	file, err = os.OpenFile(s.lockPath(), os.O_CREATE|os.O_RDWR, 0o600)
	if err == nil {
		defer file.Close()
		err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
	}
	if err == nil {
		defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		err = fn()
	}
	return err
}

func (s *Service) removeOldBackups() error {
	var (
		files []os.DirEntry
		err   error
	)
	files, err = os.ReadDir(s.backupDir)
	if err == nil {
		err = s.deleteExcess(files)
	}
	return err
}

func (s *Service) deleteExcess(entries []os.DirEntry) error {
	var (
		paths []string
		err   error
	)
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".db") {
			paths = append(paths, filepath.Join(s.backupDir, entry.Name()))
		}
	}
	if len(paths) > s.maxBackups {
		sort.Slice(paths, func(i, j int) bool {
			return fileModTime(paths[i]).Before(fileModTime(paths[j]))
		})
		err = s.removeOld(paths[:len(paths)-s.maxBackups])
	}
	return err
}

func (s *Service) removeOld(paths []string) error {
	var err error
	for _, path := range paths {
		err = os.Remove(path)
		if err != nil {
			break
		}
	}
	return err
}

func fileModTime(path string) time.Time {
	var (
		info os.FileInfo
		err  error
		out  time.Time
	)
	info, err = os.Stat(path)
	if err == nil {
		out = info.ModTime()
	}
	return out
}
