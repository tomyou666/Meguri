package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var globalRotating *rotatingWriter

// rotatingWriter はサイズベースのローテーションと flush を行う io.Writer。
type rotatingWriter struct {
	dir      string
	filename string
	maxSize  int64
	maxFiles int
	flush    FlushConfig

	mu               sync.Mutex
	file             *os.File
	size             int64
	writesSinceFlush int
	lastFlush        time.Time
}

func newRotatingWriter(dir, filename string, maxSize int64, maxFiles int, flush FlushConfig) (*rotatingWriter, error) {
	r := &rotatingWriter{
		dir:       dir,
		filename:  filename,
		maxSize:   maxSize,
		maxFiles:  maxFiles,
		flush:     flush,
		lastFlush: time.Now(),
	}
	if err := r.openFile(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *rotatingWriter) activePath() string {
	return filepath.Join(r.dir, r.filename)
}

func (r *rotatingWriter) backupPath(n int) string {
	return r.activePath() + fmt.Sprintf(".%d", n)
}

func (r *rotatingWriter) openFile() error {
	path := r.activePath()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return fmt.Errorf("stat log file: %w", err)
	}
	r.file = f
	r.size = info.Size()
	return nil
}

func (r *rotatingWriter) rotate() error {
	if r.file != nil {
		if err := r.file.Close(); err != nil {
			return fmt.Errorf("close log for rotate: %w", err)
		}
		r.file = nil
	}

	if r.maxFiles > 0 {
		oldest := r.backupPath(r.maxFiles)
		if err := os.Remove(oldest); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove oldest log: %w", err)
		}
		for i := r.maxFiles - 1; i >= 1; i-- {
			src := r.backupPath(i)
			dst := r.backupPath(i + 1)
			if err := os.Rename(src, dst); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("rotate log %d: %w", i, err)
			}
		}
		if err := os.Rename(r.activePath(), r.backupPath(1)); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("rotate active log: %w", err)
		}
	}

	r.size = 0
	return r.openFile()
}

func (r *rotatingWriter) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.maxSize > 0 && r.size+int64(len(p)) > r.maxSize {
		if err := r.rotate(); err != nil {
			return 0, err
		}
	}

	n, err := r.file.Write(p)
	if err != nil {
		return n, err
	}
	r.size += int64(n)
	r.writesSinceFlush++

	if err := r.maybeFlushLocked(); err != nil {
		return n, err
	}
	return n, nil
}

func (r *rotatingWriter) maybeFlushLocked() error {
	switch r.flush.Policy {
	case FlushImmediate:
		return r.syncLocked()
	case FlushInterval:
		if time.Since(r.lastFlush) >= r.flush.Interval {
			return r.syncLocked()
		}
	case FlushEveryN:
		if r.flush.EveryN > 0 && r.writesSinceFlush >= r.flush.EveryN {
			return r.syncLocked()
		}
	}
	return nil
}

func (r *rotatingWriter) syncLocked() error {
	if r.file == nil {
		return nil
	}
	if err := r.file.Sync(); err != nil {
		return err
	}
	r.writesSinceFlush = 0
	r.lastFlush = time.Now()
	return nil
}

func (r *rotatingWriter) flushNow() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.syncLocked()
}

func (r *rotatingWriter) close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.file == nil {
		return nil
	}
	err := r.file.Close()
	r.file = nil
	return err
}
