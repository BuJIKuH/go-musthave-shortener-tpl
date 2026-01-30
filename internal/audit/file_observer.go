package audit

import (
	"bufio"
	"context"
	"encoding/json"
	"os"

	"go.uber.org/zap"
)

type FileObserver struct {
	writer *bufio.Writer
	file   *os.File
	logger *zap.Logger
}

func NewFileObserver(path string, logger *zap.Logger) (*FileObserver, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logger.Error("failed to open audit file", zap.Error(err), zap.String("path", path))
		return nil, err
	}

	return &FileObserver{
		file:   file,
		writer: bufio.NewWriterSize(file, 4096),
		logger: logger,
	}, nil
}

func (f *FileObserver) Notify(ctx context.Context, e Event) error {
	data, err := json.Marshal(e)
	if err != nil {
		f.logger.Error(
			"failed to marshal audit event",
			zap.Error(err),
			zap.String("action", e.Action),
			zap.String("user_id", e.UserID),
		)
		return err
	}

	if _, err := f.writer.Write(data); err != nil {
		f.logger.Error("failed to write audit event", zap.Error(err))
		return err
	}

	if err := f.writer.WriteByte('\n'); err != nil {
		f.logger.Error("failed to write newline to audit file", zap.Error(err))
		return err
	}

	if err := f.writer.Flush(); err != nil {
		f.logger.Error("failed to flush audit file", zap.Error(err))
		return err
	}

	return nil
}

func (f *FileObserver) Close() error {
	if err := f.writer.Flush(); err != nil {
		f.logger.Warn("failed to flush audit file on close", zap.Error(err))
	}

	if err := f.file.Close(); err != nil {
		f.logger.Warn("failed to close audit file", zap.Error(err))
		return err
	}

	f.logger.Info("audit file observer closed")
	return nil
}
