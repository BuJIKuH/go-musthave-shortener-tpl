package audit

import (
	"context"
	"encoding/json"
	"os"
)

type FileObserver struct {
	file *os.File
}

func NewFileObserver(path string) (*FileObserver, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &FileObserver{file: f}, nil
}

func (o *FileObserver) Notify(ctx context.Context, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = o.file.Write(append(data, '\n'))
	return err
}
