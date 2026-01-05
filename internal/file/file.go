package file

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	maxLogSize    = 50 << 20       // 50MB - truncate
	idleThreshold = 24 * time.Hour // delete if idle
)

type File struct {
	in        <-chan []byte
	dir       string
	file      *os.File
	size      int64
	lastWrite time.Time
	path      string
}

func NewFile(dir string, in <-chan []byte) (*File, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create dir: %w", err)
	}

	w := &File{
		in:   in,
		dir:  dir,
		path: filepath.Join(dir, "webhook.log"),
	}

	if err := w.open(); err != nil {
		return nil, err
	}

	return w, nil
}

func (w *File) Run() {
	for data := range w.in {
		w.write(data)
	}
}

func (w *File) Close() error {
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

func (w *File) write(p []byte) {
	now := time.Now()

	if w.size > 0 && (w.size+int64(len(p)) > maxLogSize || now.Sub(w.lastWrite) > idleThreshold) {
		w.truncate()
	}

	n, _ := w.file.Write(p)
	w.size += int64(n)
	w.lastWrite = now
}

func (w *File) truncate() {
	w.file.Close()
	os.Remove(w.path)
	w.open()
}

func (w *File) open() error {
	f, err := os.OpenFile(w.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return fmt.Errorf("stat file: %w", err)
	}

	w.file = f
	w.size = info.Size()
	w.lastWrite = time.Now()
	return nil
}
