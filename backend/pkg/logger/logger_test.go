package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// InitApp のコンソール・ファイル出力を検証する。
func TestInitApp(t *testing.T) {
	dir := t.TempDir()
	t.Cleanup(func() { _ = Shutdown() })

	t.Run("コンソールとファイルへ同時出力される", func(t *testing.T) {
		var buf strings.Builder
		if err := InitApp(AppConfig{
			Console: &buf,
			Level:   slog.LevelInfo,
			FileDir: dir,
			Flush:   FlushConfig{Policy: FlushImmediate},
		}); err != nil {
			t.Fatal(err)
		}
		slog.Info("init app smoke")
		if !strings.Contains(buf.String(), "init app smoke") {
			t.Fatalf("console missing log: %q", buf.String())
		}
		data, err := os.ReadFile(filepath.Join(dir, defaultLogFilename))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "init app smoke") {
			t.Fatalf("file missing log: %q", data)
		}
	})

	t.Run("Console省略時はファイルのみ", func(t *testing.T) {
		sub := filepath.Join(dir, "fileonly")
		if err := os.Mkdir(sub, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := InitApp(AppConfig{
			Level:   slog.LevelInfo,
			FileDir: sub,
			Flush:   FlushConfig{Policy: FlushImmediate},
		}); err != nil {
			t.Fatal(err)
		}
		slog.Warn("file only")
		data, err := os.ReadFile(filepath.Join(sub, defaultLogFilename))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "file only") {
			t.Fatalf("file missing log: %q", data)
		}
	})
}

func TestInitConsole(t *testing.T) {
	var w strings.Builder
	InitConsole(&w, slog.LevelInfo)
	slog.Info("console")
	if !strings.Contains(w.String(), "console") {
		t.Fatalf("got %q", w.String())
	}
}

func TestFlush_global(t *testing.T) {
	dir := t.TempDir()
	t.Cleanup(func() { _ = Shutdown() })
	if err := InitApp(AppConfig{
		Level:   slog.LevelInfo,
		FileDir: dir,
		Flush: FlushConfig{
			Policy:   FlushInterval,
			Interval: time.Hour,
		},
	}); err != nil {
		t.Fatal(err)
	}
	slog.Info("defer flush")
	if err := Flush(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, defaultLogFilename))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "defer flush") {
		t.Fatalf("flush did not persist: %q", data)
	}
}

// rotatingWriter のローテーションと flush 挙動を検証する。
func TestRotatingWriter(t *testing.T) {
	dir := t.TempDir()

	t.Run("FlushImmediateでWrite直後にディスクへ反映される", func(t *testing.T) {
		rw, err := newRotatingWriter(dir, "test.log", 1024, 3, FlushConfig{Policy: FlushImmediate})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := rw.Write([]byte("hello\n")); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(dir, "test.log"))
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "hello\n" {
			t.Fatalf("got %q", data)
		}
		_ = rw.close()
	})

	t.Run("MaxSize超過でバックアップが生成される", func(t *testing.T) {
		sub := filepath.Join(dir, "rotate")
		if err := os.Mkdir(sub, 0o755); err != nil {
			t.Fatal(err)
		}
		rw, err := newRotatingWriter(sub, "size.log", 8, 2, FlushConfig{Policy: FlushImmediate})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := rw.Write([]byte("12345678")); err != nil {
			t.Fatal(err)
		}
		if _, err := rw.Write([]byte("9")); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(sub, "size.log.1")); err != nil {
			t.Fatalf("backup missing: %v", err)
		}
		_ = rw.close()
	})

	t.Run("FlushIntervalで経過後にSyncされる", func(t *testing.T) {
		sub := filepath.Join(dir, "interval")
		if err := os.Mkdir(sub, 0o755); err != nil {
			t.Fatal(err)
		}
		rw, err := newRotatingWriter(sub, "interval.log", 1024, 2, FlushConfig{
			Policy:   FlushInterval,
			Interval: 50 * time.Millisecond,
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := rw.Write([]byte("tick\n")); err != nil {
			t.Fatal(err)
		}
		time.Sleep(60 * time.Millisecond)
		if _, err := rw.Write([]byte("tock\n")); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(sub, "interval.log"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "tick") || !strings.Contains(string(data), "tock") {
			t.Fatalf("unexpected content: %q", data)
		}
		_ = rw.close()
	})

	t.Run("flushNowで未同期分が反映される", func(t *testing.T) {
		sub := filepath.Join(dir, "manual")
		if err := os.Mkdir(sub, 0o755); err != nil {
			t.Fatal(err)
		}
		rw, err := newRotatingWriter(sub, "manual.log", 1024, 2, FlushConfig{
			Policy:   FlushInterval,
			Interval: time.Hour,
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := rw.Write([]byte("pending\n")); err != nil {
			t.Fatal(err)
		}
		if err := rw.flushNow(); err != nil {
			t.Fatal(err)
		}
		_ = rw.close()
		data, err := os.ReadFile(filepath.Join(sub, "manual.log"))
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "pending\n" {
			t.Fatalf("got %q", data)
		}
	})
}

func TestMultiWriter_order(t *testing.T) {
	var a, b strings.Builder
	mw := io.MultiWriter(&a, &b)
	if _, err := mw.Write([]byte("x")); err != nil {
		t.Fatal(err)
	}
	if a.String() != "x" || b.String() != "x" {
		t.Fatal("multi writer failed")
	}
}
