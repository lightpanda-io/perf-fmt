package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
)

type Puller interface {
	Pull(ctx context.Context) (io.ReadCloser, error)
}

type Pusher interface {
	Push(ctx context.Context) (io.WriteCloser, error)
}

type FileIO struct {
	Path string
}

func (fio *FileIO) Pull(ctx context.Context) (io.ReadCloser, error) {
	f, err := os.Open(fio.Path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, f); err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	return io.NopCloser(&buf), nil
}

func (fio *FileIO) Push(ctx context.Context) (io.WriteCloser, error) {
	f, err := os.Create(fio.Path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	return f, nil
}
