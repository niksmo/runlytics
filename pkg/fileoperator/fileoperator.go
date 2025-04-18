// Package fileoperator provides abstraction over [os.File].
package fileoperator

import (
	"bufio"
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
	if err = fo.f.Truncate(0); err != nil {
		return err
	}
	_, err = fo.f.Seek(0, io.SeekStart)
	return err
}

// Load full reads underlying file.
func (fo *FileOperator) Load() ([]byte, error) {
	return io.ReadAll(fo.r)
}

// Save clear file and then write all passed data.
func (fo *FileOperator) Save(data []byte) (err error) {
	if err = fo.Clear(); err != nil {
		return
	}
	if _, err = fo.w.Write(data); err != nil {
		return
	}
	return fo.w.Flush()
}

// Close close underlying file.
func (fo *FileOperator) Close() error {
	return fo.f.Close()
}
