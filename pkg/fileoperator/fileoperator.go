// Package fileoperator provides abstraction over [os.File].
package fileoperator

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// FileOperator provides simple file usage.
// Read and Write is buffered.
type FileOperator struct {
	r *bufio.Reader
	w *bufio.Writer
	f *os.File
}

// New returns FileOperator pointer.
func New(file *os.File) *FileOperator {
	return &FileOperator{
		r: bufio.NewReader(file),
		w: bufio.NewWriter(file),
		f: file,
	}
}

// Clear erase underlying file and move cursor to zero position.
func (fo *FileOperator) Clear() (err error) {
	const op = "FileOperator.Clear"
	if err = fo.f.Truncate(0); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if _, err = fo.f.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Load full reads underlying file.
func (fo *FileOperator) Load() ([]byte, error) {
	const op = "FileOperator.Load"
	b, err := io.ReadAll(fo.r)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return b, nil
}

// Save clear file and then write all passed data.
func (fo *FileOperator) Save(data []byte) (err error) {
	const op = "FileOperator.Save"
	if err = fo.Clear(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if _, err = fo.w.Write(data); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err = fo.w.Flush(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// Close close underlying file.
func (fo *FileOperator) Close() error {
	const op = "FileOperator.Close"
	if err := fo.f.Close(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
