package fileoperator

import (
	"bufio"
	"io"
	"os"
)

type FileOperator struct {
	r *bufio.Reader
	w *bufio.Writer
	f *os.File
}

func New(file *os.File) *FileOperator {
	return &FileOperator{
		r: bufio.NewReader(file),
		w: bufio.NewWriter(file),
		f: file,
	}
}

func (fo *FileOperator) Clear() (err error) {
	if err = fo.f.Truncate(0); err != nil {
		return err
	}
	_, err = fo.f.Seek(0, io.SeekStart)
	return err
}

func (fo *FileOperator) Load() ([]byte, error) {
	return io.ReadAll(fo.r)
}

func (fo *FileOperator) Save(data []byte) (err error) {
	if err = fo.Clear(); err != nil {
		return
	}
	if _, err = fo.w.Write(data); err != nil {
		return
	}
	return fo.w.Flush()
}

func (fo *FileOperator) Close() error {
	return fo.f.Close()
}
