package main

import (
	"context"
	"io"
)

type Append interface {
	Append(ctx context.Context, out io.Writer, all io.Reader, one io.Reader) error
}
