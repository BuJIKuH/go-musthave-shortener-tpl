package audit

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
)

type FileObserver struct {
	writer *bufio.Writer
	file   *os.File
}

func NewFileObserver(path string) (*FileObserver, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &FileObserver{
		file:   file,
		writer: bufio.NewWriterSize(file, 4096),
	}, nil
}

func (f *FileObserver) Notify(_ context.Context, e Event) error {
	data, _ := json.Marshal(e)
	_, err := f.writer.Write(data)
	if err != nil {
		return err
	}
	f.writer.WriteByte('\n')
	f.writer.Flush()
	return nil
}

func (f *FileObserver) Close() error {
	f.writer.Flush()
	return f.file.Close()
}
