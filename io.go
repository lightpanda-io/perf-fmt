package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
)

type Puller interface {
	Pull(ctx context.Context) (io.ReadCloser, error)
}

type Pusher interface {
	Push(ctx context.Context, r io.Reader) error
}

type FileIO struct {
	Path string
}

func (fio *FileIO) Pull(ctx context.Context) (io.ReadCloser, error) {
	f, err := os.Open(fio.Path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return io.NopCloser(&bytes.Buffer{}), nil
		}

		return nil, fmt.Errorf("open file: %w", err)
	}

	return f, nil
}

func (fio *FileIO) Push(ctx context.Context, r io.Reader) error {
	f, err := os.Create(fio.Path)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
